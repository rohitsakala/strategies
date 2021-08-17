package models

import "github.com/zerodha/gokiteconnect/v4/models"

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
	StrikePrice float64
	Type        string
	LotSize     int
	LotQuantity float64
}

type Instrument struct {
	InstrumentToken int         `json:"instrument_token"`
	ExchangeToken   int         `json:"exchange_token"`
	Tradingsymbol   string      `json:"tradingsymbol"`
	Name            string      `json:"name"`
	LastPrice       float64     `json:"last_price"`
	Expiry          models.Time `json:"expiry"`
	StrikePrice     float64     `json:"strike"`
	TickSize        float64     `json:"tick_size"`
	LotSize         float64     `json:"lot_size"`
	InstrumentType  string      `json:"instrument_type"`
	Segment         string      `json:"segment"`
	Exchange        string      `json:"exchange"`
}

type Instruments []Instrument
