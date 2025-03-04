package cmd

import (
	"context"
	"fmt"
	"github.com/clambin/go-common/httputils/metrics"
	"github.com/clambin/go-common/httputils/roundtripper"
	"github.com/clambin/solaredge-monitor/oauth2redis"
	"github.com/clambin/solaredge/v2"
	"github.com/clambin/tado/v2"
	"github.com/clambin/tado/v2/oauth2store"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"net/http"
	"time"
)

func newSolarEdgeClient(subsystem string, r prometheus.Registerer, v *viper.Viper) solaredge.Client {
	solarEdgeMetrics := metrics.NewRequestMetrics(metrics.Options{Namespace: "solaredge", Subsystem: subsystem, ConstLabels: prometheus.Labels{"application": "solaredge"}})
	r.MustRegister(solarEdgeMetrics)

	return solaredge.Client{
		SiteKey: v.GetString("solaredge.token"),
		HTTPClient: &http.Client{
			Timeout:   5 * time.Second,
			Transport: roundtripper.New(roundtripper.WithRequestMetrics(solarEdgeMetrics)),
		},
	}
}

func newRedisClient(v *viper.Viper) *redis.Client {
	var redisClient *redis.Client
	if addr := v.GetString("redis.addr"); addr != "" {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     addr,
			Username: v.GetString("redis.username"),
			Password: v.GetString("redis.password"),
			DB:       v.GetInt("redis.db"),
		})
	}
	return redisClient
}

func newTadoClient(ctx context.Context, r prometheus.Registerer, redisClient *redis.Client) (*tado.ClientWithResponses, error) {
	tadoMetrics := metrics.NewRequestMetrics(metrics.Options{Namespace: "solaredge", Subsystem: "scraper", ConstLabels: prometheus.Labels{"application": "tado"}})
	r.MustRegister(tadoMetrics)

	tadoHttpClient, err := newOAuth2Client(ctx, redisClient, func(response *oauth2.DeviceAuthResponse) {
		fmt.Printf("No token found. Visit %s and log in ...\n", response.VerificationURIComplete)
	})
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

func newOAuth2Client(ctx context.Context, redisClient *redis.Client, deviceAuthCallback func(response *oauth2.DeviceAuthResponse)) (client *http.Client, err error) {
	// store to save our token
	store := oauth2redis.TokenStore{
		Key:         "github.com/clambin/solaredge-monitor/oauth2redis",
		RedisClient: redisClient,
		TTL:         30 * 24 * time.Hour,
	}
	token, err := store.Load()
	if err != nil {
		// store doesn't contain a valid token. ask the user to log in
		var devAuthResponse *oauth2.DeviceAuthResponse
		if devAuthResponse, err = tado.Config.DeviceAuth(ctx); err != nil {
			return nil, fmt.Errorf("DevAuth: %w", err)
		}
		deviceAuthCallback(devAuthResponse)
		if token, err = tado.Config.DeviceAccessToken(ctx, devAuthResponse); err != nil {
			return nil, fmt.Errorf("DeviceAccessToken: %w", err)
		}
	}
	pts := oauth2store.TokenSource{
		TokenSource: tado.Config.TokenSource(ctx, token),
		TokenStore:  store,
	}
	return oauth2.NewClient(ctx, &pts), nil
}
