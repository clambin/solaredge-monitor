package server_test

import (
	"github.com/clambin/solaredge-monitor/server"
	"github.com/clambin/solaredge-monitor/store/mockdb"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestReporter_GetFirstLast(t *testing.T) {
	reporter := server.Reporter{DB: mockdb.NewDB()}

	_, err := reporter.GetFirst()
	assert.Error(t, err)

	_, err = reporter.GetLast()
	assert.Error(t, err)

	reporter = server.Reporter{DB: mockdb.BuildDB()}

	var timestamp time.Time
	timestamp, err = reporter.GetFirst()
	assert.NoError(t, err)
	assert.NotZero(t, timestamp)

	_, err = reporter.GetLast()
	assert.NoError(t, err)
	assert.NotZero(t, timestamp)
}

func Benchmark_Scatter(b *testing.B) {
	reporter := server.Reporter{DB: mockdb.BuildDB()}

	start, _ := reporter.GetFirst()
	stop, _ := reporter.GetLast()
	assert.NotEqual(b, start, stop)

	for i := 0; i < b.N; i++ {
		_, err := reporter.Plot(server.ScatterPlot, true, start, stop)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Contour(b *testing.B) {
	reporter := server.Reporter{DB: mockdb.BuildDB()}

	start, _ := reporter.GetFirst()
	stop, _ := reporter.GetLast()
	assert.NotEqual(b, start, stop)

	for i := 0; i < b.N; i++ {
		_, err := reporter.Plot(server.ContourPlot, true, start, stop)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_Heatmap(b *testing.B) {
	reporter := server.Reporter{DB: mockdb.BuildDB()}

	start, _ := reporter.GetFirst()
	stop, _ := reporter.GetLast()
	assert.NotEqual(b, start, stop)

	for i := 0; i < b.N; i++ {
		_, err := reporter.Plot(server.HeatmapPlot, true, start, stop)
		if err != nil {
			b.Fatal(err)
		}
	}
}
