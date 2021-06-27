package feeder

import (
	"github.com/clambin/solaredge-monitor/scrape/collector"
)

func FeedMetrics(power []collector.Metric, intensity []collector.Metric, coll *collector.Collector) (err error) {
	for len(power) > 0 && len(intensity) > 0 {
		if power[0].Timestamp.Before(intensity[0].Timestamp) {
			coll.Power <- power[0]
			power = power[1:]
		} else {
			coll.Intensity <- intensity[0]
			intensity = intensity[1:]
		}
	}

	for len(power) > 0 {
		coll.Power <- power[0]
		power = power[1:]
	}

	for len(intensity) > 0 {
		coll.Intensity <- intensity[0]
		intensity = intensity[1:]
	}
	return
}
