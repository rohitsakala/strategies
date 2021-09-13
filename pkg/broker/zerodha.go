package broker

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/avast/retry-go"
	"github.com/rohitsakala/strategies/pkg/database"
	"github.com/rohitsakala/strategies/pkg/models"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	kiteconnect "github.com/zerodha/gokiteconnect/v4"
	"go.mongodb.org/mongo-driver/bson"
)

type ZerodhaBroker struct {
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

func NewZerodhaBroker(database database.Database, url, userID, password, apiKey, apiSecret, pin string) (ZerodhaBroker, error) {
	err := database.CreateCollection("credentials")
	if err != nil {
		return ZerodhaBroker{}, err
	}

	return ZerodhaBroker{
		URL:       url,
		UserID:    userID,
		Pin:       pin,
		APIKey:    apiKey,
		APISecret: apiSecret,
		Password:  password,
		Database:  database,
	}, nil
}

func (z *ZerodhaBroker) fetchAccessToken() (models.Credentials, error) {
	var data models.Credentials

	collectionRaw, err := z.Database.GetCollection(bson.D{}, "credentials")
	if err != nil {
		return models.Credentials{}, err
	}
	if len(collectionRaw) <= 0 {
		insertID, err := z.Database.InsertCollection(data, "credentials")
		if err != nil {
			return data, err
		}
		z.Filter = bson.M{
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

	z.Filter = bson.M{
		"_id": collectionRaw["_id"],
	}

	return data, nil
}

func (z *ZerodhaBroker) checkConnection(credentials models.Credentials) error {
	kc := kiteconnect.New(z.APIKey)
	kc.SetAccessToken(credentials.AccessToken)

	_, err := kc.GetUserMargins()
	if err != nil {
		return err
	}

	return nil
}

func (z *ZerodhaBroker) getAccessToken(kc *kiteconnect.Client) (string, error) {
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

	webDriver.Get(z.URL)

	userIDField, err := webDriver.FindElement(selenium.ByID, "userid")
	if err != nil {
		return "", err
	}
	userIDField.SendKeys(z.UserID)
	passwordElement, err := webDriver.FindElement(selenium.ByID, "password")
	if err != nil {
		return "", err
	}
	passwordElement.SendKeys(z.Password)
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
	pinField.SendKeys(z.Pin)
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

	data, err := kc.GenerateSession(requestToken, z.APISecret)
	if err != nil {
		return "", err
	}

	return data.AccessToken, nil
}

func (z *ZerodhaBroker) Authenticate() error {
	credentials, err := z.fetchAccessToken()
	if err != nil {
		return err
	}

	kc := kiteconnect.New(z.APIKey)
	if err := z.checkConnection(credentials); err != nil {
		accessToken, err := z.getAccessToken(kc)
		if err != nil {
			return err
		}

		credentials.AccessToken = accessToken
	}

	kc.SetAccessToken(credentials.AccessToken)
	err = z.Database.UpdateCollection(z.Filter, credentials, "credentials")
	if err != nil {
		return err
	}
	z.Client = kc

	return nil
}

func (z *ZerodhaBroker) GetLTP(symbol string) (float64, error) {
	// find instrument token of the symbol
	instruments, err := z.Client.GetInstruments()
	if err != nil {
		return -1, err
	}

	for _, instrument := range instruments {
		if instrument.Tradingsymbol == symbol {
			// find last price of the symbol
			ltp, err := z.Client.GetLTP(fmt.Sprintf("%d", instrument.InstrumentToken))
			if err != nil {
				return -1, err
			}
			return ltp[fmt.Sprintf("%d", instrument.InstrumentToken)].LastPrice, nil
		}
	}

	return -1, nil
}

func (z *ZerodhaBroker) GetInstruments(exchange string) (models.Positions, error) {
	var resultInstruments models.Positions
	var instruments kiteconnect.Instruments
	var err error

	if len(exchange) < 1 {
		instruments, err = z.Client.GetInstruments()
		if err != nil {
			return models.Positions{}, err
		}
	} else {
		instruments, err = z.Client.GetInstrumentsByExchange(exchange)
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

func (z *ZerodhaBroker) GetInstrument(symbol string, exchange string) (models.Position, error) {
	var instruments kiteconnect.Instruments
	var err error

	if len(exchange) < 1 {
		instruments, err = z.Client.GetInstruments()
		if err != nil {
			return models.Position{}, err
		}
	} else {
		instruments, err = z.Client.GetInstrumentsByExchange(exchange)
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

func (z *ZerodhaBroker) GetPositions() (models.Positions, error) {
	resultPositions := models.Positions{}

	positions, err := z.Client.GetPositions()
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

func (z *ZerodhaBroker) CheckPosition(symbol string) (bool, error) {
	positions, err := z.Client.GetPositions()
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

func (z *ZerodhaBroker) PlaceOrder(position *models.Position) error {
	var err error

	err = retry.Do(
		func() error {
			err = z.placeOrder(position)
			if err != nil {
				return err
			}
			return nil
		},
		retry.OnRetry(func(_ uint, err error) {
			log.Println(fmt.Sprintf("%s %s because %s", "Retrying placing position", position.TradingSymbol, err))
		}),
		retry.Delay(5*time.Second),
		retry.Attempts(5),
	)
	if err != nil {
		return err
	}

	return nil
}

func (z *ZerodhaBroker) placeOrder(position *models.Position) error {
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
		orderResponse, err := z.Client.PlaceOrder(kiteconnect.VarietyRegular, orderParams)
		if err != nil {
			return err
		}
		position.OrderID = orderResponse.OrderID
		if position.OrderType == kiteconnect.OrderTypeLimit {
			time.Sleep(10 * time.Second)
		} else {
			time.Sleep(1 * time.Second)
		}
	} else {
		if position.OrderType == kiteconnect.OrderTypeLimit {
			_, err = z.Client.ModifyOrder(kiteconnect.VarietyRegular, position.OrderID, orderParams)
			if err != nil {
				return err
			}
			time.Sleep(10 * time.Second)
		}
	}

	orders, err := z.Client.GetOrders()
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

func (z *ZerodhaBroker) GetOrders() (models.Positions, error) {
	var positions models.Positions
	orders, err := z.Client.GetOrders()
	if err != nil {
		return models.Positions{}, err
	}

	for _, order := range orders {
		position := models.Position{
			OrderID: order.OrderID,
			Status:  order.Status,
		}
		positions = append(positions, position)
	}

	return positions, nil
}

func (z *ZerodhaBroker) CancelOrder(position *models.Position) error {
	orders, err := z.Client.GetOrders()
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
				_, err := z.Client.CancelOrder(kiteconnect.VarietyRegular, position.OrderID, nil)
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

func (z *ZerodhaBroker) CancelOrders(positions models.RefPositions) error {
	for _, position := range positions {
		err := retry.Do(
			func() error {
				err := z.CancelOrder(position)
				if err != nil {
					return err
				}
				return nil
			},
			retry.OnRetry(func(_ uint, err error) {
				log.Println(fmt.Sprintf("%s %s because %s", "Retrying cancelling order ", position.TradingSymbol, err))
			}),
			retry.Delay(5*time.Second),
			retry.Attempts(5),
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (z *ZerodhaBroker) GetMargin() {
	allMargins, err := z.Client.GetUserMargins()
	if err != nil {
		return
	}
	fmt.Println(allMargins)
}