package reports_test

import (
	"flag"
	"github.com/clambin/solaredge-monitor/reports"
	"github.com/clambin/solaredge-monitor/store/mockdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var update = flag.Bool("update", false, "update .golden files")

func TestReporter_GetFirstLast(t *testing.T) {
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

func TestReporter(t *testing.T) {
	reporter := reports.New(mockdb.BuildDB())

	var start, stop time.Time
	var err error
	start, err = reporter.GetFirst()
	assert.NoError(t, err)
	stop, err = reporter.GetLast()
	assert.NoError(t, err)
	assert.NotEqual(t, start, stop)

	testcases := map[string]func(time.Time, time.Time) ([]byte, error){
		"summary":    reporter.Summary,
		"timeseries": reporter.TimeSeries,
		"classify":   reporter.Classify,
	}

	for name, f := range testcases {
		var img []byte
		img, err = f(start, stop)
		assert.NoError(t, err)
		assert.NotNil(t, img)

		gp := filepath.Join("testdata", t.Name()+"_"+name+"_golden.png")
		if *update {
			t.Logf("updating golden file %s", gp)
			err = os.WriteFile(gp, img, 0644)
			require.NoError(t, err, "failed to update golden file")
		}

		var golden []byte
		golden, err = os.ReadFile(gp)
		require.NoError(t, err)
		assert.Equal(t, golden, img)
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
