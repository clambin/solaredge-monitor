package collector

import (
	"context"
	"github.com/clambin/solaredge-monitor/store"
	log "github.com/sirupsen/logrus"
	"math"
	"time"
)

type Collector struct {
	Intensity chan Metric
	Power     chan Metric
	intensity Summary
	power     Summary
	interval  time.Duration
	cutOff    time.Time
	db        store.DB
}

type Metric struct {
	Timestamp time.Time
	Value     float64
}

func New(interval time.Duration, db store.DB) *Collector {
	return &Collector{
		Intensity: make(chan Metric),
		Power:     make(chan Metric),
		interval:  interval,
		db:        db,
	}
}

func (collector *Collector) Run(ctx context.Context) {
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case m := <-collector.Intensity:
			log.WithFields(log.Fields{"metric": m, "cutOff": collector.cutOff}).Debug("solar intensity metric received")
			if collector.shouldCollect(m.Timestamp) {
				collector.collect()
			}
			collector.intensity.Add(m)
		case m := <-collector.Power:
			log.WithFields(log.Fields{"metric": m, "cutOff": collector.cutOff}).Debug("power metric received")
			if collector.shouldCollect(m.Timestamp) {
				collector.collect()
			}
			collector.power.Add(m)
		}
	}
	collector.collect()
}

func (collector *Collector) shouldCollect(next time.Time) bool {
	if collector.cutOff.IsZero() {
		collector.cutOff = next.Add(collector.interval)
		return false
	}

	return next.After(collector.cutOff)
}

func (collector *Collector) collect() {
	log.WithFields(log.Fields{
		"power":     collector.power.Count,
		"intensity": collector.intensity.Count,
		"cutOff":    collector.cutOff,
	}).Debug("running collection")

	power := collector.power.Get()
	intensity := collector.intensity.Get()

	if math.IsNaN(power.Value) || math.IsNaN(intensity.Value) {
		log.Warning("impartial data collection")
		return
	}

	ts := power.Timestamp
	if intensity.Timestamp.Before(ts) {
		ts = intensity.Timestamp
	}

	measurement := store.Measurement{
		Timestamp: ts,
		Power:     power.Value,
		Intensity: intensity.Value,
	}

	if err := collector.db.Store(measurement); err == nil {
		log.WithField("measurement", measurement).Info("new entry")
	} else {
		log.WithError(err).Warning("failed to store metrics")
	}

	collector.cutOff = collector.cutOff.Add(collector.interval)

	log.WithField("cutOff", collector.cutOff).Debug("new cut-off time set")
}
