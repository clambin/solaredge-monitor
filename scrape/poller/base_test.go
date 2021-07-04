package poller_test

import (
	"context"
	"fmt"
	"github.com/clambin/solaredge-monitor/scrape/collector"
	"github.com/clambin/solaredge-monitor/scrape/poller"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestBasePoller_Run(t *testing.T) {
	collectChannel := make(chan collector.Metric)
	p := poller.NewBasePoller(25*time.Millisecond, pollFunction, collectChannel)

	ctx, cancel := context.WithCancel(context.Background())
	go p.Run(ctx)

	time.Sleep(50 * time.Millisecond)
	cancel()
	assert.Never(t, func() bool {
		_, ok := <-collectChannel
		return ok
	}, 200*time.Millisecond, 100*time.Millisecond)
}

func pollFunction(ctx context.Context) (value float64, err error) {
	<-ctx.Done()

	return 0, fmt.Errorf("deadline expired")
}
