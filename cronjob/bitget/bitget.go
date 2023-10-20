package bitget

import (
	"qq/bot"
	"qq/config"
	"qq/cronjob"
	"qq/features/bitget"
	"qq/util"
)

func init() {
	cronjob.Manager().NewCommand("bitget", func(bot bot.CronBot) error {
		if config.BgApiSecretKey() != "" && config.BgApiKey() != "" && config.BgPassphrase() != "" {
			if v := bitget.Get(false); v != "" {
				util.Bark("有变动", v, config.BarkUrls()...)
			}
		}
		return nil
	}).EveryTenSeconds()
}
