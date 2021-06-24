package poller

import (
	"github.com/clambin/solaredge-monitor/collector"
	log "github.com/sirupsen/logrus"
	"time"
)

type PollFunc func() (float64, error)

type BasePoller struct {
	Stop      chan struct{}
	ticker    *time.Ticker
	poll      PollFunc
	collector chan collector.Metric
}

func NewBasePoller(interval time.Duration, poll PollFunc, collectorChannel chan collector.Metric) *BasePoller {
	return &BasePoller{
		Stop:      make(chan struct{}),
		ticker:    time.NewTicker(interval),
		poll:      poll,
		collector: collectorChannel,
	}
}

func (poller *BasePoller) Run() {
loop:
	for {
		select {
		case <-poller.Stop:
			break loop
		case <-poller.ticker.C:
			value, err := poller.poll()
			if err == nil {
				poller.collector <- collector.Metric{
					Timestamp: time.Now(),
					Value: value,
				}
			} else {
				log.WithError(err).Warning("failed to poll data")
			}
		}
	}
	poller.ticker.Stop()
}
