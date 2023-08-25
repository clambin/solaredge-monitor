package scraper

import (
	"context"
	"github.com/clambin/go-common/taskmanager"
	"log/slog"
	"time"
)

type Poller interface {
	Poll(ctx context.Context) error
}

var _ taskmanager.Task = &daemon{}

type daemon struct {
	Interval time.Duration
	Poller
	Logger *slog.Logger
}

func (d *daemon) Run(ctx context.Context) error {
	ticker := time.NewTicker(d.Interval)
	defer ticker.Stop()

	d.Logger.Debug("starting")
	for {
		select {
		case <-ticker.C:
			if err := d.Poll(ctx); err != nil {
				d.Logger.Error("poll failed", "err", err)
			}
		case <-ctx.Done():
			d.Logger.Debug("stopping")
			return ctx.Err()
		}
	}
}
