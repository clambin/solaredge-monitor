package publisher

import (
	"context"
	solaredge "github.com/clambin/solaredge/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestSolarEdgeUpdater_GetUpdate(t *testing.T) {
	u := SolarEdgeUpdater{SolarEdgeClient: fakeSolarEdgeClient{}}

	update, err := u.GetUpdate(context.Background())
	require.NoError(t, err)
	require.Len(t, update, 1)
	want := SiteUpdate{
		ID:   1,
		Name: "my home",
		PowerOverview: solaredge.PowerOverview{
			LastUpdateTime: solaredge.Time{},
			LifeTimeData:   solaredge.EnergyOverview{Energy: 1000, Revenue: 0},
			LastYearData:   solaredge.EnergyOverview{Energy: 100, Revenue: 0},
			LastMonthData:  solaredge.EnergyOverview{Energy: 10, Revenue: 0},
			LastDayData:    solaredge.EnergyOverview{Energy: 1, Revenue: 0},
			CurrentPower:   solaredge.CurrentPower{Power: 100},
		},
		InverterUpdates: []InverterUpdate{
			{
				Name:         "foo",
				SerialNumber: "1234",
				Telemetry: solaredge.InverterTelemetry{
					L1Data:    solaredge.InverterTelemetryL1Data{AcCurrent: 1, AcVoltage: 220},
					DcVoltage: 380,
				},
			},
		},
	}
	assert.Equal(t, want, update[0])
}

var _ SolarEdgeClient = fakeSolarEdgeClient{}

type fakeSolarEdgeClient struct{}

func (f fakeSolarEdgeClient) GetSites(_ context.Context) (solaredge.GetSitesResponse, error) {
	var response solaredge.GetSitesResponse
	response.Sites.Site = []solaredge.SiteDetails{{Id: 1, Name: "my home"}}
	response.Sites.Count = len(response.Sites.Site)
	return response, nil
}

func (f fakeSolarEdgeClient) GetPowerOverview(_ context.Context, _ int) (solaredge.GetPowerOverviewResponse, error) {
	return solaredge.GetPowerOverviewResponse{
		Overview: solaredge.PowerOverview{
			LifeTimeData:  solaredge.EnergyOverview{Energy: 1000},
			LastYearData:  solaredge.EnergyOverview{Energy: 100},
			LastMonthData: solaredge.EnergyOverview{Energy: 10},
			LastDayData:   solaredge.EnergyOverview{Energy: 1},
			CurrentPower:  solaredge.CurrentPower{Power: 100},
		},
	}, nil
}

func (f fakeSolarEdgeClient) GetComponents(_ context.Context, id int) (solaredge.GetComponentsResponse, error) {
	var responses = map[int][]solaredge.Inverter{
		1: {{Name: "foo", SerialNumber: "1234"}},
	}
	var response solaredge.GetComponentsResponse
	response.Reporters.List = responses[id]
	response.Reporters.Count = len(response.Reporters.List)
	return response, nil
}

func (f fakeSolarEdgeClient) GetInverterTechnicalData(_ context.Context, _ int, serialNr string, _ time.Time, _ time.Time) (solaredge.GetInverterTechnicalDataResponse, error) {
	var responses = map[string][]solaredge.InverterTelemetry{
		"1234": {
			{
				L1Data:    solaredge.InverterTelemetryL1Data{AcCurrent: 1, AcVoltage: 220},
				DcVoltage: 380,
			},
		},
	}
	var response solaredge.GetInverterTechnicalDataResponse
	response.Data.Telemetries = responses[serialNr]
	response.Data.Count = len(response.Data.Telemetries)
	return response, nil
}
