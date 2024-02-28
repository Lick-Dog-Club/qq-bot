package task

import (
	"encoding/json"
	"fmt"
	"qq/bot"
	"qq/config"
	"qq/cronjob"
	"qq/features"
	"qq/util"
	"qq/util/random"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai/jsonschema"

	"github.com/samber/lo"

	"github.com/duc-cnzj/when-rules/zh"
	"github.com/olebedev/when"
	"github.com/olebedev/when/rules/common"
	log "github.com/sirupsen/logrus"
)

var w = when.New(nil)

func init() {
	w.Add(zh.All...)
	w.Add(common.All...)

	features.AddKeyword("listtask", "任务列表", func(bot bot.Bot, content string) error {
		bot.Send(cronjob.Manager().ListOnceCommands())
		return nil
	}, features.WithGroup("task"))
	features.AddKeyword("canceltask", "取消任务", func(bot bot.Bot, content string) error {
		find, b := lo.Find(config.Tasks(), func(item config.Task) bool {
			return item.Name == content
		})
		if b {
			filter := lo.Filter(config.Tasks(), func(item config.Task, index int) bool {
				return item.ID != find.ID
			})
			config.SyncTasks(filter)
			cronjob.Manager().RemoveOnceCommand(int(find.ID))
			bot.Send("已取消")
		} else {
			bot.Send("未找到任务")
		}
		return nil
	}, features.WithGroup("task"))
	features.AddKeyword("task", "<+content: 具体内容> 添加一次性的任务/提醒事项", func(b bot.Bot, content string) error {
		parse, err := w.Parse(content, time.Now())
		if err != nil {
			b.Send(err.Error())
			return err
		}
		if parse == nil {
			b.Send("解析失败：" + content)
			return nil
		}
		log.Println(parse.Time.String(), parse)

		after := strings.SplitAfter(content, parse.Text)
		cc := content
		if len(after) == 2 {
			cc = after[1]
		}
		b.Send(AddTask(parse.Time, cc, b))
		return nil
	}, features.WithGroup("task"), features.WithAIFunc(features.AIFuncDef{
		Properties: map[string]jsonschema.Definition{
			"time": {
				Type:        jsonschema.String,
				Description: "执行的时间，格式为 `2006-01-02 15:04:05`",
			},
			"content": {
				Type:        jsonschema.String,
				Description: "用户给的完整的内容",
			},
			"uid": {
				Type:        jsonschema.String,
				Description: "用户的 UID",
			},
			"gid": {
				Type:        jsonschema.String,
				Description: "群的 GID",
			},
		},
		Call: func(args string) (string, error) {
			var s = struct {
				Time    string `json:"time"`
				Content string `json:"content"`
				UID     string `json:"uid"`
				GID     string `json:"gid"`
			}{}

			json.Unmarshal([]byte(args), &s)
			parse, _ := time.Parse(time.DateTime, s.Time)

			return AddTask(parse, s.Content, bot.NewQQBot(&bot.Message{
				SenderUserID:  s.UID,
				IsSendByGroup: s.GID != "",
				GroupID:       s.GID,
			})), nil
		},
	}))
}

func AddTask(t time.Time, c string, b bot.Bot) string {
	var result string
	var tid int
	name := random.String(20)
	if k, v := util.GetKeywordAndContent(c); features.Match(k) {
		tid = cronjob.Manager().NewOnceCommand(name, t, func(bot.Bot) error {
			features.Run(b, k, v)
			return nil
		})
		result = fmt.Sprintf("已设置:\n时间: %s, 命令: %s\n取消任务请执行: canceltask %v", t.Format(time.DateTime), k, name)
	} else {
		tid = cronjob.Manager().NewOnceCommand(name, t, func(bot.Bot) error {
			b.Send(c)
			return nil
		})

		result = fmt.Sprintf("已设置:\n时间: %s, 内容: %s\n取消任务请执行: canceltask %v", t.Format(time.DateTime), c, name)
	}

	var uid, gid string
	if b.IsGroupMessage() {
		gid = b.UserID()
	} else {
		uid = b.UserID()
	}
	res := config.Tasks()
	res = append(res, config.Task{
		ID:      tid,
		Name:    name,
		RunAt:   t.Format(time.DateTime),
		Content: c,
		UserID:  uid,
		GroupID: gid,
	})

	config.SyncTasks(res)
	return result
}
