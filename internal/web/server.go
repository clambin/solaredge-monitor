package web

import (
	"log/slog"
	"net/http"
)

func New(repo Repository, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux, repo, logger)
	return mux
}
