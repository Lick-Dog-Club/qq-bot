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
			res += fmt.Sprintf("%s:\n%s\n", info.Phone, imaotai.Run(info.Phone))
		}
		robot.SendGroup(config.GroupID(), "茅台申购结束")
		robot.SendGroup(config.GroupID(), res)
		return nil
	}).DailyAt("9:15")
}
