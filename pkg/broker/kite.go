package broker

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/rohitsakala/strategies/pkg/database"
	"github.com/rohitsakala/strategies/pkg/models"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	kiteconnect "github.com/zerodha/gokiteconnect/v4"
	"go.mongodb.org/mongo-driver/bson"
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
	Database  database.Database
	Filter    bson.M
}

func NewKiteBroker(database database.Database, url, userID, password, apiKey, apiSecret, pin string) (KiteBroker, error) {
	err := database.CreateCollection("credentials")
	if err != nil {
		return KiteBroker{}, err
	}

	return KiteBroker{
		URL:       url,
		UserID:    userID,
		Pin:       pin,
		APIKey:    apiKey,
		APISecret: apiSecret,
		Password:  password,
		Database:  database,
	}, nil
}

func (k *KiteBroker) fetchAccessToken() (models.Credentials, error) {
	var data models.Credentials

	collectionRaw, err := k.Database.GetCollection(bson.D{}, "credentials")
	if err != nil {
		return models.Credentials{}, err
	}
	if len(collectionRaw) <= 0 {
		insertID, err := k.Database.InsertCollection(data, "credentials")
		if err != nil {
			return data, err
		}
		k.Filter = bson.M{
			"_id": insertID,
		}
		return data, nil
	}

	dataBytes, err := bson.Marshal(collectionRaw)
	if err != nil {
		return models.Credentials{}, err
	}
	err = bson.Unmarshal(dataBytes, &data)
	if err != nil {
		return models.Credentials{}, err
	}

	k.Filter = bson.M{
		"_id": collectionRaw["_id"],
	}

	return data, nil
}

func (k *KiteBroker) checkConnection(credentials models.Credentials) error {
	kc := kiteconnect.New(k.APIKey)
	kc.SetAccessToken(credentials.AccessToken)

	_, err := kc.GetUserMargins()
	if err != nil {
		return err
	}

	return nil
}

func (k *KiteBroker) getAccessToken(kc *kiteconnect.Client) (string, error) {
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
		return "", err
	}
	defer webDriver.Quit()

	webDriver.Get(k.URL)

	userIDField, err := webDriver.FindElement(selenium.ByID, "userid")
	if err != nil {
		return "", err
	}
	userIDField.SendKeys(k.UserID)
	passwordElement, err := webDriver.FindElement(selenium.ByID, "password")
	if err != nil {
		return "", err
	}
	passwordElement.SendKeys(k.Password)
	loginButton, err := webDriver.FindElement(selenium.ByCSSSelector, "button[type=submit]")
	if err != nil {
		return "", err
	}
	loginButton.Click()
	time.Sleep(1 * time.Second)

	pinField, err := webDriver.FindElement(selenium.ByID, "pin")
	if err != nil {
		return "", err
	}
	pinField.SendKeys(k.Pin)
	submitButton, err := webDriver.FindElement(selenium.ByCSSSelector, "button[type=submit]")
	if err != nil {
		return "", err
	}
	submitButton.Click()
	time.Sleep(1 * time.Second)

	webDriver.Get(kc.GetLoginURL())
	time.Sleep(1 * time.Second)

	authorizedURLString, err := webDriver.CurrentURL()
	if err != nil {
		return "", err
	}
	authorizedURL, err := url.Parse(authorizedURLString)
	if err != nil {
		return "", err
	}
	requestTokenArray, ok := authorizedURL.Query()["request_token"]
	if !ok || len(requestTokenArray[0]) < 1 {
		return "", errors.New("access token is missing")
	}
	requestToken := requestTokenArray[0]

	data, err := kc.GenerateSession(requestToken, k.APISecret)
	if err != nil {
		return "", err
	}

	return data.AccessToken, nil
}

func (k *KiteBroker) Authenticate() error {
	credentials, err := k.fetchAccessToken()
	if err != nil {
		return err
	}

	kc := kiteconnect.New(k.APIKey)
	if err := k.checkConnection(credentials); err != nil {
		accessToken, err := k.getAccessToken(kc)
		if err != nil {
			return err
		}

		credentials.AccessToken = accessToken
	}

	kc.SetAccessToken(credentials.AccessToken)
	err = k.Database.UpdateCollection(k.Filter, credentials, "credentials")
	if err != nil {
		return err
	}
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
			Expiry:         instrument.Expiry.Time,
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
				Expiry:         instrument.Expiry.Time,
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

func (k *KiteBroker) PlaceOrder(position *models.Position) error {
	var err error

	orderParams := kiteconnect.OrderParams{
		Exchange:        position.Exchange,
		Tradingsymbol:   position.TradingSymbol,
		Product:         position.Product,
		OrderType:       position.OrderType,
		TransactionType: position.TransactionType,
		Quantity:        position.Quantity,
	}

	if position.OrderType == kiteconnect.OrderTypeLimit {
		orderParams.Price = position.Price
	}

	if position.OrderType == kiteconnect.OrderTypeSL {
		orderParams.TriggerPrice = position.TriggerPrice
		orderParams.Price = position.Price
	}

	if len(position.OrderID) <= 0 {
		orderResponse, err := k.Client.PlaceOrder(kiteconnect.VarietyRegular, orderParams)
		if err != nil {
			return err
		}
		position.OrderID = orderResponse.OrderID
		time.Sleep(1 * time.Second)
	}

	if position.OrderType == kiteconnect.OrderTypeLimit {
		time.Sleep(10 * time.Second)

		_, err = k.Client.ModifyOrder(kiteconnect.VarietyRegular, position.OrderID, orderParams)
		if err != nil {
			return err
		}
	}

	orders, err := k.Client.GetOrders()
	if err != nil {
		return err
	}
	for _, order := range orders {
		if order.OrderID == position.OrderID {
			if position.OrderType == kiteconnect.OrderTypeSL {
				if order.Status == "TRIGGER PENDING" {
					position.Status = order.Status
				} else {
					return fmt.Errorf("order failed with status %s and message %s", order.Status, order.StatusMessage)
				}
			}

			if position.OrderType == kiteconnect.OrderTypeMarket {
				if order.Status == kiteconnect.OrderStatusComplete {
					position.AveragePrice = order.AveragePrice
					position.Status = order.Status
				} else {
					return fmt.Errorf("order failed with status %s and message %s", order.Status, order.StatusMessage)
				}
			}

			if position.OrderType == kiteconnect.OrderTypeLimit {
				if order.Status == kiteconnect.OrderStatusComplete {
					position.AveragePrice = order.AveragePrice
					position.Status = order.Status
				} else {
					return fmt.Errorf("order failed with status %s and message %s", order.Status, order.StatusMessage)
				}
			}
		}
	}

	return nil
}

func (k *KiteBroker) CancelOrder(position *models.Position) error {
	orders, err := k.Client.GetOrders()
	if err != nil {
		return err
	}
	for _, order := range orders {
		if order.OrderID == position.OrderID {
			if order.Status == kiteconnect.OrderStatusComplete {
				position.Status = kiteconnect.OrderStatusComplete
			} else if order.Status == kiteconnect.OrderStatusCancelled {
				position.Status = kiteconnect.OrderStatusCancelled
			} else if order.Status == "TRIGGER PENDING" {
				_, err := k.Client.CancelOrder(kiteconnect.VarietyRegular, position.OrderID, nil)
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("order failed with status %s and message %s", order.Status, order.StatusMessage)
			}
		}
	}

	return nil
}

func (k *KiteBroker) GetMargin() {
	allMargins, err := k.Client.GetUserMargins()
	if err != nil {
		return
	}
	fmt.Println(allMargins)
}
