package scraper

import (
	"context"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

// A Summarizer returns the summary of a list of samples
type Summarizer interface {
	Summarize() Sample
	Count() int
}

// Client collects measurements and returns a summary of those collected measurement
type Client struct {
	Scraper
	summary Summary
	lock    sync.RWMutex
}

var _ Summarizer = &Client{}

// Run collects measurements at the specified interval and records them for latest summarizing
func (c *Client) Run(ctx context.Context, interval time.Duration) {
	c.collect(ctx)
	ticker := time.NewTicker(interval)
	for running := true; running; {
		select {
		case <-ctx.Done():
			running = false
		case <-ticker.C:
			c.collect(ctx)
		}
	}
	ticker.Stop()
}

func (c *Client) collect(ctx context.Context) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if m, err := c.Scrape(ctx); err == nil {
		log.Debugf("received %v", m)
		c.summary.Add(m)
	} else {
		log.WithError(err).Warning("failed to measure")
	}
}

// Summarize returns a summary of collected measurements. The summary is then reset, i.e. previously collected measurements are discarded.
func (c *Client) Summarize() Sample {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.summary.Summarize()
}

// Count returns the number of measurements currently collected
func (c *Client) Count() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.summary.Count
}
