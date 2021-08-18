package broker

import "github.com/rohitsakala/strategies/pkg/models"

type Broker interface {
	Authenticate() error
	GetLTP(symbol string) (float64, error)

	// Positions
	GetPositions() (models.Positions, error)
	CheckPosition(symbol string) (bool, error)

	// Option Funcs
	GetInstruments(exchange string) (models.Positions, error)
	GetInstrument(symbol string, exchange string) (models.Position, error)

	// Orders
	PlaceOrder(position models.Position) (models.Position, error)
}
