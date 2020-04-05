package rates

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	put200Ok = promauto.NewCounter(prometheus.CounterOpts{
		Name: "put_new_rates_success_count",
		Help: "The total number of succesfully processed PUT requests",
	})
	putError = promauto.NewCounter(prometheus.CounterOpts{
		Name: "put_rate_error_count",
		Help: "The total number of PUT requests that resulted in an error",
	})
	get200Ok = promauto.NewCounter(prometheus.CounterOpts{
		Name: "get_rate_success_count",
		Help: "The total number of succesfully processed GET requests",
	})
	get404NotFound = promauto.NewCounter(prometheus.CounterOpts{
		Name: "get_rate_404_count",
		Help: "The total number of GET requests that resulted in a 404",
	})
	get400BadRequest = promauto.NewCounter(prometheus.CounterOpts{
		Name: "get_rate_400_count",
		Help: "The total number of GET requests that resulted in a 400",
	})
)

// record a successful put request
func recordPutSuccess() {
	put200Ok.Inc()
}

// record an unsuccessful put request
func recordPutFail() {
	putError.Inc()
}

// record a 404 on a get request
func recordGeRatetNotFound() {
	get404NotFound.Inc()
}

// record a successful 200 OK on a get request
func recordGetRateSuccess() {
	get200Ok.Inc()
}

// record a 400 Bad Request on a get request
func recordGetRateBadRequest() {
	get400BadRequest.Inc()
}
