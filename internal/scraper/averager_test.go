package scraper_test

import (
	"github.com/clambin/solaredge-monitor/internal/scraper"
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

func TestSummary(t *testing.T) {
	var s scraper.Averager

	assert.True(t, math.IsNaN(s.Average()))

	s.Add(100)
	assert.Equal(t, 100.0, s.Average())

	assert.True(t, math.IsNaN(s.Average()))

	s.Add(5)
	s.Add(10)
	s.Add(0)
	s.Add(5)

	assert.Equal(t, 5.0, s.Average())
}
