package gold

import (
	"fmt"
	"qq/bot"
	"qq/config"
	"qq/cronjob"
	"qq/features/gold"
)

func init() {
	cronjob.NewCommand("gold", func(bot bot.CronBot) error {
		s := run()
		if s != "" {
			bot.SendToUser(config.UserID(), s)
		}
		return nil
	}).DailyAt("12:00")
}

func run() string {
	get := gold.Get("JO_52683", 10)
	if len(get.Data) > 0 && get.Data[0].IsToday() {
		str := "涨了"
		if get.Data[0].Q70 < 0 {
			str = "跌了"
		}
		return fmt.Sprintf("今日金价: %v 元/克, %s %v 元/克。", get.Data[0].Q1, str, get.Data[0].Q70)
	}
	return ""
}
