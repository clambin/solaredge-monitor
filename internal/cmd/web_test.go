package cmd

import (
	"github.com/clambin/solaredge-monitor/internal/testutils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"net/http"
	"strconv"
	"testing"
	"time"
)

func Test_runWeb(t *testing.T) {
	ctx := t.Context()
	c, connString, err := testutils.NewTestPostgresDB(ctx, "solaredge", "username", "password")
	require.NoError(t, err)
	r, redisPort, err := testutils.NewTestRedis(ctx)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, testcontainers.TerminateContainer(c))
		require.NoError(t, testcontainers.TerminateContainer(r))
	})
	reg := prometheus.NewPedanticRegistry()
	v := getViperFromViper(viper.GetViper())
	v.Set("database.url", connString)
	v.Set("web.cache.addr", "localhost:"+strconv.Itoa(redisPort))
	v.Set("web.addr", ":8080")

	go func() {
		assert.NoError(t, runWeb(ctx, "dev", v, reg, discardLogger))
	}()

	assert.Eventually(t, func() bool {
		_, err := http.Get("http://localhost" + viper.GetString("web.addr") + "/")
		return err == nil
	}, time.Second, 100*time.Millisecond)

	_, err = http.Get("http://localhost" + viper.GetString("prometheus.addr") + "/metrics")
	assert.NoError(t, err)

}
