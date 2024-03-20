package lottery

import (
	"encoding/json"
	"qq/bot"
	"qq/config"
	"qq/features"
	"strings"

	"github.com/sashabaranov/go-openai/jsonschema"

	log "github.com/sirupsen/logrus"
)

func init() {
	features.AddKeyword("bili-lottery", "<+bilibili-cookie> bilibili 抽奖, 自动转发up主的抽奖活动", func(bot bot.Bot, content string) error {
		bot.Send(Run(func(s string) { bot.Send(s) }, content))
		return nil
	}, features.WithAIFunc(features.AIFuncDef{
		Properties: map[string]jsonschema.Definition{
			"cookie": {
				Type:        jsonschema.String,
				Description: "用户的 cookie 值",
			},
		},
		Call: func(args string) (string, error) {
			var input = struct {
				Cookie string `json:"cookie"`
			}{}
			json.Unmarshal([]byte(args), &input)
			str := ""
			Run(func(s string) {
				str += s
			}, input.Cookie)
			return str, nil
		},
	}))
}

func Run(send func(string), cookie string) string {
	user := User{
		cookie: cookiePair(cookie),
	}
	_, err := user.info()
	if err != nil {
		// 登录失败清除 cookie
		config.Set(map[string]string{"bili_cookie": ""})
		return err.Error()
	}
	send(user.me.Data.Uname + " 登录成功，现在开始处理抽奖请求~")
	user.forwards = user.myForwards(user.me.Data.Mid)
	var results noticeBodyList
	for _, in := range user.lotteryDynamics() {
		if !in.Past && !in.Forwarded {
			user.dynaRepost(int64(in.DynamicId), "拉低中奖率~")
			results = append(results, in)
			log.Println("已转发: ", in.WebUrl)
		}
	}
	return results.String()
}

type cookiePairs map[string]string

func cookiePair(raw string) cookiePairs {
	var res = cookiePairs{}
	splits := strings.Split(raw, "; ")
	for _, split := range splits {
		kvs := strings.Split(split, "=")
		if len(kvs) == 2 {
			res[kvs[0]] = kvs[1]
		}
	}
	return res
}
