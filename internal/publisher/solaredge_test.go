package publisher

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/clambin/solaredge"
	solaredge2 "github.com/clambin/solaredge-monitor/internal/publisher/solaredge"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestSolarEdgeUpdater_GetUpdate(t *testing.T) {
	c := solaredge.Client{
		HTTPClient: &http.Client{
			Transport: &fakeSolarEdgeServer{},
		},
	}

	u := SolarEdgeUpdater{SolarEdge: c}

	update, err := u.GetUpdate(context.Background())
	require.NoError(t, err)
	require.Len(t, update, 1)
	want := solaredge2.SiteUpdate{
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
		InverterUpdates: []solaredge2.InverterUpdate{
			{
				Name:         "foo",
				SerialNumber: "1234",
				Telemetry: solaredge.InverterTelemetry{
					L1Data: struct {
						AcCurrent     float64 "json:\"acCurrent\""
						AcFrequency   float64 "json:\"acFrequency\""
						AcVoltage     float64 "json:\"acVoltage\""
						ActivePower   float64 "json:\"activePower\""
						ApparentPower float64 "json:\"apparentPower\""
						CosPhi        float64 "json:\"cosPhi\""
						ReactivePower float64 "json:\"reactivePower\""
					}{AcCurrent: 1, AcVoltage: 220},
					DcVoltage: 380,
				},
			},
		},
	}
	assert.Equal(t, want, update[0])
}

func Test_fakeSolarEdgeServer(t *testing.T) {
	c := solaredge.Client{
		HTTPClient: &http.Client{
			Transport: &fakeSolarEdgeServer{},
		},
	}
	ctx := context.Background()
	sites, err := c.GetSites(ctx)
	require.NoError(t, err)
	require.Len(t, sites, 1)

	overview, err := sites[0].GetPowerOverview(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1000.0, overview.LifeTimeData.Energy)

	inverters, err := sites[0].GetInverters(ctx)
	require.NoError(t, err)
	require.Len(t, inverters, 1)
	assert.Equal(t, "foo", inverters[0].Name)

	telemetry, err := inverters[0].GetTelemetry(ctx, time.Now().Add(-5*time.Minute), time.Now())
	require.NoError(t, err)
	require.Len(t, telemetry, 1)
	assert.Equal(t, 380.0, telemetry[0].DcVoltage)
	assert.Equal(t, 220.0, telemetry[0].L1Data.AcVoltage)
	assert.Equal(t, 1.0, telemetry[0].L1Data.AcCurrent)
}

type fakeSolarEdgeServer struct{}

func (f fakeSolarEdgeServer) RoundTrip(r *http.Request) (*http.Response, error) {
	response, ok := responses[r.URL.Path]
	if !ok {
		return &http.Response{StatusCode: http.StatusNotFound, Status: "unknown path: " + r.URL.Path, Body: http.NoBody}, nil
	}
	body, err := json.Marshal(response)
	if err != nil {
		return &http.Response{StatusCode: http.StatusInternalServerError, Status: "failed to marshal response:" + err.Error()}, nil
	}
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBuffer(body))}, err
}

var responses = map[string]any{
	"/sites/list": struct {
		Sites struct {
			Count int              `json:"count"`
			Site  []solaredge.Site `json:"site"`
		} `json:"sites"`
	}{
		Sites: struct {
			Count int              `json:"count"`
			Site  []solaredge.Site `json:"site"`
		}{
			Count: 1,
			Site:  []solaredge.Site{{ID: 1, Name: "my home"}},
		}},

	"/site/1/overview": struct {
		Overview solaredge.PowerOverview `json:"overview"`
	}{
		Overview: solaredge.PowerOverview{
			LifeTimeData:  solaredge.EnergyOverview{Energy: 1000},
			LastYearData:  solaredge.EnergyOverview{Energy: 100},
			LastMonthData: solaredge.EnergyOverview{Energy: 10},
			LastDayData:   solaredge.EnergyOverview{Energy: 1},
			CurrentPower:  solaredge.CurrentPower{Power: 100},
		},
	},

	"/equipment/1/list": struct {
		Reporters struct {
			Count int                  `json:"count"`
			List  []solaredge.Inverter `json:"list"`
		} `json:"reporters"`
	}{
		Reporters: struct {
			Count int                  `json:"count"`
			List  []solaredge.Inverter `json:"list"`
		}{
			Count: 1,
			List:  []solaredge.Inverter{{Name: "foo", SerialNumber: "1234"}},
		},
	},

	"/equipment/1/1234/data": struct {
		Data struct {
			Count       int                           `json:"count"`
			Telemetries []solaredge.InverterTelemetry `json:"telemetries"`
		} `json:"data"`
	}{
		Data: struct {
			Count       int                           `json:"count"`
			Telemetries []solaredge.InverterTelemetry `json:"telemetries"`
		}{
			Count: 1,
			Telemetries: []solaredge.InverterTelemetry{
				{
					L1Data: struct {
						AcCurrent     float64 `json:"acCurrent"`
						AcFrequency   float64 `json:"acFrequency"`
						AcVoltage     float64 `json:"acVoltage"`
						ActivePower   float64 `json:"activePower"`
						ApparentPower float64 `json:"apparentPower"`
						CosPhi        float64 `json:"cosPhi"`
						ReactivePower float64 `json:"reactivePower"`
					}{
						AcCurrent: 1,
						AcVoltage: 220,
					},
					DcVoltage: 380,
				},
			},
		},
	},
}
