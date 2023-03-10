package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"qq/bot"
	"qq/cronjob"
	"qq/features"
	"strings"

	log "github.com/sirupsen/logrus"

	// _ "qq/cronjob/lifetip"
	_ "qq/cronjob/dx"
	_ "qq/cronjob/kfc"
	_ "qq/cronjob/maotai"

	//_ "qq/cronjob/picture"
	_ "qq/cronjob/tianqi"
	_ "qq/cronjob/xiaofeiquan"

	_ "qq/features/ai"
	_ "qq/features/bili-lottery"
	_ "qq/features/config"
	_ "qq/features/daxin"
	_ "qq/features/help"
	_ "qq/features/kfc"
	_ "qq/features/lifetip"
	_ "qq/features/picture"
	_ "qq/features/pixiv"
	_ "qq/features/raokouling"
	_ "qq/features/sys-update"
	_ "qq/features/version"
	_ "qq/features/weather"
	_ "qq/features/weibo"
	_ "qq/features/zhihu"
)

func init() {
	log.SetReportCaller(true)
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var message *bot.Message
		json.NewDecoder(r.Body).Decode(&message)
		if message.PostType == "meta_event" {
			return
		}
		fmt.Printf("receive %d: %v\n", message.UserID, message.Message)
		atMsg := fmt.Sprintf("[CQ:at,qq=%v]", message.SelfID)
		if (strings.Contains(message.Message, atMsg) && message.MessageType == "group") || message.MessageType == "private" {
			msg := strings.ReplaceAll(message.Message, atMsg, "")
			keyword, content := getKeywordAndContent(msg)
			if err := features.Run(message, keyword, content); err != nil {
				log.Println(err)
			}
		}
	})
	cm := cronjob.Manager()
	cm.Run(context.TODO())
	defer cm.Shutdown(context.TODO())

	log.Println("[HTTP]: start...")
	log.Println(http.ListenAndServe(":5701", nil))
}

func getKeywordAndContent(msg string) (string, string) {
	msg = trimSpace(msg)
	split := strings.SplitN(msg, " ", 2)
	if len(split) == 2 {
		return strings.ToLower(split[0]), trimSpace(split[1])
	}

	return strings.ToLower(msg), ""
}

func trimSpace(s string) string {
	return strings.TrimSpace(s)
}
