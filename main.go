package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/rohitsakala/strategies/pkg/authenticator"
	"github.com/rohitsakala/strategies/pkg/broker"
	"github.com/rohitsakala/strategies/pkg/database"
	"github.com/rohitsakala/strategies/pkg/strategy"
	"github.com/rohitsakala/strategies/pkg/utils"
	"github.com/rohitsakala/strategies/pkg/watcher"
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
		utils.SendEmail("Twelve Thirty run paniced. Immediate Attention needed", err.Error())
		fmt.Println(err)
		panic(err)
	}
	log.Printf("Connected to mongo database.")

	log.Printf("Autheticating to kite broker....")
	googleAuthenticator := authenticator.GetAuthenticator("google")
	zerodhaBroker, err := broker.GetBroker("zerodha", &mongoDatabase, googleAuthenticator)
	if err != nil {
		utils.SendEmail("Twelve Thirty run paniced. Immediate Attention needed", err.Error())
		fmt.Println(err)
		panic(err)
	}
	err = zerodhaBroker.Authenticate()
	if err != nil {
		utils.SendEmail("Twelve Thirty run paniced. Immediate Attention needed", err.Error())
		fmt.Println(err)
		panic(err)
	}
	log.Printf("Authenticated to kite broker.")

	log.Printf("Setting to Indian Standard TimeZone...")
	IndianTimeZone, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		utils.SendEmail("Twelve Thirty run paniced. Immediate Attention needed", err.Error())
		fmt.Println(err)
		panic(err)
	}
	log.Printf("Set to Indian Standard TimeZone.")

	watcher, err := watcher.NewWatcher(zerodhaBroker, *IndianTimeZone)
	if err != nil {
		utils.SendEmail("Twelve Thirty run paniced. Immediate Attention needed", err.Error())
		fmt.Println(err)
		panic(err)
	}

	log.Printf("Executing %s pm strategy with args...%s", args[1], args[2])
	strategy, err := strategy.GetStrategy(args[1], zerodhaBroker, *IndianTimeZone, &mongoDatabase, watcher, args[2])
	if err != nil {
		utils.SendEmail("Twelve Thirty run paniced. Immediate Attention needed", err.Error())
		fmt.Println(err)
		panic(err)
	}
	err = strategy.Start()
	if err != nil {
		utils.SendEmail("Twelve Thirty run paniced. Immediate Attention needed", err.Error())
		fmt.Println(err)
		panic(err)
	}
	err = strategy.Stop()
	if err != nil {
		utils.SendEmail("Twelve Thirty run paniced. Immediate Attention needed", err.Error())
		fmt.Println(err)
		panic(err)
	}
	log.Printf("Executed %s pm strategy.", args[1])
}
