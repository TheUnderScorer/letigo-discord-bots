package scheduler

import (
	"context"
	"errors"
	"fmt"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

func Init(ctx context.Context) error {
	c := cron.New()

	err := schedule(c, "DailyGreeting", "0 9 * * *", func() {
		DailyGreeting(ctx)
	})
	if err != nil {
		return err
	}

	c.Start()

	return nil
}

func schedule(c *cron.Cron, name string, spec string, f func()) error {
	_, err := c.AddFunc(spec, f)
	if err != nil {
		log.Error("failed to schedule job", zap.String("name", name), zap.String("spec", spec), zap.Error(err))
		return errors.Join(err, errors.New(fmt.Sprintf("failed to schedule job %s", name)))
	}

	return nil
}
