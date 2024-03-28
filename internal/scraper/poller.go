package scraper

import (
	"context"
	"github.com/clambin/solaredge-monitor/internal/scraper/solaredge"
	"github.com/clambin/solaredge-monitor/pkg/pubsub"
	"log/slog"
	"time"
)

type Poller struct {
	Client   SolarEdgeGetter
	Interval time.Duration
	Logger   *slog.Logger
	pubsub.Publisher[solaredge.Update]
}
type SolarEdgeGetter interface {
	GetUpdate(context.Context) (solaredge.Update, error)
}

func (p *Poller) Run(ctx context.Context) error {
	p.Logger.Debug("starting poller", "interval", p.Interval)
	defer p.Logger.Debug("stopped poller")

	for {
		start := time.Now()
		if update, err := p.Client.GetUpdate(ctx); err == nil {
			p.Publish(update)
			p.Logger.Debug("poll done", "duration", time.Since(start))
		} else {
			p.Logger.Error("failed to get solaredge data", "err", err)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(p.Interval):
		}
	}
}
