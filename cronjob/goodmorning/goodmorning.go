package goodmorning

import (
	"qq/bot"
	"qq/config"
	"qq/cronjob"
	"qq/features/goodmorning"
)

func init() {
	cronjob.Manager().NewCommand("Good morning", func(robot bot.CronBot) error {
		robot.SendTextImageToUser(config.UserID(), goodmorning.Get())
		return nil
	}).DailyAt("8")
}
