package comic

import (
	"fmt"
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
			jpegPaths := c.ToJPEG()
			for p := range jpegPaths {
				bot.SendGroup(config.GroupID(), fmt.Sprintf("[CQ:image,file=file://%s]", p))
				//os.Remove(p)
			}
		}
		return nil
	}).DailyAt("12:00")
}
