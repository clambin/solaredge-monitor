package cmd

import (
	"github.com/clambin/go-common/charmer"
	"github.com/clambin/go-common/httputils/metrics"
	"github.com/clambin/go-common/httputils/roundtripper"
	"github.com/clambin/solaredge/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log/slog"
	"net/http"
	"time"
)

var (
	configFile string
	RootCmd    = cobra.Command{
		Use:   "solaredge",
		Short: "solaredge metrics collector",
	}

	commonArguments = charmer.Arguments{
		"debug":            {Default: false, Help: "Log debug messages"},
		"pprof":            {Default: "", Help: "Address for pprof endpoint (blank: don't run pprof"},
		"prometheus.addr":  {Default: ":9090", Help: "Prometheus metrics endpoint"},
		"solaredge.token":  {Default: "", Help: "SolarEdge API token"},
		"polling.interval": {Default: 5 * time.Minute, Help: "Polling interval"},
	}

	dbArguments = charmer.Arguments{
		"database.url": {Default: "", Help: "Postgres connection string (postgres://<user>:<password>@<host>:<port>/<dbname>)"},
	}
	webArguments = charmer.Arguments{
		"web.addr":           {Default: ":8080", Help: "Web server address"},
		"web.cache.addr":     {Default: "", Help: "Redis server address"},
		"web.cache.username": {Default: "", Help: "Redis cache username"},
		"web.cache.password": {Default: "", Help: "Redis cache password"},
		"web.cache.rounding": {Default: 15 * time.Minute, Help: "Cache granularity rounding"},
		"web.cache.ttl":      {Default: time.Hour, Help: "Time to cache images"},
	}

	scrapeArguments = charmer.Arguments{
		"scrape.interval":       {Default: 15 * time.Minute, Help: "Scraper interval"},
		"tado.token.path":       {Default: "/data/tado-token.enc", Help: "Location to store the authentication token"},
		"tado.token.passphrase": {Default: "", Help: "passphrase to encrypt the stored authentication token"},
	}
)

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVar(&configFile, "config", "", "Configuration file")
	setFlags(&RootCmd, viper.GetViper(), commonArguments)
	setFlags(&webCmd, viper.GetViper(), dbArguments, webArguments)
	setFlags(&scrapeCmd, viper.GetViper(), dbArguments, scrapeArguments)
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

func setFlags(cmd *cobra.Command, v *viper.Viper, arguments ...charmer.Arguments) {
	for _, args := range arguments {
		if err := charmer.SetPersistentFlags(cmd, v, args); err != nil {
			panic(err)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func newSolarEdgeClient(subsystem string, r prometheus.Registerer, token string) solaredge.Client {
	solarEdgeMetrics := metrics.NewRequestMetrics(metrics.Options{Namespace: "solaredge", Subsystem: subsystem, ConstLabels: prometheus.Labels{"application": "solaredge"}})
	r.MustRegister(solarEdgeMetrics)

	return solaredge.Client{
		SiteKey: token,
		HTTPClient: &http.Client{
			Timeout:   5 * time.Second,
			Transport: roundtripper.New(roundtripper.WithRequestMetrics(solarEdgeMetrics)),
		},
	}
}
