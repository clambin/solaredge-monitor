package cmd

import (
	"github.com/spf13/viper"
	"strconv"
)

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
