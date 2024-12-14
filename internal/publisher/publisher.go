package publisher

import (
	"context"
	"github.com/clambin/go-common/pubsub"
	"log/slog"
	"time"
)

type Publisher[T any] struct {
	Updater[T]
	Interval time.Duration
	Logger   *slog.Logger
	pubsub.Publisher[T]
}

type Updater[T any] interface {
	GetUpdate(context.Context) (T, error)
}

func (p *Publisher[T]) Run(ctx context.Context) error {
	p.Logger.Debug("starting publisher", "interval", p.Interval)
	defer p.Logger.Debug("stopped publisher")

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
