package cmd

import (
	"context"
	"errors"
	"github.com/clambin/tado/v2"
	"github.com/clambin/tado/v2/tools"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"testing"
)

func Test_getHomeId(t *testing.T) {
	type args struct {
		resp *tado.GetMeResponse
		err  error
	}
	type want struct {
		homeId tado.HomeId
		err    assert.ErrorAssertionFunc
	}
	tests := []struct {
		name string
		args
		want
	}{
		{
			name: "success",
			args: args{
				resp: &tado.GetMeResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusOK},
					JSON200:      &tado.User{Homes: &[]tado.HomeBase{{Id: varP(tado.HomeId(1))}}},
				},
				err: nil,
			},
			want: want{homeId: tado.HomeId(1), err: assert.NoError},
		},
		{
			name: "error",
			args: args{
				err: errors.New("some error"),
			},
			want: want{err: assert.Error},
		},
		{
			name: "no homes",
			args: args{
				resp: &tado.GetMeResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusOK},
					JSON200:      &tado.User{Homes: &[]tado.HomeBase{}},
				},
				err: nil,
			},
			want: want{err: assert.Error},
		},
		{
			name: "more than one home",
			args: args{
				resp: &tado.GetMeResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusOK},
					JSON200:      &tado.User{Homes: &[]tado.HomeBase{{Id: varP(tado.HomeId(1))}, {Id: varP(tado.HomeId(2))}}},
				},
				err: nil,
			},
			want: want{homeId: 1, err: assert.NoError},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := fakeMeGetter{tt.args.resp, tt.args.err}

			got, err := getHomeId(context.Background(), f, discardLogger)
			assert.Equal(t, tt.want.homeId, got)
			tt.want.err(t, err)
		})
	}

}

var _ tools.TadoClient = fakeMeGetter{}

type fakeMeGetter struct {
	resp *tado.GetMeResponse
	err  error
}

func (f fakeMeGetter) GetMeWithResponse(_ context.Context, _ ...tado.RequestEditorFn) (*tado.GetMeResponse, error) {
	return f.resp, f.err
}

var discardLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

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

func initViperDB(v *viper.Viper, port int) {
	v.Set("database.host", "localhost")
	v.Set("database.port", port)
	v.Set("database.database", "solaredge")
	v.Set("database.username", "username")
	v.Set("database.password", "password")
}

func initViperCache(v *viper.Viper, port int) {
	v.Set("web.cache.addr", "localhost:"+strconv.Itoa(port))
}
