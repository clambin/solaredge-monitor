package testutils

import (
	"context"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type PostgresContainer struct {
	*postgres.PostgresContainer
	Port int
}

func NewTestPostgresDB(ctx context.Context, dbName, userName, password string) (*PostgresContainer, error) {
	c, err := postgres.Run(ctx,
		"postgres:16-alpine",
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
		return nil, err
	}

	connectionString, err := c.ConnectionString(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get connection string: %w", err)
	}
	parsedURL, err := url.Parse(connectionString)
	if err != nil {
		return nil, fmt.Errorf("could not parse connection string %q: %w", connectionString, err)
	}
	_, port, ok := strings.Cut(parsedURL.Host, ":")
	if !ok {
		return nil, fmt.Errorf("could not determine port number from connection string %q", connectionString)
	}
	portNr, err := strconv.Atoi(port)
	if err != nil {
		return nil, fmt.Errorf("could not parse port number %q: %w", port, err)
	}
	return &PostgresContainer{PostgresContainer: c, Port: portNr}, nil
}
