package kfc

import (
	"qq/bot"
	"qq/config"
	"qq/cronjob"
	"qq/features/kfc"
)

func init() {
	cronjob.NewCommand("kfc", func(robot bot.CronBot) error {
		robot.SendGroup(config.GroupID(), kfc.Get())
		return nil
	}).Thursdays().At("17")
}
