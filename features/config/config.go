package config

import (
	"qq/bot"
	"qq/config"
	"qq/features"
	"strings"
)

func init() {
	features.AddKeyword("cs", "设置环境变量, ex: ai_token=xxx,GroupId=xxx", func(bot bot.Bot, content string) error {
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
		bot.Send(config.Configs().String())
		return nil
	}, features.WithSysCmd(), features.WithHidden())
}
