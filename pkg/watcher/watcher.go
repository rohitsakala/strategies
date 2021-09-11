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

	orderType := position.OrderType
	switch orderType {
	case kiteconnect.OrderTypeSL:
		for _, order := range orders {
			if order.OrderID == position.OrderID {
				log.Printf("Order Status right now %s", position.Status)
				//if order.Status == "TRIGGER PENDING" {
				if order.Status == "AMO REQ RECEIVED" {
					if order.Status != position.Status {
						message := fmt.Sprintf("Order %s Changed from %s to %s", position.TradingSymbol, position.Status, order.Status)
						log.Println(message)

						position.Status = order.Status
						err := utils.SendEmail("12:30 pm Trade Update", message)
						if err != nil {
							return err
						}
						switch order.Status {
						case "OPEN":
							time.Sleep(10 * time.Second)
							// Place Limt
						}
					}
				}
			}
		}
	}

	return nil
}
