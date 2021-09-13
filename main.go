package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/rohitsakala/strategies/pkg/broker"
	"github.com/rohitsakala/strategies/pkg/database"
	"github.com/rohitsakala/strategies/pkg/strategy"
)

func main() {
	log.Printf("Getting arguments...")
	args := os.Args
	if len(args) < 2 {
		log.Printf("Need strategy name as argument")
		os.Exit(1)
	}
	log.Printf("Got %s argument.", args[1])

	log.Printf("Connecting to mongo database....")
	mongoDatabase := database.MongoDatabase{}
	err := mongoDatabase.Connect()
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	log.Printf("Connected to mongo database.")

	log.Printf("Autheticating to kite broker....")
	zerodhaBroker, err := broker.GetBroker("zerodha", &mongoDatabase)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	err = zerodhaBroker.Authenticate()
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	log.Printf("Authenticated to kite broker.")

	log.Printf("Setting to Indian Standard TimeZone...")
	IndianTimeZone, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	log.Printf("Set to Indian Standard TimeZone.")

	/*position := models.Position{
		Exchange:        kiteconnect.ExchangeNSE,
		TradingSymbol:   "TCS",
		Product:         kiteconnect.ProductCNC,
		OrderType:       kiteconnect.OrderTypeSL,
		TransactionType: kiteconnect.TransactionTypeBuy,
		Quantity:        1,
		TriggerPrice:    3850,
		Price:           3851,
		OrderID:         "210913002922239",
		Status:          "OPEN",
	}

	watcher, err := watcher.NewWatcher(zerodhaBroker, *IndianTimeZone)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	watcher.Watch(&position)*/

	log.Printf("Executing %s pm strategy...", args[1])
	strategy, err := strategy.GetStrategy(args[1], zerodhaBroker, *IndianTimeZone, &mongoDatabase, watcher)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	err = strategy.Start()
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	err = strategy.Stop()
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	log.Printf("Executed %s pm strategy.", args[1])
}
