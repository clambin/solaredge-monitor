package web_test

import (
	"errors"
	"github.com/clambin/solaredge-monitor/internal/repository"
	"github.com/clambin/solaredge-monitor/internal/web"
	"github.com/clambin/solaredge-monitor/internal/web/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestReportsHandler(t *testing.T) {
	tests := []struct {
		name     string
		args     url.Values
		wantCode int
		want     string
	}{
		{
			name: "valid",
			args: url.Values{
				"start": []string{time.Date(2023, time.August, 24, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)},
				"end":   []string{time.Date(2023, time.August, 24, 12, 0, 0, 0, time.UTC).Format(time.RFC3339)},
			},
			wantCode: http.StatusOK,
			want:     `<a href="/plot/scatter?fold=false&end=2023-08-24T12%3A00%3A00Z&start=2023-08-24T00%3A00%3A00Z">`,
		},
		{
			name:     "missing timestamps: redirect",
			wantCode: http.StatusTemporaryRedirect,
		},
		{
			name: "invalid timestamps: bad request",
			args: url.Values{
				"start": []string{"foo"},
				"end":   []string{"bar"},
			},
			wantCode: http.StatusBadRequest,
		},
	}

	r := mocks.NewRepository(t)
	r.EXPECT().GetDataRange().Return(time.Time{}, time.Time{}, nil).Maybe()
	h := web.ReportHandler(r, discardLogger)

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

func TestPlotHandler(t *testing.T) {
	tests := []struct {
		name     string
		target   string
		args     url.Values
		wantCode int
		want     string
	}{
		{
			name:   "all args present",
			target: "/plot/heatmap",
			args: url.Values{
				"start": []string{time.Date(2023, time.August, 24, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)},
				"end":   []string{time.Date(2023, time.August, 24, 12, 0, 0, 0, time.UTC).Format(time.RFC3339)},
				"fold":  []string{"false"},
			},
			wantCode: http.StatusOK,
			want:     `<img src="/plotter/heatmap?end=2023-08-24T12%3A00%3A00Z&fold=false&start=2023-08-24T00%3A00%3A00Z" alt="heatmap"/>`,
		},
		{
			name:     "arg missing: bad request",
			target:   "/plot/scatter",
			args:     url.Values{},
			wantCode: http.StatusBadRequest,
		},
		{
			name:   "invalid timestamps: bad request",
			target: "/plot/scatter",
			args: url.Values{
				"end": []string{"foo"},
			},
			wantCode: http.StatusBadRequest,
		},
	}

	h := web.PlotHandler(discardLogger)
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

func TestPlotterHandler(t *testing.T) {
	tests := []struct {
		name         string
		args         url.Values
		measurements repository.Measurements
		dbErr        error
		wantCode     int
	}{
		{
			name: "valid arguments",
			args: url.Values{
				"start": []string{time.Date(2023, time.August, 24, 0, 0, 0, 0, time.Local).Format(time.RFC3339)},
				"end":   []string{time.Date(2023, time.August, 24, 12, 0, 0, 0, time.Local).Format(time.RFC3339)},
				"fold":  []string{"true"},
			},
			measurements: repository.Measurements{
				{time.Date(2024, time.December, 23, 12, 0, 0, 0, time.UTC), 1000, 65, "SUN"},
				{time.Date(2024, time.December, 23, 12, 15, 0, 0, time.UTC), 1000, 65, "SUN"},
				{time.Date(2024, time.December, 23, 12, 30, 0, 0, time.UTC), 1000, 65, "SUN"},
				{time.Date(2024, time.December, 23, 12, 45, 0, 0, time.UTC), 1000, 65, "SUN"},
				{time.Date(2024, time.December, 23, 13, 0, 0, 0, time.UTC), 1000, 65, "SUN"},
			},
			wantCode: http.StatusOK,
		},
		{
			name: "no data",
			args: url.Values{
				"start": []string{time.Date(2023, time.August, 24, 0, 0, 0, 0, time.Local).Format(time.RFC3339)},
				"end":   []string{time.Date(2023, time.August, 24, 12, 0, 0, 0, time.Local).Format(time.RFC3339)},
				"fold":  []string{"false"},
			},
			wantCode: http.StatusOK,
		},
		{
			name:     "missing arguments",
			args:     url.Values{},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "invalid argument",
			args: url.Values{
				"end": []string{"foo"},
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "db failure",
			args: url.Values{
				"start": []string{time.Date(2023, time.August, 24, 0, 0, 0, 0, time.Local).Format(time.RFC3339)},
				"end":   []string{time.Date(2023, time.August, 24, 12, 0, 0, 0, time.Local).Format(time.RFC3339)},
				"fold":  []string{"true"},
			},
			dbErr:    errors.New("db failure"),
			wantCode: http.StatusInternalServerError,
		},
	}

	r := mocks.NewRepository(t)
	h := web.PlotterHandler(r, "scatter", discardLogger)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantCode != http.StatusBadRequest {
				r.EXPECT().Get(mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return(tt.measurements, tt.dbErr).Once()
			}

			target := url.URL{Path: "/plot/scatter", RawQuery: tt.args.Encode()}
			req, _ := http.NewRequest(http.MethodGet, target.String(), nil)
			resp := httptest.NewRecorder()
			h.ServeHTTP(resp, req)
			assert.Equal(t, tt.wantCode, resp.Code)
		})
	}
}
