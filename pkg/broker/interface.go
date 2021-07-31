package broker

type Broker interface {
	Authenticate() error
	PlaceOrder() error
}
