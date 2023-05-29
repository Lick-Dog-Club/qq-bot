package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"qq/bot"
	"qq/cronjob"
	"qq/features"
	"qq/util"
	"strings"

	log "github.com/sirupsen/logrus"

	// _ "qq/cronjob/lifetip"
	_ "qq/cronjob/dx"
	//_ "qq/cronjob/kfc"
	_ "qq/cronjob/lottery"
	_ "qq/cronjob/maotai"

	_ "qq/cronjob/picture"
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
	_ "qq/features/sysupdate"
	_ "qq/features/task"
	_ "qq/features/version"
	_ "qq/features/weather"
	_ "qq/features/weibo"
	_ "qq/features/zhihu"
)

func init() {
	log.SetReportCaller(true)
}

func main() {
	cm := cronjob.Manager()
	cm.Run(context.TODO())
	defer cm.Shutdown(context.TODO())
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
			keyword, content := util.GetKeywordAndContent(msg)
			if err := features.Run(message, keyword, content); err != nil {
				log.Println(err)
			}
		}
	})

	log.Println("[HTTP]: start...")
	log.Println(http.ListenAndServe(":5701", nil))
}
