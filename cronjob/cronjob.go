package cronjob

import (
	"context"
	"fmt"
	"qq/bot"
	"qq/config"
	"qq/features"
	"qq/util"
	"qq/util/random"
	"sort"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

var cronManager = newManager(newRobfigCronV3Runner())

var newBotFunc func(msg *bot.Message) bot.Bot = bot.NewQQBot

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
	LoadOnceTasks()
	Shutdown(context.Context) error

	List() []CommandImp
	ListOnceCommands() string
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

func (m *manager) LoadOnceTasks() {
	var newTasks []config.Task
	for _, task := range config.Tasks() {
		parse, _ := time.ParseInLocation(time.DateTime, task.RunAt, time.Local)
		if time.Now().After(parse) {
			continue
		}
		b := bot.NewQQBot(&bot.Message{
			SenderUserID:  task.UserID,
			IsSendByGroup: task.GroupID != "",
			GroupID:       task.GroupID,
		})
		tid := m.NewOnceCommand(random.String(20), parse, func(bot.Bot) error {
			if k, v := util.GetKeywordAndContent(task.Content); features.Match(k) {
				features.Run(b, k, v)
			} else {
				b.SendToUser(task.UserID, task.Content)
			}
			return nil
		})
		newTasks = append(newTasks, config.Task{
			ID:      tid,
			Name:    task.Name,
			RunAt:   task.RunAt,
			Content: task.Content,
			UserID:  task.UserID,
			GroupID: task.GroupID,
		})
	}
	config.SyncTasks(newTasks)
}

func (m *manager) NewCommand(name string, fn func(bot bot.CronBot) error) CommandImp {
	if config.DisabledCrons().Contains(name) {
		log.Println("filter disabled cron: ", name)
		return &command{expression: expression, name: name, fn: func() {}}
	}
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

func (m *manager) ListOnceCommands() (res string) {
	m.RLock()
	defer m.RUnlock()

	res = "任务列表:\n"
	for _, o := range m.onceCommands {
		res += fmt.Sprintf("ID: %d 时间：%s, 任务: %s\n", o.id, o.date.Format(time.DateTime), o.name)
	}
	return
}

func (m *manager) Run(ctx context.Context) error {
	log.Println("[Server]: start cron.")
	for _, cmd := range m.List() {
		c := cmd
		if err := m.runner.AddCommand(c.Name(), c.Expression(), func() {
			//log.Println("[RUNNING]: " + c.Name())
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

var NewCommand = Manager().NewCommand
