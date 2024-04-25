package bitgetcoin

import (
	"fmt"
	"qq/bot"
	"qq/config"
	"qq/cronjob"
	"qq/features/bitget"
	"qq/util"
	"strings"
)

var money = map[string]float64{}

func init() {
	cronjob.Manager().NewCommand("bitget-money-total-watch-coin", func(bot bot.CronBot) error {
		err := run()
		if err != nil {
			return err
		}
		return nil
	}).EveryThirtySeconds()
}

func run() error {
	if config.BgApiSecretKey() != "" && config.BgApiKey() != "" && config.BgPassphrase() != "" {
		for _, coin := range config.BgCoinWatch() {
			fmt.Println(coin)
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
