package rates

import (
	"errors"
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

	a, err := NewAPI()
	assert.Nil(t, err)
	// assert that the ratemap is seeded during instantiation
	assert.NotNil(t, a.rateMap)
	a.Put(ir)

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

	testCases := []struct {
		p    ParkingTimesRequest
		rate int
		err  error
	}{
		{
			p: ParkingTimesRequest{
				StartTime: time.Date(2020, 4, 3, 12, 30, 0, 0, time.UTC),
				EndTime:   time.Date(2020, 4, 3, 19, 30, 0, 0, time.UTC),
			},
			rate: 2000,
			err:  nil,
		},
		{
			p: ParkingTimesRequest{
				StartTime: time.Date(2020, 4, 3, 12, 30, 0, 0, time.UTC),
				EndTime:   time.Date(2020, 4, 4, 19, 30, 0, 0, time.UTC),
			},
			rate: 0,
			err:  errors.New("start and end time days do not match"),
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
