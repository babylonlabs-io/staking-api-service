package handler

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"

	cosmosMath "cosmossdk.io/math"
	"github.com/rs/zerolog/log"
)

// InfoMetrics handler returns supply information to an external client (e.g., CoinMarketCap).
// Note that the error text is returned to the client, so avoid including sensitive data in errors.
func (h *Handler) InfoMetrics(req *http.Request) (any, error) {
	key := req.URL.Query().Get("key")
	switch key {
	case "baby_total_supply":
		return h.babyTotalSupply(req)
	case "baby_circulation_supply":
		return h.babyCirculationSupply(req)
	case "":
		return nil, fmt.Errorf("GET-parameter 'key' is required")
	default:
		return nil, fmt.Errorf("wrong key parameter='%s' (please provider either 'baby_total_supply' or 'baby_circulation_supply')", key)
	}
}

func (h *Handler) babyTotalSupply(req *http.Request) (any, error) {
	if h.bbnClient == nil {
		return nil, fmt.Errorf("bbn configuration is not set")
	}
	ctx := req.Context()

	coin, err := h.bbnClient.GetTotalSupply(ctx, "ubbn")
	if err != nil {
		log.Ctx(ctx).Err(err).Msg("Failed to get total supply")
		return nil, fmt.Errorf("internal error")
	}

	const ubbnPerBabyToken = 1e6
	babyAmount := coin.Amount.Quo(cosmosMath.NewInt(ubbnPerBabyToken))
	if !babyAmount.IsInt64() {
		return nil, fmt.Errorf("cosmos.int %s overflowed int64", coin.Amount)
	}

	return babyAmount.Int64(), nil
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
		start, end := tu.StartTime, tu.EndTime

		// it's guaranteed that this pointer is non nil (see init() function that populates schedule var)
		switch f := *tu.Frequency; f {
		case daily:
			daysWithinRange := int(end.Sub(start).Hours() / 24)
			daysPassed := int(t.Sub(start).Hours() / 24)

			tokensUnlockedPerDay := tu.Amount / float64(daysWithinRange)
			availableTokens = float64(daysPassed) * tokensUnlockedPerDay
		case monthly:
			monthsWithinRange := (end.Year()-start.Year())*12 + int(end.Month()) - int(start.Month())
			monthsPassed := (t.Year()-start.Year())*12 + int(t.Month()) - int(start.Month())

			tokensUnlockedPerMonth := tu.Amount / float64(monthsWithinRange)
			availableTokens = float64(monthsPassed) * tokensUnlockedPerMonth
		default:
			panic(fmt.Errorf("unexpected frequency value %v", f))
		}
	}

	return math.Ceil(availableTokens)
}

var schedule []tokenUnlock

//go:embed unlock-schedule.json
var scheduleData []byte

func init() {
	err := json.Unmarshal(scheduleData, &schedule)
	if err != nil {
		panic(fmt.Errorf("failed to unmarshal unlock-schedule.json: %v", err))
	}

	for i, unlock := range schedule {
		if unlock.VestingType != cliff && unlock.Frequency == nil {
			panic(fmt.Errorf("non cliff token unlock %d has empty frequency", i))
		}
	}
}

func (h *Handler) babyCirculationSupply(_ *http.Request) (float64, error) {
	now := time.Now()

	var tokenInCirculation float64
	for _, unlock := range schedule {
		tokenInCirculation += unlock.availableTokensAt(now)
	}
	// correction made on 5.05.2025
	const ineligibleBTCStakingGauge = 8_919_424.02
	tokenInCirculation -= ineligibleBTCStakingGauge

	return tokenInCirculation, nil
}
