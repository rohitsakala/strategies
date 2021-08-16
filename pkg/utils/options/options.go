package options

import (
	"fmt"

	"github.com/rohitsakala/strategies/pkg/broker"
)

const (
	CURRENT_WEKLY = "current_week"
	CURRENT_MONTH = "current_month"
)

// GetSymbol will construct the symbol of the
// option according to the parameters given
func GetSymbol(expiry, strikePrice, optionType string, broker broker.Broker) (string, error) {
	if expiry == CURRENT_MONTH {
		monthExpiryDate, err := broker.GetCurrentMonthyExpiry()
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("%s%s%s", monthExpiryDate, strikePrice, optionType), nil
	}

	return "", nil
}
