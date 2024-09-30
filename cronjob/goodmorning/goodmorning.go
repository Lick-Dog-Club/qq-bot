package goodmorning

import (
	"qq/bot"
	"qq/config"
	"qq/cronjob"
	"qq/features/goodmorning"
	"time"
)

func init() {
	cronjob.NewCommand("Good morning", func(robot bot.CronBot) error {
		robot.SendTextImageToUser(config.UserID(), goodmorning.Get(time.Now()))
		return nil
	}).DailyAt("8")
}
