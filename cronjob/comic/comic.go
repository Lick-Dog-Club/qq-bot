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
	cronjob.NewCommand("haizeiwang", func(bot bot.CronBot) error {
		c := comic.Get("haizeiwang", -1)
		if c.TodayUpdated() {
			bot.SendGroup(config.GroupID(), c.Render())
			jpegPaths := c.ToJPEG()
			for p := range jpegPaths {
				bot.SendGroup(config.GroupID(), fmt.Sprintf("[CQ:image,file=file://%s]", p))
				os.Remove(p)
			}
		}
		return nil
	}).DailyAt("12:00")
}
