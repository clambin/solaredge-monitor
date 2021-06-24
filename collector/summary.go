package collector

import (
	"time"
)

type Summary struct {
	Total float64
	Count float64
	First time.Time
	Last  time.Time
}

func (summary *Summary) Add(m Metric) {
	if m.Timestamp.Before(summary.First) || summary.First.IsZero() {
		summary.First = m.Timestamp
	}
	if m.Timestamp.After(summary.Last) {
		summary.Last = m.Timestamp
	}
	summary.Total += m.Value
	summary.Count++
}

func (summary *Summary) InRange(m Metric, interval time.Duration) bool {
	if summary.First.IsZero() {
		return true
	}
	return m.Timestamp.Before(summary.First.Add(interval))
}

func (summary *Summary) Get() (result Metric) {
	result = Metric{
		Timestamp: summary.First,
		Value:     summary.Total / summary.Count,
	}
	summary.First = time.Time{}
	summary.Last = time.Time{}
	summary.Total = 0
	summary.Count = 0
	return
}
