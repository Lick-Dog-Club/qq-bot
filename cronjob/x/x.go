package x

import (
	"context"
	"qq/bot"
	"qq/config"
	"qq/cronjob"
	"qq/features/x"
	"time"

	log "github.com/sirupsen/logrus"
)

var lastTime time.Time

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
		for _, s := range config.XUsers() {
			tweets, err := m.GetTweets(context.TODO(), s, 1)
			if err != nil {
				bot.SendGroup(config.XGroupID(), err.Error())
			}
			for _, tweet := range tweets {
				func() {
					if tweet.TimeParsed.Before(lastTime) {
						return
					}
					result, f := x.RenderTweetResult(tweet)
					defer f()
					bot.SendGroup(config.XGroupID(), result)
				}()
			}
		}
		return nil
	}).EveryMinute()
}
