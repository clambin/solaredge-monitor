package html_test

import (
	"github.com/clambin/solaredge-monitor/internal/web/handlers/html"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestPlotHandler(t *testing.T) {
	h := html.PlotHandler{
		Logger: slog.Default(),
	}
	r := chi.NewRouter()
	r.Get("/plot/{plotType}", h.Handle)

	testCases := []struct {
		name     string
		target   string
		args     url.Values
		wantCode int
		want     string
	}{
		{
			name:     "default",
			target:   "/plot/scatter",
			wantCode: http.StatusOK,
			want:     "/plotter/scatter?start=",
		},
		{
			name:   "start & stop",
			target: "/plot/heatmap",
			args: url.Values{
				"start": []string{time.Date(2023, time.August, 24, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)},
				"stop":  []string{time.Date(2023, time.August, 24, 12, 0, 0, 0, time.UTC).Format(time.RFC3339)},
				"fold":  []string{"true"},
			},
			wantCode: http.StatusOK,
			want:     "/plotter/heatmap?fold=true&start=2023-08-24T00%3A00%3A00Z&stop=2023-08-24T12%3A00%3A00Z",
		},
		{
			name:   "start",
			target: "/plot/scatter",
			args: url.Values{
				"start": []string{time.Date(2023, time.August, 24, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)},
			},
			wantCode: http.StatusOK,
			want:     "/plotter/scatter?start=2023-08-24T00%3A00%3A00Z",
		},
		{
			name:   "stop",
			target: "/plot/heatmap",
			args: url.Values{
				"stop": []string{time.Date(2023, time.August, 24, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)},
			},
			wantCode: http.StatusOK,
			want:     "/plotter/heatmap?start=0001-01-01T00%3A00%3A00Z&stop=2023-08-24T00%3A00%3A00Z",
		},
		{
			name:   "error",
			target: "/plot/scatter",
			args: url.Values{
				"stop": []string{"foo"},
			},
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			target := url.URL{Path: tt.target, RawQuery: tt.args.Encode()}

			req, _ := http.NewRequest(http.MethodGet, target.String(), nil)
			resp := httptest.NewRecorder()

			r.ServeHTTP(resp, req)
			assert.Equal(t, tt.wantCode, resp.Code)

			if tt.wantCode == http.StatusOK {
				assert.Contains(t, resp.Body.String(), tt.want)
			}
		})
	}
}
