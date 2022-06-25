package sampler_test

import (
	"context"
	"errors"
	"github.com/clambin/solaredge-monitor/scrape/sampler"
	"github.com/clambin/solaredge/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSolarEdgeClient(t *testing.T) {
	api := &mocks.API{}
	c := sampler.SolarEdgeSampler{
		API: api,
	}
	ctx := context.Background()

	api.On("GetSiteIDs", mock.AnythingOfType("*context.emptyCtx")).Return([]int{1, 2, 3}, nil).Once()
	api.On("GetPowerOverview", mock.AnythingOfType("*context.emptyCtx"), 1).Return(0.0, 0.0, 0.0, 0.0, 1500.0, nil).Twice()

	sample, err := c.Sample(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1500.0, sample.Value)

	sample, err = c.Sample(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1500.0, sample.Value)

	c = sampler.SolarEdgeSampler{
		API:    api,
		SiteID: 100,
	}

	api.On("GetPowerOverview", mock.AnythingOfType("*context.emptyCtx"), 100).Return(0.0, 0.0, 0.0, 0.0, 2500.0, nil).Once()

	sample, err = c.Sample(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2500.0, sample.Value)

	c = sampler.SolarEdgeSampler{
		API: api,
	}

	api.On("GetSiteIDs", mock.AnythingOfType("*context.emptyCtx")).Return([]int{}, errors.New("fail")).Once()

	sample, err = c.Sample(ctx)
	assert.Error(t, err)

	api.On("GetSiteIDs", mock.AnythingOfType("*context.emptyCtx")).Return([]int{}, nil).Once()
	sample, err = c.Sample(ctx)
	assert.Error(t, err)

	api.On("GetSiteIDs", mock.AnythingOfType("*context.emptyCtx")).Return([]int{1, 2, 3}, nil).Once()
	api.On("GetPowerOverview", mock.AnythingOfType("*context.emptyCtx"), 1).Return(0.0, 0.0, 0.0, 0.0, 0.0, errors.New("fail")).Once()

	sample, err = c.Sample(ctx)
	assert.Error(t, err)

	mock.AssertExpectationsForObjects(t, api)
}
