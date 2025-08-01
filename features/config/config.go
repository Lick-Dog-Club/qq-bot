package config

import (
	"bufio"
	"fmt"
	"qq/bot"
	"qq/config"
	"qq/features"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
)

func init() {
	features.AddKeyword("cs-help", "", func(bot bot.Bot, content string) error {
		help := config.GetHelp(content)
		if help == "" {
			bot.Send("未找到帮助信息")
			return nil
		}
		bot.Send(help)
		return nil
	}, features.WithSysCmd(), features.WithHidden(), features.WithGroup("config"))
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
	}, features.MustAdminRole(), features.WithSysCmd(), features.WithHidden(), features.WithGroup("config"))
	features.AddKeyword("cg", "<+key|[keys: 全部keys]>显示环境变量", func(bot bot.Bot, content string) error {
		if content == "keys" {
			var keys []string
			for s := range config.Configs() {
				keys = append(keys, s)
			}
			sort.Strings(keys)
			bot.SendTextImage(strings.Join(keys, "\n"))
			return nil
		}

		if content != "" {
			bot.Send(fmt.Sprintf("%s=%s", content, config.Configs()[content]))
			return nil
		}

		path, err := bot.SendTextImage(config.Configs().String())
		if err != nil {
			log.Println(err)
		}
		fmt.Println(path)
		return nil
	}, features.WithSysCmd(), features.WithHidden(), features.WithGroup("config"), features.MustAdminRole())
	features.AddKeyword("cgall", "显示环境变量", func(bot bot.Bot, content string) error {
		bot.Send(config.Configs().String())
		return nil
	}, features.MustAdminRole(), features.WithSysCmd(), features.WithHidden(), features.WithGroup("config"))
}
