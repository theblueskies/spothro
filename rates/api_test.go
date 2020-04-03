package rates

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewAPI(t *testing.T) {
	a, err := NewAPI()
	assert.Nil(t, err)
	assert.NotNil(t, a)
	assert.NotNil(t, a.rateMap)
}

func TestBuildAndReplaceRateMap(t *testing.T) {
	rd := RateDetail{
		Days:  "mon,tues,thurs",
		Times: "0900-2100",
		TZ:    "America/Chicago",
		Price: 1500,
	}
	ir := IncomingRates{
		Rates: []RateDetail{rd},
	}

	a, err := NewAPI()
	assert.Nil(t, err)
	// assert that the ratemap is seeded during instantiation
	assert.NotNil(t, a.rateMap)
	a.buildAndReplaceRateMap(ir)

	assert.NotNil(t, a.rateMap)
	mondayRate := a.rateMap["Monday"]
	expectedMondayDayRate := []DayRate{DayRate{
		day:       "Monday",
		startTime: 900,
		endTime:   2100,
		price:     1500,
		tz:        "America/Chicago"},
	}
	assert.Equal(t, expectedMondayDayRate, mondayRate)

	// assert that the new rate was put in place. sunday is not present in the new rate
	_, ok := a.rateMap["sun"]
	assert.False(t, ok)
}

func TestGetRate(t *testing.T) {
	a, err := NewAPI()
	assert.Nil(t, err)

	p := ParkingTimesRequest{
		StartTime: time.Now(),
		EndTime:   time.Now(),
	}
	a.Get(p)

}
