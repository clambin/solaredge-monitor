package cmd

import (
	"context"
	"github.com/clambin/tado/v2"
	"github.com/clambin/tado/v2/tools"
	"github.com/spf13/viper"
	"log/slog"
)

var _ tools.TadoClient = fakeMeGetter{}

type fakeMeGetter struct {
	resp *tado.GetMeResponse
	err  error
}

func (f fakeMeGetter) GetMeWithResponse(_ context.Context, _ ...tado.RequestEditorFn) (*tado.GetMeResponse, error) {
	return f.resp, f.err
}

var discardLogger = slog.New(slog.DiscardHandler)

func varP[T any](v T) *T {
	return &v
}

func getViperFromViper(v *viper.Viper) *viper.Viper {
	n := viper.New()
	for _, k := range v.AllKeys() {
		n.Set(k, v.Get(k))
	}
	return n
}
