package rates

// Service defines the interface to get rates for a given time range
type Service interface {
	Get(InputTimes) (rate int, err error)
	Put()
}
