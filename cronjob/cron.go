package cronjob

import (
	"context"
	"qq/config"
	"strings"
	"sync"
	"time"

	"github.com/samber/lo"

	log "github.com/sirupsen/logrus"

	"github.com/robfig/cron/v3"
)

type robfigCronV3Runner struct {
	sync.RWMutex

	c        *cron.Cron
	entryMap map[string]int64
}

func newRobfigCronV3Runner() *robfigCronV3Runner {
	return &robfigCronV3Runner{
		c: cron.New(
			cron.WithLocation(time.Local),
			cron.WithSeconds(),
			cron.WithChain(
				cron.Recover(&cronLogger{}),
			),
		),
		entryMap: make(map[string]int64),
	}
}

func (c *robfigCronV3Runner) AddCommand(name string, expression string, fn func()) error {
	c.Lock()
	defer c.Unlock()
	id, err := c.c.AddFunc(expression, fn)
	if err != nil {
		return err
	}
	c.entryMap[name] = int64(id)
	return nil
}

type onceSchedule struct {
	sync.Once

	next time.Time
	c    *cron.Cron
	id   cron.EntryID
}

func newOnceSchedule(next time.Time, c *cron.Cron) cron.Schedule {
	return &onceSchedule{next: next, c: c}
}

func (o *onceSchedule) Next(t time.Time) time.Time {
	res := o.next
	o.Once.Do(func() {
		o.next = time.Time{}
	})
	return res
}

func (c *robfigCronV3Runner) Remove(id int) error {
	c.c.Remove(cron.EntryID(id))
	return nil
}

func (c *robfigCronV3Runner) AddOnceCommand(t time.Time, fn func()) int {
	c.Lock()
	defer c.Unlock()
	s := newOnceSchedule(t, c.c).(*onceSchedule)
	s.id = c.c.Schedule(s, cron.FuncJob(func() {
		cronManager.RemoveOnceCommand(int(s.id))
		filter := lo.Filter(config.Tasks(), func(item config.Task, index int) bool {
			return item.ID != int(s.id)
		})
		config.SyncTasks(filter)
		c.c.Remove(s.id)
		fn()
	}))
	return int(s.id)
}

func (c *robfigCronV3Runner) Run(ctx context.Context) error {
	go func() {
		defer func() {
			e := recover()
			if e != nil {
				log.Println(e)
			}
		}()
		c.c.Run()
	}()
	return nil
}

func (c *robfigCronV3Runner) Shutdown(ctx context.Context) error {
	stopCtx := c.c.Stop()
	select {
	case <-stopCtx.Done():
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

type cronLogger struct{}

func (c *cronLogger) Info(msg string, keysAndValues ...any) {
	log.Printf(formatString(len(keysAndValues)), append([]any{msg}, keysAndValues...)...)
}

func (c *cronLogger) Error(err error, msg string, keysAndValues ...any) {
	log.Printf("[CRON]: %v, %v=%v\n", err, msg, keysAndValues)
}

func formatString(numKeysAndValues int) string {
	var sb strings.Builder
	sb.WriteString("[CRON]: %s")
	if numKeysAndValues > 0 {
		sb.WriteString(", ")
	}
	for i := 0; i < numKeysAndValues/2; i++ {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("%v=%v")
	}
	return sb.String()
}
