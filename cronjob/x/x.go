package x

import (
	"context"
	"qq/bot"
	"qq/config"
	"qq/cronjob"
	"qq/features/x"
)

func init() {
	cronjob.Manager().NewCommand("x-users", func(bot bot.CronBot) error {
		m := x.NewManager(config.XTokens(), config.HttpProxy())
		for _, s := range config.XUsers() {
			tweets, err := m.GetTweets(context.TODO(), s, 1)
			if err != nil {
				bot.SendGroup(config.XGroupID(), err.Error())
			}
			for _, tweet := range tweets {
				bot.SendGroup(config.XGroupID(), x.RenderTweetResult(tweet))
			}
		}
		return nil
	}).EveryMinute()
}
