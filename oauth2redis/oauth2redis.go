package oauth2redis

import (
	"context"
	"encoding/json"
	"github.com/clambin/tado/v2/oauth2store"
	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"
	"time"
)

var _ oauth2store.TokenStore = &tokenStore{}

type tokenStore struct {
	RedisClient
	Key string
	TTL time.Duration
}

type RedisClient interface {
	Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
}

func NewRedisTokenStore(redisClient RedisClient, key string, ttl time.Duration) oauth2store.TokenStore {
	return &tokenStore{
		RedisClient: redisClient,
		Key:         key,
		TTL:         ttl,
	}
}

func (t tokenStore) Save(token *oauth2.Token) error {
	bytes, err := json.Marshal(token)
	if err == nil {
		err = t.RedisClient.Set(context.Background(), t.Key, bytes, t.TTL).Err()
	}
	return err
}

func (t tokenStore) Load() (*oauth2.Token, error) {
	result := t.RedisClient.Get(context.Background(), t.Key)
	if result.Err() != nil {
		return nil, result.Err()
	}
	var token oauth2.Token
	err := json.Unmarshal([]byte(result.Val()), &token)
	return &token, err
}
