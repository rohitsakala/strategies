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

// Watch ensures that any order which is being
// watched gets to completed status.
// This method is meant to run in a loop.
func (w *Watcher) Watch(position *models.Position) error {
	var orderP models.Position
	orders, err := w.Broker.GetOrders()
	if err != nil {
		return err
	}

	for _, order := range orders {
		if order.OrderID == position.OrderID {
			orderP = order
		}
	}

	switch position.OrderType {
	// TRIGGER PENDING - Send Email once if Status changes
	// COMPLETE - Send Email once if Status changes
	// OPEN - Convert to Limit Order
	case kiteconnect.OrderTypeSL:
		switch position.Status {
		case "TRIGGER PENDING", kiteconnect.OrderStatusComplete:
			if orderP.Status != position.Status {
				message := fmt.Sprintf("Order %s Changed from %s to %s", position.TradingSymbol, position.Status, orderP.Status)
				log.Println(message)
				err = utils.SendEmail("12:30 pm Trade Update", message)
				if err != nil {
					return err
				}
				position.Status = orderP.Status
			}
		case "OPEN":
			position.OrderType = kiteconnect.OrderTypeLimit
		}
	// COMPLETE - Send Email once if Status changes
	// OPEN - Complete the order
	case kiteconnect.OrderTypeLimit:
		switch position.Status {
		case kiteconnect.OrderStatusComplete:
			if orderP.Status != position.Status {
				message := fmt.Sprintf("Order %s Changed from %s to %s", position.TradingSymbol, position.Status, orderP.Status)
				log.Println(message)
				err = utils.SendEmail("12:30 pm Trade Update", message)
				if err != nil {
					return err
				}
				position.Status = orderP.Status
			}
		case "OPEN":
			err = w.Broker.PlaceOrder(position)
			if err != nil {
				return err
			}
			message := fmt.Sprintf("Order %s Changed from OPEN to %s", position.TradingSymbol, position.Status)
			log.Println(message)
			err = utils.SendEmail("12:30 pm Trade Update", message)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
