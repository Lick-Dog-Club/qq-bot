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
	fpath := filepath.Join("/data", "images", "help.png")
	once.Do(func() {
		text2png.Draw(features.BeautifulOutputLines(hidden, true), fpath)
	})

	if bot.Message().WeSendImg != nil {
		open, _ := os.Open(fpath)
		defer open.Close()
		bot.Message().WeSendImg(open)
	} else {
		bot.Send(fmt.Sprintf("[CQ:image,file=file://%s]", fpath))
	}
}
