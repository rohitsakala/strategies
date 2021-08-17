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
	CELeg := models.OptionPosition{
		Type: "CE",
	}
	strikePrice, err := options.GetATM("NIFTY 50", t.Broker)
	if err != nil {
		return err
	}
	optionType := CELeg.Type

	CEOptionSymbol, lotSize, err := options.GetSymbol("NIFTY", options.WEEK, 0, strikePrice, optionType, t.Broker)
	if err != nil {
		return err
	}
	CELeg.TradingSymbol = CEOptionSymbol
	log.Printf("CE leg Trading Symbol %s", CELeg.TradingSymbol)
	lotQuantity, err := strconv.Atoi(os.Getenv("TWELVE_THIRTY_LOT_QUANTITY"))
	if err != nil {
		return err
	}
	CELeg.LotQuantity = float64(lotQuantity)
	CELeg.LotSize = int(lotSize)

	// calculate pe leg
	PELeg := models.OptionPosition{
		Type: "PE",
	}
	strikePrice, err = options.GetATM("NIFTY 50", t.Broker)
	if err != nil {
		return err
	}
	optionType = PELeg.Type

	PEOptionSymbol, lotSize, err := options.GetSymbol("NIFTY", options.WEEK, 0, strikePrice, optionType, t.Broker)
	if err != nil {
		return err
	}
	PELeg.TradingSymbol = PEOptionSymbol
	log.Printf("PE leg Trading Symbol %s", PELeg.TradingSymbol)
	lotQuantity, err = strconv.Atoi(os.Getenv("TWELVE_THIRTY_LOT_QUANTITY"))
	if err != nil {
		return err
	}
	PELeg.LotQuantity = float64(lotQuantity)
	PELeg.LotSize = int(lotSize)

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

func (t TwelveThirtyStrategy) getATMStrike() (float64, error) {
	strikePrice, err := t.Broker.GetLTP("NIFTY 50")
	if err != nil {
		return 0, err
	}

	return strikePrice, nil
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
