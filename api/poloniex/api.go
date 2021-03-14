package poloniex

import (
	"../../utils/log"
	"crypto/hmac"
	"crypto/sha512"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var (
	// Technically 6 req/s allowed, but we're being nice / playing it safe.
	reqInterval = 200 * time.Millisecond
)

func NewPoloniexApi(key string, secret string) *Config {
	cookieJar, _ := cookiejar.New(nil)

	c := &Config{
		Key:               key,
		Secret:            secret,
		RestEndpoint:      "https://poloniex.com",
		WebsocketEndpoint: "wss://api2.poloniex.com",
		Keepalive:         true,
		Timeout:           time.Second * 60,
		Header:            make(map[string]string),
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Jar: cookieJar,
		},
		Currencies:  make(map[string]*Currency),
		Pairs:       make(map[string]*Pair),
		throttle:    time.Tick(reqInterval),
		httpTimeout: 30 * time.Second,
		debug:       false,
	}
	s := NewSocket(c.WebsocketEndpoint)
	s.Key = c.Key
	s.Secret = c.Secret
	c.Socket = s

	return c
}

func (c *Config) Setup() error {
	curr, err := c.GetCurrencies()
	if err != nil {
		return err
	}
	pairs, err := c.GetTicker()
	if err != nil {
		return err
	}
	c.Currencies = curr
	c.Pairs = pairs
	return nil
}

func (c *Config) GetCurrencies() (map[string]*Currency, error) {
	b, err := c.do("GET", "public?command=returnCurrencies", nil, false)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	r := make(map[string]*Currency)
	if err := json.Unmarshal(b, &r); err != nil {
		log.Error(err)
		return nil, err
	}
	return r, nil
}

func (c *Config) GetTicker() (map[string]*Pair, error) {
	b, err := c.do("GET", "public?command=returnTicker", nil, false)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	r := make(map[string]*Pair)
	if err := json.Unmarshal(b, &r); err != nil {
		log.Error(err)
		return nil, err
	}
	return r, nil
}

func (c *Config) GetOrderBook(market, cat string, depth int) (orderBook *OrderBook, err error) {
	// not implemented
	if cat != "bid" && cat != "ask" && cat != "both" {
		cat = "both"
	}
	if depth > 100 {
		depth = 100
	}
	if depth < 1 {
		depth = 1
	}

	r, err := c.do("GET", fmt.Sprintf("public?command=returnOrderBook&currencyPair=%s&depth=%d", strings.ToUpper(market), depth), nil, false)
	if err != nil {
		return
	}
	if err = json.Unmarshal(r, &orderBook); err != nil {
		return
	}
	if orderBook.Error != "" {
		err = errors.New(orderBook.Error)
		return
	}
	return
}

func (c *Config) GetOpenOrders(symbol string) ([]*OpenOrder, error) {
	b, err := c.doCommand("returnOpenOrders", map[string]string{"currencyPair": symbol})
	if err != nil {
		log.Error(err)
		return nil, err
	}
	r := make([]*OpenOrder, 0)
	if err := json.Unmarshal(b, &r); err != nil {
		log.Error(err)
		return nil, err
	}
	return r, nil
}

func (c *Config) GetTradeHistory(symbol string) ([]*Trade, error) {
	b, err := c.doCommand("returnTradeHistory", map[string]string{"currencyPair": symbol, "limit": "2500"})
	if err != nil {
		log.Error(err)
		return nil, err
	}
	r := make([]*Trade, 0)
	if err := json.Unmarshal(b, &r); err != nil {
		log.Error(err)
		return nil, err
	}
	return r, nil
}

func (c *Config) GetPair(symbol string) *Pair {
	if pair, ok := c.Pairs[symbol]; ok {
		return pair
	}
	return nil
}

func (c *Config) SubscribeOrderBook(symbol string, updatesCh chan<- MarketUpd, stopCh <-chan bool) error {
	if pair, ok := c.Pairs[symbol]; ok {
		return c.Socket.subscribeMarketUpdates(pair.Id, updatesCh, stopCh)
	}
	return errors.New("symbol is not supported")
}

func (c *Config) SubscribeAccount(updatesCh chan<- AccountUpd, stopCh <-chan bool) error {
	return c.Socket.subscribeAccountUpdates(updatesCh, stopCh)
}

func (c *Config) Buy(symbol string, rate float64, amount float64) (TradeOrder, error) {
	return c.trade("buy", symbol, rate, amount)
}

func (c *Config) Sell(symbol string, rate float64, amount float64) (TradeOrder, error) {
	return c.trade("sell", symbol, rate, amount)
}

func (c *Config) trade(direction string, symbol string, rate float64, amount float64) (TradeOrder, error) {
	if _, ok := c.Pairs[symbol]; !ok {
		return TradeOrder{}, errors.New("pair not found")
	}

	params := map[string]string{
		"currencyPair": symbol,
		"rate":         strconv.FormatFloat(rate, 'f', -1, 64),
		"amount":       strconv.FormatFloat(amount, 'f', -1, 64),
	}
	b, err := c.doCommand(direction, params)

	if err != nil {
		return TradeOrder{}, err
	}
	var orderResponse TradeOrder
	if err = json.Unmarshal(b, &orderResponse); err != nil {
		return TradeOrder{}, err
	}

	if orderResponse.ErrorMessage != "" {
		return TradeOrder{}, errors.New(orderResponse.ErrorMessage)
	}

	return orderResponse, nil
}

func (c *Config) doCommand(command string, payload map[string]string) (response []byte, err error) {
	if payload == nil {
		payload = make(map[string]string)
	}

	payload["command"] = command
	payload["nonce"] = strconv.FormatInt(time.Now().UnixNano(), 10)

	return c.do("POST", "tradingApi", payload, true)
}

func (c *Config) do(method, resource string, payload map[string]string, authNeeded bool) (response []byte, err error) {
	respCh := make(chan []byte)
	errCh := make(chan error)
	<-c.throttle
	go c.makeReq(method, resource, payload, authNeeded, respCh, errCh)
	response = <-respCh
	err = <-errCh
	return
}

func (c *Config) makeReq(method, resource string, payload map[string]string, authNeeded bool, respCh chan<- []byte, errCh chan<- error) {
	body := []byte{}
	connectTimer := time.NewTimer(c.httpTimeout)

	var rawurl string
	if strings.HasPrefix(resource, "http") {
		rawurl = resource
	} else {
		rawurl = fmt.Sprintf("%s/%s", c.RestEndpoint, resource)
	}

	formValues := url.Values{}
	for key, value := range payload {
		formValues.Set(key, value)
	}
	formData := formValues.Encode()

	req, err := http.NewRequest(method, rawurl, strings.NewReader(formData))
	if err != nil {
		respCh <- body
		errCh <- errors.New("You need to set API Key and API Secret to call this method")
		return
	}

	if authNeeded {
		if len(c.Key) == 0 || len(c.Secret) == 0 {
			respCh <- body
			errCh <- errors.New("You need to set API Key and API Secret to call this method")
			return
		}

		mac := hmac.New(sha512.New, []byte(c.Secret))
		_, err := mac.Write([]byte(formData))
		if err != nil {
			errCh <- err
		}
		sig := hex.EncodeToString(mac.Sum(nil))
		req.Header.Add("Key", c.Key)
		req.Header.Add("Sign", sig)
	}

	if method == "POST" || method == "PUT" {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	req.Header.Add("Accept", "application/json")

	resp, err := c.doTimeoutRequest(connectTimer, req)
	if err != nil {
		respCh <- body
		errCh <- err
		return
	}

	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		respCh <- body
		errCh <- err
		return
	}
	if resp.StatusCode != 200 {
		respCh <- body
		errCh <- errors.New(resp.Status)
		return
	}

	respCh <- body
	errCh <- nil
	close(respCh)
	close(errCh)
}

func (c *Config) doTimeoutRequest(timer *time.Timer, req *http.Request) (*http.Response, error) {
	// Do the request in the background so we can check the timeout
	type result struct {
		resp *http.Response
		err  error
	}
	done := make(chan result, 1)
	go func() {
		if c.debug {
			c.dumpRequest(req)
		}
		resp, err := c.client.Do(req)
		if c.debug {
			c.dumpResponse(resp)
		}
		done <- result{resp, err}
	}()
	// Wait for the read or the timeout
	select {
	case r := <-done:
		return r.resp, r.err
	case <-timer.C:
		return nil, errors.New("timeout on reading data from Poloniex API")
	}
}

func (c *Config) dumpRequest(r *http.Request) {
	if r == nil {
		log.Warn("dumpReq ok: <nil>")
		return
	}
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Error("dumpReq err:", err)
	} else {
		log.Success("dumpReq ok:", string(dump))
	}
}

func (c *Config) dumpResponse(r *http.Response) {
	if r == nil {
		log.Warn("dumpResponse ok: <nil>")
		return
	}
	dump, err := httputil.DumpResponse(r, true)
	if err != nil {
		log.Error("dumpResponse err:", err)
	} else {
		log.Success("dumpResponse ok:", string(dump))
	}
}
