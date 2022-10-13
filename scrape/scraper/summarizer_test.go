package scraper_test

import (
	"context"
	"errors"
	"github.com/clambin/solaredge-monitor/scrape/scraper"
	"github.com/stretchr/testify/assert"
	"math"
	"sync"
	"testing"
	"time"
)

func TestClient_Run(t *testing.T) {
	s := scraper.Client{Scraper: &measurer{}, Interval: 10 * time.Millisecond}
	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		s.Run(ctx)
		wg.Done()
	}()

	assert.Eventually(t, func() bool { return s.Count() >= 5 }, 100*time.Millisecond, 20*time.Millisecond)

	m := s.Summarize()
	assert.True(t, !math.IsNaN(m.Value))

	cancel()
	wg.Wait()
}

func TestClient_Run_Failure(t *testing.T) {
	m := measurer{err: errors.New("fail")}
	s := scraper.Client{Scraper: &m, Interval: 10 * time.Millisecond}
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	wg.Add(1)
	go func() {
		s.Run(ctx)
		wg.Done()
	}()

	assert.Never(t, func() bool { return s.Count() > 0 }, 50*time.Millisecond, 10*time.Millisecond)

	cancel()
	wg.Wait()
}

type measurer struct {
	counter float64
	err     error
}

func (m *measurer) Scrape(_ context.Context) (scraper.Sample, error) {
	m.counter++
	return scraper.Sample{
		Timestamp: time.Now(),
		Value:     m.counter,
	}, m.err
}

var _ scraper.Scraper = &measurer{}
