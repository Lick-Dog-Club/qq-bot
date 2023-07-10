package features

import (
	"fmt"
	"qq/bot"
	"sort"
	"sync"
)

var (
	defaultCommand cmd
	commands       = make(map[string]CommandImp)
	mu             sync.RWMutex
)

type Option func(cmd *cmd) error

func WithSysCmd() Option {
	return func(cmd *cmd) error {
		cmd.sysCmd = true
		return nil
	}
}

func WithGroup(group string) Option {
	return func(cmd *cmd) error {
		cmd.group = group
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
	Group() string
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

func Match(key string) bool {
	for _, c := range commands {
		if key == c.Keyword() {
			return true
		}
	}
	return false
}

func Run(bot bot.Bot, keyword string, content string) error {
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

	return command.Run(bot, content)
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

func BeautifulOutput(hidden bool, simple bool) string {
	var cmds string
	for _, imp := range AllKeywordCommands(hidden) {
		fmtStr := "%-12s\t%s\n"
		if !simple {
			fmtStr = "@bot\t" + fmtStr
		}
		cmds += fmt.Sprintf(fmtStr, imp.Keyword(), imp.Description())
	}
	return cmds
}

func AllKeywordCommands(hidden bool) []CommandImp {
	mu.RLock()
	defer mu.RUnlock()
	var cmds sortCommands
	var groupCmds = map[string]sortCommands{}

	for _, imp := range commands {
		if hidden && imp.Hidden() {
			continue
		}
		if imp.Group() != "" {
			groupCmds[imp.Group()] = append(groupCmds[imp.Group()], imp)
			continue
		}
		cmds = append(cmds, imp)
	}

	cmds = append(cmds, defaultCommand)
	sort.Sort(cmds)
	for _, imps := range groupCmds {
		sort.Sort(imps)
	}

	for _, imps := range groupCmds {
		cmds = append(cmds, imps...)
	}

	return cmds
}

type cmd struct {
	keyword string
	desc    string
	fn      commandFunc
	sysCmd  bool
	hidden  bool
	group   string
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

func (c cmd) Group() string {
	return c.group
}

func (c cmd) Description() string {
	return c.desc
}

func (c cmd) Run(bot bot.Bot, content string) error {
	return c.fn(bot, content)
}
