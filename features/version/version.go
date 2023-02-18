package help

import (
	"qq/bot"
	"qq/features"
	sys_update "qq/features/sys-update"
)

func init() {
	features.AddKeyword("version", "系统版本", func(sender bot.Bot, content string) error {
		sender.Send(sys_update.Version())
		return nil
	}, features.WithSysCmd(), features.WithHidden())
}
