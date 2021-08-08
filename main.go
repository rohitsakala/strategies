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
	callCreditSpreadStrategy := strategy.NewCallCreditSpreadStrategy(&kiteBroker, *IndianTimeZone)

	// Run Strategy
	callCreditSpreadStrategy.Start()

	// Stop Chrome browser
	err = chromeBrowser.Stop()
	if err != nil {
		panic(err)
	}
}
