package serpent_test

import (
	"github.com/clambin/solaredge-monitor/pkg/serpent"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
	"time"
)

var args = serpent.Arguments{
	"int":      {Default: 42},
	"float64":  {Default: math.Pi},
	"string":   {Default: "foo"},
	"bool":     {Default: true},
	"duration": {Default: time.Second},
}

func TestSerpent(t *testing.T) {
	var cmd cobra.Command
	v := viper.New()

	assert.NoError(t, serpent.SetDefaults(v, args))
	assert.NoError(t, serpent.SetPersistentFlags(&cmd, v, args))

	for name := range args {
		assert.NotNil(t, cmd.PersistentFlags().Lookup(name))
	}

	assert.Equal(t, args["bool"].Default.(bool), v.Get("bool"))
	assert.Equal(t, args["duration"].Default.(time.Duration), v.Get("duration"))
	assert.Equal(t, args["string"].Default.(string), v.Get("string"))
	assert.Equal(t, args["int"].Default.(int), v.Get("int"))
}

func TestSerpent_SetPersistentFlags_Fail(t *testing.T) {
	var cmd cobra.Command
	v := viper.New()

	badArgs1 := serpent.Arguments{"time": {Default: time.Now()}}
	assert.NoError(t, serpent.SetDefaults(v, badArgs1))
	assert.Error(t, serpent.SetPersistentFlags(&cmd, v, badArgs1))

	badArgs2 := serpent.Arguments{"null": {Default: nil}}
	assert.Error(t, serpent.SetDefaults(v, badArgs2))
}
