package main

import (
	"fmt"
	"os"
	"time"

	"github.com/rohitsakala/strategies/pkg/broker"
	"github.com/rohitsakala/strategies/pkg/browser"
	"github.com/rohitsakala/strategies/pkg/strategy"
)

func main() {
	// Get Chrome Browser
	chromeBrowser := browser.NewChromeBrowser("/usr/local/bin/chromedriver", 8080)

	// Start Chrome browser
	err := chromeBrowser.Start()
	if err != nil {
		panic(err)
	}

	// Get Kite Broker
	kiteBroker := broker.NewKiteBroker(
		os.Getenv("KITE_URL"), os.Getenv("KITE_USERID"), os.Getenv("KITE_PASSWORD"), os.Getenv("KITE_APIKEY"), os.Getenv("KITE_APISECRET"), os.Getenv("KITE_PIN"),
	)

	// Authenticate Kite broker
	err = kiteBroker.Authenticate()
	if err != nil {
		fmt.Println("cannot authenticate kite")
		panic(err)
	}

	// Get TimeZone
	IndianTimeZone, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		panic(err)
	}

	// Get Strategy
	twelveThirtyStrategy := strategy.NewTwelveThirtyStrategy(kiteBroker, IndianTimeZone)

	// 12:25 pm to 3:24 pm - there should be short straddle ATM
	// Check if time is between 12:25 pm to 3:20 pm

	currentTime := time.Now().In(loc)
	if currentTime.After(start) && currentTime.Before(end) {
		fmt.Println("In Between")
	} else {
		// If ATM Straddle present

	}

	// Stop Chrome browser
	err = chromeBrowser.Stop()
	if err != nil {
		panic(err)
	}
}
