package x

import (
	"context"
	"fmt"
	"qq/bot"
	"qq/config"
	"qq/cronjob"
	"qq/features/x"
	"qq/util"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var lastTime time.Time = time.Now()

func init() {
	cronjob.NewCommand("x-users", func(bot bot.CronBot) error {
		if config.XGroupID() == "" || len(config.XUsers()) == 0 {
			log.Println("XGroupID/x-users not set, skip")
			return nil
		}
		m := x.NewManager(config.XTokens(), config.HttpProxy())
		defer func() {
			lastTime = time.Now()
			log.Println("x-users done", lastTime)
		}()
		res := strings.Builder{}
		var fns []func()
		var e = x.NewAggregateError()
		for _, s := range config.XUsers() {
			tweets, err := m.GetTweets(context.TODO(), s, 1)
			if err != nil {
				e.Add(fmt.Errorf("get tweets for %s: %w", s, err))
				continue
			}
			for _, tweet := range tweets {
				func() {
					if tweet.TimeParsed.Local().Before(lastTime) {
						log.Println("skip old tweet", tweet.TimeParsed.Local().Format(time.DateTime), tweet.PermanentURL, tweet.Name)
						return
					}
					result, f := x.RenderTweetResult(tweet)
					fns = append(fns, f)
					if result != "" {
						res.WriteString(result + "\n")
					}
				}()
			}
		}
		r := res.String()
		if strings.Contains(r, "MUMU") || strings.Contains(r, "mumu") {
			util.Bark("MUMU", "MUMU📈", config.BarkUrls()...)
		}
		if r != "" {
			bot.SendGroup(config.XGroupID(), r)
			bot.SendToUser(config.UserID(), r)
		}
		if e.ToError() != nil {
			bot.SendToUser(config.UserID(), e.ToError().Error())
		}

		for _, f := range fns {
			f()
		}
		return nil
	}).EveryMinute()
}
