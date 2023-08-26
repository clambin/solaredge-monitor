package plot_test

import (
	"errors"
	"github.com/clambin/solaredge-monitor/internal/repository"
	"github.com/clambin/solaredge-monitor/internal/web/handlers/mocks"
	"github.com/clambin/solaredge-monitor/internal/web/handlers/plot"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gonum.org/v1/plot/vg/vgimg"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestPlotHandler(t *testing.T) {
	r := mocks.NewRepository(t)
	p := mocks.NewPlotter(t)
	h := plot.Handler{
		Repository: r,
		Plotter:    p,
		Logger:     slog.Default(),
	}

	m := chi.NewRouter()
	m.Get("/plot/heatmap", h.Handle)

	img := vgimg.PngCanvas{Canvas: vgimg.New(10, 10)}

	testCases := []struct {
		name     string
		args     url.Values
		dbErr    error
		plotErr  error
		wantCode int
	}{
		{
			name:     "default",
			wantCode: http.StatusOK,
		},
		{
			name: "all args",
			args: url.Values{
				"start": []string{time.Date(2023, time.August, 24, 0, 0, 0, 0, time.Local).Format(time.RFC3339)},
				"stop":  []string{time.Date(2023, time.August, 24, 12, 0, 0, 0, time.Local).Format(time.RFC3339)},
				"fold":  []string{"true"},
			},
			wantCode: http.StatusOK,
		},
		{
			name: "invalid argument",
			args: url.Values{
				"stop": []string{"foo"},
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "db failure",
			dbErr:    errors.New("db failure"),
			wantCode: http.StatusInternalServerError,
		},
		{
			name:     "plot failure",
			plotErr:  errors.New("plot failure"),
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {

			if tt.wantCode != http.StatusBadRequest {
				r.EXPECT().Get(mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return(repository.Measurements{}, tt.dbErr).Once()
				if tt.dbErr == nil {
					p.EXPECT().Plot(mock.AnythingOfType("repository.Measurements"), mock.AnythingOfType("bool")).Return(&img, tt.plotErr).Once()
				}
			}

			target := url.URL{Path: "/plot/heatmap", RawQuery: tt.args.Encode()}
			req, _ := http.NewRequest(http.MethodGet, target.String(), nil)
			resp := httptest.NewRecorder()
			m.ServeHTTP(resp, req)
			assert.Equal(t, tt.wantCode, resp.Code)
		})
	}
}
