package help

import (
	"fmt"
	"os"
	"path/filepath"
	"qq/bot"
	"qq/features"
	"qq/util/text2png"
	"sync"
)

var once sync.Once
var hideOnce sync.Once

func init() {
	features.AddKeyword("help", "帮助信息", func(sender bot.Bot, content string) error {
		showHelp(sender, true)
		return nil
	}, features.WithSysCmd(), features.WithGroup("help"))
	features.AddKeyword("hhelp", "帮助信息, 显示被隐藏的指令", func(sender bot.Bot, content string) error {
		showHelp(sender, false)
		return nil
	}, features.WithSysCmd(), features.WithHidden(), features.WithGroup("help"))
}

func showHelp(bot bot.Bot, hidden bool) {
	helpPpath := filepath.Join("/data", "images", "help.png")
	hhelpPath := filepath.Join("/data", "images", "hhelp.png")
	hideOnce.Do(func() {
		text2png.Draw(features.BeautifulOutputLines(false, true), hiddenFpath)
	})
	once.Do(func() {
		text2png.Draw(features.BeautifulOutputLines(true, true), fpath)
	})

	var p string = helpPpath
	if hidden {
		p = hhelpPath
	}

	if bot.Message().WeSendImg != nil {
		open, _ := os.Open(p)
		defer open.Close()
		bot.Message().WeSendImg(open)
	} else {
		bot.Send(fmt.Sprintf("[CQ:image,file=file://%s]", p))
	}
}
