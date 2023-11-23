package bitget

import (
	"fmt"
	"math"
	"qq/bot"
	"qq/config"
	"qq/cronjob"
	"qq/features/bitget"
	"qq/util"
)

//func init() {
//	cronjob.Manager().NewCommand("bitget", func(bot bot.CronBot) error {
//		if config.BgApiSecretKey() != "" && config.BgApiKey() != "" && config.BgPassphrase() != "" {
//			if v := bitget.Get(false); v != "" {
//				bot.SendToUser(config.UserID(), v)
//				util.Bark("有变动", v, config.BarkUrls()...)
//			}
//		}
//		return nil
//	}).EveryTenSeconds()
//}

var money float64

func init() {
	cronjob.Manager().NewCommand("bitget-money-total", func(bot bot.CronBot) error {
		if config.BgApiSecretKey() != "" && config.BgApiKey() != "" && config.BgPassphrase() != "" {
			total, err := bitget.MoneyTotal()
			if err != nil {
				fmt.Println(err)
				return nil
			}
			if money == 0 {
				money = total
				return nil
			}
			if math.Abs(total-money) > util.ToFloat64(config.BgMoneyDiff()) {
				s := fmt.Sprintf("资产变动: %.2f u, 目前一共 %.2f u", total-money, total)
				util.Bark("money", s, config.BarkUrls()...)
				bot.SendToUser(config.UserID(), s)
			}
			money = total
		}
		return nil
	}).EveryThirtySeconds()
}
