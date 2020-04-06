package rates

import (
	"time"

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
	putBadRequest = promauto.NewCounter(prometheus.CounterOpts{
		Name: "put_rate_bad_request_count",
		Help: "The total number of PUT requests that were a bad request",
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

	requestDurationGet = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "request_duration_seconds_get",
		Help:    "Histogram for the runtime of calling GET rates.",
		Buckets: prometheus.ExponentialBuckets(0.00001, 1.5, 10),
	})
	requestDurationPut = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "request_duration_seconds_put",
		Help:    "Histogram for the runtime of calling PUT rates.",
		Buckets: prometheus.ExponentialBuckets(0.0001, 1.5, 10),
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

// record a Bad put request (400 error)
func recordPutBadRequest() {
	putBadRequest.Inc()
}

// record a 404 on a get request
func recordGetRatetNotFound() {
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

func init() {
	// These metrics have to be registered to be exposed:
	prometheus.MustRegister(requestDurationGet)
	prometheus.MustRegister(requestDurationPut)
}

// record latency of Get calls
func recordGetLatency(d time.Duration) {
	requestDurationGet.Observe(d.Seconds())
}

// record latency of Put calls
func recordPutLatency(d time.Duration) {
	requestDurationPut.Observe(d.Seconds())
}
