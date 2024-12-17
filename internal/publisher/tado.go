package publisher

import (
	"context"
	"fmt"
	"github.com/clambin/tado/v2"
	"github.com/clambin/tado/v2/tools"
	"net/http"
)

type TadoUpdater struct {
	Client WeatherGetter
	HomeId tado.HomeId
}

type WeatherGetter interface {
	GetWeatherWithResponse(ctx context.Context, homeId tado.HomeId, reqEditors ...tado.RequestEditorFn) (*tado.GetWeatherResponse, error)
}

func (c TadoUpdater) GetUpdate(ctx context.Context) (*tado.Weather, error) {
	resp, err := c.Client.GetWeatherWithResponse(ctx, c.HomeId)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("tado: %w", tools.HandleErrors(resp.HTTPResponse, map[int]any{
			http.StatusUnauthorized: resp.JSON401,
			http.StatusForbidden:    resp.JSON403,
		}))
	}
	return resp.JSON200, nil
}
