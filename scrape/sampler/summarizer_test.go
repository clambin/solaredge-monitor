package sampler_test

import (
	"context"
	"errors"
	"github.com/clambin/solaredge-monitor/scrape/sampler"
	"github.com/stretchr/testify/assert"
	"math"
	"sync"
	"testing"
	"time"
)

func TestClient_Run(t *testing.T) {
	s := sampler.Client{Sampler: &measurer{}}
	ctx, cancel := context.WithCancel(context.Background())

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		s.Run(ctx, 10*time.Millisecond)
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
	s := sampler.Client{Sampler: &m}
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	wg.Add(1)
	go func() {
		s.Run(ctx, 10*time.Millisecond)
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

func (m *measurer) Sample(_ context.Context) (sampler.Sample, error) {
	m.counter++
	return sampler.Sample{
		Timestamp: time.Now(),
		Value:     m.counter,
	}, m.err
}

var _ sampler.Sampler = &measurer{}
