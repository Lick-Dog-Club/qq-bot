package features

import (
	"fmt"
	"math"
	"qq/bot"
	"qq/config"
	"qq/features/stock/tools"
	"sort"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/samber/lo"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	log "github.com/sirupsen/logrus"
)

var (
	defaultCommand cmd
	commands       = make(map[string]CommandImp)
	mu             sync.RWMutex
)

func AllFuncCalls() (res []tools.Tool) {
	mu.RLock()
	defer mu.RUnlock()
	for _, imp := range commands {
		if imp.Enabled() && imp.AiDefine() != nil {
			res = append(res, tools.Tool{
				Name: imp.Keyword(),
				Define: openai.Tool{
					Type: openai.ToolTypeFunction,
					Function: &openai.FunctionDefinition{
						Name:        imp.Keyword(),
						Description: imp.Description(),
						Parameters: &jsonschema.Definition{
							Type:       jsonschema.Object,
							Properties: imp.AiDefine().Properties,
							Required:   lo.Keys(imp.AiDefine().Properties),
						},
					},
				},
			})
		}
	}
	return
}

func CallFunc(keyword string, args string) (string, error) {
	log.Printf("call '%s', args '%s'\n", keyword, args)
	mu.RLock()
	defer mu.RUnlock()
	if imp, ok := commands[keyword]; ok {
		if imp.Enabled() {
			res, err := imp.AiDefine().Call(args)
			fmt.Println(res)
			return res, err
		}
	}
	return fmt.Sprintf("func '%s' 已禁用，无法调用该方法", keyword), nil
}

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

type AIFuncDef struct {
	Properties map[string]jsonschema.Definition
	Call       func(args string) (string, error)
}

func WithAI() Option {
	return func(cmd *cmd) error {
		cmd.hasAi = true
		return nil
	}
}
func WithAIFunc(define AIFuncDef) Option {
	return func(cmd *cmd) error {
		cmd.aiDefine = &define
		cmd.hasAi = true
		return nil
	}
}

type commandFunc func(bot bot.Bot, content string) error

type CommandImp interface {
	Enabled() bool
	Hidden() bool
	IsSysCmd() bool
	Keyword() string
	Group() string
	Description() string
	Run(bot bot.Bot, content string) error
	AiDefine() *AIFuncDef
	HasAI() bool
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
		if ok {
			return
		}
		command = defaultCommand
		content = keyword + " " + content
	}()

	if !command.Enabled() {
		bot.Send(fmt.Sprintf("指令 '%s' 未开启", command.Keyword()))
		return nil
	}

	return command.Run(bot, content)
}

type sortCommands []CommandImp

func (s sortCommands) MaxLen() int {
	var m float64
	for _, imp := range s {
		m = math.Max(m, float64(utf8.RuneCountInString(imp.Description())))
	}
	return int(m)
}

func (s sortCommands) Len() int {
	return len(s)
}

func (s sortCommands) Less(i, j int) bool {
	if s[i].IsSysCmd() == s[j].IsSysCmd() {
		return utf8.RuneCountInString(s[i].Description()) > utf8.RuneCountInString(s[j].Description())
	}
	return !s[i].IsSysCmd() && s[j].IsSysCmd()
}

func (s sortCommands) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func BeautifulOutput(hidden bool, simple bool) string {
	return strings.Join(BeautifulOutputLines(hidden, simple), "\n")
}

func BeautifulOutputLines(hidden bool, simple bool) []string {
	var cmds []string
	for _, imp := range AllKeywordCommands(hidden) {
		var aiEnabled = "x"
		if imp.HasAI() {
			aiEnabled = "y"
		}
		fmtStr := "%-16s\t%s"
		if !simple {
			fmtStr = "@bot\t" + fmtStr
		}
		cmds = append(cmds, fmt.Sprintf(fmtStr, fmt.Sprintf("(%s)%s", aiEnabled, imp.Keyword()), imp.Description()))
	}
	return cmds
}

type GroupCommands []sortCommands

func (g GroupCommands) Len() int {
	return len(g)
}

func (g GroupCommands) Less(i, j int) bool {
	return g[i].MaxLen() < g[j].MaxLen()
}

func (g GroupCommands) Swap(i, j int) {
	g[i], g[j] = g[j], g[i]
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

	sort.Sort(cmds)
	var sortGcmds GroupCommands
	for _, imps := range groupCmds {
		sort.Sort(imps)
		sortGcmds = append(sortGcmds, imps)
	}
	sort.Sort(sortGcmds)

	for _, imps := range sortGcmds {
		cmds = append(imps, cmds...)
	}

	return append(cmds, defaultCommand)
}

type cmd struct {
	keyword  string
	desc     string
	fn       commandFunc
	sysCmd   bool
	hidden   bool
	group    string
	disabled bool
	aiDefine *AIFuncDef
	hasAi    bool
}

func (c cmd) IsSysCmd() bool {
	return c.sysCmd
}

func (c cmd) Hidden() bool {
	return c.hidden
}

func (c cmd) Enabled() bool {
	return !config.DisabledCmds().Contains(c.keyword)
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

func (c cmd) AiDefine() *AIFuncDef {
	return c.aiDefine
}
func (c cmd) HasAI() bool {
	return c.hasAi
}

func (c cmd) Run(bot bot.Bot, content string) error {
	return c.fn(bot, content)
}
