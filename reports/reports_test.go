package reports_test

import (
	"github.com/clambin/solaredge-monitor/reports"
	"github.com/clambin/solaredge-monitor/store/mockdb"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestServer_Empty(t *testing.T) {
	_, err := reports.MakeGraph(nil, true)
	assert.Error(t, err)
}

func TestServer_Summary(t *testing.T) {
	reporter := reports.New(mockdb.BuildDB())

	start, _ := reporter.GetFirst()
	stop, _ := reporter.GetLast()
	assert.NotEqual(t, start, stop)

	image, err := reporter.Summary(start, stop)

	assert.NoError(t, err)
	assert.NotNil(t, image)
}

func TestServer_TimeSeries(t *testing.T) {
	reporter := reports.New(mockdb.BuildDB())

	start, _ := reporter.GetFirst()
	stop, _ := reporter.GetLast()
	assert.NotEqual(t, start, stop)

	image, err := reporter.TimeSeries(start, stop)

	assert.NoError(t, err)
	assert.NotNil(t, image)
}

func Benchmark(b *testing.B) {
	reporter := reports.New(mockdb.BuildDB())

	start, _ := reporter.GetFirst()
	stop, _ := reporter.GetLast()
	assert.NotEqual(b, start, stop)

	image, err := reporter.TimeSeries(start, stop)

	assert.NoError(b, err)
	assert.NotNil(b, image)

}
