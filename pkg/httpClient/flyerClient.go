package httpClient

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type Client interface {
	GetQuote(symbol string) (SymbolQuote, error)
}

type flyerClient struct {
	client      *http.Client
	url         string
	appId       string
	accessToken string
}

func NewFyerHttpClient(url string) Client {
	return &flyerClient{client: new(http.Client), url: url}
}

func (c *flyerClient) GetQuote(symbol string) (SymbolQuote, error) {
	req, err := http.NewRequest("GET", c.url+"/quotes/symbol="+symbol, nil)
	if err != nil {
		return SymbolQuote{}, err
	}
	req.Header.Add("Authorization", strings.Join([]string{c.appId, c.accessToken}, ":"))
	resp, err := c.client.Do(req)
	if err != nil {
		return SymbolQuote{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return SymbolQuote{}, fmt.Errorf(fmt.Sprintf("unsuccessful response oode - %v", resp))
	}
	rp := make(map[string]json.RawMessage)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return SymbolQuote{}, err
	}
	err = json.Unmarshal(body, &rp)
	if err != nil {
		return SymbolQuote{}, err
	}
	if _, ok := rp["s"]; !ok {
		return SymbolQuote{}, errors.New("unsuccessful status code from fyer client")
	}
	var status string
	err = json.Unmarshal(rp["s"], &status)
	if err != nil {
		return SymbolQuote{}, err
	}
	if status != "ok" {
		return SymbolQuote{}, errors.New("unsuccessful status code from fyer client")
	}
	symbolQuotes := make([]SymbolQuoteResponse, 0)
	return symbolQuotes[0].Quote, nil
}
