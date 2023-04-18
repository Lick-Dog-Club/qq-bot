package help

import (
	"qq/bot"
	"qq/features"
	"qq/features/sysupdate"
)

func init() {
	features.AddKeyword("version", "系统版本", func(sender bot.Bot, content string) error {
		sender.Send(sysupdate.Version())
		return nil
	}, features.WithSysCmd(), features.WithHidden())
}
