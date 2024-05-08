package web_test

import (
	"github.com/clambin/solaredge-monitor/internal/repository"
	"github.com/clambin/solaredge-monitor/internal/web"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewHTTPServer(t *testing.T) {
	tests := []struct {
		name           string
		target         string
		wantStatusCode int
	}{
		{
			name:           "home",
			target:         "/",
			wantStatusCode: http.StatusSeeOther,
		},
		{
			name:           "report",
			target:         "/report",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "scatter",
			target:         "/plotter/scatter?fold=true",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "heatmap",
			target:         "/plotter/heatmap?fold=false",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "contour",
			target:         "/plotter/contour",
			wantStatusCode: http.StatusSeeOther,
		},
	}

	r := repo{measurements: makeMeasurements(100)}
	s := web.New(r, slog.Default())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			req, _ := http.NewRequest(http.MethodGet, tt.target, nil)
			resp := httptest.NewRecorder()
			s.ServeHTTP(resp, req)
			assert.Equal(t, tt.wantStatusCode, resp.Code)
		})
	}
}

func makeMeasurements(count int) repository.Measurements {
	measurements := make(repository.Measurements, count)
	timestamp := time.Date(2023, time.August, 25, 0, 0, 0, 0, time.UTC)
	for i := range count {
		measurements[i] = repository.Measurement{
			Timestamp: timestamp,
			Power:     1000,
			Intensity: 25,
			Weather:   "SUNNY",
		}
		timestamp = timestamp.Add(time.Minute)
	}
	return measurements
}

func BenchmarkHTTPServer(b *testing.B) {
	r := repo{measurements: makeMeasurements(100)}
	s := web.New(r, slog.Default())

	for range b.N {
		req, _ := http.NewRequest(http.MethodGet, "/plot/heatmap", nil)
		resp := httptest.NewRecorder()
		s.ServeHTTP(resp, req)
		if resp.Code != http.StatusOK {
			b.Fail()
		}
	}
}

var _ web.Repository = repo{}

type repo struct {
	measurements repository.Measurements
}

func (r repo) Get(_, _ time.Time) (repository.Measurements, error) {
	return r.measurements, nil
}
