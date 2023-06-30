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
				var t string = "空"
				if alert.isMore {
					t = "多"
				}
				bot.SendToUser(config.UserID(), alert.msg)
				// 五分钟之内如果开过一次仓就不再继续开仓
				// 因为防止误开，比如暴跌之后的回调，也可能出发报警，但不需要开仓
				if cn.currentAlert.date.IsZero() || cn.currentAlert.date.Add(15*time.Minute).Before(alert.date) {
					bot.SendToUser(config.UserID(), fmt.Sprintf("开%s，价格是 %v", t, alert.openPrice))
					cn.currentAlert = alert
				}
				util.Bark("BTC", alert.msg, config.BarkUrls()...)
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

	currentAlert alertBody
}

func toFloat64(s string) float64 {
	float, _ := strconv.ParseFloat(s, 64)
	return float
}

type alertBody struct {
	msg string
	// 开多？
	isMore    bool
	openPrice float64
	date      time.Time
}

func (cn *ContractNotifier) Alert() (alertBody, bool) {
	res, _ := cn.client.NewListPricesService().Symbol("BTCUSDT").Do(context.TODO())
	if len(res) > 0 {
		cn.prices = append(cn.prices, toFloat64(res[0].Price))
	}
	return cn.alert()
}

func max[T constraints.Ordered](items []T) (idx int, res T) {
	if len(items) < 1 {
		return
	}
	res = items[0]
	for i := 1; i < len(items); i++ {
		if items[i] > res {
			res = items[i]
			idx = i
		}
	}
	return
}
func min[T constraints.Ordered](items []T) (idx int, res T) {
	if len(items) < 1 {
		return
	}
	res = items[0]
	for i := 1; i < len(items); i++ {
		if items[i] < res {
			res = items[i]
			idx = i
		}
	}
	return
}

func (cn *ContractNotifier) alert() (alertBody, bool) {
	if len(cn.prices) < 2 {
		return alertBody{}, false
	}
	maxIdx, maxPrice := max(cn.prices)
	minIdx, minPrice := min(cn.prices)
	if maxPrice-minPrice > toFloat64(config.BinanceDiff()) && time.Now().Sub(cn.alertAt).Seconds() > 30 {
		cn.prices = nil
		cn.alertAt = time.Now()
		var openPrice = minPrice - 100
		if maxIdx > minIdx {
			openPrice = maxPrice + 100
		}
		return alertBody{
			msg:       fmt.Sprintf("BTC 出现异动，当前最低值为 %.0f, 最高为 %.0f", minPrice, maxPrice),
			isMore:    maxIdx < minIdx,
			openPrice: openPrice,
			date:      time.Now(),
		}, true
	}
	if len(cn.prices) > 50 {
		cn.prices = nil
	}
	return alertBody{}, false
}
