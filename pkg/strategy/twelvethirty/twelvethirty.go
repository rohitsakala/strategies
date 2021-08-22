package twelvethirty

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/avast/retry-go"
	"github.com/rohitsakala/strategies/pkg/broker"
	"github.com/rohitsakala/strategies/pkg/database"
	"github.com/rohitsakala/strategies/pkg/models"
	"github.com/rohitsakala/strategies/pkg/utils/duration"
	"github.com/rohitsakala/strategies/pkg/utils/options"
	kiteconnect "github.com/zerodha/gokiteconnect/v4"
	"go.mongodb.org/mongo-driver/bson"
)

type TwelveThirtyStrategy struct {
	StartTime time.Time
	EndTime   time.Time
	Broker    broker.Broker
	TimeZone  time.Location
	Database  database.Database
	Filter    bson.M
}

func NewTwelveThirtyStrategy(broker broker.Broker, timeZone time.Location, database database.Database) (TwelveThirtyStrategy, error) {
	// Create a collection in the database
	err := database.CreateCollection("twelvethirty")
	if err != nil {
		return TwelveThirtyStrategy{}, err
	}

	return TwelveThirtyStrategy{
		StartTime: time.Date(time.Now().In(&timeZone).Year(), time.Now().In(&timeZone).Month(), time.Now().In(&timeZone).Day(), 12, 25, 0, 0, &timeZone),
		EndTime:   time.Date(time.Now().In(&timeZone).Year(), time.Now().In(&timeZone).Month(), time.Now().In(&timeZone).Day(), 15, 20, 0, 0, &timeZone),
		Broker:    broker,
		TimeZone:  timeZone,
		Database:  database,
	}, nil
}

func (t TwelveThirtyStrategy) fetchData() (TwelveThiryStrategyPositions, error) {
	collectionRaw, err := t.Database.GetCollection(bson.D{}, "twelvethirty")
	if err != nil {
		return TwelveThiryStrategyPositions{}, err
	}
	if len(collectionRaw) <= 0 {
		return TwelveThiryStrategyPositions{}, err
	}

	var data TwelveThiryStrategyPositions

	dataBytes, err := bson.Marshal(collectionRaw)
	if err != nil {
		return TwelveThiryStrategyPositions{}, err
	}
	err = bson.Unmarshal(dataBytes, &data)
	if err != nil {
		return TwelveThiryStrategyPositions{}, err
	}

	t.Filter = bson.M{
		"_id": collectionRaw["_id"],
	}

	return data, nil
}

func (t TwelveThirtyStrategy) Start() error {
	var data TwelveThiryStrategyPositions
	data, err := t.fetchData()
	if err != nil {
		return err
	}

	log.Printf("Waiting for 12:25 pm to 12:35 pm....")
	startTime := time.Date(time.Now().In(&t.TimeZone).Year(), time.Now().In(&t.TimeZone).Month(), time.Now().In(&t.TimeZone).Day(), 12, 25, 0, 0, &t.TimeZone)
	endTime := time.Date(time.Now().In(&t.TimeZone).Year(), time.Now().In(&t.TimeZone).Month(), time.Now().In(&t.TimeZone).Day(), 12, 35, 0, 0, &t.TimeZone)

	for {
		if !duration.ValidateTime(startTime, endTime, t.TimeZone) {
			log.Printf("Time : %v", time.Now().In(&t.TimeZone))
			time.Sleep(1 * time.Minute)
		} else {
			break
		}
	}

	ceLeg, err := t.calculateLeg("CE")
	if err != nil {
		return err
	}
	data.SellCEOptionPosition = ceLeg
	log.Printf("Calculating CE Leg.... %s %d", ceLeg.TradingSymbol, ceLeg.Quantity)
	peLeg, err := t.calculateLeg("PE")
	if err != nil {
		return err
	}
	data.SellPEOptionPoistion = peLeg
	log.Printf("Calculating PE Leg.... %s %d", peLeg.TradingSymbol, peLeg.Quantity)

	dataBytes, err := bson.Marshal(data)
	if err != nil {
		return err
	}
	var dataMap bson.M
	err = bson.Unmarshal(dataBytes, &dataMap)
	if err != nil {
		return err
	}
	dataMapFull := bson.M{
		"$set": dataMap,
	}
	err = t.Database.UpdateCollection(t.Filter, dataMapFull, "twelvethirty")
	if err != nil {
		return err
	}

	err = t.placeLeg(&ceLeg, "Retrying placing leg")
	if err != nil {
		return err
	}
	log.Printf("Placing CE Leg with Avg Price %f", ceLeg.AveragePrice)
	err = t.placeLeg(&peLeg, "Retrying placing leg")
	if err != nil {
		return err
	}
	log.Printf("Placing PE Leg with Avg Price %f", peLeg.AveragePrice)

	ceStopLossLeg, err := t.calculateStopLossLeg(ceLeg)
	if err != nil {
		return err
	}
	data.SellCEStopLossOptionPosition = ceStopLossLeg
	peStopLossLeg, err := t.calculateStopLossLeg(peLeg)
	if err != nil {
		return err
	}
	data.SellPEStopLossOptionPosition = peStopLossLeg

	dataBytes, err = bson.Marshal(data)
	if err != nil {
		return err
	}
	err = bson.Unmarshal(dataBytes, &dataMap)
	if err != nil {
		return err
	}
	dataMapFull = bson.M{
		"$set": dataMap,
	}
	err = t.Database.UpdateCollection(t.Filter, dataMapFull, "twelvethirty")
	if err != nil {
		return err
	}

	err = t.placeLeg(&peStopLossLeg, "Retrying placing stoploss leg")
	if err != nil {
		return err
	}
	log.Printf("Placing PE StopLoss Leg with Trigger Price %f", peStopLossLeg.TriggerPrice)
	err = t.placeLeg(&ceStopLossLeg, "Retrying placing stoploss leg")
	if err != nil {
		return err
	}
	log.Printf("Placing CE StopLoss Leg with Trigger Price %f", ceStopLossLeg.TriggerPrice)

	startTime = time.Date(time.Now().In(&t.TimeZone).Year(), time.Now().In(&t.TimeZone).Month(), time.Now().In(&t.TimeZone).Day(), 15, 20, 0, 0, &t.TimeZone)
	endTime = time.Date(time.Now().In(&t.TimeZone).Year(), time.Now().In(&t.TimeZone).Month(), time.Now().In(&t.TimeZone).Day(), 15, 25, 0, 0, &t.TimeZone)

	log.Printf("Waiting for 3:20 to 3:25 pm....")
	for {
		if !duration.ValidateTime(startTime, endTime, t.TimeZone) {
			log.Printf("Time : %v", time.Now().In(&t.TimeZone))
			time.Sleep(1 * time.Minute)
		} else {
			break
		}
	}

	log.Printf("Cancelling all pending orders and current positions....")
	orderList := models.Positions{ceStopLossLeg, peStopLossLeg}
	err = t.cancelOrders(orderList)
	if err != nil {
		return err
	}

	log.Printf("Exiting current positions....")
	positionList := models.Positions{ceLeg, peLeg}
	err = t.cancelPositions(positionList)
	if err != nil {
		return err
	}

	data.SellPEOptionPoistion = models.Position{}
	data.SellCEOptionPosition = models.Position{}
	data.SellPEStopLossOptionPosition = models.Position{}
	data.SellCEStopLossOptionPosition = models.Position{}

	dataBytes, err = bson.Marshal(data)
	if err != nil {
		return err
	}
	err = bson.Unmarshal(dataBytes, &dataMap)
	if err != nil {
		return err
	}
	dataMapFull = bson.M{
		"$set": dataMap,
	}
	err = t.Database.UpdateCollection(t.Filter, dataMapFull, "twelvethirty")
	if err != nil {
		return err
	}

	return nil
}

func (t TwelveThirtyStrategy) cancelOrders(positions models.Positions) error {
	for _, position := range positions {
		err := retry.Do(
			func() error {
				err := t.Broker.CancelOrder(position)
				if err != nil {
					return err
				}
				return nil
			},
			retry.OnRetry(func(n uint, err error) {
				log.Println(fmt.Sprintf("%s because %s", "Retrying cancelling", err))
			}),
			retry.Delay(5*time.Second),
			retry.Attempts(5),
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t TwelveThirtyStrategy) cancelPositions(positions models.Positions) error {
	for _, position := range positions {
		position.TransactionType = kiteconnect.TransactionTypeBuy
		position.OrderID = ""
		err := t.placeLeg(&position, "Retrying cancelling leg")
		if err != nil {
			return err
		}
	}

	return nil
}

func (t TwelveThirtyStrategy) calculateLeg(optionType string) (models.Position, error) {
	leg := models.Position{
		Type:            optionType,
		Exchange:        kiteconnect.ExchangeNFO,
		TransactionType: "SELL",
		Product:         kiteconnect.ProductMIS,
		OrderType:       kiteconnect.OrderTypeMarket,
	}
	strikePrice, err := options.GetATM("NIFTY 50", t.Broker)
	if err != nil {
		return models.Position{}, err
	}

	legSymbol, err := options.GetSymbol("NIFTY", options.WEEK, 0, strikePrice, optionType, t.Broker)
	if err != nil {
		return models.Position{}, err
	}
	leg.TradingSymbol = legSymbol
	lotSize, err := options.GetLotSize(legSymbol, t.Broker)
	if err != nil {
		return models.Position{}, err
	}
	leg.LotSize = lotSize

	lotQuantity, err := strconv.Atoi(os.Getenv("TWELVE_THIRTY_LOT_QUANTITY"))
	if err != nil {
		return models.Position{}, err
	}
	leg.Quantity = lotQuantity * lotSize

	return leg, nil
}

func (t TwelveThirtyStrategy) calculateStopLossLeg(leg models.Position) (models.Position, error) {
	leg.TransactionType = kiteconnect.TransactionTypeBuy
	leg.Product = kiteconnect.ProductMIS
	leg.OrderType = kiteconnect.OrderTypeSLM
	leg.OrderID = ""
	leg.Status = ""

	stopLossPercentage := 30

	expiryDate := leg.Expiry.Time
	now := time.Now().In(&t.TimeZone)
	diff := now.Sub(expiryDate)

	switch int(diff.Hours() / 24) {
	case 0:
		stopLossPercentage = 70
	case 1:
		stopLossPercentage = 40
	}
	stopLossPrice := leg.AveragePrice * float64(stopLossPercentage) / 100
	stopLossPrice = stopLossPrice + leg.AveragePrice
	leg.TriggerPrice = float64(int(stopLossPrice*10)) / 10

	return leg, nil
}

func (t TwelveThirtyStrategy) placeLeg(leg *models.Position, retryMsg string) error {
	err := retry.Do(
		func() error {
			err := t.Broker.PlaceOrder(leg)
			if err != nil {
				return err
			}
			return nil
		},
		retry.OnRetry(func(n uint, err error) {
			log.Println(fmt.Sprintf("%s because %s", retryMsg, err))
		}),
		retry.Delay(5*time.Second),
		retry.Attempts(5),
	)
	if err != nil {
		return err
	}

	return nil
}
