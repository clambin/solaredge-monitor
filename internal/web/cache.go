package web

import (
	"bytes"
	"context"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type ImageCache struct {
	Namespace string
	Rounding  time.Duration
	TTL       time.Duration
	Client    RedisClient
}

type RedisClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
}

func (c *ImageCache) Get(ctx context.Context, plotType string, args arguments) (string, []byte, error) {
	key := c.getKey(plotType, args)
	content, err := c.Client.Get(ctx, key).Result()
	return key, []byte(content), err
}

func (c *ImageCache) getKey(plotType string, args arguments) string {
	roundedStart := args.start.Truncate(c.Rounding)
	roundedStop := args.stop.Truncate(c.Rounding)
	folded := "false"
	if args.fold {
		folded = "true"
	}
	return strings.Join([]string{
		c.Namespace,
		plotType,
		folded,
		roundedStart.Format(time.RFC3339),
		roundedStop.Format(time.RFC3339),
	}, "|")
}

func (c *ImageCache) Set(ctx context.Context, key string, content []byte) error {
	return c.Client.Set(ctx, key, content, c.TTL).Err()
}

func (c *ImageCache) Middleware(plotType string, logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if c == nil {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			args, err := parseArguments(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			key, img, err := c.Get(r.Context(), plotType, args)
			if err == nil && len(img) > 0 {
				logger.Debug("serving image from cache", "key", key)
				w.Header().Set("ContentType", "image/png")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(img)
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

var _ http.ResponseWriter = &teeResponseWriter{}

type teeResponseWriter struct {
	http.ResponseWriter
	dupe bytes.Buffer
}

func (t *teeResponseWriter) Write(bytes []byte) (int, error) {
	_, _ = t.dupe.Write(bytes)
	return t.ResponseWriter.Write(bytes)
}
