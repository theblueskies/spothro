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
	// When the new rates are received, the map is built out with the key of days
	// This let's the service quickly shortlist the rates that could be applicable for a given time range.
	// Instead of a map, an immutable trie could also have been used - https://github.com/hashicorp/go-immutable-radix
	// All times are stored as UTC

	// m will contain the new rate map.
	m := make(map[string][]DayRate)

	// Iterate over the new rates and process them
	for _, r := range ir.Rates {
		// Split the time range and establish a start time and end time
		timeRange := strings.Split(r.Times, "-")
		startTime, err := strconv.Atoi(timeRange[0])
		if err != nil {
			return err
		}
		startTimeHours := startTime / 100 // Hours
		startTimeMins := startTime % 100  // Minutes
		endTime, err := strconv.Atoi(timeRange[1])
		if err != nil {
			return err
		}
		endTimeHours := endTime / 100 // Hours
		endTimeMins := endTime % 100  // Minutes
		if err != nil {
			return err
		}
		// Iterate over all the days in an input rate detail and make entries in
		// the map based on the key of the weekday
		for _, day := range strings.Split(r.Days, ",") {
			properWeekdayName, ok := dayMap[day]
			if !ok {
				return fmt.Errorf("abbreviated day not present: %s", day)
			}
			// build time with localized timezone that's present in the input
			localizedTime, err := TimeIn(time.Now(), r.TZ)
			if err != nil {
				return err
			}
			for {
				s := localizedTime.Weekday().String()
				if s != properWeekdayName {
					localizedTime = localizedTime.AddDate(0, 0, 1) // Increment by one day until we hit the desired weekday
					continue
				}
				break
			}

			// Get UTC times of the corresponding localized times
			utcStartTime := time.Date(localizedTime.Year(), localizedTime.Month(), localizedTime.Day(), startTimeHours, startTimeMins, 0, 0, localizedTime.Location()).UTC()
			utcEndTime := time.Date(localizedTime.Year(), localizedTime.Month(), localizedTime.Day(), endTimeHours, endTimeMins, 0, 0, localizedTime.Location()).UTC()
			// Get 2400 layout times of the corresponding UTC times
			armyStartTime := a.armyTime(utcStartTime)
			armyEndTime := a.armyTime(utcEndTime)
			// if armyEndTime > 2400, then it starts back from 0. We need to add 2400
			// to maintain the contiguous time block for time comparisons
			if armyEndTime < armyStartTime {
				armyEndTime += 2400
			}

			// Populate struct with time range and rate for a specific weekday
			dr := DayRate{
				day:       utcStartTime.Weekday().String(),
				startTime: float32(armyStartTime),
				endTime:   float32(armyEndTime),
				price:     r.Price,
				tz:        "UTC",
			}
			// Check if there is an existing key of the weekday in the map
			v, ok := m[properWeekdayName]
			// If there is a key, then append the new rate detail to the key
			if ok {
				v = append(v, dr)
				m[properWeekdayName] = v
				continue
			}
			// If there was no key, then create a new entry for it in the map
			m[properWeekdayName] = []DayRate{dr}
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
	// Rates will not span multiple days
	if p.StartTime.Day() != p.EndTime.Day() || p.StartTime.Month() != p.EndTime.Month() || p.StartTime.Year() != p.EndTime.Year() {
		return 0, errors.New("unavailable")
	}
	// Transform localized time to UTC time
	utcStart := p.StartTime.UTC()
	utcEnd := p.EndTime.UTC()
	// Get UTC army time (2400 hour layout)
	startHours := a.armyTime(utcStart)
	endHours := a.armyTime(utcEnd)

	// Get the rates for the specific weekday
	weekday := p.StartTime.Weekday().String()
	rates, ok := a.rateMap[weekday]
	if !ok {
		return 0, errors.New("unavailable")
	}

	// Check if the parking time range is contained within the defined ranges of rates
	for _, r := range rates {
		if startHours >= r.startTime && endHours <= r.endTime {
			return r.price, nil
		}
	}

	// Return error of unavailable when the parking time range was not found among the rates
	return 0, errors.New("unavailable")
}

// armyTime returns a 2400 layout time in UTC timezone
func (a *API) armyTime(tm time.Time) float32 {
	// armyTime is a 24 hour clock in the format: 1240 this means the clock time is 12:40pm
	tm = tm.UTC() // Get UTC time
	h, min, sec := tm.Clock()
	minAsHours := float32(min)
	secAsHours := float32(sec) / 3600
	armyTime := float32(h)*100 + minAsHours + secAsHours

	return armyTime
}

// TimeIn returns the time in UTC if the name is "" or "UTC".
// It returns the local time if the name is "Local".
// Otherwise, the name is taken to be a location name in
// the IANA Time Zone database, such as "Africa/Lagos".
func TimeIn(t time.Time, timeZone string) (time.Time, error) {
	loc, err := time.LoadLocation(timeZone)
	if err == nil {
		t = t.In(loc)
	}
	return t, err
}
