package picture

import (
	"qq/bot"
	"qq/config"
	"qq/cronjob"
	"qq/features/picture"
)

func init() {
	cronjob.Manager().NewCommand("setu", func(robot bot.Bot) error {
		robot.SendGroup(config.GroupId(), picture.Url())
		robot.SendGroup(config.GroupId(), "每日一图~")
		return nil
	}).At("16:15")
}
