package main

import (
	"log"
	"os"
	"time"

	"github.com/rohitsakala/strategies/pkg/broker"
	"github.com/rohitsakala/strategies/pkg/browser"
	"github.com/rohitsakala/strategies/pkg/database"
	"github.com/rohitsakala/strategies/pkg/strategy/twelvethirty"
)

func main() {
	// Connect Mongo Database
	mongoDatabase := database.MongoDatabase{}
	err := mongoDatabase.Connect()
	if err != nil {
		panic(err)
	}

	// Get Kite Broker
	kiteBroker, err := broker.NewKiteBroker(&mongoDatabase,
		os.Getenv("KITE_URL"), os.Getenv("KITE_USERID"), os.Getenv("KITE_PASSWORD"), os.Getenv("KITE_APIKEY"), os.Getenv("KITE_APISECRET"), os.Getenv("KITE_PIN"),
	)
	if err != nil {
		panic(err)
	}

	// Authenticate Kite broker
	err = kiteBroker.Authenticate()
	if err != nil {
		panic(err)
	}

	// Get Chrome Browser
	chromeBrowser := browser.NewChromeBrowser("/usr/local/bin/chromedriver", 8080)

	// Start Chrome browser
	err = chromeBrowser.Start()
	if err != nil {
		panic(err)
	}

	// Get TimeZone
	IndianTimeZone, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		panic(err)
	}

	// Get Strategy
	twelvethirtyStrategy, err := twelvethirty.NewTwelveThirtyStrategy(&kiteBroker, *IndianTimeZone, &mongoDatabase)
	if err != nil {
		panic(err)
	}

	// Run Strategy
	log.Printf("Starting Twelve Thiry PM trade....")
	err = twelvethirtyStrategy.Start()
	if err != nil {
		panic(err)
	}
}
