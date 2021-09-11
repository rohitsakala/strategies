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

const (
	TwelveThirtyStrategyDatabaseName = "twelvethirty"
)

type TwelveThirtyStrategy struct {
	EntryStartTime time.Time
	EntryEndTime   time.Time
	ExitStartTime  time.Time
	ExitEndTime    time.Time
	Data           TwelveThiryStrategyPositions
	Broker         broker.Broker
	TimeZone       time.Location
	Database       database.Database
	Filter         bson.M
}

func NewTwelveThirtyStrategy(broker broker.Broker, timeZone time.Location, database database.Database) (TwelveThirtyStrategy, error) {
	err := database.CreateCollection(TwelveThirtyStrategyDatabaseName)
	if err != nil {
		return TwelveThirtyStrategy{}, err
	}

	return TwelveThirtyStrategy{
		EntryStartTime: time.Date(time.Now().In(&timeZone).Year(), time.Now().In(&timeZone).Month(), time.Now().In(&timeZone).Day(), 12, 25, 0, 0, &timeZone),
		EntryEndTime:   time.Date(time.Now().In(&timeZone).Year(), time.Now().In(&timeZone).Month(), time.Now().In(&timeZone).Day(), 12, 35, 0, 0, &timeZone),
		ExitStartTime:  time.Date(time.Now().In(&timeZone).Year(), time.Now().In(&timeZone).Month(), time.Now().In(&timeZone).Day(), 15, 25, 0, 0, &timeZone),
		ExitEndTime:    time.Date(time.Now().In(&timeZone).Year(), time.Now().In(&timeZone).Month(), time.Now().In(&timeZone).Day(), 15, 30, 0, 0, &timeZone),
		Broker:         broker,
		TimeZone:       timeZone,
		Database:       database,
	}, nil
}

func (t *TwelveThirtyStrategy) fetchData() error {
	collectionRaw, err := t.Database.GetCollection(bson.D{}, TwelveThirtyStrategyDatabaseName)
	if err != nil {
		return err
	}
	if len(collectionRaw) <= 0 {
		insertID, err := t.Database.InsertCollection(t.Data, TwelveThirtyStrategyDatabaseName)
		if err != nil {
			return err
		}
		t.Filter = bson.M{
			"_id": insertID,
		}

		return nil
	}

	dataBytes, err := bson.Marshal(collectionRaw)
	if err != nil {
		return err
	}
	err = bson.Unmarshal(dataBytes, &t.Data)
	if err != nil {
		return err
	}
	t.Filter = bson.M{
		"_id": collectionRaw["_id"],
	}

	return nil
}

func (t *TwelveThirtyStrategy) Start() error {
	log.Printf("Waiting for 12:25 pm to 12:35 pm....")
	for {
		if !duration.ValidateTime(t.EntryStartTime, t.EntryEndTime, t.TimeZone) {
			time.Sleep(1 * time.Minute)
			log.Printf("Time : %v", time.Now().In(&t.TimeZone))
		} else {
			log.Printf("Time : %v", time.Now().In(&t.TimeZone))
			break
		}
	}
	log.Printf("Entering 12:25 pm to 12:35 pm.")

	err := t.fetchData()
	if err != nil {
		return err
	}

	strikePrice, err := options.GetATM("NIFTY 50", t.Broker)
	if err != nil {
		return err
	}

	if t.Data.SellCEOptionPosition.TradingSymbol == "" {
		t.Data.SellCEOptionPosition, err = t.calculateLeg("CE", strikePrice)
		if err != nil {
			return err
		}
		log.Printf("Calculating CE Leg.... %s %d", t.Data.SellCEOptionPosition.TradingSymbol, t.Data.SellCEOptionPosition.Quantity)
		err = t.placeLeg(&t.Data.SellCEOptionPosition, "Retrying placing leg")
		if err != nil {
			return err
		}
		log.Printf("Placing CE Leg with Avg Price %f", t.Data.SellCEOptionPosition.AveragePrice)
	}
	if t.Data.SellPEOptionPoistion.TradingSymbol == "" {
		t.Data.SellPEOptionPoistion, err = t.calculateLeg("PE", strikePrice)
		if err != nil {
			return err
		}
		log.Printf("Calculating PE Leg.... %s %d", t.Data.SellPEOptionPoistion.TradingSymbol, t.Data.SellPEOptionPoistion.Quantity)
		err = t.placeLeg(&t.Data.SellPEOptionPoistion, "Retrying placing leg")
		if err != nil {
			return err
		}
		log.Printf("Placing PE Leg with Avg Price %f", t.Data.SellPEOptionPoistion.AveragePrice)
	}
	err = t.Database.UpdateCollection(t.Filter, t.Data, "twelvethirty")
	if err != nil {
		return err
	}

	if t.Data.SellCEStopLossOptionPosition.TradingSymbol == "" {
		t.Data.SellCEStopLossOptionPosition, err = t.calculateStopLossLeg(t.Data.SellCEOptionPosition)
		if err != nil {
			return err
		}
		err = t.placeLeg(&t.Data.SellCEStopLossOptionPosition, "Retrying placing stoploss leg")
		if err != nil {
			return err
		}
		log.Printf("Placing CE StopLoss Leg with Trigger Price %f", t.Data.SellCEStopLossOptionPosition.TriggerPrice)
	}
	if t.Data.SellPEStopLossOptionPosition.TradingSymbol == "" {
		t.Data.SellPEStopLossOptionPosition, err = t.calculateStopLossLeg(t.Data.SellPEOptionPoistion)
		if err != nil {
			return err
		}
		err = t.placeLeg(&t.Data.SellPEStopLossOptionPosition, "Retrying placing stoploss leg")
		if err != nil {
			return err
		}
		log.Printf("Placing PE StopLoss Leg with Trigger Price %f", t.Data.SellPEStopLossOptionPosition.TriggerPrice)
	}
	if err = t.Database.UpdateCollection(t.Filter, t.Data, "twelvethirty"); err != nil {
		return err
	}

	t.WaitAndWatch()

	return nil
}

func (t *TwelveThirtyStrategy) Stop() error {
	log.Printf("Cancelling all pending orders...")
	stopLossLegs := models.RefPositions{&t.Data.SellCEStopLossOptionPosition, &t.Data.SellPEStopLossOptionPosition}
	err := t.Broker.CancelOrders(stopLossLegs)
	if err != nil {
		return err
	}
	log.Printf("Cancelled all pending orders.")

	log.Printf("Exiting all current positions...")
	positionList := models.Positions{}
	if t.Data.SellCEStopLossOptionPosition.Status != kiteconnect.OrderStatusComplete {
		positionList = append(positionList, t.Data.SellCEOptionPosition)
	}
	if t.Data.SellPEOptionPoistion.Status != kiteconnect.OrderStatusComplete {
		positionList = append(positionList, t.Data.SellPEOptionPoistion)
	}
	err = t.cancelPositions(positionList)
	if err != nil {
		return err
	}
	log.Printf("Exited all current positions.")

	t.Data.SellPEOptionPoistion = models.Position{}
	t.Data.SellCEOptionPosition = models.Position{}
	t.Data.SellPEStopLossOptionPosition = models.Position{}
	t.Data.SellCEStopLossOptionPosition = models.Position{}
	err = t.Database.UpdateCollection(t.Filter, t.Data, "twelvethirty")
	if err != nil {
		return err
	}

	return nil
}

func (t *TwelveThirtyStrategy) WaitAndWatch() {
	log.Printf("Waiting for 3:25 to 3:30 pm....")
	for {
		if !duration.ValidateTime(t.ExitStartTime, t.ExitEndTime, t.TimeZone) {
			time.Sleep(1 * time.Minute)

			log.Printf("Time : %v", time.Now().In(&t.TimeZone))
		} else {
			log.Printf("Time : %v", time.Now().In(&t.TimeZone))
			break
		}
	}
}

func (t *TwelveThirtyStrategy) cancelPositions(positions models.Positions) error {
	for _, position := range positions {
		position.TransactionType = kiteconnect.TransactionTypeBuy
		position.Status = ""
		position.OrderID = ""
		err := t.placeLeg(&position, fmt.Sprintf("%s %s", "Retrying cancelling position ", position.TradingSymbol))
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *TwelveThirtyStrategy) calculateLeg(optionType string, strikePrice float64) (models.Position, error) {
	leg := models.Position{
		Type:            optionType,
		Exchange:        kiteconnect.ExchangeNFO,
		TransactionType: "SELL",
		Product:         kiteconnect.ProductNRML,
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
	leg.Product = kiteconnect.ProductNRML
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
	stopLossPrice := leg.AveragePrice * float64(stopLossPercentage) / 100
	stopLossPrice = stopLossPrice + leg.AveragePrice
	leg.TriggerPrice = float64(int(stopLossPrice*10)) / 10
	leg.Price = float64(int(leg.TriggerPrice) + 5)

	return leg, nil
}

func (t *TwelveThirtyStrategy) placeLeg(leg *models.Position, retryMsg string) error {
	var err error

	err = retry.Do(
		func() error {
			if leg.OrderType == kiteconnect.OrderTypeLimit {
				leg.Price, err = options.GetLTPNoFreak(leg.TradingSymbol, t.Broker)
				if err != nil {
					return err
				}
			}
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
