package tianqi

import (
	"qq/bot"
	"qq/config"
	"qq/cronjob"
	"qq/features/weather"
)

func init() {
	cronjob.Manager().NewCommand("tianqi", func(robot bot.CronBot) error {
		robot.SendGroup(config.GroupID(), weather.Get("杭州"))
		return nil
	}).DailyAt("8:15")
}
