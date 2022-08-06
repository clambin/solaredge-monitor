package scraper

import (
	"time"
)

type Summary struct {
	Total float64
	Count int
	First time.Time
	Last  time.Time
}

func (summary *Summary) Add(m Sample) {
	if m.Timestamp.Before(summary.First) || summary.First.IsZero() {
		summary.First = m.Timestamp
	}
	if m.Timestamp.After(summary.Last) {
		summary.Last = m.Timestamp
	}
	summary.Total += m.Value
	summary.Count++
}

func (summary *Summary) Summarize() (result Sample) {
	result = Sample{
		Timestamp: summary.First,
		Value:     summary.Total / float64(summary.Count),
	}
	summary.First = time.Time{}
	summary.Last = time.Time{}
	summary.Total = 0
	summary.Count = 0
	return
}
