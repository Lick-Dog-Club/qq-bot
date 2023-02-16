package maotai

import (
	"qq/bot"
	"qq/cronjob"
)

func init() {
	cronjob.Manager().NewCommand("maotai", func(bot bot.Bot) error {
		bot.Send("开始申购茅台了~")
		return nil
	}).DailyAt("9:00")
}
