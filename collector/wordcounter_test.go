package collector_test

import (
	"github.com/clambin/solaredge-monitor/collector"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWordCounter(t *testing.T) {
	c := collector.WordCounter{}

	assert.Empty(t, c.GetMostUsed())

	c.Add("foo")
	c.Add("bar")
	c.Add("snafu")
	c.Add("foo")
	c.Add("bar")
	c.Add("foo")

	assert.Equal(t, "foo", c.GetMostUsed())
	assert.Empty(t, c.GetMostUsed())
}
