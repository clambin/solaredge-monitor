package publisher

import (
	"context"
	"fmt"
	"github.com/clambin/go-common/pubsub"
	"github.com/clambin/tado/v2"
	"log/slog"
	"sync/atomic"
	"time"
)

type Publisher[T any] struct {
	Updater[T]
	Logger *slog.Logger
	pubsub.Publisher[T]
	Interval   time.Duration
	lastUpdate atomic.Value
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
			p.lastUpdate.Store(time.Now())
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

func (p *Publisher[T]) IsHealthy(_ context.Context) error {
	lastUpdate := p.lastUpdate.Load()
	if lastUpdate == nil {
		return fmt.Errorf("no data received from %s", p.getSource())
	}
	if noData := time.Since(lastUpdate.(time.Time)); noData > 5*p.Interval {
		return fmt.Errorf("no data received from %s since %v", p.getSource(), noData)
	}
	return nil
}

func (p *Publisher[T]) getSource() string {
	var t T
	var ptr any = t
	switch ptr.(type) {
	case SolarEdgeUpdate:
		return "SolarEdge"
	case *tado.Weather:
		return "Tado"
	default:
		return "unknown source"
	}
}
