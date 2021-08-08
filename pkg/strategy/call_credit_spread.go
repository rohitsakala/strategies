package strategy

import (
	"log"
	"time"

	"github.com/rohitsakala/strategies/pkg/broker"
	"github.com/rohitsakala/strategies/pkg/utils/math"
	"github.com/rohitsakala/strategies/pkg/utils/options"
)

type CallCreditSpreadStrategy struct {
	SellingPEStopLoss           int
	SellingPEStopLossPercentage int
	SellingPEStopLossMultiple   int
	Broker                      broker.Broker
	TimeZone                    time.Location
}

func NewCallCreditSpreadStrategy(broker broker.Broker, timeZone time.Location) CallCreditSpreadStrategy {
	return CallCreditSpreadStrategy{
		Broker:                      broker,
		TimeZone:                    timeZone,
		SellingPEStopLossMultiple:   500,
		SellingPEStopLossPercentage: 10,
	}
}

func (c CallCreditSpreadStrategy) Start() error {
	// Start PE Selling leg

	// Get NIFTY 50 LTP
	LTP, err := c.Broker.GetLTP("NIFTY 50")
	if err != nil {
		return err
	}
	log.Printf("NIFTY 50 LTP : %d", LTP)

	// Get Floor 10 Percent of NIFTY 50
	floorLTP := math.GetFloorAfterPercentage(LTP, c.SellingPEStopLossPercentage, c.SellingPEStopLossMultiple)
	log.Printf("NIFTY Floor 10 percent LTP : %d", floorLTP)

	// Check if position already exists
	options.GetSymbol("NIFTY", options.CURRENT_MONTH, string(floorLTP), "PE", c.Broker)
	c.Broker.CheckPosition()

	// Start CE Buying leg
	// Start CE Selling leg
	return nil
}

func (c CallCreditSpreadStrategy) Stop() error {
	return nil
}
