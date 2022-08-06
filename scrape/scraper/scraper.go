package scraper

import (
	"context"
	"time"
)

type Scraper interface {
	Scrape(ctx context.Context) (Sample, error)
}

type Sample struct {
	Timestamp time.Time
	Value     float64
}
