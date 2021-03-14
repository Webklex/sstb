package app

import (
	"../utils/config"
	"../utils/filesystem"
	"../utils/log"
	"../utils/values"
	"./notifier"
	"context"
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"math/big"
	"reflect"
	"strings"
	"sync"
	"time"
)

var (
	decType             = reflect.TypeOf(values.Float{})
	orderBookDecodeHook = mapstructure.ComposeDecodeHookFunc(decimalDecodeHook)
)

func decimalDecodeHook(from reflect.Type, to reflect.Type, v interface{}) (interface{}, error) {
	if to == decType {
		if str, ok := v.(string); ok {
			val, _ := new(big.Float).SetString(str)
			return values.Float{Float: *val}, nil
		}
		return nil, errors.New(fmt.Sprintf("cannot decode %s to decimal", from.String()))
	}
	return v, nil
}

func NewDefaultJob() *Job {
	j := &Job{
		Config:     config.DefaultConfig(),
		Symbol:     "",
		Primary:    "",
		ProviderId: "",
		Volume:     *values.NewEmptyFloat(),
		Step:       *values.NewEmptyFloat(),
		Fee:        *values.NewEmptyFloat(),
		stepSize:   values.NewFloatFromFloat64(0.00000001),
		Alert: &Alert{
			Buy:     true,
			Sell:    true,
			Idle:    0,
			Summary: true,
		},
		lastOperation: time.Now(),
		mx:            sync.Mutex{},
		orders:        make(map[int64]*Order),
		balance:       make(map[string]*values.Float),
		NotifierIds:   make([]string, 0),
		Notifier:      make([]*notifier.Notifier, 0),
	}
	j.SetContext(j)

	return j
}

func NewJobFromFile(filepath string) *Job {
	j := NewDefaultJob()

	j.Load(filepath)
	j.File = filepath
	j.Id = filesystem.FileNameWithoutExtension(filepath)
	j.Init()

	return j
}

func (j *Job) Init() {
	sec := strings.Replace(j.Symbol, j.Primary, "", 1)
	sec = strings.Replace(sec, "_", "", 1)
	sec = strings.Replace(sec, "-", "", 1)
	sec = strings.Replace(sec, "/", "", 1)

	j.Secondary = sec

	j.balance[j.Primary] = values.NewEmptyFloat()
	j.balance[j.Secondary] = values.NewEmptyFloat()
}

func (j *Job) Start() {
	if j.Provider.Exchange == "poloniex" {
		j.StartPoloniex()
	} else {
		j.StartBinance()
	}
}

func (j *Job) StartPoloniex() {
	if orders, err := j.PoloniexClient.GetOpenOrders(j.Symbol); err == nil {
		j.parsePolOpenOrders(orders)
	}

	j.WatchPolMarket()
}

func (j *Job) StartBinance() {
	j.setBinanceStepSize()
	j.setBinanceBalance()
	if orders, err := j.BinanceClient.NewListOpenOrdersService().Symbol(j.Symbol).Do(context.Background()); err == nil {
		j.parseBinOpenOrders(orders)
	}

	j.WatchBinMarket()
}

func (j *Job) getStep(d string) *values.Float {
	if d == "sell" {
		if j.SellStep.Gt(values.ZeroFloat) {
			return &j.SellStep
		}
	} else {
		if j.BuyStep.Gt(values.ZeroFloat) {
			return &j.BuyStep
		}
	}
	return &j.Step
}

func (j *Job) setProvider(p *Provider) {
	j.mx.Lock()
	j.Provider = p

	if j.Provider.Exchange == "poloniex" {
		j.PoloniexClient = j.Provider.NewPoloniexClient()
	} else {
		j.BinanceClient = j.Provider.NewBinanceClient()
	}

	j.mx.Unlock()
}

func (j *Job) AttachOrder(o *Order) {
	j.mx.Lock()
	if _, ok := j.orders[o.Id]; !ok {
		j.orders[o.Id] = o
		log.Success(fmt.Sprintf("%s ORDER REGISTERED: %d", strings.ToUpper(j.Provider.Name), o.Id))
	}
	j.mx.Unlock()
}

func (j *Job) DetachOrder(num int64) {
	j.mx.Lock()
	if _, ok := j.orders[num]; ok {
		delete(j.orders, num)
	}
	j.mx.Unlock()
	log.Warn(fmt.Sprintf("%s ORDER REMOVED: %d", strings.ToUpper(j.Provider.Name), num))
}

func (j *Job) GetOrder(on int64) (*Order, error) {
	j.mx.Lock()
	o, ok := j.orders[on]
	j.mx.Unlock()
	if !ok {
		return nil, errors.New("order not found")
	}
	return o, nil
}

func (j *Job) getBalance(asset string) *values.Float {
	j.mx.Lock()
	balance := values.NewEmptyFloat()
	if b, ok := j.balance[asset]; ok {
		balance = b
	}
	j.mx.Unlock()

	return balance
}

func (j *Job) setBalance(asset string, value *values.Float) {
	j.mx.Lock()
	j.balance[asset] = value
	j.mx.Unlock()
}

func (j *Job) addBalance(asset string, value *values.Float) {
	balance := j.getBalance(asset)
	j.setBalance(asset, balance.Add(value))
}

func (j *Job) subBalance(asset string, value *values.Float) {
	balance := j.getBalance(asset)
	j.setBalance(asset, balance.Sub(value))
}

func (j *Job) SendIdleAlert() {
	text := fmt.Sprintf("#### %s on %s is idling\n", strings.ToUpper(j.Symbol), strings.ToUpper(j.Provider.Name))
	j.Notify(text)
	j.touch()
}

func (j *Job) touch() {
	j.mx.Lock()
	j.lastOperation = time.Now()
	j.mx.Unlock()
}

func (j *Job) Notify(msg string) {
	for _, n := range j.Notifier {
		go n.Send(msg)
	}
}

func (j *Job) NotifyOrder(pf float64, amt float64, total float64, dif float64, direction string) {
	text := fmt.Sprintf("#### %s %s order placed on %s\n", strings.ToUpper(direction), strings.ToUpper(j.Symbol), strings.ToUpper(j.Provider.Name))
	text = text + `
| Gain  | Price | Amount | Total  |
|:------|:------|:-------|:-------|`
	text = text + fmt.Sprintf("\n| %.8f | %.8f | %.8f | %.8f |", dif, pf, amt, total)

	if (j.Alert.Sell && direction == "sell") || (j.Alert.Buy && direction == "buy") {
		j.Notify(text)
	}

	j.touch()
}
