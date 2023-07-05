package help

import (
	"qq/bot"
	"qq/features"
)

func init() {
	features.AddKeyword("help", "帮助信息", func(sender bot.Bot, content string) error {
		showHelp(sender, true)
		return nil
	}, features.WithSysCmd())
	features.AddKeyword("hhelp", "帮助信息, 显示被隐藏的指令", func(sender bot.Bot, content string) error {
		showHelp(sender, false)
		return nil
	}, features.WithSysCmd(), features.WithHidden())
}

func showHelp(sender bot.Bot, hidden bool) {
	sender.Send(features.BeautifulOutput(hidden))
}
