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

func TestImageCache(t *testing.T) {
	redisClient := mocks.NewRedisClient(t)
	c := ImageCache{
		Namespace: "github.com/clambin/solaredge-monitor",
		Rounding:  15 * time.Minute,
		TTL:       time.Hour,
		Client:    redisClient,
	}
	ctx := context.Background()
	args := arguments{
		start: time.Date(2024, time.June, 19, 12, 0, 0, 0, time.UTC),
		stop:  time.Date(2024, time.June, 19, 13, 0, 0, 0, time.UTC),
	}
	const wantKey = `github.com/clambin/solaredge-monitor|foo|false|2024-06-19T12:00:00Z|2024-06-19T13:00:00Z`
	var wantValue = []byte("hello world")

	redisClient.EXPECT().Get(ctx, wantKey).RunAndReturn(func(ctx context.Context, _ string) *redis.StringCmd {
		cmd := redis.NewStringCmd(ctx)
		cmd.SetErr(redis.Nil)
		return cmd
	}).Once()

	key, _, err := c.Get(ctx, "foo", args)
	assert.ErrorIs(t, err, redis.Nil)
	assert.Equal(t, wantKey, key)

	redisClient.EXPECT().Set(ctx, wantKey, wantValue, c.TTL).RunAndReturn(func(ctx context.Context, _ string, _ interface{}, _ time.Duration) *redis.StatusCmd {
		return redis.NewStatusCmd(ctx)
	}).Once()

	assert.NoError(t, c.Set(ctx, key, wantValue))

	redisClient.EXPECT().Get(ctx, wantKey).RunAndReturn(func(ctx context.Context, _ string) *redis.StringCmd {
		cmd := redis.NewStringCmd(ctx)
		cmd.SetErr(nil)
		cmd.SetVal(string(wantValue))
		return cmd
	}).Once()

	_, value, err := c.Get(ctx, "foo", args)
	assert.NoError(t, err)
	assert.Equal(t, wantValue, value)

}

func TestImageCache_Middleware(t *testing.T) {
	redisClient := mocks.NewRedisClient(t)
	c := ImageCache{Client: redisClient, TTL: time.Hour}

	f := c.Middleware("scatter", slog.Default())(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("hello world"))
	}))

	ctx := context.Background()
	const wantKey = "|scatter|false|0001-01-01T00:00:00Z|0001-01-01T00:00:00Z"
	redisClient.EXPECT().Get(ctx, wantKey).RunAndReturn(func(ctx context.Context, _ string) *redis.StringCmd {
		cmd := redis.NewStringCmd(ctx)
		cmd.SetErr(redis.Nil)
		return cmd
	}).Once()
	redisClient.EXPECT().Set(ctx, wantKey, []byte("hello world"), time.Hour).RunAndReturn(func(ctx context.Context, _ string, _ interface{}, _ time.Duration) *redis.StatusCmd {
		cmd := redis.NewStatusCmd(ctx)
		return cmd
	})

	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/foo", nil)
	w := httptest.NewRecorder()
	f.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "hello world", w.Body.String())

	redisClient.EXPECT().Get(ctx, wantKey).RunAndReturn(func(ctx context.Context, _ string) *redis.StringCmd {
		cmd := redis.NewStringCmd(ctx)
		cmd.SetVal("hello world")
		return cmd
	}).Once()

	r, _ = http.NewRequestWithContext(ctx, http.MethodGet, "/foo", nil)
	w = httptest.NewRecorder()
	f.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "hello world", w.Body.String())

}

func TestImageCache_getKey(t *testing.T) {
	tests := []struct {
		name string
		args arguments
		want string
	}{
		{
			name: "baseline",
			args: arguments{
				start: time.Date(2024, time.June, 19, 12, 0, 0, 0, time.UTC),
				stop:  time.Date(2024, time.June, 19, 13, 0, 0, 0, time.UTC),
			},
			want: `github.com/clambin/solaredge-monitor|foo|false|2024-06-19T12:00:00Z|2024-06-19T13:00:00Z`,
		},
		{
			name: "rounded",
			args: arguments{
				start: time.Date(2024, time.June, 19, 12, 10, 0, 0, time.UTC),
				stop:  time.Date(2024, time.June, 19, 13, 0, 10, 0, time.UTC),
			},
			want: `github.com/clambin/solaredge-monitor|foo|false|2024-06-19T12:00:00Z|2024-06-19T13:00:00Z`,
		},
		{
			name: "folded",
			args: arguments{
				start: time.Date(2024, time.June, 19, 12, 10, 0, 0, time.UTC),
				stop:  time.Date(2024, time.June, 19, 13, 0, 10, 0, time.UTC),
				fold:  true,
			},
			want: `github.com/clambin/solaredge-monitor|foo|true|2024-06-19T12:00:00Z|2024-06-19T13:00:00Z`,
		},
	}
	c := ImageCache{
		Namespace: "github.com/clambin/solaredge-monitor",
		Rounding:  15 * time.Minute,
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, c.getKey("foo", tt.args))
		})
	}
}
