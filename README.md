[![Rohit_Twelve_Thirty](https://github.com/rohitsakala/strategies/actions/workflows/rohit_twelve_thirty.yml/badge.svg?branch=master)](https://github.com/rohitsakala/strategies/actions/workflows/rohit_twelve_thirty.yml)


# Status 

Testing the algorithm till October end. 
Test Results can be seen here - https://github.com/rohitsakala/strategies/actions


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
export KITE_APIKEY={value}
export KITE_APISECRET={value}
export TWELVE_THIRTY_LOT_QUANTITY={value}
export GOOGLE_AUTHENTICATOR_SECRET_KEY=NSXXYXPNX6X62G7SK636NYY37UB2U7LW
export EMAIL_ADDRESS={value}
export EMAIL_PASSWORD={value}
```

```bash
go run main.go twelvethirty
```

### Run on Linux

```bash
wget -q -O - https://dl-ssl.google.com/linux/linux_signing_key.pub | sudo apt-key add - 
sudo sh -c 'echo "deb https://dl.google.com/linux/chrome/deb/ stable main" >> /etc/apt/sources.list.d/google.list'
sudo apt-get update
sudo apt-get install google-chrome-stable unzip -y
sudo snap install go --classic
wget -qO - https://www.mongodb.org/static/pgp/server-5.0.asc | sudo apt-key add -
echo "deb [ arch=amd64,arm64 ] https://repo.mongodb.org/apt/ubuntu focal/mongodb-org/5.0 multiverse" | sudo tee /etc/apt/sources.list.d/mongodb-org-5.0.list
sudo apt-get update
sudo apt-get install -y mongodb-org
sudo systemctl start mongod
sudo systemctl daemon-reload
sudo systemctl status mongod
sudo systemctl enable mongod
chromedriver --url-base=/wd/hub --port=8080 &
```

```bash
export MONGO_URL="mongodb://localhost:27017"
export KITE_URL=https://kite.zerodha.com/
export KITE_USERID={value}
export KITE_PASSWORD={value}
export KITE_APIKEY={value}
export KITE_APISECRET={value}
export TWELVE_THIRTY_LOT_QUANTITY={value}
export GOOGLE_AUTHENTICATOR_SECRET_KEY=NSXXYXPNX6X62G7SK636NYY37UB2U7LW
export EMAIL_ADDRESS={value}
export EMAIL_PASSWORD={value}
```

```bash
go run main.go twelvethirty
```



# TODO's

- Email Alerts instead of using Sensibull.
- Add Fyers broker support.
- Add a GitHub action to cleanup mongo database

# FAQ's

1.How are freak trades avoided ?
- The code only places Limit and Stop Loss Limit orders. Freak trades happen only in Market orders.
