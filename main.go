package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"qq/bot"
	"qq/config"
	"qq/cronjob"
	"qq/features"
	"qq/features/webot"
	"qq/util"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	// _ "qq/cronjob/lifetip"
	//_ "qq/cronjob/kfc"
	_ "qq/cronjob/autoupdate"
	_ "qq/cronjob/ddys"
	_ "qq/cronjob/lpr"

	// _ "qq/cronjob/btc"
	_ "qq/cronjob/comic"
	_ "qq/cronjob/dx"
	_ "qq/cronjob/lottery"
	_ "qq/cronjob/maotai"

	//_ "qq/cronjob/zaoan"

	//_ "qq/cronjob/picture"
	_ "qq/cronjob/bitget"
	_ "qq/cronjob/goodmorning"

	//_ "qq/cronjob/trainticketleft"
	_ "qq/cronjob/xiaofeiquan"

	_ "qq/features/ai"
	_ "qq/features/bililottery"
	_ "qq/features/bitget"
	_ "qq/features/comic"
	_ "qq/features/config"
	_ "qq/features/daxin"
	_ "qq/features/ddys"
	_ "qq/features/geo"
	_ "qq/features/goodmorning"
	_ "qq/features/googlesearch"
	_ "qq/features/help"
	_ "qq/features/holiday"
	_ "qq/features/imaotai"
	_ "qq/features/jin10"
	_ "qq/features/kfc"
	_ "qq/features/lifetip"
	_ "qq/features/lpr"
	_ "qq/features/picture"
	_ "qq/features/pixiv"
	_ "qq/features/raokouling"
	_ "qq/features/star"
	_ "qq/features/stock"
	_ "qq/features/sysupdate"
	_ "qq/features/task"
	_ "qq/features/telegram"
	_ "qq/features/trainticket"
	_ "qq/features/version"
	_ "qq/features/weather"
	_ "qq/features/webot"
	_ "qq/features/weibo"
	_ "qq/features/zhihu"
	_ "qq/features/zuan"
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
	cm.LoadOnceTasks()
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
			fmt.Printf("key: %v, content: %v\n", keyword, content)
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

	brithCry()
	log.Println("[HTTP]: start...")
	log.Println(http.ListenAndServe(":5701", nil))
}

func brithCry() {
	go func() {
		time.Sleep(15 * time.Second)
		for _, s := range config.AdminIDs().List() {
			bot.NewQQBot(&bot.Message{}).SendToUser(s, fmt.Sprintf("%s 系统已启动", time.Now().Format(time.DateTime)))
			webot.Run(bot.NewQQBot(&bot.Message{SenderUserID: config.UserID()}))
		}
	}()
}

func printREADME() {
	fmt.Printf(mdTemplate, features.BeautifulOutput(true, false))
}

var mdTemplate = `
# QQ Bot

[![build-docker](https://github.com/Lick-Dog-Club/qq-bot/actions/workflows/build.yaml/badge.svg)](https://github.com/Lick-Dog-Club/qq-bot/actions/workflows/build.yaml)

> qq 机器人
>
> 详细咨询请加 QQ 1025434218, 注明来源

## 指令 (y代表可ai交互的指令)

` + "```text\n%s\n```" + `

## Example

![识图+画图](./images/seedraw.jpg)

![车票+热搜](./images/2.png)
`
