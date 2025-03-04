package web_test

import (
	"errors"
	"github.com/clambin/solaredge-monitor/internal/repository"
	"github.com/clambin/solaredge-monitor/internal/web"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var discardLogger = slog.New(slog.DiscardHandler)

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
			wantStatusCode: http.StatusTemporaryRedirect,
		},
		{
			name:           "scatter",
			target:         "/plotter/scatter?fold=true",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "heatmap",
			target:         "/plotter/heatmap?fold=false",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "contour",
			target:         "/plotter/contour",
			wantStatusCode: http.StatusSeeOther,
		},
	}

	r := repo{measurements: makeMeasurements(100)}
	s := web.New(r, nil, slog.Default())

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

var _ web.Repository = repo{}

type repo struct {
	measurements repository.Measurements
}

func (r repo) GetDataRange() (time.Time, time.Time, error) {
	if len(r.measurements) == 0 {
		return time.Time{}, time.Time{}, errors.New("no data")
	}
	return r.measurements[0].Timestamp, r.measurements[len(r.measurements)-1].Timestamp, nil
}

func (r repo) Get(_, _ time.Time) (repository.Measurements, error) {
	return r.measurements, nil
}
