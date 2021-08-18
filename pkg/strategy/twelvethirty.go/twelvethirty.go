package twelvethirty

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/rohitsakala/strategies/pkg/broker"
	"github.com/rohitsakala/strategies/pkg/database"
	"github.com/rohitsakala/strategies/pkg/models"
	"github.com/rohitsakala/strategies/pkg/utils/options"
	kiteconnect "github.com/zerodha/gokiteconnect/v4"
)

type TwelveThirtyStrategy struct {
	StartTime time.Time
	EndTime   time.Time
	Broker    broker.Broker
	TimeZone  time.Location
}

func NewTwelveThirtyStrategy(broker broker.Broker, timeZone time.Location, database database.Database) (TwelveThirtyStrategy, error) {
	// Create a collection in the database
	err := database.CreateCollection("twelvethirty")
	if err != nil {
		return TwelveThirtyStrategy{}, err

	}

	return TwelveThirtyStrategy{
		StartTime: time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 12, 25, 0, 0, &timeZone),
		EndTime:   time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 15, 20, 0, 0, &timeZone),
		Broker:    broker,
		TimeZone:  timeZone,
	}, nil
}

func (t TwelveThirtyStrategy) Start() error {
	// calculate CE and PE leg
	// check if there is data in the database
	// If data
	//    makeGroundTruth
	// else
	//    makeGroundTruth
	// Report Success and Failure

	// calculate ce leg
	ceLeg, err := t.calculateLeg("CE")
	if err != nil {
		return nil
	}
	log.Printf("%v", ceLeg)

	peLeg, err := t.calculateLeg("PE")
	if err != nil {
		return nil
	}
	log.Printf("%v", peLeg)

	// place the legs
	ceLeg, err = t.placeLeg(ceLeg)
	if err != nil {
		return err
	}
	log.Printf("Placed CE Leg with Avg Price %f", ceLeg.AveragePrice)
	peLeg, err = t.placeLeg(peLeg)
	if err != nil {
		return err
	}
	log.Printf("Placed PE Leg with Avg Price %f", peLeg.AveragePrice)

	// calculate the stoplosses
	ceStopLossPrice, err := t.getStopLoss(ceLeg.AveragePrice)
	if err != nil {
		return err
	}
	peStopLossPrice, err := t.getStopLoss(peLeg.AveragePrice)
	if err != nil {
		return err
	}

	// place the orders

	currentTime := time.Now()
	if currentTime.After(t.StartTime) && currentTime.Before(t.EndTime) {
		// Check if positions are already present

	} else {
		// Check if positions are present
		t.positionsPresent()
	}

	return nil
}

func (t TwelveThirtyStrategy) Stop() error {
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

func (t TwelveThirtyStrategy) getATMStrike() (float64, error) {
	strikePrice, err := t.Broker.GetLTP("NIFTY 50")
	if err != nil {
		return 0, err
	}

	return strikePrice, nil
}

func (t TwelveThirtyStrategy) placeLeg(leg models.Position) (models.Position, error) {
	position, err := t.Broker.PlaceOrder(leg)
	if err != nil {
		return models.Position{}, err
	}

	return position, nil
}

func (t TwelveThirtyStrategy) getStopLoss(leg models.Position) (models.Position, error) {
	date := time.Now().Date()
	switch leg.Expiry.Time. {
	case "Monday":
		stopLossPercentage = 30

	}

	return position, nil
}

func (t TwelveThirtyStrategy) positionsPresent() (bool, error) {
	strikePrice, err := t.getATMStrike()
	if err != nil {
		return false, err
	}

	var atmStrikePrice int

	moduleValue := strikePrice - 50
	if moduleValue > 25 {
		difference := 50 - moduleValue
		atmStrikePrice = int(strikePrice + difference)
	} else {
		atmStrikePrice = int(strikePrice - moduleValue)
	}

	// Weekly or Monthly ?
	fmt.Println(atmStrikePrice)

	return true, nil
}
