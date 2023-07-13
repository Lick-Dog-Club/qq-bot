package comic

import (
	"qq/bot"
	"qq/config"
	"qq/cronjob"
	"qq/features/comic"
)

func init() {
	cronjob.Manager().NewCommand("haizeiwang", func(bot bot.CronBot) error {
		c := comic.Get("haizeiwang")
		if c.TodayUpdated() {
			bot.SendGroup(config.GroupID(), c.Render())
		}
		return nil
	}).DailyAt("12:00")
}
