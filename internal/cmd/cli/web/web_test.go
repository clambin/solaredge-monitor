package web

import (
	"context"
	"github.com/clambin/go-common/charmer"
	"github.com/clambin/solaredge-monitor/internal/testutils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"net/http"
	"strconv"
	"testing"
	"time"
)

func Test_run(t *testing.T) {
	dbenv, ok := testutils.DBEnv()
	if !ok {
		t.SkipNow()
	}
	initViper(viper.GetViper(), dbenv)

	cmd := cobra.Command{}
	ctx, cancel := context.WithCancel(context.Background())
	cmd.SetContext(ctx)
	charmer.SetLogger(&cmd, slog.Default())
	ch := make(chan error)
	go func() {
		ch <- run(&cmd, nil)
	}()

	assert.Eventually(t, func() bool {
		_, err := http.Get("http://localhost" + viper.GetString("web.addr") + "/")
		return err == nil
	}, time.Second, 100*time.Millisecond)

	_, err := http.Get("http://localhost" + viper.GetString("prometheus.addr") + "/metrics")
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
