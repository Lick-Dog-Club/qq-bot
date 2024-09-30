package ddys

import (
	"qq/bot"
	"qq/config"
	"qq/cronjob"
	"qq/features/ddys"
	"time"
)

func init() {
	cronjob.NewCommand("ddys-once-day-me", func(bot bot.CronBot) error {
		for _, m := range ddys.Get("dm", 14*time.Hour) {
			bot.SendToUser(config.UserID(), m.String())
		}
		return nil
	}).DailyAt("9:10")

	cronjob.NewCommand("ddys-hourly-me", func(bot bot.CronBot) error {
		for _, m := range ddys.Get("dm", 1*time.Hour) {
			bot.SendToUser(config.UserID(), m.String())
		}
		return nil
	}).DailyAt("10-20").HourlyAt([]int{10})

	cronjob.NewCommand("ddys-once-day", func(bot bot.CronBot) error {
		for _, m := range ddys.Get("dy", 14*time.Hour) {
			bot.SendGroup(config.GroupID(), m.String())
		}
		return nil
	}).DailyAt("9:10")

	cronjob.NewCommand("ddys-hourly", func(bot bot.CronBot) error {
		for _, m := range ddys.Get("dy", 1*time.Hour) {
			bot.SendGroup(config.GroupID(), m.String())
		}
		return nil
	}).DailyAt("10-20").HourlyAt([]int{10})
}
