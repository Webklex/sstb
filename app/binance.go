package app

import (
	"../utils/log"
	"../utils/values"
	"context"
	"encoding/json"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"strings"
	"time"
)

func (j *Job) parseBinOpenOrders(orders []*binance.Order) {
	for _, o := range orders {
		j.AttachBinOrder(o)
	}
}

func (j *Job) cancelAllBinOrders() {
	if _, err := j.BinanceClient.NewCancelOpenOrdersService().Symbol(j.Symbol).Do(context.Background()); err != nil {
		log.Error(err)
		return
	}
}

func (j *Job) AttachBinOrder(o *binance.Order) {
	j.mx.Lock()
	if _, ok := j.orders[o.OrderID]; !ok {
		j.orders[o.OrderID] = &Order{
			Id:     o.OrderID,
			Volume: values.NewFloatFromString(o.OrigQuantity),
			Price:  values.NewFloatFromString(o.Price),
			Side:   string(o.Side),
			Status: string(o.Status),
			Date:   time.Time{},
		}
		j.orders[o.OrderID].Total = j.orders[o.OrderID].Volume.Mul(j.orders[o.OrderID].Price)
		j.orders[o.OrderID].Fee = j.orders[o.OrderID].Total.Div(values.HundredFloat).Mul(&j.Fee)
		log.Success(fmt.Sprintf("%s ORDER REGISTERED: %d", strings.ToUpper(j.Provider.Name), o.OrderID))
	}
	j.mx.Unlock()
}

func (j *Job) validateAmount(amount *values.Float) *values.Float {
	a := values.NewEmptyFloat()
	for {
		if a.Gt(amount) {
			return a.Sub(j.stepSize)
		}
		a = a.Add(j.stepSize)
	}
}

func (j *Job) wsHandler() func(message []byte) {

	return func(message []byte) {
		evt := &BinanceEvent{}
		if err := json.Unmarshal(message, evt); err != nil {
			log.Debug(string(message))
			log.Error(err)
			return
		}

		if evt.OrderType == binance.OrderTypeMarket {
			log.Debug("SKIPPED: " + string(message))
			return
		}

		if evt.EventType == "executionReport" && evt.Symbol == j.Symbol {
			to := &binance.Order{
				Symbol:                   evt.Symbol,
				OrderID:                  evt.OrderId,
				ClientOrderID:            evt.ClientOrderId,
				Price:                    evt.Price.ToString(),
				OrigQuantity:             evt.Quantity.ToString(),
				ExecutedQuantity:         evt.LastExecutedQuantity.ToString(),
				CummulativeQuoteQuantity: evt.CumulativeQuoteQuantity.ToString(),
				Status:                   evt.Status,
				TimeInForce:              evt.TimeInForce,
				Type:                     evt.OrderType,
				Side:                     evt.Side,
				StopPrice:                evt.StopPrice.ToString(),
				IcebergQuantity:          evt.IcebergQuantity.ToString(),
				Time:                     evt.TimeCreated,
				UpdateTime:               evt.EventTime,
				IsWorking:                false,
				IsIsolated:               false,
			}

			if to.Status == binance.OrderStatusTypeNew {
				j.AttachBinOrder(to)
			} else if to.Status == binance.OrderStatusTypeCanceled {
				j.DetachOrder(to.OrderID)
			} else if to.Status == binance.OrderStatusTypeFilled {
				if to.Side == binance.SideTypeBuy {
					// Create a new sell order
					price := evt.Price.Add(j.getStep("sell"))

					dif := evt.Quantity.Div(values.HundredFloat)
					buyFee := dif.Mul(&j.Fee)

					availableAmount := evt.Quantity.Sub(buyFee)
					sellAmount := j.validateAmount(availableAmount)

					if j.balance[j.Secondary].Gt(buyFee) {
						sellAmount = &evt.Quantity

						j.subBalance(j.Secondary, buyFee)
						log.Info(fmt.Sprintf("%s LEND %.8f %s", strings.ToUpper(j.Provider.Name), buyFee, j.Secondary))
					} else {
						dif := availableAmount.Sub(sellAmount)

						j.addBalance(j.Secondary, dif)
						log.Info(fmt.Sprintf("%s GAVE %.8f %s", strings.ToUpper(j.Provider.Name), dif, j.Secondary))
					}

					amt, _ := sellAmount.Float64()

					total := price.Mul(sellAmount)

					pf, _ := price.Float64()
					tot, _ := total.Float64()

					ps := price.ToString()
					as := sellAmount.ToString()

					order, err := j.BinanceClient.NewCreateOrderService().Symbol(j.Symbol).
						Side(binance.SideTypeSell).Type(binance.OrderTypeLimit).
						TimeInForce(binance.TimeInForceTypeGTC).Quantity(as).
						Price(ps).Do(context.Background())
					if err != nil {
						log.Debug(string(message))
						log.Debug(j.Provider.Name, evt.OrderId, "sell", pf, amt, ps, as, tot)
						log.Error(err)
						return
					}

					log.Success(fmt.Sprintf("%s ORDER CREATED: %d", strings.ToUpper(j.Provider.Name), order.OrderID))

					diff := total.Sub(evt.Quantity.Mul(&evt.Price))
					d, _ := diff.Float64()

					go j.SaveOrder(&Order{
						Id:     evt.OrderId,
						Volume: &evt.Quantity,
						Price:  &evt.Price,
						Total:  total,
						Fee:    buyFee,
						Side:   "buy",
						Status: "filled",
						Date:   time.Now(),
					})

					go j.NotifyOrder(pf, amt, tot, d, "sell")

				} else if to.Side == binance.SideTypeSell {
					// Create a new buy order
					price := evt.Price.Sub(j.getStep("buy"))

					amount := j.validateAmount(j.Volume.Div(price))
					total := price.Mul(amount)

					pf, _ := price.Float64()
					amt, _ := amount.Float64()
					tot, _ := total.Float64()

					ps := price.ToString()
					as := amount.ToString()

					order, err := j.BinanceClient.NewCreateOrderService().Symbol(j.Symbol).
						Side(binance.SideTypeBuy).Type(binance.OrderTypeLimit).
						TimeInForce(binance.TimeInForceTypeGTC).Quantity(as).
						Price(ps).Do(context.Background())
					if err != nil {
						log.Debug(string(message))
						log.Debug(j.Provider.Name, evt.OrderId, "buy", pf, amt, ps, as, tot)
						log.Error(err)
						return
					}

					log.Success(fmt.Sprintf("%s ORDER CREATED: %d", strings.ToUpper(j.Provider.Name), order.OrderID))

					diff := evt.Quantity.Mul(&evt.Price).Sub(total)
					d, _ := diff.Float64()

					dif := evt.Quantity.Div(values.HundredFloat)
					sellFee := dif.Mul(&j.Fee)

					go j.SaveOrder(&Order{
						Id:     evt.OrderId,
						Volume: &evt.Quantity,
						Price:  &evt.Price,
						Total:  total,
						Fee:    sellFee,
						Side:   "sell",
						Status: "filled",
						Date:   time.Now(),
					})

					go j.NotifyOrder(pf, amt, tot, d, "buy")
				}
			}
		} else if evt.EventType == "outboundAccountPosition" {
			/**
			{
			  "e": "outboundAccountPosition", //Event type
			  "E": 1564034571105,             //Event Time
			  "u": 1564034571073,             //Time of last account update
			  "B": [                          //Balances Array
			    {
			      "a": "ETH",                 //Asset
			      "f": "10000.000000",        //Free
			      "l": "0.000000"             //Locked
			    }
			  ]
			}
			*/
			if evt.Balances != nil {
				for _, b := range evt.Balances {
					if b.Free.Lt(j.getBalance(b.Asset)) {
						j.setBalance(b.Asset, &b.Free)
						log.Info(fmt.Sprintf("%s AVAILABLE BALANCE %.8f %s", strings.ToUpper(j.Provider.Name), b.Free.ToFloat(), b.Asset))
					}
				}
			}
		}
		log.Debug(j.Provider.Name)
		log.Debug(string(message))
	}
}

func (j *Job) setBinanceBalance() {
	if acc, err := j.BinanceClient.NewGetAccountService().Do(context.Background()); err == nil {
		for _, b := range acc.Balances {
			j.balance[b.Asset] = values.NewFloatFromString(b.Free)
		}
	}
}

func (j *Job) setBinanceStepSize() {
	if ex, err := j.BinanceClient.NewExchangeInfoService().Do(context.Background()); err == nil {
		for _, s := range ex.Symbols {
			if s.Symbol == j.Symbol {
				for _, f := range s.Filters {
					if ft, ok := f["filterType"]; ok {
						if ft == "LOT_SIZE" {
							j.mx.Lock()
							j.stepSize = values.NewFloatFromString(f["stepSize"].(string))
							j.mx.Unlock()
							return
						}
					}
				}
			}
		}
	}
}

func (j *Job) KeepListenKeyAlive(listenKey string, done chan struct{}, stop chan struct{}) {
	ticker := time.NewTicker(time.Minute * 30)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-stop:
			return
		case <-ticker.C:
			if err := j.BinanceClient.NewKeepaliveUserStreamService().ListenKey(listenKey).Do(context.Background()); err != nil {
				log.Access(j.Provider.Name)
				log.Error(err)
				stop <- struct{}{}
				return
			}
		}
	}
}

func (j *Job) WatchBinMarket() {
	for {
		listenKey, err := j.BinanceClient.NewStartUserStreamService().Do(context.Background())
		if err == nil {
			log.Success(fmt.Sprintf("Subscribing to %s account update events..", strings.ToUpper(j.Provider.Name)))
			doneC, stopC, err := binance.WsUserDataServe(listenKey, j.wsHandler(), func(err error) {
				log.Error(err)
			})
			if err != nil {
				log.Error(err)
				stopC <- struct{}{}
			} else {
				go j.KeepListenKeyAlive(listenKey, doneC, stopC)
				<-doneC
			}
		} else {
			log.Error(fmt.Sprintf("Subscribing to %s account update events failed", strings.ToUpper(j.Provider.Name)))
			log.Error(err)
			time.Sleep(time.Second)
		}
	}
}
