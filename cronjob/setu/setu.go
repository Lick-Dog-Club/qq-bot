package maotai

import (
	"qq/bot"
	"qq/config"
	"qq/cronjob"
	"qq/features/picture"
	"strconv"
)

func toInt(s string) int {
	atoi, _ := strconv.Atoi(s)
	return atoi
}

func init() {
	cronjob.Manager().NewCommand("setu", func(robot bot.Bot) error {
		msg := &bot.Message{GroupID: toInt(config.GroupId)}
		robot.SendMsg(msg, picture.Url())
		robot.SendMsg(msg, "每日一涩图~")
		return nil
	}).At("16:15")
}
