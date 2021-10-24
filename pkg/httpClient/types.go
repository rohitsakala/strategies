package httpClient

import "time"

type CMD struct {
	TimeOfDay     time.Time `json:"t"`
	Open          float64   `json:"o"`
	High          float64   `json:"h"`
	Low           float64   `json:"l"`
	Close         float64   `json:"c"`
	Volume        int       `json:"v"`
	TimeFormatted string    `json:"tf"`
}

type SymbolQuote struct {
	S                string    `json:"s"`
	Change           float64   `json:"ch"`
	PercentageChange float64   `json:"chp"`
	LTP              float64   `json:"lp"`
	Spread           float64   `json:"spread"`
	AskingPrice      float64   `json:"ask"`
	BiddingPrice     float64   `json:"bid"`
	OpenPrice        float64   `json:"open_price"`
	HighPrice        float64   `json:"high_price"`
	LowPrice         float64   `json:"low_price"`
	PreviousClose    float64   `json:"prev_close_price"`
	Volume           int       `json:"volume"`
	ShortName        string    `json:"short_name"`
	Exchange         string    `json:"exchange"`
	Description      string    `json:"description"`
	OriginalName     string    `json:"original_name"`
	Symbol           string    `json:"symbol"`
	TimeOfDay        time.Time `json:"tt"`
	FYToken          string    `json:"fyToken"`
	CMD              CMD       `json:"cmd"`
}

type SymbolQuoteResponse struct {
	S     string      `json:"s"`
	Name  string      `json:"n"`
	Quote SymbolQuote `json:"v"`
}
