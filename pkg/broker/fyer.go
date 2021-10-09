package broker

import (
	"fmt"
	ogRek "github.com/kisielk/og-rek"
	"github.com/rohitsakala/strategies/pkg/httpClient"
	"net/http"
	"time"
)

type fyerBroker struct {
	url         string
	userId      string
	appId       string
	accessToken string
	client      httpClient.Client
	TimeZone    time.Location
}

func NewFyerBroker(url string, userId string, appId string) Broker {
	return &fyerBroker{
		url:      url,
		userId:   userId,
		appId:    appId,
		client:   nil,
		TimeZone: time.Location{},
	}
}

func (z *fyerBroker) GetLTP(symbol string) (float64, error) {
	// find instrument token of the symbol
	quote, err := z.client.GetQuote(symbol)
	if err != nil {
		return 0, err
	}
	return quote.LTP, nil
}
