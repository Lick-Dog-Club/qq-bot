package picture

import (
	"qq/bot"
	"qq/config"
	"qq/cronjob"
	"qq/features/pixiv"
)

func init() {
	cronjob.Manager().NewCommand("setu", func(robot bot.Bot) error {
		image, err := pixiv.Image("n")
		if err == nil {
			robot.SendGroup(config.GroupId(), image)
			robot.SendGroup(config.GroupId(), "每日一图~")
		}
		return nil
	}).Weekdays().At("9,13,17").HourlyAt([]int{30})
}
