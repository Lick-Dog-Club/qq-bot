package maotai

import (
	"fmt"
	"qq/bot"
	"qq/config"
	"qq/cronjob"
	"qq/features/imaotai"
	"qq/util"
)

func init() {
	cronjob.Manager().NewCommand("maotai", func(robot bot.CronBot) error {
		var res string
		for _, info := range config.MaoTaiInfoMap() {
			if info.Expired() {
				res += fmt.Sprintf("%s: token已过期，需要重新登陆\n", util.FuzzyPhone(info.Phone))
				continue
			}
			res += fmt.Sprintf("%s:\n%s\n", util.FuzzyPhone(info.Phone), imaotai.Run(info.Phone))
		}
		robot.SendGroup(config.GroupID(), "茅台申购结束")
		robot.SendGroup(config.GroupID(), res)
		return nil
	}).DailyAt("9:10")
}
