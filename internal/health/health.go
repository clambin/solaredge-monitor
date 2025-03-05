package health

import (
	"context"
	"log/slog"
	"net/http"
)

type Component interface {
	IsHealthy(context.Context) error
}

type IsHealthyFunc func(context.Context) error

func (c IsHealthyFunc) IsHealthy(ctx context.Context) error {
	return c(ctx)
}

func Probe(logger *slog.Logger, components ...Component) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, component := range components {
			if err := component.IsHealthy(r.Context()); err != nil {
				logger.Warn("health check failed", "err", err)
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
		}
		w.WriteHeader(http.StatusOK)
	})
}
