package models

type Position struct {
	TradingSymbol string  `json:"tradingsymbol"`
	Exchange      string  `json:"exchange"`
	Product       string  `json:"product"`
	AveragePrice  float64 `json:"average_price"`
	Value         float64 `json:"value"`
	Quantity      int     `json:"quantity"`
	BuyPrice      float64 `json:"buy_price"`
	SellPrice     float64 `json:"sell_price"`
	StoplossPrice string  `json:"stoploss_price"`
	TargetPrice   string  `json:"target_price"`
}

type PositionList []Position

type OptionPosition struct {
	Position
	StrikePrice int
	Type        string
}
