package options

import (
	"sort"

	"github.com/rohitsakala/strategies/pkg/broker"
	"github.com/rohitsakala/strategies/pkg/models"
	"github.com/rohitsakala/strategies/pkg/utils/math"
)

const (
	WEEK  = "week"
	MONTH = "month"
)

type InstrumentSorter []models.Instrument

func (s InstrumentSorter) Len() int      { return len(s) }
func (s InstrumentSorter) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s InstrumentSorter) Less(i, j int) bool {
	return s[i].Expiry.Before(s[j].Expiry.Time)
}

// GetSymbol will construct the symbol of the
// option according to the parameters given
func GetSymbol(symbol, expiryType string, expiryOffset int, strikePrice float64, optionType string, broker broker.Broker) (string, error) {
	instruments, err := broker.GetInstruments("NFO")
	if err != nil {
		return "", err
	}

	filteredInstruments := models.Instruments{}
	for _, instrument := range instruments {
		if instrument.Segment == "NFO-OPT" && instrument.StrikePrice == strikePrice && instrument.Exchange == "NFO" && instrument.InstrumentType == optionType {
			filteredInstruments = append(filteredInstruments, instrument)
		}
	}
	sort.Sort(InstrumentSorter(filteredInstruments))

	switch expiryType {
	case MONTH:
		resultSymbol := filteredInstruments[0].Tradingsymbol
		month := filteredInstruments[0].Expiry.Month()
		for i := 1; i < len(filteredInstruments); i++ {
			if filteredInstruments[i].Expiry.Month() != month {
				return symbol, nil
			}
			resultSymbol = filteredInstruments[i].Tradingsymbol
		}

		return resultSymbol, nil
	case WEEK:
		return filteredInstruments[0].Tradingsymbol, nil
	}

	return "", nil
}

// GetLotSize will return lotsize of the symbol
func GetLotSize(symbol string, broker broker.Broker) (float64, error) {
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
