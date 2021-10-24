package broker

import (
	"github.com/rohitsakala/strategies/pkg/database"
	"github.com/rohitsakala/strategies/pkg/httpClient"
	"github.com/rohitsakala/strategies/pkg/models"
)

type FyerBroker struct {
	url    string
	userId string
	appId  string
	client httpClient.Client
}

func NewFyerBroker(database database.Database, url, userID, password, apiKey, apiSecret, appId string) (FyerBroker, error) {
	err := database.CreateCollection("credentials")
	if err != nil {
		return FyerBroker{}, err
	}

	return FyerBroker{
		url:    url,
		userId: userID,
		appId:  appId,
		client: nil,
	}, nil
}

func (f *FyerBroker) IsMarketOpen() (bool, error) {
	return false, nil
}

func (f *FyerBroker) GetOrderID(position models.Position) (string, error) {
	return "", nil
}

func (f *FyerBroker) Authenticate() error {
	return nil
}

func (f *FyerBroker) CancelOrder(position *models.Position) error {
	return nil
}

// Positions
func (f *FyerBroker) GetPositions() (models.Positions, error) {
	return nil, nil
}

func (f *FyerBroker) CheckPosition(symbol string) (bool, error) {
	return false, nil
}

func (f *FyerBroker) GetInstruments(exchange string) (models.Positions, error) {
	return models.Positions{}, nil
}

func (f *FyerBroker) GetInstrument(symbol string, exchange string) (models.Position, error) {
	return models.Position{}, nil
}

// Orders
func (f *FyerBroker) GetOrders() (models.Positions, error) {
	return models.Positions{}, nil
}
func (f *FyerBroker) PlaceOrder(position *models.Position) error {
	return nil
}
func (f *FyerBroker) CancelOrders(positions models.RefPositions) error {
	return nil
}

// Margin
func (f *FyerBroker) GetMargin() {
}

func (f *FyerBroker) GetLTP(symbol string) (float64, error) {
	quote, err := f.client.GetQuote(symbol)
	if err != nil {
		return 0, err
	}
	return quote.LTP, nil
}

func (f *FyerBroker) GetLTPNoFreak(symbol string) (float64, error) {
	return 0, nil
}
