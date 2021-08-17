package broker

import "github.com/rohitsakala/strategies/pkg/models"

type Broker interface {
	Authenticate() error
	PlaceOrder() error
	GetLTP(symbol string) (float64, error)

	// Positions
	GetPositions() (models.PositionList, error)
	CheckPosition(symbol string) (bool, error)

	// Option Funcs
	GetInstruments(exchange string) (models.Instruments, error)
}
