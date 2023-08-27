package web_test

import (
	"github.com/clambin/solaredge-monitor/internal/repository"
	"github.com/clambin/solaredge-monitor/internal/web"
	"github.com/clambin/solaredge-monitor/internal/web/handlers/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewHTTPServer(t *testing.T) {
	repo := mocks.NewRepository(t)
	measurements := makeMeasurements(100)
	repo.EXPECT().Get(mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return(measurements, nil)
	s := web.NewHTTPServer(repo, slog.Default())

	testCases := []struct {
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
			wantStatusCode: http.StatusNotFound,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, tt.target, nil)
			resp := httptest.NewRecorder()
			s.Router.ServeHTTP(resp, req)
			assert.Equal(t, tt.wantStatusCode, resp.Code)
		})
	}
}

func makeMeasurements(count int) repository.Measurements {
	measurements := make(repository.Measurements, count)
	timestamp := time.Date(2023, time.August, 25, 0, 0, 0, 0, time.UTC)
	for i := 0; i < count; i++ {
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
	repo := mocks.NewRepository(b)
	measurements := makeMeasurements(100)
	repo.EXPECT().Get(mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return(measurements, nil)
	s := web.NewHTTPServer(repo, slog.Default())

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest(http.MethodGet, "/plot/heatmap", nil)
		resp := httptest.NewRecorder()
		s.Router.ServeHTTP(resp, req)
		if resp.Code != http.StatusOK {
			b.Fail()
		}
	}

}
