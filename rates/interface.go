package rates

// Service defines the interface to get rates for a given time range
type Service interface {
	Get(ParkingTimesRequest) (rate int, err error)
	Put()
}
