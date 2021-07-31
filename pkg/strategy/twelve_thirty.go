package strategy

import (
	"time"

	"github.com/rohitsakala/strategies/pkg/broker"
)

type TwelveThirtyStrategy struct {
	StartTime time.Time
	EndTime   time.Time
	Broker    broker.Broker
	TimeZone  time.Location
}

func NewTwelveThirtyStrategy(broker broker.Broker, timeZone time.Location) TwelveThirtyStrategy {
	return TwelveThirtyStrategy{
		StartTime: time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 12, 25, 0, 0, &timeZone),
		EndTime:   time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 15, 20, 0, 0, &timeZone),
		Broker:    broker,
		TimeZone:  timeZone,
	}
}

func (t TwelveThirtyStrategy) Start() {
	currentTime := time.Now()
	if currentTime.After(t.StartTime) && currentTime.Before(t.EndTime) {
		// Check if positions are already present

	} else {

	}
}

func (t TwelveThirtyStrategy) Stop() error {
	return nil
}

func (t TwelveThirtyStrategy) getATMStrike() (int, error) {
	t.Broker.
}

func (t TwelveThirtyStrategy) positionsPresent() (bool, error) {

}
