[![Rohit_Twelve_Thirty](https://github.com/rohitsakala/strategies/actions/workflows/rohit_twelve_thirty.yml/badge.svg?branch=master)](https://github.com/rohitsakala/strategies/actions/workflows/rohit_twelve_thirty.yml)

TODO

Margin Check and give time


# Strategy

## NIFTY 12:30 PM Strategy 

Sell ATM CE AND PE Weekly Nifty Options at 12:30 pm and square off at 3:25 pm.

### Run on Mac

```bash
brew install chromedriver
chromedriver --url-base=/wd/hub --port=8080 &
```

```bash
brew install mongodb-community
brew services start mongodb-community
```

```bash
export MONGO_URL="mongodb://localhost:27017"
export KITE_URL=https://kite.zerodha.com/
export KITE_USERID={value}
export KITE_PASSWORD={value}
export KITE_PIN={value}
export KITE_APIKEY={value}
export KITE_APISECRET={value}
export TWELVE_THIRTY_LOT_QUANTITY={value}
```

```bash
go run main.go twelvethirty
```



# FAQ's

1.How are freak trades avoided ?
- The code only places Limit and Stop Loss Limit orders. Freak trades happen only in Market orders.
