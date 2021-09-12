package broker

import (
	"os"

	"github.com/rohitsakala/strategies/pkg/database"
)

func GetBroker(name string, database database.Database) (Broker, error) {
	switch name {
	case "zerodha":
		zerodhaBroker, err := NewZerodhaBroker(database,
			os.Getenv("KITE_URL"), os.Getenv("KITE_USERID"), os.Getenv("KITE_PASSWORD"), os.Getenv("KITE_APIKEY"), os.Getenv("KITE_APISECRET"), os.Getenv("KITE_PIN"),
		)
		if err != nil {
			return nil, err
		}
		return &zerodhaBroker, nil
	}

	return nil, nil
}
