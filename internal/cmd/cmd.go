package cmd

import (
	"context"
	"github.com/clambin/go-common/charmer"
	"github.com/clambin/solaredge-monitor/internal/scraper"
	"github.com/clambin/solaredge-monitor/internal/scraper/solaredge"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log/slog"
	"time"
)

type Poller interface {
	Run(context.Context) error
	scraper.Publisher[solaredge.Update]
}

var (
	configFile string
	RootCmd    = cobra.Command{
		Use:   "solaredge",
		Short: "solaredge metrics collector",
		PreRun: func(cmd *cobra.Command, args []string) {
			charmer.SetTextLogger(cmd, viper.GetBool("debug"))
		},
	}

	commonArguments = charmer.Arguments{
		"debug":              {Default: false, Help: "Log debug messages"},
		"prometheus.addr":    {Default: ":9090", Help: "Prometheus metrics endpoint"},
		"database.host":      {Default: "postgres", Help: "Postgres database host"},
		"database.port":      {Default: 5432, Help: "Postgres database port"},
		"database.database":  {Default: "solar", Help: "Postgres database name"},
		"database.username":  {Default: "solar", Help: "Postgres database username"},
		"database.password":  {Default: "", Help: "Postgres database password"},
		"polling.token":      {Default: "", Help: "SolarEdge API token"},
		"polling.interval":   {Default: 5 * time.Minute, Help: "Polling interval"},
		"scrape.interval":    {Default: 15 * time.Minute, Help: "Scraper interval"},
		"tado.username":      {Default: "", Help: "Tado API username"},
		"tado.password":      {Default: "", Help: "Tado API password"},
		"tado.secret":        {Default: "", Help: "Tado API secret"},
		"web.addr":           {Default: ":8080", Help: "Web server address"},
		"web.cache.addr":     {Default: "", Help: "Redis server address"},
		"web.cache.username": {Default: "", Help: "Redis cache username"},
		"web.cache.password": {Default: "", Help: "Redis cache password"},
		"web.cache.rounding": {Default: 15 * time.Minute, Help: "Cache granularity rounding"},
		"web.cache.ttl":      {Default: time.Hour, Help: "Time to cache images"},
	}
)

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVar(&configFile, "config", "", "Configuration file")
	if err := charmer.SetPersistentFlags(&RootCmd, viper.GetViper(), commonArguments); err != nil {
		panic(err)
	}
	RootCmd.AddCommand(&webCmd, &exportCmd, &scrapeCmd)
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

	viper.SetEnvPrefix("SOLAREDGE")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		slog.Warn("failed to read config file", "err", err)
	}
}
