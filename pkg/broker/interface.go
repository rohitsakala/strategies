package broker

type Broker interface {
	Authenticate() error
	PlaceOrder() error
	GetLTP(instrument string) (int, error)
	CheckPosition(instrument string) (bool, error)

	// Option Funcs
	GetCurrentMonthyExpiry() (string, error)
}
