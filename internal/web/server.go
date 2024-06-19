package web

import (
	"log/slog"
	"net/http"
)

func New(repo Repository, imageCache *ImageCache, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux, repo, imageCache, logger)
	return mux
}
