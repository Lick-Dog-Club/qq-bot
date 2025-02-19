package autoclean

import (
	"fmt"
	"os"
	"qq/bot"
	"qq/config"
	"qq/cronjob"
	"time"
)

func init() {
	cronjob.Manager().NewCommand("auto-clean", func(bot bot.CronBot) error {
		// 清理 /data/images/ 下面所有的图片
		os.RemoveAll(config.ImageDir)
		os.MkdirAll(config.ImageDir, 0755)
		bot.SendToUser(config.UserID(), fmt.Sprintf("自动清理图片 %s", time.Now().Format(time.DateTime)))
		return nil
	}).Daily()
}
