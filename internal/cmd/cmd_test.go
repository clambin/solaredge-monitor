package cmd

import (
	"context"
	solaredge2 "github.com/clambin/solaredge"
	"github.com/clambin/solaredge-monitor/internal/scraper/solaredge"
	"github.com/spf13/viper"
	"strconv"
	"time"
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
func feed(ctx context.Context, ch chan solaredge.Update, count int, interval time.Duration) {
	for range count {
		ch <- solaredge.Update{{
			ID:   1,
			Name: "my home",
			PowerOverview: solaredge2.PowerOverview{
				LastUpdateTime: solaredge2.Time(time.Date(2024, time.December, 12, 12, 0, 0, 0, time.UTC)),
				LifeTimeData:   solaredge2.EnergyOverview{Energy: 1000},
				LastYearData:   solaredge2.EnergyOverview{Energy: 100},
				LastMonthData:  solaredge2.EnergyOverview{Energy: 10},
				LastDayData:    solaredge2.EnergyOverview{Energy: 1},
				CurrentPower:   solaredge2.CurrentPower{Power: 500},
			},
		}}
		select {
		case <-ctx.Done():
			return
		case <-time.After(interval):
		}
	}
}
