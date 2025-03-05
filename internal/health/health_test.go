package health

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthProbe(t *testing.T) {
	up := componentFunc(func() error { return nil })
	down := componentFunc(func() error { return errors.New("error") })

	tests := []struct {
		name       string
		components []Component
		want       int
	}{
		{"no probes", nil, http.StatusOK},
		{"up", []Component{up}, http.StatusOK},
		{"down", []Component{down}, http.StatusServiceUnavailable},
		{"partial", []Component{up, down, down}, http.StatusServiceUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := slog.New(slog.DiscardHandler) //slog.New(slog.NewTextHandler(os.Stdout, nil))
			h := Probe(l, tt.components...)

			r, _ := http.NewRequest("GET", "/health", nil)
			w := httptest.NewRecorder()

			h.ServeHTTP(w, r)
			assert.Equal(t, tt.want, w.Code)
		})
	}
}

type componentFunc func() error

func (c componentFunc) IsHealthy() error {
	return c()
}
