package config

import (
	"bufio"
	"fmt"
	"qq/bot"
	"qq/config"
	"qq/features"
	"strings"

	log "github.com/sirupsen/logrus"
)

func init() {
	features.AddKeyword("cs", "设置环境变量, ex: ai_token=xxx, 多个用换行隔开", func(bot bot.Bot, content string) error {
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
			bot.Send(fmt.Sprintf("%s", scanner.Err()))
		}

		bot.Send("已设置: \n" + config.Set(conf).String())
		return nil
	}, features.WithSysCmd(), features.WithHidden(), features.WithGroup("config"))
	features.AddKeyword("cg", "<+key|[keys: 全部keys]>显示环境变量", func(bot bot.Bot, content string) error {
		if bot.IsFromAdmin() {
			if content == "keys" {
				var keys []string
				for s, _ := range config.Configs() {
					keys = append(keys, s)
				}
				bot.Send(strings.Join(keys, "\n"))
				return nil
			}

			if content != "" {
				bot.Send(config.Configs()[content])
				return nil
			}

			path, err := bot.SendTextImage(config.Configs().String())
			if err != nil {
				log.Println(err)
			}
			fmt.Println(path)
			return nil
		}
		bot.Send("未授权")
		return nil
	}, features.WithSysCmd(), features.WithHidden(), features.WithGroup("config"))
	features.AddKeyword("cgall", "显示环境变量", func(bot bot.Bot, content string) error {
		if bot.IsFromAdmin() {
			if _, err := bot.SendTextImage(config.Configs().String()); err != nil {
				log.Println(err)
			}
			return nil
		}
		bot.Send("未授权")
		return nil
	}, features.WithSysCmd(), features.WithHidden(), features.WithGroup("config"))
}
