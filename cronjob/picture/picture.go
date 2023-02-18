package picture

import (
	"qq/bot"
	"qq/config"
	"qq/cronjob"
	"qq/features/picture"
	"qq/features/pixiv"
)

func init() {
	cronjob.Manager().NewCommand("setu", func(robot bot.Bot) error {
		robot.SendGroup(config.GroupID(), picture.Url())
		robot.SendGroup(config.GroupID(), "每日一图~")
		return nil
	}).Weekdays().At("9,13,17").HourlyAt([]int{30})
	cronjob.Manager().NewCommand("tome", func(robot bot.Bot) error {
		uid := config.UserID()
		if uid == "" {
			return nil
		}
		image, err := pixiv.Image("r18_ai")
		if err == nil {
			robot.SendToUser(uid, image)
		}
		return nil
	}).DailyAt("9-18").HourlyAt([]int{5})
}
