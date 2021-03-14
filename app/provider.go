package app

import (
	"../api/poloniex"
	"../utils/log"
	"context"
	"github.com/adshao/go-binance/v2"
	"time"
)

type Provider struct {
	Name     string `json:"name"`
	Exchange string `json:"exchange"`
	Key      string `json:"key"`
	Secret   string `json:"secret"`
}

func (p *Provider) NewPoloniexClient() *poloniex.Config {
	if p.Key != p.Secret {
		client := poloniex.NewPoloniexApi(p.Key, p.Secret)
		if err := client.Setup(); err != nil {
			log.Error(err)
			return nil
		}
		return client
	}
	return nil
}

func (p *Provider) NewBinanceClient() *binance.Client {

	if p.Key != p.Secret {
		client := binance.NewClient(p.Key, p.Secret)

		t, err := client.NewServerTimeService().Do(context.Background())
		if err != nil {
			log.Error(err)
		} else {
			nt := time.Now().Unix() * 1000

			if nt > t {
				client.TimeOffset = t - nt
			} else {
				client.TimeOffset = nt - t
			}

			return client
		}
	}

	return nil
}
