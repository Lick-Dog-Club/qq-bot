package main

import (
	"encoding/json"
	"fmt"
	"github.com/lithammer/dedent"
	"log"
	"net/http"
	"qq/ai"
	lottery "qq/bili-lottery"
	"qq/bot"
	"strings"
)

func main() {
	//log.Fatal(request(chat, "你是谁?"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var message bot.Message
		json.NewDecoder(r.Body).Decode(&message)
		if message.PostType == "meta_event" {
			return
		}
		log.Printf("receive %#v\n", message)
		atMsg := fmt.Sprintf("[CQ:at,qq=%v]", message.SelfID)
		if strings.Contains(message.Message, atMsg) && message.GroupID > 0 {
			msg := strings.ReplaceAll(message.Message, atMsg, "")
			switch {
			case isKeyword(msg, "抽奖"):
				cookie := content(msg, "抽奖")
				bot.Send(message, lottery.Run(message, cookie).String())
			case isKeyword(msg, "help"):
				bot.Send(message, dedent.Dedent(`
					@bot 抽奖 <bilibili-cookie>: 自动转发up主的抽奖活动
					@bot help: 帮助界面
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
