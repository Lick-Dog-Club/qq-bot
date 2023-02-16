package cronjob

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

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
	log.Printf("[CRON]: %v", err)
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
