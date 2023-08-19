package maotai

import (
	"qq/bot"
	"qq/config"
	"qq/cronjob"
	"qq/features/imaotai"
)

func init() {
	cronjob.Manager().NewCommand("maotai", func(robot bot.CronBot) error {
		robot.SendTextImageToGroup(config.GroupID(), imaotai.ReservationAll())
		robot.SendGroup(config.GroupID(), "茅台申购结束")

		return nil
	}).DailyAt("9:10")
	cronjob.Manager().NewCommand("maotai-reward", func(robot bot.CronBot) error {
		robot.SendTextImageToGroup(config.GroupID(), imaotai.ReceiveAllReward())
		robot.SendGroup(config.GroupID(), "茅台自动领取小游戏奖励")

		return nil
	}).DailyAt("15:00")
}
