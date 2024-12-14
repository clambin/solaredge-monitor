package testutils

import (
	"context"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	postgresImage = "postgres:12-alpine"
	redisImage    = "redis:7.4.1"
)

func NewTestPostgresDB(ctx context.Context, dbName, userName, password string) (testcontainers.Container, int, error) {
	c, err := postgres.Run(ctx,
		postgresImage,
		postgres.WithDatabase(dbName),
		postgres.WithUsername(userName),
		postgres.WithPassword(password),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)
	if err != nil {
		return nil, 0, err
	}
	connectionString, err := c.ConnectionString(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("could not get connection string: %w", err)
	}
	port, err := getPortNr(connectionString)
	return c, port, err
}

func NewTestRedis(ctx context.Context) (*redis.RedisContainer, int, error) {
	c, err := redis.Run(ctx, redisImage)
	if err != nil {
		return nil, 0, err
	}
	connectionString, err := c.ConnectionString(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("could not get connection string: %w", err)
	}
	port, err := getPortNr(connectionString)
	return c, port, err
}

func getPortNr(connectionString string) (int, error) {
	parsedURL, err := url.Parse(connectionString)
	if err != nil {
		return 0, fmt.Errorf("could not parse connection string %q: %w", connectionString, err)
	}
	_, port, ok := strings.Cut(parsedURL.Host, ":")
	if !ok {
		return 0, fmt.Errorf("could not determine port number from connection string %q", connectionString)
	}
	portNr, err := strconv.Atoi(port)
	if err != nil {
		return 0, fmt.Errorf("could not parse port number %q: %w", port, err)
	}
	return portNr, nil
}
