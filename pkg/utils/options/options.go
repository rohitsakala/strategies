package options

import (
	"sort"
	"strings"

	"github.com/rohitsakala/strategies/pkg/broker"
	"github.com/rohitsakala/strategies/pkg/models"
	"github.com/rohitsakala/strategies/pkg/utils/math"
)

const (
	WEEK  = "week"
	MONTH = "month"
)

type PositionSorter []models.Position

func (s PositionSorter) Len() int      { return len(s) }
func (s PositionSorter) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s PositionSorter) Less(i, j int) bool {
	return s[i].Expiry.Before(s[j].Expiry.Time)
}

// GetSymbol will construct the symbol of the
// option according to the parameters given
func GetSymbol(symbol, expiryType string, expiryOffset int, strikePrice float64, optionType string, broker broker.Broker) (string, error) {
	instruments, err := broker.GetInstruments("NFO")
	if err != nil {
		return "", err
	}

	filteredInstruments := models.Positions{}
	for _, instrument := range instruments {
		if strings.HasPrefix(instrument.TradingSymbol, symbol) && instrument.Segment == "NFO-OPT" && instrument.StrikePrice == strikePrice && instrument.Exchange == "NFO" && instrument.InstrumentType == optionType {
			filteredInstruments = append(filteredInstruments, instrument)
		}
	}
	sort.Sort(PositionSorter(filteredInstruments))

	switch expiryType {
	case MONTH:
		resultSymbol := filteredInstruments[0].TradingSymbol
		month := filteredInstruments[0].Expiry.Month()
		for i := 1; i < len(filteredInstruments); i++ {
			if filteredInstruments[i].Expiry.Month() != month {
				return symbol, nil
			}
			resultSymbol = filteredInstruments[i].TradingSymbol
		}

		return resultSymbol, nil
	case WEEK:
		return filteredInstruments[0].TradingSymbol, nil
	}

	return "", nil
}

// GetLotSize will return lotsize of the symbol
func GetLotSize(symbol string, broker broker.Broker) (int, error) {
	instrument, err := broker.GetInstrument(symbol, "NFO")
	if err != nil {
		return -1, err
	}

	return instrument.LotSize, nil
}

func GetATM(symbol string, broker broker.Broker) (float64, error) {
	ltp, err := broker.GetLTP(symbol)
	if err != nil {
		return -1, err
	}
	return math.GetNearestMultiple(ltp, 50), nil
}
