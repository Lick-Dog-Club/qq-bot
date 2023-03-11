package picture

import (
	"fmt"
	"os"

	"qq/bot"
	"qq/config"
	"qq/cronjob"
	"qq/features/pixiv"
)

func init() {
	// cronjob.Manager().NewCommand("setu", func(robot bot.CronBot) error {
	// 	robot.SendGroup(config.GroupID(), picture.Url())
	// 	return nil
	// }).Weekdays().At("9,13,17").HourlyAt([]int{30})
	cronjob.Manager().NewCommand("tome", func(robot bot.CronBot) error {
		uid := config.UserID()
		if uid == "" {
			return nil
		}
		image, err := pixiv.Image("")
		if err == nil {
			robot.SendToUser(uid, fmt.Sprintf("[CQ:image,file=file://%s]", image))
			os.Remove(image)
		}
		return nil
	}).DailyAt("8-23").HourlyAt([]int{10})
}
