package feeder

import (
	"context"
	"fmt"
	"github.com/clambin/solaredge-monitor/scrape/collector"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
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

type Feed struct {
	client v1.API
}

func New(prometheusURL string) (*Feed, error) {
	client, err := api.NewClient(api.Config{Address: prometheusURL})

	if err != nil {
		return nil, err
	}
	return &Feed{client: v1.NewAPI(client)}, nil
}

func (feed *Feed) call(ctx context.Context, query string, start, end time.Time) (result model.Matrix, err error) {
	var (
		response model.Value
		warnings v1.Warnings
	)

	response, warnings, err = feed.client.QueryRange(ctx, query, v1.Range{
		Start: start,
		End:   end,
		Step:  time.Minute,
	})

	if len(warnings) > 0 {
		err = fmt.Errorf("warnings: %v", warnings)
	}

	if err != nil {
		return nil, err
	}

	return response.(model.Matrix), err
}
