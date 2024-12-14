package cmd

import (
	"context"
	"github.com/clambin/solaredge-monitor/internal/testutils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"net/http"
	"testing"
	"time"
)

func Test_runWeb(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	c, dbPort, err := testutils.NewTestPostgresDB(ctx, "solaredge", "username", "password")
	require.NoError(t, err)
	r, redisPort, err := testutils.NewTestRedis(ctx)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, testcontainers.TerminateContainer(c))
		require.NoError(t, testcontainers.TerminateContainer(r))
	})
	reg := prometheus.NewPedanticRegistry()
	v := getViperFromViper(viper.GetViper())
	initViperDB(v, dbPort)
	initViperCache(v, redisPort)

	ch := make(chan error)
	go func() {
		ch <- runWeb(ctx, "dev", v, reg, discardLogger)
	}()

	assert.Eventually(t, func() bool {
		_, err := http.Get("http://localhost" + viper.GetString("web.addr") + "/")
		return err == nil
	}, time.Second, 100*time.Millisecond)

	_, err = http.Get("http://localhost" + viper.GetString("prometheus.addr") + "/metrics")
	assert.NoError(t, err)
	cancel()
	assert.NoError(t, <-ch)
}
