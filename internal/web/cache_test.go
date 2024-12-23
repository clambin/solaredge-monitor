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

/*
	func TestImageCache(t *testing.T) {
		redisClient := mocks.NewRedisClient(t)
		c := ImageCache{
			Namespace: "github.com/clambin/solaredge-monitor",
			Rounding:  15 * time.Minute,
			TTL:       time.Hour,
			Client:    redisClient,
		}
		ctx := context.Background()
		//	args := arguments{
		//		start: ,
		//		end:   time.Date(2024, time.June, 19, 13, 0, 0, 0, time.UTC),
		//	}
		const wantKey = `github.com/clambin/solaredge-monitor|foo|false|2024-06-19T12:00:00Z|2024-06-19T13:00:00Z`
		var wantValue = []byte("hello world")

		redisClient.EXPECT().Get(ctx, wantKey).RunAndReturn(func(ctx context.Context, _ string) *redis.StringCmd {
			cmd := redis.NewStringCmd(ctx)
			cmd.SetErr(redis.Nil)
			return cmd
		}).Once()

		key := c.getKey(
			"foo",
			time.Date(2024, time.June, 19, 12, 0, 0, 0, time.UTC),
			time.Date(2024, time.June, 19, 13, 0, 0, 0, time.UTC),
			false,
		)
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
*/
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

/*
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
				end:   time.Date(2024, time.June, 19, 13, 0, 0, 0, time.UTC),
			},
			want: `github.com/clambin/solaredge-monitor|foo|false|2024-06-19T12:00:00Z|2024-06-19T13:00:00Z`,
		},
		{
			name: "rounded",
			args: arguments{
				start: time.Date(2024, time.June, 19, 12, 10, 0, 0, time.UTC),
				end:   time.Date(2024, time.June, 19, 13, 0, 10, 0, time.UTC),
			},
			want: `github.com/clambin/solaredge-monitor|foo|false|2024-06-19T12:00:00Z|2024-06-19T13:00:00Z`,
		},
		{
			name: "folded",
			args: arguments{
				start: time.Date(2024, time.June, 19, 12, 10, 0, 0, time.UTC),
				end:   time.Date(2024, time.June, 19, 13, 0, 10, 0, time.UTC),
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
*/
