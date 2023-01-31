package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"qq/ai"
	lottery "qq/bili-lottery"
	"qq/bot"
	"qq/picture"
	"qq/weather"
	"strings"
	"time"

	"github.com/lithammer/dedent"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var message bot.Message
		json.NewDecoder(r.Body).Decode(&message)
		if message.PostType == "meta_event" {
			return
		}
		log.Printf("receive %#v\n", message)
		atMsg := fmt.Sprintf("[CQ:at,qq=%v]", message.SelfID)
		if (strings.Contains(message.Message, atMsg) && message.MessageType == "group") || message.MessageType == "private" {
			msg := strings.ReplaceAll(message.Message, atMsg, "")
			switch {
			case isKeyword(msg, "涩图"):
				msgID := bot.Send(message, picture.Url())
				go func() {
					bot.Send(message, "图片即将在 30s 之后撤回，要保存的赶紧了~")
					time.Sleep(30 * time.Second)
					bot.DeleteMsg(msgID)
				}()
			case isKeyword(msg, "天气"):
				city := content(msg, "天气")
				if city == "" {
					city = "杭州"
				}
				bot.Send(message, weather.Get(city))
			case isKeyword(msg, "抽奖"):
				cookie := content(msg, "抽奖")
				bot.Send(message, lottery.Run(message, cookie).String())
			case isKeyword(msg, "help"):
				bot.Send(message, dedent.Dedent(`
					@bot 抽奖 <bilibili-cookie>: 自动转发up主的抽奖活动
					@bot 涩图: 返回动漫图片~
					@bot help: 帮助界面
					@bot 天气 <城市: 默认杭州>: 查询城市天气
					@bot default: ai 自动回答
				`))
			default:
				bot.Send(message, ai.Request(msg))
			}
		}
	})
	log.Println("start...")
	log.Println(http.ListenAndServe(":5701", nil))
}

func isKeyword(msg, k string) bool {
	split := strings.Split(trimSpace(msg), " ")
	if len(split) > 0 {
		return split[0] == k
	}
	return false
}

func content(msg, k string) string {
	return trimSpace(strings.TrimPrefix(trimSpace(msg), k))
}

func trimSpace(s string) string {
	return strings.TrimSpace(s)
}
