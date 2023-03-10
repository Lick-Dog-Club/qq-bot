package dx

import (
	"qq/bot"
	"qq/config"
	"qq/cronjob"
	"qq/features/daxin"
)

func init() {
	cronjob.Manager().NewCommand("daxin", func(robot bot.CronBot) error {
		robot.SendGroup(config.GroupID(), daxin.Get())
		return nil
	}).Weekdays().At("9:00")
}
