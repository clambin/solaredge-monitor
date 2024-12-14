package poller

import (
	"context"
	"github.com/clambin/go-common/pubsub"
	"log/slog"
	"time"
)

type Poller[T any] struct {
	Updater[T]
	Interval time.Duration
	Logger   *slog.Logger
	pubsub.Publisher[T]
}

type Updater[T any] interface {
	GetUpdate(context.Context) (T, error)
}

func (p *Poller[T]) Run(ctx context.Context) error {
	p.Logger.Debug("starting poller", "interval", p.Interval)
	defer p.Logger.Debug("stopped poller")

	for {
		start := time.Now()
		if update, err := p.GetUpdate(ctx); err == nil {
			p.Publish(update)
			p.Logger.Debug("poll done", "duration", time.Since(start))
		} else {
			p.Logger.Error("failed to get update", "err", err)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(p.Interval):
		}
	}
}
