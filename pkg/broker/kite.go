package broker

import (
	"errors"
	"fmt"
	"net/url"
	"time"

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

func (k *KiteBroker) PlaceOrder() error {
	return nil
}

func (k *KiteBroker) GetLTP(symbol string) (int, error) {
	// find instrument token of the symbol
	instruments, err := k.Client.GetInstrumentsByExchange("NSE")
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
			return int(ltp[fmt.Sprintf("%d", instrument.InstrumentToken)].LastPrice), nil
		}
	}

	return -1, nil
}

func (k *KiteBroker) GetCurrentMonthyExpiry(symbol string) (bool, error) {

}
