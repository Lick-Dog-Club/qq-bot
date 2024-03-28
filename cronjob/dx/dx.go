package dx

import (
	"qq/bot"
	"qq/config"
	"qq/cronjob"
	"qq/features/daxin"
	"qq/util"
)

func init() {
	cronjob.Manager().NewCommand("daxin", func(robot bot.CronBot) error {
		get, b := daxin.Get()
		if b {
			util.Bark("今日有新债", get, config.BarkUrls()...)
		}
		robot.SendGroup(config.GroupID(), get)
		return nil
	}).Weekdays().At("9:00")
}
