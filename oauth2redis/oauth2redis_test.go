package oauth2redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"
	"sync/atomic"
	"testing"
	"time"
)

func TestTokenStore(t *testing.T) {
	var c fakeRedisClient
	ts := &TokenStore{
		RedisClient: &c,
		Key:         "token",
		TTL:         1 * time.Minute,
	}

	tok := &oauth2.Token{AccessToken: "abc", Expiry: time.Now().Add(time.Hour)}
	if err := ts.Save(tok); err != nil {
		t.Fatal(err)
	}
	tok2, err := ts.Load()
	if err != nil {
		t.Fatal(err)
	}
	if tok2.AccessToken != tok.AccessToken {
		t.Errorf("access token mismatch: want %q, got %q", tok.AccessToken, tok2.AccessToken)
	}
}

var _ RedisClient = &fakeRedisClient{}

type fakeRedisClient struct {
	value atomic.Value
}

func (f *fakeRedisClient) Set(ctx context.Context, _ string, value any, _ time.Duration) *redis.StatusCmd {
	f.value.Store(string(value.([]byte)))
	return redis.NewStatusCmd(ctx)
}

func (f *fakeRedisClient) Get(ctx context.Context, _ string) *redis.StringCmd {
	cmd := redis.NewStringCmd(ctx)
	value := f.value.Load()
	if value != nil {
		cmd.SetVal(value.(string))
	} else {
		cmd.SetErr(redis.Nil)
	}
	return cmd
}
