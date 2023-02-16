package lottery

import (
	"log"
	"qq/bot"
	"qq/features"
	"strings"
)

func init() {
	features.AddKeyword("抽奖", "+<bilibili-cookie> 自动转发up主的抽奖活动", func(bot bot.Bot, content string) error {
		bot.Send(run(bot, content))
		return nil
	})
}

func run(bot bot.Bot, cookie string) string {
	user := User{
		cookie: cookiePair(cookie),
	}
	user.info()
	bot.Send(user.me.Data.Uname + " 登录成功，现在开始处理抽奖请求~")
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
