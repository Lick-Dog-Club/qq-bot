package lottery

import (
	"fmt"
	"log"
	"qq/bot"
	"strings"
)

func Run(message bot.Message, cookie string) fmt.Stringer {
	user := User{
		cookie: cookiePair(cookie),
	}
	user.Me()
	bot.Send(message, user.me.Data.Uname+" 登录成功，现在开始处理抽奖请求~")
	user.forwards = user.MyForwards(user.me.Data.Mid)
	var results NoticeBodyList
	for _, in := range user.LotteryDynamics() {
		if !in.Past && !in.Forwarded {
			user.DynaRepost(int64(in.DynamicId), "拉低中奖率~")
			results = append(results, in)
			log.Println("已转发: ", in.WebUrl)
		}
	}
	return results
}

type Pair map[string]string

func cookiePair(raw string) Pair {
	var res = Pair{}
	splits := strings.Split(raw, "; ")
	for _, split := range splits {
		kvs := strings.Split(split, "=")
		if len(kvs) == 2 {
			res[kvs[0]] = kvs[1]
		}
	}
	return res
}
