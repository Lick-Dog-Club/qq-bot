package lpr

import (
	"fmt"
	"qq/bot"
	"qq/config"
	"qq/cronjob"
	"qq/features/lpr"
	"time"
)

func init() {
	cronjob.Manager().NewCommand("lpr", func(robot bot.CronBot) error {
		lprs := lpr.Get()
		if len(lprs) > 0 && time.Now().Format("2006-01-02") == lprs[0].Date.Format("2006-01-02") {
			robot.SendGroup(
				config.GroupID(),
				fmt.Sprintf("LPR 调整了，日期 %v 一年 '%v%%' 五年 '%v%%'",
					lprs[0].Date.Format("2006-01-02"),
					lprs[0].OneYear,
					lprs[0].FiveYear,
				),
			)
		}
		return nil
	}).DailyAt("09:00")
}
