package features

import (
	"fmt"
	"qq/bot"
	"sort"
	"sync"
)

var (
	newBotFunc func(msg *bot.Message) bot.Bot = bot.NewBot

	defaultCommand cmd
	commands       = make(map[string]CommandImp)
	mu             sync.RWMutex
)

func SetNewBotFunc(fn func(msg *bot.Message) bot.Bot) {
	newBotFunc = fn
}

type Option func(cmd *cmd) error

func WithSysCmd() Option {
	return func(cmd *cmd) error {
		cmd.sysCmd = true
		return nil
	}
}

func WithHidden() Option {
	return func(cmd *cmd) error {
		cmd.hidden = true
		return nil
	}
}

type commandFunc func(bot bot.Bot, content string) error

type CommandImp interface {
	Hidden() bool
	IsSysCmd() bool
	Keyword() string
	Description() string
	Run(bot bot.Bot, content string) error
}

func AddKeyword(keyword, desc string, fn commandFunc, opts ...Option) {
	mu.Lock()
	defer mu.Unlock()
	_, ok := commands[keyword]
	if ok {
		panic(fmt.Sprintf("关键字: %s 已经注册过了!", keyword))
	}

	c := cmd{
		keyword: keyword,
		desc:    desc,
		fn:      fn,
	}
	for _, opt := range opts {
		opt(&c)
	}
	commands[keyword] = c
}

func SetDefault(desc string, fn commandFunc) {
	mu.Lock()
	defer mu.Unlock()
	defaultCommand = cmd{
		fn:      fn,
		desc:    desc,
		keyword: "default",
	}
}

func Run(msg *bot.Message, keyword string, content string) error {
	var command CommandImp
	func() {
		mu.RLock()
		defer mu.RUnlock()
		var ok bool
		command, ok = commands[keyword]
		if !ok {
			command = defaultCommand
			content = keyword + " " + content
		}
	}()
	return command.Run(newBotFunc(msg), content)
}

type sortCommands []CommandImp

func (s sortCommands) Len() int {
	return len(s)
}

func (s sortCommands) Less(i, j int) bool {
	if s[i].IsSysCmd() == s[j].IsSysCmd() {
		return len([]rune(s[i].Keyword()+s[i].Description())) > len([]rune(s[j].Keyword()+s[j].Description()))
	}
	return !s[i].IsSysCmd() && s[j].IsSysCmd()
}

func (s sortCommands) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func AllKeywordCommands(hidden bool) []CommandImp {
	mu.RLock()
	defer mu.RUnlock()
	var cmds sortCommands
	for _, imp := range commands {
		if hidden && imp.Hidden() {
			continue
		}
		cmds = append(cmds, imp)
	}

	cmds = append(cmds, defaultCommand)
	sort.Sort(cmds)

	return cmds
}

type cmd struct {
	keyword string
	desc    string
	fn      commandFunc
	sysCmd  bool
	hidden  bool
}

func (c cmd) IsSysCmd() bool {
	return c.sysCmd
}

func (c cmd) Hidden() bool {
	return c.hidden
}

func (c cmd) Keyword() string {
	return c.keyword
}

func (c cmd) Description() string {
	return c.desc
}

func (c cmd) Run(bot bot.Bot, content string) error {
	return c.fn(bot, content)
}
