package config

import (
	"bufio"
	"fmt"
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

		scanner := bufio.NewScanner(strings.NewReader(content))
		for scanner.Scan() {
			n := strings.SplitN(scanner.Text(), "=", 2)
			if len(n) == 2 {
				conf[n[0]] = n[1]
			}
		}
		if scanner.Err() != nil {
			bot.Send(fmt.Sprintf("%e", scanner.Err()))
		}

		config.Set(conf)
		bot.Send("已设置: \n" + config.Configs().String())
		return nil
	}, features.WithSysCmd(), features.WithHidden(), features.WithGroup("config"))
	features.AddKeyword("cg", "显示环境变量", func(bot bot.Bot, content string) error {
		if bot.UserID() == config.UserID() {
			bot.Send(config.Configs().String())
			return nil
		}
		bot.Send("未授权")
		return nil
	}, features.WithSysCmd(), features.WithHidden(), features.WithGroup("config"))
}
