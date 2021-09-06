package options

import (
	"errors"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/avast/retry-go"
	"github.com/rohitsakala/strategies/pkg/broker"
	"github.com/rohitsakala/strategies/pkg/models"
	"github.com/rohitsakala/strategies/pkg/utils/maths"
)

const (
	WEEK  = "week"
	MONTH = "month"
)

type PositionSorter []models.Position

func (s PositionSorter) Len() int      { return len(s) }
func (s PositionSorter) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s PositionSorter) Less(i, j int) bool {
	return s[i].Expiry.Before(s[j].Expiry)
}

// GetSymbol will construct the symbol of the
// option according to the parameters given
func GetSymbol(symbol, expiryType string, expiryOffset int, strikePrice float64, optionType string, broker broker.Broker) (string, error) {
	var instruments models.Positions
	var filteredInstruments models.Positions
	var err error

	err = retry.Do(
		func() error {
			instruments, err = broker.GetInstruments("NFO")
			if err != nil {
				return err
			}
			if len(instruments) <= 0 {
				return errors.New("instruments is empty")
			}

			filteredInstruments = models.Positions{}
			for _, instrument := range instruments {
				if strings.HasPrefix(instrument.TradingSymbol, symbol) && instrument.Segment == "NFO-OPT" && instrument.StrikePrice == strikePrice && instrument.Exchange == "NFO" && instrument.InstrumentType == optionType {
					filteredInstruments = append(filteredInstruments, instrument)
				}
			}
			sort.Sort(PositionSorter(filteredInstruments))

			if len(filteredInstruments) <= 0 {
				return errors.New("filtered instruments is empty")
			}

			return nil
		},
		retry.OnRetry(func(n uint, err error) {
			log.Println(fmt.Sprintf("%s because %s", "Retrying getting instruments from NFO", err))
		}),
		retry.Delay(5*time.Second),
		retry.Attempts(5),
	)
	if err != nil {
		return "", err
	}

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

// GetExpiry will return expiry date according to
// the parameters passed in the function
func GetExpiry(symbol, expiryType string, expiryOffset int, strikePrice float64, optionType string, broker broker.Broker) (time.Time, error) {
	var instruments models.Positions
	var err error

	err = retry.Do(
		func() error {
			instruments, err = broker.GetInstruments("NFO")
			if err != nil {
				return err
			}
			if len(instruments) <= 0 {
				return errors.New("instruments is empty")
			}

			return nil
		},
		retry.OnRetry(func(n uint, err error) {
			log.Println(fmt.Sprintf("%s because %s", "Retrying getting instruments from NFO", err))
		}),
		retry.Delay(5*time.Second),
		retry.Attempts(5),
	)
	if err != nil {
		return time.Time{}, err
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
		expiry := filteredInstruments[0].Expiry
		month := filteredInstruments[0].Expiry.Month()
		for i := 1; i < len(filteredInstruments); i++ {
			if filteredInstruments[i].Expiry.Month() != month {
				if expiryOffset == 0 {
					return expiry, nil
				} else {
					expiry = filteredInstruments[i].Expiry
					month = filteredInstruments[i].Expiry.Month()
					expiryOffset--
				}
			}
		}

		return expiry, nil
	case WEEK:
		return filteredInstruments[0].Expiry, nil
	}

	return time.Time{}, nil
}

// GetLotSize will return lotsize of the symbol
func GetLotSize(symbol string, broker broker.Broker) (int, error) {
	var instrument models.Position
	var err error

	err = retry.Do(
		func() error {
			instrument, err = broker.GetInstrument(symbol, "NFO")
			if err != nil {
				return err
			}
			if instrument.TradingSymbol == "" {
				return errors.New("instruments is empty")
			}

			return nil
		},
		retry.OnRetry(func(n uint, err error) {
			log.Println(fmt.Sprintf("%s %s because %s", "Retrying getting lot size for symbol", symbol, err))
		}),
		retry.Delay(5*time.Second),
		retry.Attempts(5),
	)
	if err != nil {
		return -1, err
	}

	return instrument.LotSize, nil
}

// GetATM gives the ATM strike price pf
func GetATM(symbol string, broker broker.Broker) (float64, error) {
	var ltp float64
	var err error

	err = retry.Do(
		func() error {
			ltp, err = broker.GetLTP(symbol)
			if err != nil {
				return err
			}
			return nil
		},
		retry.OnRetry(func(n uint, err error) {
			log.Println(fmt.Sprintf("%s %s because %s", "Retrying getting ATM for symbol", symbol, err))
		}),
		retry.Delay(5*time.Second),
		retry.Attempts(5),
	)
	if err != nil {
		return -1, err
	}

	return maths.GetNearestMultiple(ltp, 50), nil
}

func GetLTPNoFreak(symbol string, broker broker.Broker) (float64, error) {
	var newPrice float64

	err := retry.Do(
		func() error {
			oldPrice, err := broker.GetLTP(symbol)
			if err != nil {
				return err
			}
			for i := 0; i < 5; i++ {
				newPrice, err = broker.GetLTP(symbol)
				if err != nil {
					return err
				}
				diff := math.Abs(float64(newPrice - oldPrice))
				delta := (diff / float64(oldPrice)) * 100
				if delta > 5 {
					return errors.New("freaky price detected")
				}
				oldPrice = newPrice
				time.Sleep(1 * time.Second)
			}

			return nil
		},
		retry.OnRetry(func(n uint, err error) {
			log.Println(fmt.Sprintf("%s %s because %s", "Retrying getting LTP for symbol", symbol, err))
		}),
		retry.Delay(5*time.Second),
		retry.Attempts(5),
	)
	if err != nil {
		return -1, err
	}

	return newPrice + 5, nil
}

// GetLTP gives the LTP of the symbol
func GetLTP(symbol string, broker broker.Broker) (float64, error) {
	var ltp float64
	var err error

	err = retry.Do(
		func() error {
			ltp, err = broker.GetLTP(symbol)
			if err != nil {
				return err
			}
			return nil
		},
		retry.OnRetry(func(n uint, err error) {
			log.Println(fmt.Sprintf("%s %s because %s", "Retrying getting LTP for symbol", symbol, err))
		}),
		retry.Delay(5*time.Second),
		retry.Attempts(5),
	)
	if err != nil {
		return -1, err
	}

	return ltp, nil
}
