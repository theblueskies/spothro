package rates

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// dayMap is used to convert the abbreviated days from json to full names of the days
var dayMap = map[string]string{
	"mon":   "Monday",
	"tues":  "Tuesday",
	"wed":   "Wednesday",
	"thurs": "Thursday",
	"fri":   "Friday",
	"sat":   "Saturday",
	"sun":   "Sunday",
}

// DayRate is used to store the price for a given time range for a specific day
type DayRate struct {
	day       string
	startTime int
	endTime   int
	price     int
	tz        string
}

// ParkingTimesRequest is used to deserialize and hold the input time ranges
type ParkingTimesRequest struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// API implements the interface to get rates and store new rates
type API struct {
	rateMap map[string][]DayRate
	mu      sync.Mutex
}

// NewAPI returns a new instance of API. It is seeded with the default JSON data file
func NewAPI() (*API, error) {
	seedRatesJSON, err := os.Open("seed_rates.json")
	if err != nil {
		return nil, err
	}
	defer seedRatesJSON.Close()

	bytes, _ := ioutil.ReadAll(seedRatesJSON)
	var ir IncomingRates
	json.Unmarshal(bytes, &ir)

	a := &API{}
	a.buildAndReplaceRateMap(ir)

	return a, nil
}

// IncomingRates defines the json struct for new incoming rates
type IncomingRates struct {
	Rates []RateDetail `json:"rates"`
}

// RateDetail holds the rate details of the new incoming rates
type RateDetail struct {
	Days  string `json:"days"`
	Times string `json:"times"`
	TZ    string `json:"tz"`
	Price int    `json:"price"`
}

// buildAndReplaceRateMap creates a new rate map indexed on days
func (a *API) buildAndReplaceRateMap(ir IncomingRates) error {
	m := make(map[string][]DayRate)
	for _, r := range ir.Rates {
		timeRange := strings.Split(r.Times, "-")
		startTime, err := strconv.Atoi(timeRange[0])
		if err != nil {
			return err
		}
		endTime, err := strconv.Atoi(timeRange[1])
		if err != nil {
			return err
		}
		for _, day := range strings.Split(r.Days, ",") {
			properDayName, ok := dayMap[day]
			if !ok {
				return fmt.Errorf("abbreviated day not present: %s", day)
			}
			dr := DayRate{
				day:       properDayName,
				startTime: startTime,
				endTime:   endTime,
				price:     r.Price,
				tz:        r.TZ,
			}
			v, ok := m[properDayName]
			if ok {
				v = append(v, dr)
				m[properDayName] = v
				continue
			}
			m[properDayName] = []DayRate{dr}
		}
	}
	a.mu.Lock()
	a.rateMap = m
	a.mu.Unlock()
	return nil
}

// Get returns the rate of parking for a given time range
func (a *API) Get(p ParkingTimesRequest) (rate int, err error) {
	return 0, nil
}
