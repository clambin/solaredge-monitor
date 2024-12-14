package publisher

import (
	"context"
	"fmt"
	"github.com/clambin/tado/v2"
	"net/http"
)

type TadoUpdater struct {
	TadoClient WeatherGetter
	HomeId     tado.HomeId
}

type WeatherGetter interface {
	GetWeatherWithResponse(ctx context.Context, homeId tado.HomeId, reqEditors ...tado.RequestEditorFn) (*tado.GetWeatherResponse, error)
}

func (c TadoUpdater) GetUpdate(ctx context.Context) (*tado.Weather, error) {
	resp, err := c.TadoClient.GetWeatherWithResponse(ctx, c.HomeId)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("tado: %s", resp.Status())
	}
	return resp.JSON200, nil
}
