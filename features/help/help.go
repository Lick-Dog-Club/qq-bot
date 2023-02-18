package help

import (
	"fmt"
	"qq/bot"
	"qq/features"
	"strings"
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
	var res []string
	for _, command := range features.AllKeywordCommands(hidden) {
		res = append(res, fmt.Sprintf("%s: %s", command.Keyword(), command.Description()))
	}
	sender.Send(strings.Join(res, "\n"))
}
