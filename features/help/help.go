package help

import (
	"fmt"
	"qq/bot"
	"qq/features"
	"strings"
)

func init() {
	features.AddKeyword("help", "帮助信息", func(sender bot.Bot, content string) error {
		var res []string
		for _, command := range features.AllKeywordCommands() {
			res = append(res, fmt.Sprintf("%s: %s", command.Keyword(), command.Description()))
		}
		sender.Send(strings.Join(res, "\n"))
		return nil
	}, features.WithSysCmd())
}
