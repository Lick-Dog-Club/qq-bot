package lottery

import (
	"qq/bot"
	"qq/config"
	"qq/cronjob"
	lottery "qq/features/bili-lottery"

	log "github.com/sirupsen/logrus"
)

func init() {
	cronjob.Manager().NewCommand("lottery", func(robot bot.CronBot) error {
		cookie := config.BiliCookie()
		log.Printf("开始处理抽奖: uid: %s\n", config.UserID())
		if cookie != "" {
			lottery.Run(func(s string) { robot.SendToUser(config.UserID(), s) }, cookie)
		}
		return nil
	}).DailyAt("15:20")
}
