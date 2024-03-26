package scraper

import (
	"context"
	"github.com/clambin/go-common/taskmanager"
	"log/slog"
	"time"
)

var _ taskmanager.Task = &daemon{}

type daemon struct {
	Interval time.Duration
	Poller
	Logger *slog.Logger
}

type Poller interface {
	Poll(ctx context.Context) error
}

func (d *daemon) Run(ctx context.Context) error {
	ticker := time.NewTicker(d.Interval)
	defer ticker.Stop()

	d.Logger.Debug("starting")
	defer d.Logger.Debug("stopping")

	for {
		if err := d.Poll(ctx); err != nil {
			d.Logger.Error("poll failed", "err", err)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
