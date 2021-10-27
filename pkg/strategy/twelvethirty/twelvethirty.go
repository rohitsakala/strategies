package twelvethirty

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/rohitsakala/strategies/pkg/broker"
	"github.com/rohitsakala/strategies/pkg/models"
	"github.com/rohitsakala/strategies/pkg/utils"
	"github.com/rohitsakala/strategies/pkg/utils/duration"
	"github.com/rohitsakala/strategies/pkg/utils/options"
	"github.com/rohitsakala/strategies/pkg/watcher"
	kiteconnect "github.com/zerodha/gokiteconnect/v4"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	TwelveThirtyStrategyDatabaseName = "twelvethirty"
)

type TwelveThirtyStrategy struct {
	EntryStartTime  time.Time
	EntryEndTime    time.Time
	ExitStartTime   time.Time
	ExitEndTime     time.Time
	Data            TwelveThiryStrategyPositions
	Broker          broker.Broker
	TimeZone        time.Location
	Filter          bson.M
	Watcher         watcher.Watcher
	ProductType     string
	StopLossVariant string
}

func NewTwelveThirtyStrategy(broker broker.Broker, timeZone time.Location, watcher watcher.Watcher, productType, stopLossVariant string) (TwelveThirtyStrategy, error) {
	return TwelveThirtyStrategy{
		EntryStartTime:  time.Date(time.Now().In(&timeZone).Year(), time.Now().In(&timeZone).Month(), time.Now().In(&timeZone).Day(), 12, 28, 0, 0, &timeZone),
		EntryEndTime:    time.Date(time.Now().In(&timeZone).Year(), time.Now().In(&timeZone).Month(), time.Now().In(&timeZone).Day(), 15, 20, 0, 0, &timeZone),
		ExitStartTime:   time.Date(time.Now().In(&timeZone).Year(), time.Now().In(&timeZone).Month(), time.Now().In(&timeZone).Day(), 15, 20, 0, 0, &timeZone),
		ExitEndTime:     time.Date(time.Now().In(&timeZone).Year(), time.Now().In(&timeZone).Month(), time.Now().In(&timeZone).Day(), 15, 30, 0, 0, &timeZone),
		Broker:          broker,
		TimeZone:        timeZone,
		Watcher:         watcher,
		ProductType:     productType,
		StopLossVariant: stopLossVariant,
	}, nil
}

func (t *TwelveThirtyStrategy) Start() error {
	// Check if markets are open today ?
	open, err := t.Broker.IsMarketOpen()
	if err != nil {
		return err
	}
	if !open {
		log.Println("Market is closed")
		return nil
	}

	log.Printf("Waiting for 12:28 pm to 15:20 pm....")
	for {
		if !duration.ValidateTime(t.EntryStartTime, t.EntryEndTime, t.TimeZone) {
			time.Sleep(1 * time.Minute)
			log.Printf("Time : %v", time.Now().In(&t.TimeZone))
		} else {
			log.Printf("Time : %v", time.Now().In(&t.TimeZone))
			break
		}
	}
	log.Printf("Entering 12:28 pm to 15:20 pm.")

	strikePrice, err := options.GetATM("NIFTY 50", t.Broker)
	if err != nil {
		return err
	}

	t.Data.BuyCEOptionPosition, err = t.calculateLeg("CE", strikePrice+500, kiteconnect.TransactionTypeBuy)
	if err != nil {
		return err
	}
	log.Printf("Calculating Buy CE Leg.... %s %d", t.Data.BuyCEOptionPosition.TradingSymbol, t.Data.BuyCEOptionPosition.Quantity)
	err = t.Broker.PlaceOrder(&t.Data.BuyCEOptionPosition)
	if err != nil {
		return err
	}
	log.Printf("Placing Buy CE Leg with Avg Price %f", t.Data.BuyCEOptionPosition.AveragePrice)
	err = utils.SendEmail("Twelve Thirty PM Trade Update", fmt.Sprintf("Placed Buy CE Leg with Avg Price %f", t.Data.BuyCEOptionPosition.AveragePrice))
	if err != nil {
		return err
	}

	t.Data.BuyPEOptionPoistion, err = t.calculateLeg("PE", strikePrice-500, kiteconnect.TransactionTypeBuy)
	if err != nil {
		return err
	}
	log.Printf("Calculating Buy PE Leg.... %s %d", t.Data.BuyPEOptionPoistion.TradingSymbol, t.Data.BuyPEOptionPoistion.Quantity)
	err = t.Broker.PlaceOrder(&t.Data.BuyPEOptionPoistion)
	if err != nil {
		return err
	}
	log.Printf("Placing Buy PE Leg with Avg Price %f", t.Data.BuyPEOptionPoistion.AveragePrice)
	err = utils.SendEmail("Twelve Thirty PM Trade Update", fmt.Sprintf("Placed Buy PE Leg with Avg Price %f", t.Data.BuyPEOptionPoistion.AveragePrice))
	if err != nil {
		return err
	}

	t.Data.SellCEOptionPosition, err = t.calculateLeg("CE", strikePrice, kiteconnect.TransactionTypeSell)
	if err != nil {
		return err
	}
	log.Printf("Calculating CE Leg.... %s %d", t.Data.SellCEOptionPosition.TradingSymbol, t.Data.SellCEOptionPosition.Quantity)
	err = t.Broker.PlaceOrder(&t.Data.SellCEOptionPosition)
	if err != nil {
		return err
	}
	log.Printf("Placing CE Leg with Avg Price %f", t.Data.SellCEOptionPosition.AveragePrice)
	err = utils.SendEmail("Twelve Thirty PM Trade Update", fmt.Sprintf("Placed CE Leg with Avg Price %f", t.Data.SellCEOptionPosition.AveragePrice))
	if err != nil {
		return err
	}

	t.Data.SellPEOptionPoistion, err = t.calculateLeg("PE", strikePrice, kiteconnect.TransactionTypeSell)
	if err != nil {
		return err
	}
	log.Printf("Calculating PE Leg.... %s %d", t.Data.SellPEOptionPoistion.TradingSymbol, t.Data.SellPEOptionPoistion.Quantity)
	err = t.Broker.PlaceOrder(&t.Data.SellPEOptionPoistion)
	if err != nil {
		return err
	}
	log.Printf("Placing PE Leg with Avg Price %f", t.Data.SellPEOptionPoistion.AveragePrice)
	err = utils.SendEmail("Twelve Thirty PM Trade Update", fmt.Sprintf("Placed PE Leg with Avg Price %f", t.Data.SellPEOptionPoistion.AveragePrice))
	if err != nil {
		return err
	}

	t.Data.SellCEStopLossOptionPosition, err = t.calculateStopLossLeg(t.Data.SellCEOptionPosition)
	if err != nil {
		return err
	}
	err = t.Broker.PlaceOrder(&t.Data.SellCEStopLossOptionPosition)
	if err != nil {
		return err
	}
	log.Printf("Placing CE StopLoss Leg with Trigger Price %f", t.Data.SellCEStopLossOptionPosition.TriggerPrice)
	err = utils.SendEmail("Twelve Thirty PM Trade Update", fmt.Sprintf("Placed CE Stop Loss Leg with Trigger Price %f", t.Data.SellCEStopLossOptionPosition.TriggerPrice))
	if err != nil {
		return err
	}

	t.Data.SellPEStopLossOptionPosition, err = t.calculateStopLossLeg(t.Data.SellPEOptionPoistion)
	if err != nil {
		return err
	}
	err = t.Broker.PlaceOrder(&t.Data.SellPEStopLossOptionPosition)
	if err != nil {
		return err
	}
	log.Printf("Placing PE StopLoss Leg with Trigger Price %f", t.Data.SellPEStopLossOptionPosition.TriggerPrice)
	err = utils.SendEmail("Twelve Thirty PM Trade Update", fmt.Sprintf("Placed PE Stop Loss Leg with Trigger Price %f", t.Data.SellPEStopLossOptionPosition.TriggerPrice))
	if err != nil {
		return err
	}

	err = t.WaitAndWatch()
	if err != nil {
		return err
	}

	return nil
}

func (t *TwelveThirtyStrategy) Stop() error {
	// Check if markets are open today ?
	open, err := t.Broker.IsMarketOpen()
	if err != nil {
		return err
	}
	if !open {
		log.Println("Market is closed")
		return nil
	}
	log.Printf("Cancelling all pending orders...")
	stopLossLegs := models.RefPositions{&t.Data.SellCEStopLossOptionPosition, &t.Data.SellPEStopLossOptionPosition}
	err = t.Broker.CancelOrders(stopLossLegs)
	if err != nil {
		return err
	}
	log.Printf("Cancelled all pending orders.")
	err = utils.SendEmail("Twelve Thirty PM Trade Update", fmt.Sprintf("Cancelled Stop Loss orders %s %s", t.Data.SellPEStopLossOptionPosition.TradingSymbol, t.Data.SellCEStopLossOptionPosition.TradingSymbol))
	if err != nil {
		return err
	}

	log.Printf("Exiting all current positions...")
	positionList := models.Positions{}
	if t.Data.SellCEStopLossOptionPosition.Status != kiteconnect.OrderStatusComplete {
		positionList = append(positionList, t.Data.SellCEOptionPosition)
	}
	if t.Data.SellPEStopLossOptionPosition.Status != kiteconnect.OrderStatusComplete {
		positionList = append(positionList, t.Data.SellPEOptionPoistion)
	}
	positionList = append(positionList, t.Data.BuyPEOptionPoistion)
	positionList = append(positionList, t.Data.BuyCEOptionPosition)
	err = t.cancelPositions(positionList)
	if err != nil {
		return err
	}
	log.Printf("Exited all current positions.")
	for _, position := range positionList {
		err = utils.SendEmail("Twelve Thirty PM Trade Update", fmt.Sprintf("Cancelled position %s", position.TradingSymbol))
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *TwelveThirtyStrategy) WaitAndWatch() error {
	log.Printf("Waiting for 15:20 to 15:30 pm....")
	for {
		if !duration.ValidateTime(t.ExitStartTime, t.ExitEndTime, t.TimeZone) {
			time.Sleep(1 * time.Minute)
			err := t.Watcher.Watch(&t.Data.SellCEStopLossOptionPosition)
			if err != nil {
				return err
			}
			err = t.Watcher.Watch(&t.Data.SellPEStopLossOptionPosition)
			if err != nil {
				return err
			}
			log.Printf("Time : %v", time.Now().In(&t.TimeZone))
		} else {
			log.Printf("Time : %v", time.Now().In(&t.TimeZone))
			break
		}
	}

	return nil
}

func (t *TwelveThirtyStrategy) cancelPositions(positions models.Positions) error {
	for _, position := range positions {
		if position.TransactionType == kiteconnect.TransactionTypeBuy {
			position.TransactionType = kiteconnect.TransactionTypeSell
		} else if position.TransactionType == kiteconnect.TransactionTypeSell {
			position.TransactionType = kiteconnect.TransactionTypeBuy
		}
		position.Status = ""
		position.OrderID = ""
		err := t.Broker.PlaceOrder(&position)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *TwelveThirtyStrategy) calculateLeg(optionType string, strikePrice float64, transactionType string) (models.Position, error) {
	leg := models.Position{
		Type:            optionType,
		Exchange:        kiteconnect.ExchangeNFO,
		TransactionType: transactionType,
		Product:         t.ProductType,
		OrderType:       kiteconnect.OrderTypeLimit,
	}

	legSymbol, err := options.GetSymbol("NIFTY", options.WEEK, 0, strikePrice, optionType, t.Broker)
	if err != nil {
		return models.Position{}, err
	}
	leg.TradingSymbol = legSymbol

	leg.LotSize, err = options.GetLotSize(legSymbol, t.Broker)
	if err != nil {
		return models.Position{}, err
	}

	lotQuantity, err := strconv.Atoi(os.Getenv("TWELVE_THIRTY_LOT_QUANTITY"))
	if err != nil {
		return models.Position{}, err
	}
	leg.Quantity = lotQuantity * leg.LotSize

	leg.Expiry, err = options.GetExpiry("NIFTY", options.WEEK, 0, strikePrice, optionType, t.Broker)
	if err != nil {
		return models.Position{}, err
	}

	return leg, nil
}

func (t *TwelveThirtyStrategy) calculateStopLossLeg(leg models.Position) (models.Position, error) {
	leg.TransactionType = kiteconnect.TransactionTypeBuy
	leg.Product = t.ProductType
	leg.OrderType = kiteconnect.OrderTypeSL
	leg.OrderID = ""
	leg.Status = ""

	stopLossPercentage := 30

	expiryDate := leg.Expiry
	now := time.Now().In(&t.TimeZone)
	diff := expiryDate.Sub(now)

	if int(diff.Hours()) < 0 {
		stopLossPercentage = 70
	} else if int(diff.Hours()/24) == 0 {
		stopLossPercentage = 40
	}
	// If it is fixed then we take 30 percent SL percentage always
	if t.StopLossVariant == "fixed" {
		stopLossPercentage = 30
	}
	stopLossPrice := leg.AveragePrice * float64(stopLossPercentage) / 100
	stopLossPrice = stopLossPrice + leg.AveragePrice
	leg.TriggerPrice = float64(int(stopLossPrice*10)) / 10
	leg.Price = float64(int(leg.TriggerPrice) + 5)

	return leg, nil
}
