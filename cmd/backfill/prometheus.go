package main

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"time"
)

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

	response, warnings, err = feed.client.QueryRange(ctx, query,v1.Range{
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
