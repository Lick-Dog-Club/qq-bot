package webot

import (
	"qq/bot"
	"testing"

	_ "qq/features/ai"
	_ "qq/features/bili-lottery"
	_ "qq/features/config"
	_ "qq/features/daxin"
	_ "qq/features/help"
	_ "qq/features/kfc"
	_ "qq/features/lifetip"
	_ "qq/features/picture"
	_ "qq/features/pixiv"
	_ "qq/features/raokouling"
	_ "qq/features/sysupdate"
	_ "qq/features/task"
	_ "qq/features/version"
	_ "qq/features/weather"
	_ "qq/features/weibo"
	_ "qq/features/zhihu"
)

func TestRunWechat(t *testing.T) {
	RunWechat(bot.NewDummyBot())
	select {}
}
