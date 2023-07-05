package config

import (
	"qq/bot"
	"qq/config"
	"qq/features"
	"strings"
)

func init() {
	features.AddKeyword("cs", "设置环境变量, ex: ai_token=xxx,group_id=xxx", func(bot bot.Bot, content string) error {
		if content == "me" {
			config.Set(map[string]string{"user_id": bot.UserID()})
			bot.Send("已设置: user_id=" + bot.UserID())
			return nil
		}
		var conf = map[string]string{}
		split := strings.Split(content, ",")
		for _, s := range split {
			n := strings.SplitN(s, "=", 2)
			if len(n) == 2 {
				conf[n[0]] = n[1]
			}
		}
		config.Set(conf)
		bot.Send("已设置: " + content)
		return nil
	}, features.WithSysCmd(), features.WithHidden())
	features.AddKeyword("cg", "显示环境变量", func(bot bot.Bot, content string) error {
		if bot.UserID() == config.UserID() {
			bot.Send(config.Configs().String())
		}
		bot.Send("未授权")
		return nil
	}, features.WithSysCmd(), features.WithHidden())
}
