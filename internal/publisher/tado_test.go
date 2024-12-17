package publisher

import (
	"context"
	"errors"
	"github.com/clambin/tado/v2"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestTadoUpdater_GetUpdate(t *testing.T) {
	type want struct {
		temperature float32
		intensity   float32
		weather     tado.WeatherState
	}
	tests := []struct {
		name string
		resp fakeWeatherGetter
		err  assert.ErrorAssertionFunc
		want
	}{
		{
			name: "success",
			resp: fakeWeatherGetter{resp: &tado.GetWeatherResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusOK},
				JSON200: &tado.Weather{
					OutsideTemperature: &tado.TemperatureDataPoint{Celsius: varP(float32(18))},
					SolarIntensity:     &tado.PercentageDataPoint{Percentage: varP(float32(25))},
					WeatherState:       &tado.WeatherStateDataPoint{Value: varP(tado.DRIZZLE)},
				},
			}},
			err:  assert.NoError,
			want: want{18, 25, tado.DRIZZLE},
		},
		{
			name: "failure",
			resp: fakeWeatherGetter{err: errors.New("some error")},
			err:  assert.Error,
		},
		{
			name: "tado error",
			resp: fakeWeatherGetter{resp: &tado.GetWeatherResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusForbidden},
				JSON403:      &tado.ErrorResponse{Errors: &[]tado.Error{{Code: varP("auth"), Title: varP("bad auth")}}},
			}},
			err: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := TadoUpdater{
				Client: tt.resp,
				HomeId: 1,
			}
			u, err := c.GetUpdate(context.Background())
			tt.err(t, err)
			if err != nil {
				return
			}
			assert.Equal(t, tt.want.temperature, *u.OutsideTemperature.Celsius)
			assert.Equal(t, tt.want.intensity, *u.SolarIntensity.Percentage)
			assert.Equal(t, tt.want.weather, *u.WeatherState.Value)
		})
	}
}

var _ WeatherGetter = fakeWeatherGetter{}

type fakeWeatherGetter struct {
	resp *tado.GetWeatherResponse
	err  error
}

func (f fakeWeatherGetter) GetWeatherWithResponse(_ context.Context, _ tado.HomeId, _ ...tado.RequestEditorFn) (*tado.GetWeatherResponse, error) {
	return f.resp, f.err
}
