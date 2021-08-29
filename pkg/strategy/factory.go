package strategy

import (
	"time"

	"github.com/rohitsakala/strategies/pkg/broker"
	"github.com/rohitsakala/strategies/pkg/database"
	"github.com/rohitsakala/strategies/pkg/strategy/callcreditspread"
	"github.com/rohitsakala/strategies/pkg/strategy/twelvethirty"
)

func GetStrategy(name string, broker broker.Broker, timeZone time.Location, database database.Database) (Strategy, error) {
	switch name {
	case "twelvethirty":
		twelvethirtyStrategy, err := twelvethirty.NewTwelveThirtyStrategy(broker, timeZone, database)
		if err != nil {
			return nil, err
		}
		return &twelvethirtyStrategy, nil
	case "callcreditspread":
		callcreditspread, err := callcreditspread.NewCallCreditSpreadStrategy(broker, timeZone, database)
		if err != nil {
			return nil, err
		}
		return &callcreditspread, nil
	}

	return nil, nil
}
