package scheduler

import (
	"app/logging"
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
	_, err := c.AddFunc(spec, wrapScheduleFn(name, f))
	if err != nil {
		log.Error("failed to schedule job", zap.String("name", name), zap.String("spec", spec), zap.Error(err))
		return errors.Join(err, fmt.Errorf("failed to schedule job %s", name))
	}

	return nil
}

func wrapScheduleFn(name string, f func()) func() {
	return func() {
		log := logging.Get().With(zap.String("name", name))
		log.Info("starting schedule")
		f()
		log.Info("finished schedule")
	}
}
