package rates

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRouter(t *testing.T) {
	m := &mockService{}
	r := NewRouter(m)
	assert.NotNil(t, r)
}

func TestRouterHealth(t *testing.T) {
	m := &mockService{}
	r := NewRouter(m)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, req)

	var b PutResponse
	_ = json.Unmarshal(w.Body.Bytes(), &b)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "ok", b.Status)
}

func TestPutRatesHandler(t *testing.T) {
	rd := RateDetail{
		Days:  "mon,tues,thurs",
		Times: "0900-2100",
		TZ:    "America/Chicago",
		Price: 1500,
	}
	newRates := IncomingRates{
		Rates: []RateDetail{rd},
	}
	testCases := []struct {
		name          string
		m             *mockService
		newRates      IncomingRates
		putCallCount  int
		outStatusCode int
		outResponse   PutResponse
	}{
		{name: "put rates success",
			m:             &mockService{},
			newRates:      newRates,
			putCallCount:  1,
			outStatusCode: 200,
			outResponse: PutResponse{
				Status:  "success",
				Message: "Successfully updated rates",
			},
		},
		{name: "put rates api fail",
			m: &mockService{
				err: errors.New("Simulating error setting new rates"),
			},
			newRates:      newRates,
			putCallCount:  0,
			outStatusCode: 500,
			outResponse: PutResponse{
				Status:  "error",
				Message: "Simulating error setting new rates",
			},
		},
	}

	for _, tt := range testCases {
		r := NewRouter(tt.m)
		jsonRates, err := json.Marshal(tt.newRates)
		assert.Nil(t, err)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/rates", bytes.NewBuffer(jsonRates))
		r.ServeHTTP(w, req)

		var b PutResponse
		err = json.Unmarshal(w.Body.Bytes(), &b)
		assert.Nil(t, err)

		assert.Equal(t, tt.outStatusCode, w.Code)
		assert.Equal(t, tt.outResponse, b)
		assert.Equal(t, tt.putCallCount, tt.m.putCallCount)
	}
}

func TestGetRateHandler(t *testing.T) {
	m := &mockService{}
	r := NewRouter(m)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/rate", nil)
	q := req.URL.Query()
	q.Add("start_time", "2015-07-01T07:20:00-05:00")
	q.Add("end_time", "2015-07-01T08:00:00-05:00")

	req.URL.RawQuery = q.Encode()
	r.ServeHTTP(w, req)

	var b PutResponse
	err := json.Unmarshal(w.Body.Bytes(), &b)
	assert.Nil(t, err)

	assert.Equal(t, 200, w.Code)
}

type mockService struct {
	Service
	putCallCount int
	getCallCount int
	rate         int
	err          error
}

func (m *mockService) Put(ir IncomingRates) error {
	if m.err != nil {
		return m.err
	}
	m.putCallCount++
	return nil
}

func (m *mockService) Get(p ParkingTimesRequest) (rate int, err error) {
	if m.err != nil {
		return 0, m.err
	}
	m.getCallCount++
	return m.rate, nil
}
