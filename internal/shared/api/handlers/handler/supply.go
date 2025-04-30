package handler

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// IntegrationSupply handler returns supply information to an external client (e.g., CoinMarketCap).
// Note that the error text is returned to the client, so avoid including sensitive data in errors.
func (h *Handler) IntegrationSupply(req *http.Request) (any, error) {
	supplyType := req.URL.Query().Get("type")
	switch supplyType {
	case "total":
		return h.totalSupply(req)
	case "circulation":
		return h.circulationSupply(req)
	default:
		return nil, fmt.Errorf("wrong type parameter='%s' (please provider either 'total' or 'circulation')", supplyType)
	}
}

func (h *Handler) totalSupply(req *http.Request) (any, error) {
	if h.bbnClient == nil {
		return nil, fmt.Errorf("bbn configuration is not set")
	}
	ctx := req.Context()

	coin, err := h.bbnClient.GetTotalSupply(ctx)
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to get total supply")
		return nil, fmt.Errorf("internal error")
	}

	return coin.Amount.Uint64(), nil
}

type vestingFrequency string

const (
	daily   vestingFrequency = "daily"
	monthly vestingFrequency = "monthly"
)

type vestingType string

const (
	linear       vestingType = "linear"
	cliff        vestingType = "cliff"
	inflationary vestingType = "inflationary"
)

type tokenUnlock struct {
	StartTime   time.Time         `json:"startTime"`
	EndTime     *time.Time        `json:"endTime"`
	Amount      float64           `json:"amount"`
	Frequency   *vestingFrequency `json:"frequency"`
	VestingType vestingType       `json:"vestingType"`
}

func (tu tokenUnlock) availableTokensAt(t time.Time) float64 {
	// first normalize time
	t = t.UTC()
	vestingStarted := tu.StartTime.Before(t)
	if !vestingStarted {
		return 0
	}

	var availableTokens float64
	if tu.VestingType == cliff {
		availableTokens = tu.Amount
	} else { // note that inflationary considered as linear as well
		// if endTime already passed just return full amount
		if tu.EndTime.Before(t) {
			return tu.Amount
		}

		if tu.Frequency == nil {
			panic(fmt.Errorf("non cliff token unlock contains frequency"))
		}

		start, end := tu.StartTime, tu.EndTime

		switch f := *tu.Frequency; f {
		case daily:
			daysWithinRange := int(end.Sub(start).Hours() / 24)
			daysPassed := int(t.Sub(start).Hours() / 24)

			tokensUnlockedPerDay := tu.Amount / float64(daysWithinRange)
			availableTokens = float64(daysPassed) * tokensUnlockedPerDay
		case monthly:
			monthsWithinRange := (end.Year()-start.Year())*12 + int(end.Month()) - int(start.Month())
			monthsPassed := (t.Year()-start.Year())*12 + int(t.Month()) - int(start.Month())
			// todo check days handling

			tokensUnlockedPerMonth := tu.Amount / float64(monthsWithinRange)
			availableTokens = float64(monthsPassed) * tokensUnlockedPerMonth
		default:
			panic(fmt.Errorf("unexpected frequency value %v", f))
		}
	}

	return availableTokens
}

var schedule []tokenUnlock

//go:embed unlock-schedule.json
var scheduleData []byte

func init() {
	err := json.Unmarshal(scheduleData, &schedule)
	if err != nil {
		panic(fmt.Errorf("failed to unmarshal unlock-schedule.json: %v", err))
	}
}

func (h *Handler) circulationSupply(req *http.Request) (any, error) {
	now := time.Now()

	var tokenInCirculation float64
	for _, unlock := range schedule {
		tokenInCirculation += unlock.availableTokensAt(now)
	}

	return tokenInCirculation, nil
}
