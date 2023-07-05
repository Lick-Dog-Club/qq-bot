package main

import (
	"context"
	"encoding/json"
	"flag"
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
	_ "qq/cronjob/btc"
	_ "qq/cronjob/ddys"
	_ "qq/cronjob/lottery"
	_ "qq/cronjob/maotai"
	_ "qq/cronjob/picture"
	_ "qq/cronjob/tianqi"
	_ "qq/cronjob/xiaofeiquan"

	_ "qq/features/ai"
	_ "qq/features/bili-lottery"
	_ "qq/features/config"
	_ "qq/features/daxin"
	_ "qq/features/ddys"
	_ "qq/features/help"
	_ "qq/features/imaotai"
	_ "qq/features/kfc"
	_ "qq/features/lifetip"
	_ "qq/features/picture"
	_ "qq/features/pixiv"
	_ "qq/features/raokouling"
	_ "qq/features/sysupdate"
	_ "qq/features/task"
	_ "qq/features/version"
	_ "qq/features/weather"
	_ "qq/features/webot"
	_ "qq/features/weibo"
	_ "qq/features/zhihu"
)

var genDoc bool

func init() {
	log.SetReportCaller(true)
	flag.BoolVar(&genDoc, "doc", false, "-doc")
}

func main() {
	flag.Parse()
	if genDoc {
		printREADME()
		return
	}

	cm := cronjob.Manager()
	cm.Run(context.TODO())
	defer cm.Shutdown(context.TODO())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var message *bot.QQMessage
		json.NewDecoder(r.Body).Decode(&message)
		if message.PostType == "meta_event" {
			return
		}
		fmt.Printf("receive %d: %v\n", message.UserID, message.Message)
		atMsg := fmt.Sprintf("[CQ:at,qq=%v]", message.SelfID)
		if (strings.Contains(message.Message, atMsg) && message.MessageType == "group") || message.MessageType == "private" {
			msg := strings.ReplaceAll(message.Message, atMsg, "")
			keyword, content := util.GetKeywordAndContent(msg)
			var (
				gid string
				sid string
			)
			if message.UserID > 0 {
				sid = fmt.Sprintf("%d", message.UserID)
			}
			if message.GroupID > 0 {
				gid = fmt.Sprintf("%d", message.GroupID)
			}
			if err := features.Run(bot.NewQQBot(&bot.Message{
				SenderUserID:  sid,
				Message:       content,
				IsSendByGroup: message.MessageType == "group",
				GroupID:       gid,
			}), keyword, content); err != nil {
				log.Println(err)
			}
		}
	})

	log.Println("[HTTP]: start...")
	log.Println(http.ListenAndServe(":5701", nil))
}

func printREADME() {
	fmt.Println(fmt.Sprintf(mdTemplate, features.BeautifulOutput(true)))
}

var mdTemplate = `
# QQ-bot

[![build-docker](https://github.com/Lick-Dog-Club/qq-bot/actions/workflows/build.yaml/badge.svg)](https://github.com/Lick-Dog-Club/qq-bot/actions/workflows/build.yaml)

> qq 机器人

## 指令

` + "```text\n%s```"
