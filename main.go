package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/rohitsakala/strategies/pkg/broker"
	"github.com/rohitsakala/strategies/pkg/database"
	"github.com/rohitsakala/strategies/pkg/strategy/twelvethirty"
)

func main() {
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

	log.Printf("Executing Twelve Thirty pm strategy...")
	twelvethirtyStrategy, err := twelvethirty.NewTwelveThirtyStrategy(&kiteBroker, *IndianTimeZone, &mongoDatabase)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	err = twelvethirtyStrategy.Start()
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	log.Printf("Executed Twelve Thirty pm strategy.")
}
