package jin10

import (
	"fmt"
	"qq/bot"
	"qq/config"
	"qq/cronjob"
	"qq/features/jin10"
	"qq/util"
	"time"
)

func init() {
	cronjob.NewCommand("jin10", func(bot bot.CronBot) error {
		bot.SendToUser(config.UserID(), fmt.Sprintf(`
今日大事件: %s
%s

明日大事件: %s
%s
`,
			util.Today().Format("2006-01-02"),
			jin10.Get(util.Today()),
			util.Today().Add(24*time.Hour).Format("2006-01-02"),
			jin10.Get(util.Today().Add(24*time.Hour))),
		)
		return nil
	}).DailyAt("09:30")
	cronjob.NewCommand("jin10-watch", func(bot bot.CronBot) error {
		for _, item := range jin10.BigEvents(time.Now()) {
			if item.PubTime.Local().Format("2006-01-02 15:04") == time.Now().Add(-1*time.Minute).Format("2006-01-02 15:04") &&
				item.Actual != nil {
				bot.SendToUser(config.UserID(), item.Render())
				util.Bark(
					item.AffectStr(),
					fmt.Sprintf("%s%s, 预测值: %v, 公布值: %s", item.Country, item.Name, item.Consensus, *item.Actual),
					config.BarkUrls()...,
				)
			}
		}
		return nil
	}).EveryMinute()
}
