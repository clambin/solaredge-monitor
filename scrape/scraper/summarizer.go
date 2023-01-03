package scraper

import (
	"context"
	"golang.org/x/exp/slog"
	"sync"
	"time"
)

// A Summarizer returns the summary of a list of samples
type Summarizer interface {
	Summarize() Sample
	Count() int
	Run(ctx context.Context)
}

// Client collects measurements and returns a summary of those collected measurement
type Client struct {
	Scraper
	Interval time.Duration
	summary  Summary
	lock     sync.RWMutex
}

var _ Summarizer = &Client{}

// Run collects measurements at the specified interval and records them for latest summarizing
func (c *Client) Run(ctx context.Context) {
	ticker := time.NewTicker(c.Interval)
	defer ticker.Stop()

	c.collect(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.collect(ctx)
		}
	}
}

func (c *Client) collect(ctx context.Context) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if m, err := c.Scraper.Scrape(ctx); err == nil {
		c.summary.Add(m)
	} else {
		slog.Error("failed to measure", err)
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
