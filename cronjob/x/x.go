package x

import (
	"context"
	"fmt"
	"qq/bot"
	"qq/config"
	"qq/cronjob"
	"qq/features/x"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var lastTime time.Time = time.Now()

func init() {
	cronjob.Manager().NewCommand("x-users", func(bot bot.CronBot) error {
		if config.XGroupID() == "" || len(config.XUsers()) == 0 {
			log.Println("XGroupID/x-users not set, skip")
			return nil
		}
		m := x.NewManager(config.XTokens(), config.HttpProxy())
		startAt := time.Now()
		defer func() {
			lastTime = startAt
			log.Println("x-users done", lastTime)
		}()
		res := strings.Builder{}
		var e = x.NewAggregateError()
		for _, s := range config.XUsers() {
			tweets, err := m.GetTweets(context.TODO(), s, 1)
			if err != nil {
				e.Add(fmt.Errorf("get tweets for %s: %w", s, err))
			}
			for _, tweet := range tweets {
				func() {
					if tweet.TimeParsed.Before(lastTime) {
						log.Println("skip old tweet", tweet.TimeParsed, tweet.PermanentURL, tweet.Username)
						return
					}
					result, f := x.RenderTweetResult(tweet)
					defer f()
					if result != "" {
						res.WriteString(result + "\n")
					}
				}()
			}
		}
		r := res.String()
		if r != "" {
			bot.SendGroup(config.XGroupID(), r)
			bot.SendToUser(config.UserID(), r)
		}
		if e.ToError() != nil {
			bot.SendToUser(config.UserID(), e.ToError().Error())
		}

		return nil
	}).EveryMinute()
}
