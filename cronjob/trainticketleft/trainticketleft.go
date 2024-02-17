package trainticketleft

import (
	"qq/bot"
	"qq/config"
	"qq/cronjob"
	"qq/features/trainticket"
	"qq/util"
	"time"
)

func init() {
	cronjob.Manager().NewCommand("trainticketleft-taimushan-shangzhoudong", func(bot bot.CronBot) error {
		parse, _ := time.Parse("2006-01-02", "2024-02-18")
		if time.Now().After(parse) {
			return nil
		}
		var date = []string{"2024-02-17", "2024-02-18"}
		for _, d := range date {
			search := trainticket.Search(trainticket.SearchInput{
				From:           trainticket.GetStationCode("太姥山"),
				To:             trainticket.GetStationCode("杭州东"),
				Date:           d,
				OnlyShowTicket: true,
			})
			if len(search) > 0 {
				util.Bark("有车票", search.String(), config.BarkUrls()...)
				bot.SendToUser(config.UserID(), search.String())
			}
		}
		return nil
	}).EveryFifteenSeconds()
}
