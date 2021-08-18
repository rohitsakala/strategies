package callcreditspread

import (
	"github.com/rohitsakala/strategies/pkg/models"
)

type CallCreditSpreadStrategyPositions struct {
	SellPEOptionPoistion  models.Position
	BuyCEOptionPosition   models.Position
	SellCEOptionsPosition models.Position
}
