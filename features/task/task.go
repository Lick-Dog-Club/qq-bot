package task

import (
	"fmt"
	"qq/bot"
	"qq/config"
	"qq/cronjob"
	"qq/features"
	"qq/util"
	"strconv"
	"strings"
	"time"

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
		atoi, err := strconv.Atoi(strings.TrimSpace(content))
		if err != nil {
			bot.Send(err.Error())
			return err
		}
		filter := lo.Filter(config.Tasks(), func(item config.Task, index int) bool {
			return item.ID != atoi
		})
		config.SyncTasks(filter)
		cronjob.Manager().RemoveOnceCommand(int(atoi))
		bot.Send("已取消")
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

		var tid int
		after := strings.SplitAfter(content, parse.Text)
		cc := content
		if len(after) == 2 {
			if k, v := util.GetKeywordAndContent(after[1]); features.Match(k) {
				cc = after[1]
				tid = cronjob.Manager().NewOnceCommand(content, parse.Time, func(bot.Bot) error {
					features.Run(b, k, v)
					return nil
				})
				b.Send(fmt.Sprintf("已设置:\n时间: %s, 命令: %s\n取消任务请执行: canceltask %d", parse.Time.Format(time.DateTime), k, tid))
				return nil
			}
		}
		tid = cronjob.Manager().NewOnceCommand(content, parse.Time, func(bot.Bot) error {
			b.Send(content)
			return nil
		})
		var uid, gid string
		if b.IsGroupMessage() {
			gid = b.UserID()
		} else {
			uid = b.UserID()
		}
		res := config.Tasks()
		res = append(res, config.Task{
			ID:      tid,
			RunAt:   parse.Time.Format(time.DateTime),
			Content: cc,
			UserID:  uid,
			GroupID: gid,
		})

		config.SyncTasks(res)

		b.Send(fmt.Sprintf("已设置:\n时间: %s, 内容: %s\n取消任务请执行: canceltask %d", parse.Time.Format(time.DateTime), content, tid))
		return nil
	}, features.WithGroup("task"))
}
