package feeder

import (
	"context"
	"fmt"
	"github.com/clambin/solaredge-monitor/scrape/collector"
	"github.com/prometheus/common/model"
	"time"
)

func GetPowerMetrics(prometheusURL string, start, stop time.Time) (metrics []collector.Metric, err error) {
	return getMetrics(prometheusURL, "avg(solaredge_current_power)", start, stop)
}

func GetIntensityMetrics(prometheusURL string, start, stop time.Time) (metrics []collector.Metric, err error) {
	return getMetrics(prometheusURL, "avg(tado_solar_intensity_percentage)", start, stop)
}

func getMetrics(prometheusURL string, query string, start, stop time.Time) (metrics []collector.Metric, err error) {
	var client *Feed
	client, err = New(prometheusURL)

	if err != nil {
		return nil, fmt.Errorf("cannot connect to prometheus: %s", err.Error())
	}

	var results model.Matrix
	if results, err = client.call(context.Background(), query, start, stop); err == nil {
		for _, entry := range results {
			for _, pair := range entry.Values {
				metrics = append(metrics, collector.Metric{Timestamp: pair.Timestamp.Time(), Value: float64(pair.Value)})
			}
		}
	}
	return
}

func FeedMetrics(power []collector.Metric, intensity []collector.Metric, coll *collector.Collector) (err error) {
	for len(power) > 0 && len(intensity) > 0 {
		if power[0].Timestamp.Before(intensity[0].Timestamp) {
			coll.Power <- power[0]
			power = power[1:]
		} else {
			coll.Intensity <- intensity[0]
			intensity = intensity[1:]
		}
	}

	for len(power) > 0 {
		coll.Power <- power[0]
		power = power[1:]
	}

	for len(intensity) > 0 {
		coll.Intensity <- intensity[0]
		intensity = intensity[1:]
	}
	return
}
