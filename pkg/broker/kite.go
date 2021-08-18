package broker

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/rohitsakala/strategies/pkg/models"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	kiteconnect "github.com/zerodha/gokiteconnect/v4"
)

type KiteBroker struct {
	URL       string
	Password  string
	UserID    string
	Pin       string
	APIKey    string
	APISecret string
	Client    *kiteconnect.Client
	TimeZone  time.Location
}

func NewKiteBroker(url, userID, password, apiKey, apiSecret, pin string) KiteBroker {
	return KiteBroker{
		URL:       url,
		UserID:    userID,
		Pin:       pin,
		APIKey:    apiKey,
		APISecret: apiSecret,
		Password:  password,
	}
}

func (k *KiteBroker) Authenticate() error {
	// Connect to the chromedriver
	caps := selenium.Capabilities{"browserName": "chrome"}
	chromeCaps := chrome.Capabilities{
		Path: "",
		Args: []string{
			"--headless",
			"--no-sandbox",
		},
	}
	caps.AddChrome(chromeCaps)
	webDriver, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", 8080))
	if err != nil {
		return err
	}
	defer webDriver.Quit()

	// Go to Zerodha login page
	webDriver.Get(k.URL)

	// Enter userID
	userIDField, err := webDriver.FindElement(selenium.ByID, "userid")
	if err != nil {
		return err
	}
	userIDField.SendKeys(k.UserID)

	// Enter password
	passwordElement, err := webDriver.FindElement(selenium.ByID, "password")
	if err != nil {
		return err
	}
	passwordElement.SendKeys(k.Password)

	// Click login button
	loginButton, err := webDriver.FindElement(selenium.ByCSSSelector, "button[type=submit]")
	if err != nil {
		return err
	}
	loginButton.Click()
	time.Sleep(1 * time.Second)

	// Enter PIN
	pinField, err := webDriver.FindElement(selenium.ByID, "pin")
	if err != nil {
		return err
	}
	pinField.SendKeys(k.Pin)

	// Click submit button
	submitButton, err := webDriver.FindElement(selenium.ByCSSSelector, "button[type=submit]")
	if err != nil {
		return err
	}
	submitButton.Click()
	time.Sleep(1 * time.Second)

	// Create a new Kite connect instance
	kc := kiteconnect.New(k.APIKey)

	// Visit LoginURL for access token
	webDriver.Get(kc.GetLoginURL())
	time.Sleep(1 * time.Second)

	// Get request token
	authorizedURLString, err := webDriver.CurrentURL()
	if err != nil {
		return err
	}
	authorizedURL, err := url.Parse(authorizedURLString)
	if err != nil {
		return err
	}
	requestTokenArray, ok := authorizedURL.Query()["request_token"]
	if !ok || len(requestTokenArray[0]) < 1 {
		return errors.New("access token is missing")
	}
	requestToken := requestTokenArray[0]

	// Get user session
	data, err := kc.GenerateSession(requestToken, k.APISecret)
	if err != nil {
		return err
	}

	// Set access token
	kc.SetAccessToken(data.AccessToken)
	k.Client = kc

	return nil
}

func (k *KiteBroker) GetLTP(symbol string) (float64, error) {
	// find instrument token of the symbol
	instruments, err := k.Client.GetInstruments()
	if err != nil {
		return -1, err
	}

	for _, instrument := range instruments {
		if instrument.Tradingsymbol == symbol {
			// find last price of the symbol
			ltp, err := k.Client.GetLTP(fmt.Sprintf("%d", instrument.InstrumentToken))
			if err != nil {
				return -1, err
			}
			return ltp[fmt.Sprintf("%d", instrument.InstrumentToken)].LastPrice, nil
		}
	}

	return -1, nil
}

func (k *KiteBroker) GetInstruments(exchange string) (models.Positions, error) {
	var resultInstruments models.Positions
	var instruments kiteconnect.Instruments
	var err error

	if len(exchange) < 1 {
		instruments, err = k.Client.GetInstruments()
		if err != nil {
			return models.Positions{}, err
		}
	} else {
		instruments, err = k.Client.GetInstrumentsByExchange(exchange)
		if err != nil {
			return models.Positions{}, err
		}
	}
	for _, instrument := range instruments {
		resultInstrument := models.Position{
			TradingSymbol:  instrument.Tradingsymbol,
			Expiry:         instrument.Expiry,
			Segment:        instrument.Segment,
			Exchange:       instrument.Exchange,
			InstrumentType: instrument.InstrumentType,
			StrikePrice:    instrument.StrikePrice,
			LotSize:        int(instrument.LotSize),
		}
		resultInstruments = append(resultInstruments, resultInstrument)
	}

	return resultInstruments, nil
}

func (k *KiteBroker) GetInstrument(symbol string, exchange string) (models.Position, error) {
	var instruments kiteconnect.Instruments
	var err error

	if len(exchange) < 1 {
		instruments, err = k.Client.GetInstruments()
		if err != nil {
			return models.Position{}, err
		}
	} else {
		instruments, err = k.Client.GetInstrumentsByExchange(exchange)
		if err != nil {
			return models.Position{}, err
		}
	}
	for _, instrument := range instruments {
		if symbol == instrument.Tradingsymbol {
			resultInstrument := models.Position{
				TradingSymbol:  instrument.Tradingsymbol,
				Expiry:         instrument.Expiry,
				Segment:        instrument.Segment,
				Exchange:       instrument.Exchange,
				InstrumentType: instrument.InstrumentType,
				StrikePrice:    instrument.StrikePrice,
				LotSize:        int(instrument.LotSize),
			}
			return resultInstrument, nil
		}
	}

	return models.Position{}, nil
}

func (k *KiteBroker) GetPositions() (models.Positions, error) {
	resultPositions := models.Positions{}

	positions, err := k.Client.GetPositions()
	if err != nil {
		return models.Positions{}, err
	}
	for _, position := range positions.Net {
		resultPositon := models.Position{
			TradingSymbol: position.Tradingsymbol,
			Exchange:      position.Exchange,
			Product:       position.Product,
			AveragePrice:  position.AveragePrice,
			Value:         position.Value,
			BuyPrice:      position.BuyPrice,
			SellPrice:     position.SellPrice,
		}
		resultPositions = append(resultPositions, resultPositon)
	}

	return resultPositions, nil
}

func (k *KiteBroker) CheckPosition(symbol string) (bool, error) {
	positions, err := k.Client.GetPositions()
	if err != nil {
		return false, err
	}
	for _, position := range positions.Net {
		if position.Tradingsymbol == symbol {
			return true, nil
		}
	}

	return false, nil
}

func (k *KiteBroker) PlaceOrder(position models.Position) (models.Position, error) {
	orderParams := kiteconnect.OrderParams{
		Exchange:        position.Exchange,
		Tradingsymbol:   position.TradingSymbol,
		Product:         position.Product,
		OrderType:       position.OrderType,
		TransactionType: position.TransactionType,
		Quantity:        position.Quantity,
	}
	fmt.Println(position)
	orderResponse, err := k.Client.PlaceOrder(kiteconnect.VarietyRegular, orderParams)
	if err != nil {
		return models.Position{}, err
	}

	orders, err := k.Client.GetOrders()
	if err != nil {
		return models.Position{}, err
	}
	for _, order := range orders {
		if order.OrderID == orderResponse.OrderID {
			if order.Status == kiteconnect.OrderStatusComplete {
				position.AveragePrice = order.AveragePrice
				position.OrderID = order.OrderID
				position.Status = order.Status
			} else {
				return models.Position{}, fmt.Errorf("order failed with status %s and message %s", order.Status, order.StatusMessage)
			}
		}
	}

	return models.Position{}, nil
}
