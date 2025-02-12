package web

import (
	"context"
	"github.com/clambin/solaredge-monitor/internal/web/mocks"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestImageCache_Middleware(t *testing.T) {
	ctx := context.Background()
	const wantKey = "|scatter|false|2024-11-22T00:00:00Z|2024-12-22T00:00:00Z"
	const response = "hello world"

	tests := []struct {
		name           string
		redisClient    func(*testing.T) *mocks.RedisClient
		args           string
		wantStatusCode int
		wantBody       string
	}{
		{
			name: "miss",
			redisClient: func(t *testing.T) *mocks.RedisClient {
				c := mocks.NewRedisClient(t)
				c.EXPECT().
					Get(ctx, wantKey).
					RunAndReturn(func(ctx context.Context, _ string) *redis.StringCmd {
						cmd := redis.NewStringCmd(ctx)
						cmd.SetErr(redis.Nil)
						return cmd
					}).
					Once()
				c.EXPECT().
					Set(ctx, wantKey, []byte(response), time.Hour).
					RunAndReturn(func(ctx context.Context, _ string, _ interface{}, _ time.Duration) *redis.StatusCmd {
						cmd := redis.NewStatusCmd(ctx)
						return cmd
					}).
					Once()
				return c
			},
			args:           "?start=2024-11-22&end=2024-12-22&fold=false",
			wantStatusCode: http.StatusOK,
			wantBody:       response,
		},
		{
			name: "hit",
			redisClient: func(t *testing.T) *mocks.RedisClient {
				c := mocks.NewRedisClient(t)
				c.EXPECT().
					Get(ctx, wantKey).
					RunAndReturn(func(ctx context.Context, _ string) *redis.StringCmd {
						cmd := redis.NewStringCmd(ctx)
						cmd.SetVal(response)
						return cmd
					}).
					Once()
				return c
			},
			args:           "?start=2024-11-22&end=2024-12-22&fold=false",
			wantStatusCode: http.StatusOK,
			wantBody:       response,
		},
		{
			name: "invalid args",
			redisClient: func(t *testing.T) *mocks.RedisClient {
				return mocks.NewRedisClient(t)
			},
			args:           "?start=foo",
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rd := tt.redisClient(t)
			c := ImageCache{Client: rd, TTL: time.Hour}

			f := c.Middleware("scatter", slog.Default())(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(response))
			}))

			r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost/scatter"+tt.args, nil)
			w := httptest.NewRecorder()
			f.ServeHTTP(w, r)
			assert.Equal(t, tt.wantStatusCode, w.Code)
			if w.Code == http.StatusOK {
				assert.Equal(t, response, w.Body.String())
			}
		})
	}
}
