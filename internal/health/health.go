package health

import (
	"errors"
	"log/slog"
	"net/http"
)

type Component interface {
	IsHealthy() error
}

func Probe(logger *slog.Logger, components ...Component) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var errs error
		for _, component := range components {
			if err := component.IsHealthy(); err != nil {
				errs = errors.Join(errs, err)
			}
		}
		if errs != nil {
			logger.Warn("Health check failed", "err", errs)
			http.Error(w, errs.Error(), http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("OK\n"))
	})
}
