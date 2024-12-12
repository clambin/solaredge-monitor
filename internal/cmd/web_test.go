package cmd

import (
	"context"
	"github.com/clambin/go-common/charmer"
	"github.com/clambin/solaredge-monitor/internal/testutils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
	"net/http"
	"strconv"
	"testing"
	"time"
)

func Test_run(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	c, err := testutils.NewTestPostgresDB(ctx, "solaredge", "solaredge", "solaredge")
	require.NoError(t, err)
	initViper(viper.GetViper(), map[string]string{
		"pg_host":     "localhost",
		"pg_port":     strconv.Itoa(c.Port),
		"pg_database": "solaredge",
		"pg_user":     "solaredge",
		"pg_password": "solaredge",
	})

	cmd := cobra.Command{}
	cmd.SetContext(ctx)
	charmer.SetLogger(&cmd, slog.Default())
	ch := make(chan error)
	go func() {
		ch <- runWeb(&cmd, nil)
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

func initViper(v *viper.Viper, dbenv map[string]string) {
	v.Set("database.host", dbenv["pg_host"])
	port, _ := strconv.Atoi(dbenv["pg_port"])
	v.Set("database.port", port)
	v.Set("database.database", dbenv["pg_database"])
	v.Set("database.username", dbenv["pg_user"])
	v.Set("database.password", dbenv["pg_password"])
	v.Set("prometheus.addr", ":9090")
	v.Set("web.addr", ":8080")
}
