package autoupdate

import (
	"qq/bot"
	"qq/config"
	"qq/cronjob"
	"qq/features/sysupdate"
	"strings"
)

type upBot struct {
	bot bot.CronBot
	uid string
}

func (b *upBot) Send(s string) string {
	if b.uid != "" && strings.HasPrefix(s, "更新到最新版") {
		return b.bot.SendToUser(b.uid, s)
	}
	return ""
}

func newUpBot(bot bot.CronBot, uid string) *upBot {
	return &upBot{bot: bot, uid: uid}
}

func init() {
	cronjob.Manager().NewCommand("auto-update", func(bot bot.CronBot) error {
		sysupdate.UpdateVersion(newUpBot(bot, config.UserID()))
		return nil
	}).EveryFifteenMinutes()
}
