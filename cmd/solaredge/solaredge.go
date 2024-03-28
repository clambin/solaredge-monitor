package main

import (
	"github.com/clambin/solaredge-monitor/internal/cmd/cli/export"
	"github.com/clambin/solaredge-monitor/internal/cmd/cli/scrape"
	"github.com/clambin/solaredge-monitor/internal/cmd/cli/web"
	"github.com/clambin/solaredge-monitor/pkg/serpent"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log/slog"
	"os"
	"time"
)

var (
	version = "change_me"

	configFile string
	cmd        = cobra.Command{
		Use:   "solaredge",
		Short: "solaredge metrics collector",
	}
)

func main() {
	if err := cmd.Execute(); err != nil {
		slog.Error("failed to start", "err", err)
		os.Exit(1)
	}
}

var arguments = serpent.Arguments{
	"debug":             {Default: false, Help: "Log debug messages"},
	"prometheus.addr":   {Default: ":9090", Help: "Prometheus metrics endpoint"},
	"database.host":     {Default: "postgres", Help: "Postgres database host"},
	"database.port":     {Default: 5432, Help: "Postgres database port"},
	"database.database": {Default: "solar", Help: "Postgres database name"},
	"database.username": {Default: "solar", Help: "Postgres database username"},
	"database.password": {Default: "", Help: "Postgres database password"},
	"polling.token":     {Default: "", Help: "SolarEdge API token"},
	"polling.interval":  {Default: 5 * time.Minute, Help: "Polling interval"},
	"scrape.interval":   {Default: 15 * time.Minute, Help: "Scraper interval"},
	"tado.username":     {Default: "", Help: "Tado API username"},
	"tado.password":     {Default: "", Help: "Tado API password"},
	"tado.secret":       {Default: "", Help: "Tado API secret"},
	"web.addr":          {Default: ":8080", Help: "Web server address"},
}

func init() {
	cobra.OnInitialize(initConfig)
	cmd.Version = version
	cmd.PersistentFlags().StringVar(&configFile, "config", "", "Configuration file")
	if err := serpent.SetPersistentFlags(&cmd, viper.GetViper(), arguments); err != nil {
		panic(err)
	}
	cmd.AddCommand(&web.Cmd, &export.Cmd, &scrape.Cmd)
}

func initConfig() {
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.AddConfigPath("/etc/solaredge/")
		viper.AddConfigPath("$HOME/.solaredge")
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
	}

	if err := serpent.SetDefaults(viper.GetViper(), arguments); err != nil {
		panic(err)
	}

	viper.SetEnvPrefix("SOLAREDGE")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		slog.Warn("failed to read config file", "err", err)
	}
}
