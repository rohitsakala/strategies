package options

import "github.com/rohitsakala/strategies/pkg/broker"

const (
	CURRENT_WEKLY = "current_week"
	CURRENT_MONTH = "current_month"
)

// GetSymbol will construct the symbol of the
// option according to the parameters given
func GetSymbol(symbol, expiry, strikePrice, optionType string, broker broker.Broker) string {
	if expiry == CURRENT_MONTH {
		broker.GetCurrentMonthyExpiry()
	}
}
