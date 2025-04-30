package handler

import (
	"fmt"
	"net/http"
	"time"
)

func (h *Handler) CMCTotalSupply(http.ResponseWriter, *http.Request) {

}

type vestingFrequency int

const (
	daily vestingFrequency = iota + 1
	monthly
)

type vestingType int

const (
	linear vestingType = iota + 1
	cliff
	inflationary
)

type tokenUnlock struct {
	name        string // we don't use this value,
	startTime   time.Time
	endTime     time.Time
	amount      float64
	frequency   vestingFrequency
	vestingType vestingType
}

func (tu tokenUnlock) availableTokensAt(t time.Time) float64 {
	// first normalize time
	t = t.UTC()
	vestingStarted := tu.startTime.Before(t)
	if !vestingStarted {
		return 0
	}

	var availableTokens float64
	if tu.vestingType == cliff {
		availableTokens = tu.amount
	} else { // note that inflationary considered as linear as well
		// if endTime already passed just return full amount
		if tu.endTime.Before(t) {
			return tu.amount
		}

		start, end := tu.startTime, tu.endTime

		switch f := tu.frequency; f {
		case daily:
			daysWithinRange := int(end.Sub(start).Hours() / 24)
			daysPassed := int(t.Sub(start).Hours() / 24)

			tokensUnlockedPerDay := tu.amount / float64(daysWithinRange)
			availableTokens = float64(daysPassed) * tokensUnlockedPerDay
		case monthly:
			monthsWithinRange := (end.Year()-start.Year())*12 + int(end.Month()) - int(start.Month())
			monthsPassed := (t.Year()-start.Year())*12 + int(t.Month()) - int(start.Month())
			// todo check days handling

			tokensUnlockedPerMonth := tu.amount / float64(monthsWithinRange)
			availableTokens = float64(monthsPassed) * tokensUnlockedPerMonth
		default:
			panic(fmt.Errorf("unexpected frequency value %v", f))
		}
	}

	return availableTokens
}

var schedule = []tokenUnlock{
	{name: "Community 1", startTime: parseTime("2025-04-10 10:00:00"), endTime: time.Time{}, amount: 1378400000, frequency: 0, vestingType: cliff},
	{name: "Community 2", startTime: parseTime("2025-10-10 10:00:00"), endTime: time.Time{}, amount: 121600000, frequency: 0, vestingType: cliff},
	{name: "Ecosystem 1", startTime: parseTime("2025-04-10 10:00:00"), endTime: time.Time{}, amount: 450000000, frequency: 0, vestingType: cliff},
	{name: "Ecosystem 2", startTime: parseTime("2026-04-10 10:00:00"), endTime: parseTime("2028-04-10 09:59:59"), amount: 1350000000, frequency: monthly, vestingType: linear},
	{name: "R&D & Operations 1", startTime: parseTime("2025-04-10 10:00:00"), endTime: time.Time{}, amount: 450000000, frequency: 0, vestingType: cliff},
	{name: "R&D & Operations 2", startTime: parseTime("2026-04-10 10:00:00"), endTime: parseTime("2028-04-10 09:59:59"), amount: 1350000000, frequency: monthly, vestingType: linear},
	{name: "Core Team 1", startTime: parseTime("2026-04-10 10:00:00"), endTime: time.Time{}, amount: 187500000, frequency: 0, vestingType: cliff},
	{name: "Core Team 2", startTime: parseTime("2026-05-10 10:00:00"), endTime: parseTime("2029-04-10 09:59:59"), amount: 1312500000, frequency: monthly, vestingType: linear},
	{name: "Advisors 1", startTime: parseTime("2026-04-10 10:00:00"), endTime: time.Time{}, amount: 43750000, frequency: 0, vestingType: cliff},
	{name: "Advisors 2", startTime: parseTime("2026-05-10 10:00:00"), endTime: parseTime("2029-04-10 09:59:59"), amount: 306250000, frequency: monthly, vestingType: linear},
	{name: "Investors 1", startTime: parseTime("2026-04-10 10:00:00"), endTime: time.Time{}, amount: 381250000, frequency: 0, vestingType: cliff},
	{name: "Investors 2", startTime: parseTime("2026-05-10 10:00:00"), endTime: parseTime("2029-04-10 09:59:59"), amount: 2668750000, frequency: monthly, vestingType: linear},
	{name: "Inflation Year 1", startTime: parseTime("2025-04-03 06:36:14"), endTime: parseTime("2026-04-03 06:36:13"), amount: 800000000, frequency: daily, vestingType: inflationary},
	{name: "Inflation Year 2", startTime: parseTime("2026-04-03 06:36:14"), endTime: parseTime("2027-04-03 06:36:13"), amount: 864000000, frequency: daily, vestingType: inflationary},
	{name: "Inflation Year 3", startTime: parseTime("2027-04-03 06:36:14"), endTime: parseTime("2028-04-03 06:36:13"), amount: 869120000, frequency: daily, vestingType: inflationary},
	{name: "Inflation Year 4", startTime: parseTime("2028-04-03 06:36:14"), endTime: parseTime("2029-04-03 06:36:13"), amount: 938649600, frequency: daily, vestingType: inflationary},
}

func (h *Handler) CMCCirculationSupply(http.ResponseWriter, *http.Request) {
	now := time.Now()

	var tokenInCirculation float64
	for _, unlock := range schedule {
		tokenInCirculation += unlock.availableTokensAt(now)
	}

	// todo write this value
}

func parseTime(value string) time.Time {
	t, err := time.ParseInLocation(time.DateTime, value, time.UTC)
	if err != nil {
		panic(err)
	}
	return t
}
