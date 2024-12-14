package cmd

import (
	"context"
	"errors"
	"github.com/clambin/go-common/httputils/metrics"
	"github.com/clambin/go-common/httputils/roundtripper"
	"github.com/clambin/solaredge"
	"github.com/clambin/tado/v2"
	"github.com/clambin/tado/v2/tools"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"log/slog"
	"net/http"
	"time"
)

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

func getHomeId(ctx context.Context, client *tado.ClientWithResponses, logger *slog.Logger) (tado.HomeId, error) {
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
