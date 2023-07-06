package maotai

import (
	"fmt"
	"qq/bot"
	"qq/config"
	"qq/cronjob"
	"qq/features/imaotai"
)

func init() {
	cronjob.Manager().NewCommand("maotai", func(robot bot.CronBot) error {
		var res string
		for _, info := range config.MaoTaiInfoMap() {
			if info.Expired() {
				res += fmt.Sprintf("%s: token已过期，需要重新登陆\n", fuzzyPhone(info.Phone))
				continue
			}
			res += fmt.Sprintf("%s:\n%s\n", fuzzyPhone(info.Phone), imaotai.Run(info.Phone))
		}
		robot.SendGroup(config.GroupID(), "茅台申购结束")
		robot.SendGroup(config.GroupID(), res)
		return nil
	}).DailyAt("9:30")
}

func fuzzyPhone(phone string) string {
	if len(phone) == 11 {
		phone = phone[0:3] + "****" + phone[7:11]
	}
	return phone
}
