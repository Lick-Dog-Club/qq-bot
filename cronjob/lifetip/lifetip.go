package lifetip

import (
	"qq/bot"
	"qq/config"
	"qq/cronjob"
	"qq/features/lifetip"
)

func init() {
	cronjob.Manager().NewCommand("lifetip", func(robot bot.CronBot) error {
		robot.SendGroup(config.GroupID(), lifetip.Tip())
		return nil
	}).Weekdays().DailyAt("9:20")
}
