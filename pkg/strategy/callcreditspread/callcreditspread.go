package callcreditspread

import (
	"log"
	"time"

	"github.com/rohitsakala/strategies/pkg/broker"
	"github.com/rohitsakala/strategies/pkg/database"
	"github.com/rohitsakala/strategies/pkg/models"
	"github.com/rohitsakala/strategies/pkg/utils/math"
	"github.com/rohitsakala/strategies/pkg/utils/options"
	"go.mongodb.org/mongo-driver/bson"
)

type CallCreditSpreadStrategy struct {
	SellingPEStopLoss              int
	SellingPEStrikePricePercentage int
	SellingPEStopLossMultiple      int
	Broker                         broker.Broker
	TimeZone                       time.Location
	Database                       database.Database
}

func NewCallCreditSpreadStrategy(broker broker.Broker, timeZone time.Location, database database.Database) (CallCreditSpreadStrategy, error) {
	// Create a collection in the database
	err := database.CreateCollection("callcreditspread")
	if err != nil {
		return CallCreditSpreadStrategy{}, err

	}

	return CallCreditSpreadStrategy{
		Broker:                         broker,
		TimeZone:                       timeZone,
		SellingPEStopLossMultiple:      500,
		SellingPEStrikePricePercentage: 13,
		Database:                       database,
	}, nil
}

func (c *CallCreditSpreadStrategy) Start() error {
	// Start PE Selling leg
	sellPEPosition := models.Position{Type: "PE"}

	// Check if database has PE Selling leg
	collectionRaw, err := c.Database.GetCollection(bson.D{}, "callcreditspread")
	if err != nil {
		return err
	}
	if collectionRaw == nil {
	}

	// Get NIFTY 50 LTP
	LTP, err := c.Broker.GetLTP("NIFTY 50")
	if err != nil {
		return err
	}
	log.Printf("NIFTY 50 LTP : %f", LTP)

	// Get Floor 10 Percent of NIFTY 50
	floorLTP := math.GetFloorAfterPercentage(LTP, c.SellingPEStrikePricePercentage, c.SellingPEStopLossMultiple)
	sellPEPosition.StrikePrice = floorLTP
	log.Printf("NIFTY Floor 10 percent LTP : %f", floorLTP)

	// Construct PE symbol
	PEOptionSymbol, err := options.GetSymbol("NIFTY", "MONTH", 1, floorLTP, "PE", c.Broker)
	if err != nil {
		return err
	}
	sellPEPosition.TradingSymbol = PEOptionSymbol
	log.Printf("NIFTY PE Symbol : %s", PEOptionSymbol)

	sellPEPosition.SellPrice, err = c.Broker.GetLTP(sellPEPosition.TradingSymbol)
	if err != nil {
		return err
	}
	log.Printf("NIFTY PE Symbol LTP : %f", sellPEPosition.SellPrice)

	// Check if position already exists
	_, err = c.Broker.CheckPosition(sellPEPosition.TradingSymbol)
	if err != nil {
		return err
	}

	// Start CE Buying leg
	// Start CE Selling leg
	return nil
}
