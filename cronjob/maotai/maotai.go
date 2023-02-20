package maotai

import (
	"qq/bot"
	"qq/config"
	"qq/cronjob"
)

func init() {
	cronjob.Manager().NewCommand("maotai", func(robot bot.CronBot) error {
		robot.SendGroup(config.GroupID(), "开始申购茅台了~")
		return nil
	}).DailyAt("9:15")
}
