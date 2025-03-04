package web

import (
	"bytes"
	"context"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ImageCache struct {
	Client    RedisClient
	Namespace string
	Rounding  time.Duration
	TTL       time.Duration
}

type RedisClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd
}

func (c *ImageCache) Middleware(plotType string, logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if c == nil {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start, end, fold, err := parsePlotterArguments(r)
			if err != nil {
				http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
				return
			}

			key := c.getKey(plotType, start, end, fold)
			var content []byte
			if content, err = c.Client.Get(r.Context(), key).Bytes(); err == nil && len(content) > 0 {
				logger.Debug("serving image from cache", "key", key)
				w.Header().Set("ContentType", "image/png")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(content)
				return
			}

			duper := &teeResponseWriter{ResponseWriter: w}
			next.ServeHTTP(duper, r)

			if duper.dupe.Len() > 0 {
				if err = c.Set(r.Context(), key, duper.dupe.Bytes()); err != nil {
					logger.Error("failed to cache image", "key", key, "err", err)
				}
			}
		})
	}
}

func (c *ImageCache) getKey(plotType string, start, end time.Time, fold bool) string {
	return strings.Join([]string{
		c.Namespace,
		plotType,
		strconv.FormatBool(fold),
		start.Truncate(c.Rounding).Format(time.RFC3339),
		end.Truncate(c.Rounding).Format(time.RFC3339),
	}, "|")
}

func (c *ImageCache) Set(ctx context.Context, key string, content []byte) error {
	return c.Client.Set(ctx, key, content, c.TTL).Err()
}

var _ http.ResponseWriter = &teeResponseWriter{}

type teeResponseWriter struct {
	http.ResponseWriter
	dupe bytes.Buffer
}

func (t *teeResponseWriter) Write(bytes []byte) (int, error) {
	_, _ = t.dupe.Write(bytes)
	return t.ResponseWriter.Write(bytes)
}
