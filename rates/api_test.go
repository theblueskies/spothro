package rates

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewAPI(t *testing.T) {
	a, err := NewAPI("seed_rates.json")
	assert.Nil(t, err)
	assert.NotNil(t, a)
	assert.NotNil(t, a.rateMap)
}

func TestPut(t *testing.T) {
	rd := RateDetail{
		Days:  "mon,tues,thurs",
		Times: "0900-2100",
		TZ:    "America/Chicago",
		Price: 1500,
	}
	ir := IncomingRates{
		Rates: []RateDetail{rd},
	}

	a, err := NewAPI("seed_rates.json")
	assert.Nil(t, err)
	// assert that the ratemap is seeded during instantiation
	assert.NotNil(t, a.rateMap)
	a.Put(ir)

	assert.NotNil(t, a.rateMap)
	mondayRate := a.rateMap["Monday"]
	// the expected rate has been transformed to UTC time
	expectedMondayDayRate := []DayRate{DayRate{
		day:       "Monday",
		startTime: float32(1400),
		endTime:   float32(2600),
		price:     1500,
		tz:        "UTC"},
	}
	assert.Equal(t, expectedMondayDayRate, mondayRate)
	assert.Equal(t, expectedMondayDayRate[0].endTime, mondayRate[0].endTime)

	// assert that the new rate was put in place. sunday is not present in the new rate
	_, ok := a.rateMap["sun"]
	assert.False(t, ok)
}

func TestGetRate(t *testing.T) {
	a, err := NewAPI("seed_rates.json")
	assert.Nil(t, err)

	loc, e := time.LoadLocation("Asia/Calcutta") // +5:30 UTC
	assert.Nil(t, e)
	testCases := []struct {
		p    ParkingTimesRequest
		rate int
		err  error
	}{
		{
			p: ParkingTimesRequest{
				StartTime: time.Date(2020, 4, 3, 14, 30, 0, 0, time.UTC),
				EndTime:   time.Date(2020, 4, 3, 19, 30, 0, 0, time.UTC),
			},
			rate: 2000,
			err:  nil,
		},
		{
			p: ParkingTimesRequest{
				StartTime: time.Date(2020, 4, 3, 8, 30, 0, 0, loc),
				EndTime:   time.Date(2020, 4, 3, 20, 30, 0, 0, loc),
			},
			rate: 0,
			err:  errors.New("unavailable"),
		},
		{
			p: ParkingTimesRequest{
				StartTime: time.Date(2020, 4, 4, 07, 00, 0, 0, time.UTC),
				EndTime:   time.Date(2020, 4, 4, 20, 00, 0, 0, time.UTC),
			},
			rate: 0,
			err:  errors.New("unavailable"),
		},
	}
	for _, tt := range testCases {
		rate, err := a.Get(tt.p)
		assert.Equal(t, tt.err, err)
		assert.Equal(t, tt.rate, rate)
	}
}

func TestGetRateWhenNoRatePresent(t *testing.T) {
	a, err := NewAPI("seed_rates.json")
	assert.Nil(t, err)

	a.Put(IncomingRates{}) // Clear out all rates
	p := ParkingTimesRequest{
		StartTime: time.Date(2020, 4, 4, 07, 00, 0, 0, time.UTC),
		EndTime:   time.Date(2020, 4, 4, 20, 00, 0, 0, time.UTC),
	}
	rate, err := a.Get(p)
	assert.Equal(t, errors.New("could not find rates for day Saturday"), err)
	assert.Equal(t, 0, rate)
}

func TestArmytime(t *testing.T) {
	a, err := NewAPI("seed_rates.json")
	assert.Nil(t, err)

	tm := time.Date(2020, 4, 4, 12, 15, 0, 0, time.UTC)
	armyTime := a.armyTime(tm)
	assert.Equal(t, float32(1215), armyTime)

	tm = time.Date(2020, 4, 4, 12, 15, 23, 101, time.UTC)
	armyTime = a.armyTime(tm)
	assert.Equal(t, float32(1215.0063), armyTime)
}
