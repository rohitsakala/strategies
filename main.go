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
	kiteBroker, err := broker.NewKiteBroker(&mongoDatabase,
		os.Getenv("KITE_URL"), os.Getenv("KITE_USERID"), os.Getenv("KITE_PASSWORD"), os.Getenv("KITE_APIKEY"), os.Getenv("KITE_APISECRET"), os.Getenv("KITE_PIN"),
	)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	err = kiteBroker.Authenticate()
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

	log.Printf("Executing %s pm strategy...", args[1])
	strategy, err := strategy.GetStrategy(args[1], &kiteBroker, *IndianTimeZone, &mongoDatabase)
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
