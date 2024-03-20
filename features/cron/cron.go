package cron

import (
	"qq/bot"
	"qq/cronjob"
	"qq/features"
	"strings"
)

func init() {
	features.AddKeyword("list-cron", "列出所有定时任务", func(bot bot.Bot, content string) error {
		s := strings.Builder{}
		for _, v := range cronjob.Manager().List() {
			s.WriteString(v.Name() + "\n")
		}
		bot.Send(s.String())
		return nil
	}, features.WithGroup("cron"))
}
