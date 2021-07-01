package reports_test

import (
	"github.com/clambin/solaredge-monitor/reports"
	"github.com/clambin/solaredge-monitor/store/mockdb"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestServer_GetFirstLast(t *testing.T) {
	reporter := reports.New(mockdb.NewDB())

	_, err := reporter.GetFirst()
	assert.Error(t, err)

	_, err = reporter.GetLast()
	assert.Error(t, err)

	reporter = reports.New(mockdb.BuildDB())

	var timestamp time.Time
	timestamp, err = reporter.GetFirst()
	assert.NoError(t, err)
	assert.NotZero(t, timestamp)

	_, err = reporter.GetLast()
	assert.NoError(t, err)
	assert.NotZero(t, timestamp)
}

func TestServer_Reports(t *testing.T) {
	reporter := reports.New(mockdb.BuildDB())

	var start, stop time.Time
	var err error
	start, err = reporter.GetFirst()
	assert.NoError(t, err)
	stop, err = reporter.GetLast()
	assert.NoError(t, err)
	assert.NotEqual(t, start, stop)

	for _, f := range []func(time.Time, time.Time) ([]byte, error){reporter.Summary, reporter.TimeSeries, reporter.Classify} {
		var img []byte
		img, err = f(start, stop)
		assert.NoError(t, err)
		assert.NotNil(t, img)
		assert.Greater(t, len(img), 0)
	}

	reporter = reports.New(mockdb.BadDB())
	for _, f := range []func(time.Time, time.Time) ([]byte, error){reporter.Summary, reporter.TimeSeries, reporter.Classify} {
		_, err = f(start, stop)
		assert.Error(t, err)
	}
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
