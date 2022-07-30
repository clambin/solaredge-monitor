package reports

import (
	"fmt"
	"github.com/clambin/solaredge-monitor/store"
	"time"
)

func (r *Reporter) GetFirst() (first time.Time, err error) {
	var measurements []store.Measurement

	measurements, err = r.db.GetAll()

	if err == nil && len(measurements) == 0 {
		err = fmt.Errorf("no entries found")
	}

	if err != nil {
		return
	}

	return measurements[0].Timestamp, nil
}

func (r *Reporter) GetLast() (first time.Time, err error) {
	var measurements []store.Measurement

	measurements, err = r.db.GetAll()

	if err == nil && len(measurements) == 0 {
		err = fmt.Errorf("no entries found")
	}

	if err != nil {
		return
	}

	return measurements[len(measurements)-1].Timestamp, nil
}
