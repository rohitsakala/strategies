package callcreditspread

import (
	"github.com/rohitsakala/strategies/pkg/models"
)

type CallCreditSpreadStrategyPositions struct {
	SellPEOptionPoistion  models.OptionPosition
	BuyCEOptionPosition   models.OptionPosition
	SellCEOptionsPosition models.OptionPosition
}
