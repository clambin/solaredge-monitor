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
	cutOff    time.Time
	db        store.DB
}

type Metric struct {
	Timestamp time.Time
	Value     float64
}

const (
	Intensity = iota
	Power
)

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
			collector.process(m, Intensity)
		case m := <-collector.Power:
			collector.process(m, Power)
		}
	}
	collector.collect()
}

func (collector *Collector) process(m Metric, source int) {
	log.WithFields(log.Fields{
		"metric": m,
		"source": source,
		"cutOff": collector.cutOff,
	}).Debug("metric received")

	if collector.shouldCollect(m.Timestamp, source) {
		collector.collect()
	}

	switch source {
	case Intensity:
		collector.intensity.Add(m)
	case Power:
		collector.power.Add(m)
	}
}

func (collector *Collector) shouldCollect(next time.Time, source int) bool {
	if collector.cutOff.IsZero() {
		collector.cutOff = next.Add(collector.interval)
		return false
	}

	if source == Power && collector.intensity.Count == 0 {
		return false
	}

	if source == Intensity && collector.power.Count == 0 {
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

	collector.cutOff = collector.cutOff.Add(collector.interval)

	log.WithField("cutOff", collector.cutOff).Debug("new cut-off time set")
}
