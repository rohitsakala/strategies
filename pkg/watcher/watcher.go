package watcher

import (
	"fmt"
	"log"
	"time"

	"github.com/rohitsakala/strategies/pkg/broker"
	"github.com/rohitsakala/strategies/pkg/models"
	"github.com/rohitsakala/strategies/pkg/utils"
	kiteconnect "github.com/zerodha/gokiteconnect/v4"
)

type Watcher struct {
	Broker   broker.Broker
	TimeZone time.Location
}

func NewWatcher(broker broker.Broker, timeZone time.Location) (Watcher, error) {
	return Watcher{
		Broker:   broker,
		TimeZone: timeZone,
	}, nil
}

func (w *Watcher) Watch(position *models.Position) error {
	orders, err := w.Broker.GetOrders()
	if err != nil {
		return err
	}

	switch position.OrderType {
	case kiteconnect.OrderTypeSL:
		for _, order := range orders {
			if order.OrderID == position.OrderID {
				switch position.Status {
				case "TRIGGER PENDING":
					if order.Status != position.Status {
						message := fmt.Sprintf("Order %s Changed from %s to %s", position.TradingSymbol, position.Status, order.Status)
						log.Println(message)
						utils.SendEmail("12:30 pm Trade Update", message)
						position.Status = order.Status
					}
				case "OPEN":
					position.OrderType = kiteconnect.OrderTypeLimit
					err = w.Broker.PlaceOrder(position)
					if err != nil {
						return err
					}
					message := fmt.Sprintf("Order %s Changed from OPEN to %s", position.TradingSymbol, position.Status)
					log.Println(message)
					utils.SendEmail("12:30 pm Trade Update", message)
				}
			}
		}
	}

	return nil
}
