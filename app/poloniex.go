package app

import (
	"../api/poloniex"
	"../utils/log"
	"../utils/values"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"strings"
	"time"
)

func (j *Job) parsePolOpenOrders(orders []*poloniex.OpenOrder) {
	for _, o := range orders {
		j.AttachPolOrder(o)
	}
}

func (j *Job) parsePolTradeOrdersUpdate(arr []interface{}) (poloniex.OrderBookUpd, error) {
	var tradeOrder poloniex.OrderBookUpd
	toDecode := []interface{}{&tradeOrder.Price, &tradeOrder.Size}
	obookDec, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:     &toDecode,
		DecodeHook: orderBookDecodeHook,
	})
	return tradeOrder, obookDec.Decode(arr)
}

func (j *Job) AttachPolOrder(o *poloniex.OpenOrder) {
	j.mx.Lock()
	if o.Total.ToFloat() <= 0 {
		o.Total = *o.Amount.Mul(&o.Rate)
	}
	if _, ok := j.orders[o.OrderNumber]; !ok {
		j.orders[o.OrderNumber] = &Order{
			Id:     o.OrderNumber,
			Volume: &o.Amount,
			Price:  &o.Rate,
			Total:  &o.Total,
			Fee:    o.Total.Div(values.HundredFloat).Mul(&j.Fee),
			Side:   o.Type,
			Status: "",
			Date:   time.Time{},
		}
		log.Success(fmt.Sprintf("%s ORDER REGISTERED: %d", strings.ToUpper(j.Provider.Name), o.OrderNumber))
	}
	j.mx.Unlock()
}

func (j *Job) HandlePolTradeOrder(to *poloniex.TradeOrder) {
	o, err := j.GetOrder(to.Number)
	if err != nil {
		log.Error(fmt.Sprintf("%s ORDER NOT RELATED: %d", strings.ToUpper(j.Provider.Name), to.Number))
		return
	}

	filled := to.Amount.Eq(values.ZeroFloat) || to.Amount.Lt(values.ZeroFloat)

	if (to.Type == "f" || to.Type == "s") && filled {
		// Order is fulfilled
		// Order is known
		if o.Side == "sell" {
			// create a buy order
			price := o.Price.Sub(j.getStep("buy"))

			amount := j.Volume.Div(price)
			total := price.Mul(amount)

			diff := o.Total.Sub(total)

			dif, _ := diff.Float64()
			pf, _ := price.Float64()
			amt, _ := amount.Float64()
			tot, _ := total.Float64()

			to, err := j.PoloniexClient.Buy(j.Symbol, pf, amt)
			if err != nil {
				log.Error(err)
				log.Warn("Idle and try again..")
				time.Sleep(time.Second)

				to, err = j.PoloniexClient.Buy(j.Symbol, pf, amt)
				if err != nil {
					log.Debug(j.Provider, o.Id, "buy", pf, amount, tot)
					log.Error(err)
					return
				}
			}
			log.Success(fmt.Sprintf("%s ORDER CREATED: %d", strings.ToUpper(j.Provider.Name), to.Number))


			sellFee := o.Volume.Div(values.HundredFloat).Mul(&j.Fee)

			go j.SaveOrder(&Order{
				Id:     o.Id,
				Volume: o.Volume,
				Price:  o.Price,
				Total:  o.Total,
				Fee:    sellFee,
				Side:   "sell",
				Status: "filled",
				Date:   time.Now(),
			})

			go j.NotifyOrder(pf, amt, tot, dif, "buy")
		} else {
			// create a sell order
			price := o.Price.Add(j.getStep("sell"))

			fee := o.Volume.Div(values.HundredFloat).Mul(&j.Fee)
			amount := o.Volume.Sub(fee)
			total := price.Mul(amount)
			diff := total.Sub(o.Total)

			pf, _ := price.Float64()
			amt, _ := amount.Float64()
			tot, _ := total.Float64()
			dif, _ := diff.Float64()

			to, err := j.PoloniexClient.Sell(j.Symbol, pf, amt)
			if err != nil {
				log.Error(err)
				log.Warn("Idle and try again..")
				time.Sleep(time.Second)

				to, err = j.PoloniexClient.Sell(j.Symbol, pf, amt)
				if err != nil {
					log.Error(err)
					log.Debug(j.Provider, o.Id, "sell", pf, amount, tot)
					return
				}
			}

			log.Success(fmt.Sprintf("%s ORDER CREATED: %d", strings.ToUpper(j.Provider.Name), to.Number))
			buyFee := o.Volume.Div(values.HundredFloat).Mul(&j.Fee)

			go j.SaveOrder(&Order{
				Id:     o.Id,
				Volume: o.Volume,
				Price:  o.Price,
				Total:  o.Total,
				Fee:    buyFee,
				Side:   "buy",
				Status: "filled",
				Date:   time.Now(),
			})

			go j.NotifyOrder(pf, amt, tot, dif, "sell")
		}
	} else if to.Type == "c" && filled {
		j.DetachOrder(o.Id)
	} else {
		log.Info(fmt.Sprintf("%s ORDER UPDATE: %s %d @ %s - %.8f", strings.ToUpper(j.Provider.Name), j.Symbol, to.Number, to.Type, to.Amount.ToFloat()))
	}
}

func (j *Job) handleAccountUpdates(upd poloniex.AccountUpd, pair *poloniex.Pair) {
	for _, no := range upd.NewOrders {
		if no.Symbol == pair.Id {
			j.AttachPolOrder(&no)
		}
	}

	for _, to := range upd.TradeOrders {
		j.HandlePolTradeOrder(&to)
	}
}

func (j *Job) WatchPolMarket() {
	AcUpdChan := make(chan poloniex.AccountUpd, 128)
	stopChan := make(chan bool)

	pair := j.PoloniexClient.GetPair(j.Symbol)
	if pair == nil {
		return
	}

	go func() {
		for upd := range AcUpdChan {
			go j.handleAccountUpdates(upd, pair)
		}
	}()

	go func() {
		for {
			log.Success(fmt.Sprintf("Subscribing to %s account update events..", strings.ToUpper(j.Provider.Name)))
			if err := j.PoloniexClient.SubscribeAccount(AcUpdChan, stopChan); err != nil {
				log.Error("client: sub error: %v", err)
				time.Sleep(time.Second)
			}
		}
	}()

	for {
		select {
		case <-stopChan:
			return
		}
	}
}
