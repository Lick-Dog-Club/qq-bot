package cronjob

import (
	"context"
	"fmt"
	"log"
	"qq/bot"
	"sort"
	"sync"
)

var cronManager = newManager(newRobfigCronV3Runner())

var newBotFunc func(msg *bot.Message) bot.Bot = bot.NewBot

func SetNewBotFunc(fn func(msg *bot.Message) bot.Bot) {
	newBotFunc = fn
}

func Manager() CronManager {
	return cronManager
}

type CronRunner interface {
	AddCommand(name string, expression string, fn func()) error
	Run(context.Context) error
	Shutdown(context.Context) error
}

type CronManager interface {
	NewCommand(name string, fn func(bot bot.Bot) error) CommandImp
	Run(ctx context.Context) error
	Shutdown(context.Context) error

	List() []CommandImp
}

type manager struct {
	runner CronRunner
	sync.RWMutex
	commands map[string]*command
}

func newManager(runner CronRunner) *manager {
	return &manager{commands: make(map[string]*command), runner: runner}
}

func (m *manager) NewCommand(name string, fn func(bot bot.Bot) error) CommandImp {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.commands[name]; ok {
		panic(fmt.Sprintf("[CRON]: job %s already exists", name))
	}
	cmd := &command{expression: expression, name: name, fn: func() {
		_ = fn(newBotFunc(nil))
	}}
	m.commands[name] = cmd
	return cmd
}

func (m *manager) Run(ctx context.Context) error {
	log.Println("[Server]: start cron.")
	for _, cmd := range m.List() {
		c := cmd
		if err := m.runner.AddCommand(c.name, c.Expression(), func() {
			log.Println("[RUNNING]: " + c.name)
			c.Func()()
		}); err != nil {
			return err
		}
	}

	return m.runner.Run(ctx)
}

func (m *manager) List() []CommandImp {
	m.RLock()
	defer m.RUnlock()
	var cmds []CommandImp
	for _, c := range m.commands {
		cmds = append(cmds, &command{
			name:       c.name,
			expression: c.expression,
			fn:         c.fn,
		})
	}
	sort.Sort(sortCommand(cmds))

	return cmds
}

func (m *manager) Shutdown(ctx context.Context) error {
	return m.runner.Shutdown(ctx)
}

type sortCommand []CommandImp

func (s sortCommand) Len() int {
	return len(s)
}

func (s sortCommand) Less(i, j int) bool {
	return s[i].Name() < s[j].Name()
}

func (s sortCommand) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
