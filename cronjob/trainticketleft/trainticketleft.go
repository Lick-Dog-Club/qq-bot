package trainticketleft

import (
	"qq/bot"
	"qq/cronjob"
)

func init() {
	cronjob.Manager().NewCommand("trainticketleft", func(bot bot.CronBot) error {
		//trainticket.Search()
		//bot.SendToUser(config.UserID(), "")
		return nil
	}).EveryFiveSeconds()
}
