package help

import (
	"qq/bot"
	"qq/features"
	"qq/features/sysupdate"
)

func init() {
	features.AddKeyword("version", "获取系统版本", func(sender bot.Bot, content string) error {
		sender.Send(sysupdate.Version())
		return nil
	}, features.WithSysCmd(), features.WithHidden(), features.WithAIFunc(features.AIFuncDef{
		Call: func(args string) (string, error) {
			return sysupdate.Version(), nil
		},
	}))
}
