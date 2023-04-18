package cronjob

import (
	"context"
	"fmt"
	"qq/bot"
	"sort"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
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
	AddOnceCommand(t time.Time, fn func()) int
	Remove(int) error
	Run(context.Context) error
	Shutdown(context.Context) error
}

type CronManager interface {
	NewCommand(name string, fn func(bot bot.CronBot) error) CommandImp
	NewOnceCommand(name string, t time.Time, fn func(bot bot.Bot) error) int
	Run(ctx context.Context) error
	RemoveOnceCommand(id int) error

	Shutdown(context.Context) error

	List() []CommandImp
}

type manager struct {
	runner CronRunner
	sync.RWMutex
	commands     map[string]*command
	onceCommands map[int]*onceCommand
}

func newManager(runner CronRunner) *manager {
	return &manager{commands: make(map[string]*command), runner: runner, onceCommands: map[int]*onceCommand{}}
}

func (m *manager) NewCommand(name string, fn func(bot bot.CronBot) error) CommandImp {
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

func (m *manager) NewOnceCommand(name string, t time.Time, fn func(bot bot.Bot) error) int {
	m.Lock()
	defer m.Unlock()
	cmd := &onceCommand{
		name: name,
		date: t,
		fn: func() {
			_ = fn(newBotFunc(nil))
		},
	}
	cmd.id = m.runner.AddOnceCommand(cmd.date, cmd.fn)
	m.onceCommands[cmd.id] = cmd
	return cmd.id
}

func (m *manager) RemoveOnceCommand(id int) error {
	m.Lock()
	defer m.Unlock()
	delete(m.onceCommands, id)
	return m.runner.Remove(id)
}

func (m *manager) Run(ctx context.Context) error {
	log.Println("[Server]: start cron.")
	for _, cmd := range m.List() {
		c := cmd
		if err := m.runner.AddCommand(c.Name(), c.Expression(), func() {
			log.Println("[RUNNING]: " + c.Name())
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
