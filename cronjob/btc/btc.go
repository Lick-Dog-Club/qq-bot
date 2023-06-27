package btc

import (
	"context"
	"fmt"
	"qq/bot"
	"qq/config"
	"qq/cronjob"
	"qq/util"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	"golang.org/x/exp/constraints"
)

var cn = &ContractNotifier{}

func init() {
	cronjob.Manager().NewCommand("btc notify", func(bot bot.CronBot) error {
		if config.BinanceKey() != "" && config.BinanceSecret() != "" {
			cn.client = binance.NewFuturesClient(config.BinanceKey(), config.BinanceSecret())
			if alert, ok := cn.Alert(); ok {
				bot.SendToUser(config.UserID(), alert)
				util.Bark("BTC", alert, config.BarkUrls()...)
			}
		}
		return nil
	}).EveryTwoSeconds()
}

// ContractNotifier 监控合约
type ContractNotifier struct {
	client  *futures.Client
	prices  []float64
	alertAt time.Time
}

func toFloat64(s string) float64 {
	float, _ := strconv.ParseFloat(s, 64)
	return float
}
func (cn *ContractNotifier) Alert() (string, bool) {
	res, _ := cn.client.NewListPricesService().Symbol("BTCUSDT").Do(context.TODO())
	if len(res) > 0 {
		cn.prices = append(cn.prices, toFloat64(res[0].Price))
	}
	return cn.alert()
}

func max[T constraints.Ordered](items []T) (res T) {
	if len(items) < 1 {
		return
	}
	res = items[0]
	for i := 1; i < len(items); i++ {
		if items[i] > res {
			res = items[i]
		}
	}
	return
}
func min[T constraints.Ordered](items []T) (res T) {
	if len(items) < 1 {
		return
	}
	res = items[0]
	for i := 1; i < len(items); i++ {
		if items[i] < res {
			res = items[i]
		}
	}
	return
}

func (cn *ContractNotifier) alert() (string, bool) {
	if len(cn.prices) < 2 {
		return "", false
	}
	maxPrice := max(cn.prices)
	minPrice := min(cn.prices)
	if maxPrice-minPrice > toFloat64(config.BinanceDiff()) && time.Now().Sub(cn.alertAt).Seconds() > 30 {
		cn.prices = nil
		cn.alertAt = time.Now()
		return fmt.Sprintf("BTC 出现异动，当前最低值为 %.0f, 最高为 %.0f", minPrice, maxPrice), true
	}
	if len(cn.prices) > 50 {
		cn.prices = nil
	}
	return "", false
}
