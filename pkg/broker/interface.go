package broker

import "github.com/rohitsakala/strategies/pkg/models"

type Broker interface {
	Authenticate() error
	GetLTP(symbol string) (float64, error)

	// Positions
	GetPositions() (models.Positions, error)
	CheckPosition(symbol string) (bool, error)

	GetInstruments(exchange string) (models.Positions, error)
	GetInstrument(symbol string, exchange string) (models.Position, error)

	// Orders
	GetOrders() (models.Positions, error)
	PlaceOrder(position *models.Position) error
	CancelOrder(position *models.Position) error
	CancelOrders(positions models.RefPositions) error

	// Margin
	GetMargin()
}
