package web_test

import (
	"errors"
	"github.com/clambin/solaredge-monitor/internal/repository"
	"github.com/clambin/solaredge-monitor/internal/web"
	mocks2 "github.com/clambin/solaredge-monitor/internal/web/mocks"
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
	tests := []struct {
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

	h := web.PlotHandler(slog.Default())
	r := http.NewServeMux()
	r.Handle("GET /plot/{plotType}", h)

	for _, tt := range tests {
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

func TestReportsHandler(t *testing.T) {
	tests := []struct {
		name     string
		args     url.Values
		wantCode int
		want     string
	}{
		{
			name:     "default",
			wantCode: http.StatusOK,
			want:     "/plotter/scatter?fold=false&",
		},
		{
			name: "start & stop",
			args: url.Values{
				"start": []string{time.Date(2023, time.August, 24, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)},
				"stop":  []string{time.Date(2023, time.August, 24, 12, 0, 0, 0, time.UTC).Format(time.RFC3339)},
			},
			wantCode: http.StatusOK,
			want:     "/plotter/scatter?fold=false&start=2023-08-24T00%3A00%3A00Z&stop=2023-08-24T12%3A00%3A00Z",
		},
		{
			name: "start",
			args: url.Values{
				"start": []string{time.Date(2023, time.August, 24, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)},
			},
			wantCode: http.StatusOK,
			want:     "/plotter/scatter?fold=false&start=2023-08-24T00%3A00%3A00Z",
		},
		{
			name: "stop",
			args: url.Values{
				"stop": []string{time.Date(2023, time.August, 24, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)},
			},
			wantCode: http.StatusOK,
			want:     "/plotter/scatter?fold=false&start=0001-01-01T00%3A00%3A00Z&stop=2023-08-24T00%3A00%3A00Z",
		},
		{
			name: "error",
			args: url.Values{
				"stop": []string{"foo"},
			},
			wantCode: http.StatusBadRequest,
		},
	}

	h := web.ReportHandler(slog.Default())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := url.URL{Path: "/", RawQuery: tt.args.Encode()}

			req, _ := http.NewRequest(http.MethodGet, target.String(), nil)
			resp := httptest.NewRecorder()

			h.ServeHTTP(resp, req)
			assert.Equal(t, tt.wantCode, resp.Code)

			if tt.wantCode == http.StatusOK {
				assert.Contains(t, resp.Body.String(), tt.want)
			}
		})
	}
}

func TestPlotterHandler(t *testing.T) {
	tests := []struct {
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

	r := mocks2.NewRepository(t)
	p := mocks2.NewPlotter(t)
	h := web.PlotterHandler(r, p, slog.Default())

	img := vgimg.PngCanvas{Canvas: vgimg.New(10, 10)}

	for _, tt := range tests {
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
			h.ServeHTTP(resp, req)
			assert.Equal(t, tt.wantCode, resp.Code)
		})
	}
}
