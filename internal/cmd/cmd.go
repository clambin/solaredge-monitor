package cmd

import (
	"context"
	"errors"
	"github.com/clambin/go-common/charmer"
	"github.com/clambin/go-common/httputils/metrics"
	"github.com/clambin/go-common/httputils/roundtripper"
	"github.com/clambin/solaredge"
	"github.com/clambin/tado/v2"
	"github.com/clambin/tado/v2/tools"
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
		"scrape.interval": {Default: 15 * time.Minute, Help: "Scraper interval"},
		"tado.username":   {Default: "", Help: "Tado API username"},
		"tado.password":   {Default: "", Help: "Tado API password"},
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

func newSolarEdgeClient(subsystem string, r prometheus.Registerer, v *viper.Viper) solaredge.Client {
	solarEdgeMetrics := metrics.NewRequestMetrics(metrics.Options{Namespace: "solaredge", Subsystem: subsystem, ConstLabels: prometheus.Labels{"application": "solaredge"}})
	r.MustRegister(solarEdgeMetrics)

	return solaredge.Client{
		Token: v.GetString("polling.token"),
		HTTPClient: &http.Client{
			Timeout:   5 * time.Second,
			Transport: roundtripper.New(roundtripper.WithRequestMetrics(solarEdgeMetrics)),
		},
	}
}

func newTadoClient(ctx context.Context, r prometheus.Registerer, v *viper.Viper) (*tado.ClientWithResponses, error) {
	tadoMetrics := metrics.NewRequestMetrics(metrics.Options{Namespace: "solaredge", Subsystem: "scraper", ConstLabels: prometheus.Labels{"application": "tado"}})
	r.MustRegister(tadoMetrics)

	tadoHttpClient, err := tado.NewOAuth2Client(ctx, v.GetString("tado.username"), v.GetString("tado.password"))
	if err != nil {
		return nil, err
	}
	origTP := tadoHttpClient.Transport
	tadoHttpClient.Transport = roundtripper.New(
		roundtripper.WithRequestMetrics(tadoMetrics),
		roundtripper.WithRoundTripper(origTP),
	)
	return tado.NewClientWithResponses(tado.ServerURL, tado.WithHTTPClient(tadoHttpClient))
}

func getHomeId(ctx context.Context, client tools.TadoClient, logger *slog.Logger) (tado.HomeId, error) {
	homes, err := tools.GetHomes(ctx, client)
	if err != nil {
		return 0, err
	}
	if len(homes) == 0 {
		return 0, errors.New("no Tado Homes found")
	}
	homeId := *homes[0].Id
	if len(homes) > 1 {
		logger.Warn("Tado account has more than one home registered. Using first one", "homeId", homeId)
	}
	return homeId, nil
}
