package models

import "time"

type Position struct {
	TradingSymbol   string    `json:"tradingsymbol"`
	Exchange        string    `json:"exchange"`
	Product         string    `json:"product"`
	AveragePrice    float64   `json:"average_price"`
	Value           float64   `json:"value"`
	Quantity        int       `json:"quantity"`
	BuyPrice        float64   `json:"buy_price"`
	SellPrice       float64   `json:"sell_price"`
	StoplossPrice   float64   `json:"stoploss_price"`
	Price           float64   `json:"price"`
	TriggerPrice    float64   `json:"trigger_price"`
	TargetPrice     string    `json:"target_price"`
	StrikePrice     float64   `json:"strike_price"`
	Type            string    `json:"type"`
	LotSize         int       `json:"lot_size"`
	InstrumentToken int       `json:"instrument_token"`
	ExchangeToken   int       `json:"exchange_token"`
	Name            string    `json:"name"`
	LastPrice       float64   `json:"last_price"`
	Expiry          time.Time `json:"expiry"`
	TickSize        float64   `json:"tick_size"`
	InstrumentType  string    `json:"instrument_type"`
	Segment         string    `json:"segment"`
	OrderType       string    `json:"order_type"`
	TransactionType string    `json:"transaction_type"`
	OrderID         string    `json:"order_id"`
	Status          string    `json:"status"`
}

type Positions []Position

type Credentials struct {
	AccessToken string
}
