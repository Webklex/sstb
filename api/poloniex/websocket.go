package poloniex

import (
	"../../utils/log"
	"../../utils/values"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
	"math/big"
	"net/url"
	"reflect"
	"strconv"
	"sync"
	"time"
)

var (
	floatDecType        = reflect.TypeOf(values.Float{})
	stringDecType       = reflect.TypeOf(values.String{})
	orderBookDecodeHook = mapstructure.ComposeDecodeHookFunc(decimalDecodeHook)
)

const (
	Sell = OpType(0)
	Buy  = OpType(1)
)

func decimalDecodeHook(from reflect.Type, to reflect.Type, v interface{}) (interface{}, error) {
	if to == stringDecType {
		return values.String{Value: fmt.Sprintf("%v", v)}, nil
	}
	if to == floatDecType {
		if str, ok := v.(string); ok {
			val, _ := new(big.Float).SetString(str)
			return values.Float{*val}, nil
		}
		return nil, errors.New(fmt.Sprintf("cannot decode %s to decimal", from.String()))
	}
	return v, nil
}

func NewSocket(endpoint string) *Socket {
	result := &Socket{
		Endpoint: endpoint,
		write:    make(chan wsCmd, 16),
		ws:       make(chan bool, 1),
		subs:     make(map[int]wsSub),
	}
	result.cv = sync.NewCond(&result.mx)
	go result.writeLoop()
	go result.loop()
	return result
}

func (s *Socket) writeLoop() {
	for cmd := range s.write {
		conn, err := s.checkClient()
		if err != nil {
			asyncErr(cmd.result, err)
			continue
		}
		asyncErr(cmd.result, conn.WriteJSON(cmd.message))
	}
}

func (s *Socket) loop() {
	for range s.ws {
		s.mx.Lock()
		cl := s.conn
		s.mx.Unlock()
		if cl != nil {
			continue
		}
		cl, err := s.makeWsClient()
		s.mx.Lock()
		s.conn = cl
		s.err = err
		s.mx.Unlock()
		// notify all clients about connect attempt, successful or not.
		s.cv.Broadcast()
	}
}

func (s *Socket) makeWsClient() (*websocket.Conn, error) {
	cl, _, err := websocket.DefaultDialer.Dial(s.Endpoint, nil)
	return cl, err
}

func (s *Socket) checkClient() (*websocket.Conn, error) {
	s.mx.Lock()
	defer s.mx.Unlock()
	// do not wait for connection forever, return an error on bad attempt.
	if s.conn == nil {
		if s.ws == nil {
			return nil, errors.New("client closed")
		}
		select {
		case s.ws <- true:
		default:
		}
		s.cv.Wait()
		if s.conn == nil {
			err := s.err
			if err == nil {
				err = errors.New("connection error")
			}
			return nil, err
		} else {
			go s.readLoop(s.conn)
		}
	}
	return s.conn, nil
}

func (s *Socket) readLoop(conn *websocket.Conn) {
	hbChan := make(chan struct{})
	defer close(hbChan)
	go s.timeout(hbChan)
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			_ = s.closeWithErr(err)
			return
		}
		hbChan <- struct{}{}
		if err := s.handleMessage(message); err != nil {
			log.Error(err)
		}
	}
}

func (s *Socket) timeout(hbChan chan struct{}) {
	const hbTimeout = 5 * time.Second
	for {
		select {
		case _, ok := <-hbChan:
			if !ok {
				return
			}
		case <-time.After(hbTimeout):
			_ = s.closeWithErr(errors.New("ws timeout"))
			return
		}
	}
}

func (s *Socket) closeWithErr(err error) error {
	s.mx.Lock()
	conn := s.conn
	s.conn = nil
	for _, sub := range s.subs {
		asyncErr(sub.err, err)
	}
	s.mx.Unlock()
	if conn != nil {
		_ = withTimeout(func() error {
			return conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		}, func(e error) {
			log.Warn("timeout writing ws close message")
		}, 2*time.Second)
		return conn.Close()
	}
	return nil
}

func (s *Socket) handleMessage(message []byte) error {
	var msg wsMessage
	arr := []interface{}{&msg.channel, &msg.seq, &msg.data}
	if err := json.Unmarshal(message, &arr); err != nil {
		return err
	}
	return s.parseMessage(msg)
}

func (s *Socket) parseMessage(msg wsMessage) error {
	id, err := msg.channel.Int64()
	if err != nil {
		return err
	}
	seq, err := msg.seq.Int64()

	s.mx.Lock()
	handler := s.subs[int(id)].handler
	s.mx.Unlock()
	if handler == nil {
		return nil
	}

	var mu MarketUpd
	var au AccountUpd
	for _, el := range msg.data {
		if len(el) < 2 {
			continue
		}
		typ, ok := el[0].(string)
		if !ok {
			continue
		}
		switch typ {
		case "p", "b", "n", "m": // order book updates, or new trades
			if err := s.handleAccountEvents(typ, el, &au); err != nil {
				fmt.Println(el)
				log.Error("message parsing error: %v", err)
			}
		case "i", "o", "t": // order book updates, or new trades
			if id == 1000 {
				if err := s.handleAccountEvents(typ, el, &au); err != nil {
					fmt.Println(el)
					log.Error("message parsing error: %v", err)
				}
			} else if seq > 0 {
				if err := s.handleOrderBookOrTrades(typ, el, &mu); err != nil {
					fmt.Println(el)
					log.Error("message parsing error: %v", err)
				}
			}
		}
	}

	if id == 1000 {
		au.Seq = seq
		handler(au)
	} else if len(mu.OrderBooks)+len(mu.Trades) > 0 && seq > 0 {
		mu.Seq = seq
		handler(mu)
	}
	return nil
}

func (s *Socket) handleAccountEvents(typ string, i []interface{}, mu *AccountUpd) error {
	switch typ {
	case "p":
		// -> Gets called if a new order is being created
		update, err := s.parsePendingOpenOrdersUpdate(i)
		if err == nil {
			mu.PendingOrders = append(mu.PendingOrders, update)
		}
		return err
	case "n":
		// -> Gets called if a new order is being created
		update, err := s.parseNewOpenOrdersUpdate(i)
		if err == nil {
			mu.NewOrders = append(mu.NewOrders, update)
		}
		return err
	case "o": // first
		// o updates represent an order update.
		update, err := s.parseTradeOrdersUpdate(i)
		if err == nil {
			mu.TradeOrders = append(mu.TradeOrders, update)
		}
		return err
	case "b": // second
		// b updates represent an available balance update.
	case "t": // third
	case "f": // fourth
	case "m":
		// m updates represent an margin position update.

	}

	return nil
}

func (s *Socket) parseNewOpenOrdersUpdate(arr []interface{}) (OpenOrder, error) {
	var order OpenOrder
	var typ string
	toDecode := []interface{}{&typ, &order.Symbol, &order.OrderNumber, &order.TypeNum, &order.Rate, &order.Amount}
	obookDec, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:     &toDecode,
		DecodeHook: orderBookDecodeHook,
	})
	err := obookDec.Decode(arr)
	if err == nil {
		if order.TypeNum.Value == "0" && order.Type == "" {
			order.Type = "sell"
		} else if order.Type == "" {
			order.Type = "buy"
		}
		if order.Type == "" {
			return order, errors.New("invalid order type")
		}
	}
	return order, err
}

func (s *Socket) parsePendingOpenOrdersUpdate(arr []interface{}) (OpenOrder, error) {
	var order OpenOrder
	var typ string
	toDecode := []interface{}{&typ, &order.OrderNumber, &order.Symbol, &order.Rate, &order.Amount, &order.TypeNum}

	obookDec, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:     &toDecode,
		DecodeHook: orderBookDecodeHook,
	})
	err := obookDec.Decode(arr)
	if err == nil {
		if order.TypeNum.Value == "0" && order.Type == "" {
			order.Type = "sell"
		} else if order.Type == "" {
			order.Type = "buy"
		}
		if order.Type == "" {
			return order, errors.New("invalid order type")
		}
	}
	return order, err
}

func (s *Socket) parseTradeOrdersUpdate(arr []interface{}) (TradeOrder, error) {
	var tradeOrder TradeOrder
	var typ string
	toDecode := []interface{}{&typ, &tradeOrder.Number, &tradeOrder.Amount, &tradeOrder.Type}
	obookDec, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:     &toDecode,
		DecodeHook: orderBookDecodeHook,
	})
	return tradeOrder, obookDec.Decode(arr)
}

func (s *Socket) handleOrderBookOrTrades(typ string, i []interface{}, mu *MarketUpd) error {
	switch typ {
	case "i":
		updates, err := s.parseOrderBookInitial(i[1])
		if err == nil {
			mu.Initial = true
			mu.OrderBooks = updates
		}
		return err
	case "o":
		update, err := s.parseOrderBookUpdate(i)
		if err == nil {
			mu.OrderBooks = append(mu.OrderBooks, update)
		}
		return err
	case "t":
		update, err := s.parseTradeUpdate(i)
		if err == nil {
			mu.Trades = append(mu.Trades, update)
		}
		return err
	}
	return nil
}

func (s *Socket) parseOrderBookInitial(i interface{}) ([]OrderBookUpd, error) {
	oBookUpd := struct {
		CurrencyPair string
		OrderBook    []map[*values.Float]values.Float
	}{}
	orderBookDec, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:     &oBookUpd,
		DecodeHook: orderBookDecodeHook,
	})
	if err := orderBookDec.Decode(i); err != nil {
		return nil, err
	}
	if len(oBookUpd.OrderBook) != 2 {
		return nil, errors.New("invalid order book structure")
	}
	var updates []OrderBookUpd
	fill := func(m map[*values.Float]values.Float, typ OpType) {
		for price, size := range m {
			updates = append(updates, OrderBookUpd{
				Type:  typ,
				Price: *price,
				Size:  size,
			})
		}
	}
	fill(oBookUpd.OrderBook[0], Sell)
	fill(oBookUpd.OrderBook[1], Buy)
	return updates, nil
}

func (s *Socket) parseOrderBookUpdate(arr []interface{}) (OrderBookUpd, error) {
	var oBookUpd OrderBookUpd
	var typ string
	toDecode := []interface{}{&typ, &oBookUpd.Type, &oBookUpd.Price, &oBookUpd.Size}
	orderBookDec, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:     &toDecode,
		DecodeHook: orderBookDecodeHook,
	})
	return oBookUpd, orderBookDec.Decode(arr)
}

func (s *Socket) parseTradeUpdate(arr []interface{}) (TradeUpd, error) {
	var tradeUpd TradeUpd
	var typ string
	toDecode := []interface{}{&typ, &tradeUpd.TradeID, &tradeUpd.Type, &tradeUpd.Price, &tradeUpd.Size, &tradeUpd.Date}
	obookDec, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:     &toDecode,
		DecodeHook: orderBookDecodeHook,
	})
	return tradeUpd, obookDec.Decode(arr)
}

func (s *Socket) subscribeMarketUpdates(id int, updatesCh chan<- MarketUpd, stopCh <-chan bool) error {
	s.mx.Lock()
	if _, found := s.subs[id]; found {
		s.mx.Unlock()
		return errors.New("already subscribed")
	}
	errChan := make(chan error, 1)
	s.subs[id] = wsSub{
		handler: s.makeMarketUpdateHandler(updatesCh),
		err:     errChan,
	}
	s.mx.Unlock()
	defer s.unsubscribeChan(id)
	if err := s.cmd("subscribe", id); err != nil {
		return err
	}
	select {
	case <-stopCh:
		return nil
	case err := <-errChan:
		return err
	}
}

func (s *Socket) subscribeAccountUpdates(updatesCh chan<- AccountUpd, stopCh <-chan bool) error {
	s.mx.Lock()
	id := 1000
	if _, found := s.subs[id]; found {
		s.mx.Unlock()
		return errors.New("already subscribed")
	}
	errChan := make(chan error, 1)
	s.subs[id] = wsSub{
		handler: s.makeAccountUpdateHandler(updatesCh),
		err:     errChan,
	}
	s.mx.Unlock()
	defer s.unsubscribeChan(id)
	if err := s.authCmd(errChan); err != nil {
		return err
	}
	select {
	case <-stopCh:
		return nil
	case err := <-errChan:
		return err
	}
}

func (s *Socket) makeMarketUpdateHandler(updatesCh chan<- MarketUpd) func(interface{}) {
	return func(i interface{}) {
		mu, ok := i.(MarketUpd)
		if !ok {
			return
		}
		select {
		case updatesCh <- mu:
		default:
		}
	}
}

func (s *Socket) makeAccountUpdateHandler(updatesCh chan<- AccountUpd) func(interface{}) {
	return func(i interface{}) {
		mu, ok := i.(AccountUpd)
		if !ok {
			return
		}
		select {
		case updatesCh <- mu:
		default:
		}
	}
}

func (s *Socket) unsubscribeChan(symbol int) {
	err := withTimeout(func() error {
		return s.cmd("unsubscribe", symbol)
	}, func(err error) {
		s.mx.Lock()
		sub, found := s.subs[symbol]
		if found {
			asyncErr(sub.err, errors.New("subscription error"))
		}
		s.mx.Unlock()
		log.Warn("timeout writing ws unsubscribe message. the last error was %v", err)
	}, 2*time.Second)
	if err != nil {
		log.Error("unsubscribe error: %v", err)
	}
	s.mx.Lock()
	delete(s.subs, symbol)
	s.mx.Unlock()
}

func (s *Socket) cmd(command string, channel int) error {
	cmd := wsCmd{
		message: map[string]interface{}{
			"command": command,
			"channel": channel,
		},
		result: make(chan error, 1),
	}
	s.write <- cmd
	return <-cmd.result
}

func (s *Socket) authCmd(errCh chan<- error) error {
	formValues := url.Values{}
	formValues.Set("nonce", strconv.FormatInt(time.Now().UnixNano(), 10))
	formData := formValues.Encode()

	mac := hmac.New(sha512.New, []byte(s.Secret))
	_, err := mac.Write([]byte(formData))
	if err != nil {
		errCh <- err
	}
	sig := hex.EncodeToString(mac.Sum(nil))

	cmd := wsCmd{
		message: map[string]interface{}{
			"command": "subscribe",
			"channel": "1000",
			"key":     s.Key,
			"payload": formData,
			"sign":    sig,
		},
		result: make(chan error, 1),
	}
	s.write <- cmd
	return <-cmd.result
}

func asyncErr(ch chan<- error, err error) {
	select {
	case ch <- err:
	default:
	}
}

func withTimeout(f func() error, tmFunc func(error), timeout time.Duration) error {
	errs := make(chan error)
	go func() {
		err := f()
		select {
		case errs <- err:
		default:
			if tmFunc != nil {
				tmFunc(err)
			}
		}
	}()
	select {
	case err := <-errs:
		return err
	case <-time.After(timeout):
		return errors.New("timeout")
	}
}
