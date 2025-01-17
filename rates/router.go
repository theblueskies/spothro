package rates

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// PutResponse defines the response to updating with new rates
type PutResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// RateResponse defines the response to getting a specific rate for a time span
type RateResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Rate    int    `json:"rate"`
}

// NewRouter returns a router with the registered endpoints
// It takes an interface as a parameter
// This prevents the implementations from being tightly coupled to each other
func NewRouter(s Service) *gin.Engine {
	r := gin.Default()
	r.Use(cors.Default())
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "rates app online",
		})
	})
	r.PUT("/rates", PutRates(s))
	r.GET("/rate", GetRate(s))
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	return r
}

// PutRates is a wrapper around the Service Put function
func PutRates(s Service) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		tm := time.Now()
		var ir IncomingRates
		// Bind the json data to the struct
		err := c.ShouldBindWith(&ir, binding.JSON)
		if err != nil {
			recordPutBadRequest()
			recordPutLatency(time.Since(tm))

			c.JSON(400, PutResponse{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
		// Call the Put function of the service to store the new rates and replace the older rates
		err = s.Put(ir)
		if err != nil {
			// record stats
			recordPutFail()
			recordPutLatency(time.Since(tm))

			// If there was an error in Put, then return a 500 with an error response
			c.JSON(500, PutResponse{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}
		// If there was no error and the rates were successfully stored, then
		// return a 200 with a success response.

		// record stats
		recordPutSuccess()
		recordPutLatency(time.Since(tm))

		// Per RFC7231, when a PUT is updating a resource, it should return a 200 and not a 201.
		// The rates are already pre-seeded by seed_rates.json at the time of startup
		// Any subsequent PUT calls with new rates, is updating the resource and not creating it.
		// https://tools.ietf.org/html/rfc7231#section-4.3.4
		c.JSON(200, PutResponse{
			Status:  "success",
			Message: "Successfully updated rates",
		})
	}
	return gin.HandlerFunc(fn)
}

// GetRate is a wrapper around the Service Get function
func GetRate(s Service) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		tm := time.Now()
		var p ParkingTimesRequest
		// Bind the query params to the struct
		err := c.Bind(&p)
		// If there was an error in binding, then return a 400 with a response about the error
		if err != nil {
			// record stats
			recordGetRateBadRequest()
			recordGetLatency(time.Since(tm))

			c.JSON(400, RateResponse{
				Status:  "error",
				Message: err.Error(),
				Rate:    0,
			})
			return
		}
		// Call the Get function of the service to attempt to retrieve the rate for the given time range
		rate, err := s.Get(p)
		// If there was an error, return a 404 (not found) with a response containing the error
		// When a rate is "unavailable", it will be sent as the value of Message
		if err != nil {
			// record stats
			recordGetRatetNotFound()
			recordGetLatency(time.Since(tm))

			c.JSON(404, RateResponse{
				Status:  "error",
				Message: err.Error(),
				Rate:    0,
			})
			return
		}
		// If the service finds the rate, then it returns a  200 and send
		// a response containing the rate

		// record stats
		recordGetRateSuccess()
		recordGetLatency(time.Since(tm))

		c.JSON(200, RateResponse{
			Status:  "success",
			Message: "success retrieving rate",
			Rate:    rate,
		})
	}
	return gin.HandlerFunc(fn)
}
