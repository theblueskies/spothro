package rates

import (
	"encoding/json"
	"errors"
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
	startTime float32
	endTime   float32
	price     int
	tz        string
}

// ParkingTimesRequest is used to deserialize and hold the input time ranges
type ParkingTimesRequest struct {
	StartTime time.Time `form:"start_time" json:"start_time"`
	EndTime   time.Time `form:"end_time" json:"end_time"`
}

// API implements the interface to get rates and store new rates
type API struct {
	rateMap map[string][]DayRate
	mu      sync.Mutex
}

// NewAPI returns a new instance of API. It is seeded with the default JSON data file
func NewAPI(seedRatesFile string) (*API, error) {
	seedRatesJSON, err := os.Open(seedRatesFile)
	if err != nil {
		return nil, err
	}
	defer seedRatesJSON.Close()

	bytes, _ := ioutil.ReadAll(seedRatesJSON)
	var ir IncomingRates
	json.Unmarshal(bytes, &ir)

	a := &API{}
	a.Put(ir)

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

// Put creates a new rate map with key of days
func (a *API) Put(ir IncomingRates) error {
	// m will contain the new rate map.
	// When the new rates are received, the map is built out with the key of days
	// This let's the service quickly shortlist the rates that could be applicable for a given time range.
	// Instead of a map, an immutable trie could als have been used - https://github.com/hashicorp/go-immutable-radix

	m := make(map[string][]DayRate)

	// Iterate over the new rates and process them
	for _, r := range ir.Rates {
		// Split the time range and establish a start time and end time
		timeRange := strings.Split(r.Times, "-")
		startTime, err := strconv.Atoi(timeRange[0])
		if err != nil {
			return err
		}
		endTime, err := strconv.Atoi(timeRange[1])
		if err != nil {
			return err
		}
		// Iterate over all the days in a rate detail and make entries in
		// the map based on the key of the weekday
		for _, day := range strings.Split(r.Days, ",") {
			// The incoming rates come in as an abbreviated form - mon, tues, wed
			// The service needs to resolve the abbreviated form into the full name of the weekday
			// This is essential because the time package, returns only the full form of
			// the weekday eg: Monday, Tuesday, Wednesday etc
			properDayName, ok := dayMap[day]
			if !ok {
				return fmt.Errorf("abbreviated day not present: %s", day)
			}
			// Populate struct with time range and rate for a specific weekday
			dr := DayRate{
				day:       properDayName,
				startTime: float32(startTime),
				endTime:   float32(endTime),
				price:     r.Price,
				tz:        r.TZ,
			}
			// Check if there is an existing key of the weekday in the map
			v, ok := m[properDayName]
			// If there is a key, then append the new rate detail to the key
			if ok {
				v = append(v, dr)
				m[properDayName] = v
				continue
			}
			// If there was no key, then create a new entry for it in the map
			m[properDayName] = []DayRate{dr}
		}
	}
	// Lock it with a mutex before swapping the maps
	a.mu.Lock()
	a.rateMap = m
	a.mu.Unlock()
	return nil
}

// Get returns the rate of parking for a given time range
func (a *API) Get(p ParkingTimesRequest) (rate int, err error) {
	// If the parking time range span over more than the same day then return error
	if p.StartTime.Day() != p.EndTime.Day() || p.StartTime.Month() != p.EndTime.Month() || p.StartTime.Year() != p.EndTime.Year() {
		return 0, errors.New("start and end time days do not match")
	}
	// Get the rates for the specific weekday
	weekday := p.StartTime.Weekday()
	rates, ok := a.rateMap[weekday.String()]
	if !ok {
		return 0, fmt.Errorf("could not find rates for day %s", weekday)
	}

	// Check if the parking time range is contained within the defined ranges of rates
	for _, r := range rates {
		startHours := a.armyTime(p.StartTime)
		endHours := a.armyTime(p.EndTime)

		if startHours >= r.startTime && endHours <= r.endTime {
			return r.price, nil
		}
	}

	// Return error of unavailable when the parking time range was not found among the rates
	return 0, errors.New("unavailable")
}

func (a *API) armyTime(tm time.Time) float32 {
	// armyTime is a 24 hour clock in the format: 1240 this means the clock time is 12:40pm
	h, min, sec := tm.Clock()
	minAsHours := float32(min)
	secAsHours := float32(sec) / 3600
	armyTime := float32(h)*100 + minAsHours + secAsHours

	return armyTime
}
