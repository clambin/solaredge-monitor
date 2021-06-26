package reports

import (
	"fmt"
	"github.com/clambin/solaredge-monitor/store"
	log "github.com/sirupsen/logrus"
	"math"
	"os"
	"path"
	"time"
)

type Server struct {
	db              store.DB
	imagesDirectory string
}

func New(imagesDirectory string, db store.DB) *Server {
	return &Server{
		db:              db,
		imagesDirectory: imagesDirectory,
	}
}

func (server *Server) ImagesDirectory() string {
	return server.imagesDirectory
}

func (server *Server) Overview(start, stop time.Time) (err error) {
	var measurements []store.Measurement

	if measurements, err = server.db.Get(start, stop); err != nil {
		return err
	}

	powerAverage, powerVariance, intensityAverage, intensityVariance := analyzeMeasurements(measurements)
	log.WithFields(log.Fields{
		"start":             start,
		"stop":              stop,
		"measurement":       len(measurements),
		"powerAverage":      powerAverage,
		"powerVariance":     powerVariance,
		"intensityAverage":  intensityAverage,
		"intensityVariance": intensityVariance,
	}).Debug("building graphs")

	if err = saveGraph(measurements, path.Join(server.imagesDirectory, "week.png"), false); err != nil {
		return err
	}

	return saveGraph(measurements, path.Join(server.imagesDirectory, "summary.png"), true)
}

func analyzeMeasurements(measurements []store.Measurement) (powerAverage, powerVariance, intensityAverage, intensityVariance float64) {
	type info struct {
		total float64
		count int
		min   float64
		max   float64
	}
	power := info{total: 0.0, count: 0, min: math.Inf(+1), max: math.Inf(-1)}
	intensity := info{total: 0.0, count: 0, min: math.Inf(+1), max: math.Inf(-1)}

	for _, entry := range measurements {
		power.total += entry.Power
		power.count++
		if entry.Power < power.min {
			power.min = entry.Power
		}
		if entry.Power > power.max {
			power.max = entry.Power
		}

		intensity.total += entry.Intensity
		intensity.count++
		if entry.Intensity < intensity.min {
			intensity.min = entry.Intensity
		}
		if entry.Intensity > intensity.max {
			intensity.max = entry.Intensity
		}

	}

	powerAverage = power.total / float64(power.count)
	powerVariance = power.max - power.min
	intensityAverage = intensity.total / float64(intensity.count)
	intensityVariance = intensity.max - intensity.min

	return
}

func saveGraph(measurements []store.Measurement, filename string, fold bool) (err error) {
	img := MakeGraph(measurements, fold)

	var w *os.File
	if w, err = os.Create(path.Join(filename)); err != nil {
		return err
	}

	_, err = img.WriteTo(w)
	_ = w.Close()

	return
}

func (server *Server) GetFirst() (first time.Time, err error) {
	var measurements []store.Measurement

	if measurements, err = server.db.GetAll(); err != nil {
		return time.Time{}, err
	}

	if len(measurements) == 0 {
		return time.Time{}, fmt.Errorf("no entries found")
	}

	return measurements[0].Timestamp, nil
}

func (server *Server) GetLast() (first time.Time, err error) {
	var measurements []store.Measurement

	if measurements, err = server.db.GetAll(); err != nil {
		return time.Time{}, err
	}

	if len(measurements) == 0 {
		return time.Time{}, fmt.Errorf("no entries found")
	}

	return measurements[len(measurements)-1].Timestamp, nil
}
