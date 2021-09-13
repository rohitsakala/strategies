package twelvethirty

import (
	"github.com/rohitsakala/strategies/pkg/models"
)

type TwelveThiryStrategyPositions struct {
	SellPEOptionPoistion         models.Position
	SellCEOptionPosition         models.Position
	BuyPEOptionPoistion          models.Position
	BuyCEOptionPosition          models.Position
	SellPEStopLossOptionPosition models.Position
	SellCEStopLossOptionPosition models.Position
}
