package bitgetcoin

import (
	"encoding/json"
	"fmt"
	"math"
	"qq/bot"
	"qq/config"
	"qq/cronjob"
	"qq/features/bitget"
	"qq/util"
	"strings"
)

var money = map[string]float64{}

func init() {
	cronjob.Manager().NewCommand("bitget-watch-coin", func(bot bot.CronBot) error {
		err := run(bot)
		if err != nil {
			return err
		}
		return nil
	}).EveryThirtySeconds()
	cronjob.Manager().NewCommand("bitget-money-goal", func(bot bot.CronBot) error {
		err := runGoal(bot)
		if err != nil {
			return err
		}
		return nil
	}).EveryThirtySeconds()
}

func runGoal(bot bot.CronBot) error {
	if config.BgApiSecretKey() != "" && config.BgApiKey() != "" && config.BgPassphrase() != "" {
		goal := config.BgGoal()
		for _, coin := range goal {
			fmt.Println(coin)
			usdt, err := bitget.TransUsdt(coin.Name)
			if err != nil {
				return err
			}
			coinName := strings.TrimSuffix(coin.Name, "USDT_SPBL")
			if coin.Price < 0 {
				if usdt <= math.Abs(coin.Price) {
					str := fmt.Sprintf("%s 跌, 当前价格为 %v", coinName, usdt)
					fmt.Println(str)
					bot.SendToUser(config.UserID(), str)
					util.Bark(fmt.Sprintf("监控 %s 价格", coinName), str, config.BarkUrls()...)
				}
			}
			if coin.Price > 0 {
				if usdt >= math.Abs(coin.Price) {
					str := fmt.Sprintf("%s 涨, 当前价格为 %v", coinName, usdt)
					fmt.Println(str)
					bot.SendToUser(config.UserID(), str)
					util.Bark(fmt.Sprintf("监控 %s 价格", coinName), str, config.BarkUrls()...)
				}
			}
		}
	}
	return nil
}

func run(bot bot.CronBot) error {
	if config.BgApiSecretKey() != "" && config.BgApiKey() != "" && config.BgPassphrase() != "" {
		for _, coin := range config.BgCoinWatch() {
			//fmt.Println(coin)
			usdt, err := bitget.TransUsdt(coin.Name)
			if err != nil {
				return err
			}
			f, ok := money[coin.Name]
			if !ok {
				money[coin.Name] = usdt
				return nil
			}
			coinName := strings.TrimSuffix(coin.Name, "USDT_SPBL")
			for _, f2 := range coin.Rate {
				fmt.Printf("目标价格 %v, 当前价格%v\n", f*(1+f2), usdt)
				if f2 < 0 {
					if f*(1+f2) >= usdt {
						str := fmt.Sprintf("%s 跌幅超过 %v%%, 当前价格为 %v", coinName, f2*100, usdt)
						fmt.Println(str)
						util.Bark(fmt.Sprintf("监控 %s 价格下跌", coinName), str, config.BarkUrls()...)
						for _, buy := range config.BgBuyCoin() {
							if buy.Coin == coin.Name && buy.PriceBelow >= usdt && usdt > 0 {
								buyPrice := usdt * 0.99
								spot, err := bitget.BuySpot(coinName, fmt.Sprintf("%v", buyPrice), config.BgOneHandUSDT())
								marshal, _ := json.Marshal(spot)
								bot.SendToUser(config.UserID(), fmt.Sprintf("购买 %s\n当前价格 %v\n买入价格 %v\n结果: %v\nerror: %v", coinName, usdt, buyPrice, string(marshal), err))
							}
						}
					}
				}
				if f2 > 0 {
					if f*(1+f2) <= usdt {
						str := fmt.Sprintf("%s 涨幅超过 %v%%, 当前价格为 %v", coinName, f2*100, usdt)
						fmt.Println(str)
						util.Bark(fmt.Sprintf("监控 %s 价格上涨", coinName), str, config.BarkUrls()...)
					}
				}
			}
			money[coin.Name] = usdt
		}
	}
	return nil
}
