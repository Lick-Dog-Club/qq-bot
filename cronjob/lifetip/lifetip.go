package lifetip

import (
	"qq/bot"
	"qq/config"
	"qq/cronjob"
	"qq/features/lifetip"
)

func init() {
	cronjob.Manager().NewCommand("lifetip", func(robot bot.Bot) error {
		robot.SendGroup(config.GroupId(), lifetip.Tip())
		return nil
	}).DailyAt("9:20")
}
