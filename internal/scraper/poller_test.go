package scraper_test

import (
	"context"
	"github.com/clambin/solaredge-monitor/internal/scraper"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"testing"
	"time"
)

func TestPoller(t *testing.T) {
	c := client{update: testUpdate}
	p := scraper.Poller{
		Client:   &c,
		Interval: time.Millisecond,
		Logger:   slog.Default(),
	}
	ch := p.Subscribe()
	defer p.Unsubscribe(ch)

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error)
	go func() { errCh <- p.Run(ctx) }()

	assert.Equal(t, testUpdate, <-ch)
	assert.Equal(t, testUpdate, <-ch)

	cancel()
	assert.NoError(t, <-errCh)
}
