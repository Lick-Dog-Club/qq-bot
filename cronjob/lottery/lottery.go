package lottery

import (
	"qq/bot"
	"qq/config"
	"qq/cronjob"
	lottery "qq/features/bililottery"
)

func init() {
	cronjob.Manager().NewCommand("lottery", func(robot bot.CronBot) error {
		cookie := config.BiliCookie()
		uid := config.UserID()
		//log.Printf("开始处理抽奖: uid: %s\n", config.UserID())
		if cookie != "" {
			robot.SendToUser(uid, lottery.Run(func(s string) { robot.SendToUser(uid, s) }, cookie))
			robot.SendToUser(uid, "抽奖结束")
		}
		return nil
	}).DailyAt("09:10")
}
