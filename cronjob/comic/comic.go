package comic

import (
	"fmt"
	"os"
	"qq/bot"
	"qq/config"
	"qq/cronjob"
	"qq/features/comic"
)

func init() {
	cronjob.Manager().NewCommand("haizeiwang", func(bot bot.CronBot) error {
		c := comic.Get("haizeiwang")
		if c.TodayUpdated() {
			bot.SendGroup(config.GroupID(), c.Render())
			jpegPath := c.ToJPEG()
			bot.SendGroup(config.GroupID(), fmt.Sprintf("[CQ:image,file=file://%s]", jpegPath))
			os.Remove(jpegPath)
		}
		return nil
	}).DailyAt("12:00")
}
