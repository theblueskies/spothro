package rates

import (
	"fmt"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
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
	return r
}

// PutRates is a wrapper around the Service Put function
func PutRates(s Service) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		var ir IncomingRates
		err := c.ShouldBindWith(&ir, binding.JSON)
		if err != nil {
			c.JSON(400, PutResponse{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}

		err = s.Put(ir)
		if err != nil {
			c.JSON(500, PutResponse{
				Status:  "error",
				Message: err.Error(),
			})
			return
		}

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
		var p ParkingTimesRequest
		err := c.Bind(&p)
		if err != nil {
			c.JSON(400, RateResponse{
				Status:  "error",
				Message: err.Error(),
				Rate:    0,
			})
			return
		}
		h := p.StartTime.Hour()
		m := p.StartTime.Minute()
		fmt.Println(h, m)
		rate, err := s.Get(p)
		if err != nil {
			c.JSON(500, RateResponse{
				Status:  "error",
				Message: "error retrieving rate",
				Rate:    0,
			})
			return
		}
		c.JSON(200, RateResponse{
			Status:  "success",
			Message: "success retrieving rate",
			Rate:    rate,
		})
	}
	return gin.HandlerFunc(fn)
}
