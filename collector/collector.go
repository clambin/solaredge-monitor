package collector

import (
	"github.com/clambin/solaredge-monitor/store"
	log "github.com/sirupsen/logrus"
	"time"
)

type Collector struct {
	Stop      chan struct{}
	Intensity chan Metric
	Power     chan Metric
	intensity Summary
	power     Summary
	interval  time.Duration
	db        store.DB
}

type Metric struct {
	Timestamp time.Time
	Value     float64
}

func New(interval time.Duration, db store.DB) *Collector {
	return &Collector{
		Stop:      make(chan struct{}),
		Intensity: make(chan Metric),
		Power:     make(chan Metric),
		interval:  interval,
		db:        db,
	}
}

func (collector *Collector) Run() {
loop:
	for {
		select {
		case <-collector.Stop:
			break loop
		case m := <-collector.Intensity:
			if !collector.intensity.InRange(m, collector.interval) {
				collector.process()
			}
			collector.intensity.Add(m)
		case m := <-collector.Power:
			if !collector.power.InRange(m, collector.interval) {
				collector.process()
			}
			collector.power.Add(m)
		}
	}
	collector.process()
}

func (collector *Collector) process() {
	if collector.power.Count == 0 || collector.intensity.Count == 0 {
		log.WithFields(log.Fields{
			"power":     collector.power.Count,
			"intensity": collector.intensity.Count,
		}).Debug("one or more metrics have no data. skipping processing")
		return
	}

	power := collector.power.Get()
	intensity := collector.intensity.Get()

	log.Debugf("power: %v, intensity: %v", power, intensity)

	ts := intensity.Timestamp
	if power.Timestamp.Before(ts) {
		ts = power.Timestamp
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
}
