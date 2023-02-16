package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"qq/bot"
	"qq/features"
	"strings"

	_ "qq/features/ai"
	_ "qq/features/bili-lottery"
	_ "qq/features/help"
	_ "qq/features/picture"
	_ "qq/features/sys-update"
	_ "qq/features/version"
	_ "qq/features/weather"
	_ "qq/features/weibo"
	_ "qq/features/zhihu"
)

func main() {
	features.SetNewBotFunc(bot.NewDummyBot)
	features.Run(bot.Message{}, "help", "")
	return
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
			keyword, content := getKeywordAndContent(msg)
			if err := features.Run(message, keyword, content); err != nil {
				log.Println(err)
			}
		}
	})

	log.Println("start...")
	log.Println(http.ListenAndServe(":5701", nil))
}

func getKeywordAndContent(msg string) (string, string) {
	msg = trimSpace(msg)
	split := strings.SplitN(msg, " ", 2)
	if len(split) == 2 {
		return strings.ToLower(split[0]), trimSpace(split[1])
	}

	return msg, ""
}

func trimSpace(s string) string {
	return strings.TrimSpace(s)
}
