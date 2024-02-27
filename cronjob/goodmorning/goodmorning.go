package goodmorning

import (
	"qq/bot"
	"qq/config"
	"qq/cronjob"
	"qq/features/goodmorning"
)

func init() {
	cronjob.Manager().NewCommand("Good morning", func(robot bot.CronBot) error {
		robot.SendToUser(config.UserID(), goodmorning.Get())
		return nil
	}).DailyAt("8")
}
