package html_test

import (
	"github.com/clambin/solaredge-monitor/internal/web/handlers/html"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestReportsHandler(t *testing.T) {
	h := html.ReportHandler{
		Logger: slog.Default(),
	}

	testCases := []struct {
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

	for _, tt := range testCases {
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
