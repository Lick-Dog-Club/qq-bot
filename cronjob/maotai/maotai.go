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
		robot.SendGroup(config.GroupID(), fmt.Sprintf("[CQ:image,file=file://%s]", imaotai.ReservationAll()))
		robot.SendGroup(config.GroupID(), "茅台申购结束")

		return nil
	}).DailyAt("9:10")
}
