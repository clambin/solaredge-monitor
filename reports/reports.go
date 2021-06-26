package reports

import (
	"fmt"
	"github.com/clambin/solaredge-monitor/store"
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

	if err = saveGraph(measurements, path.Join(server.imagesDirectory, "week.png"), false); err != nil {
		return err
	}

	return saveGraph(measurements, path.Join(server.imagesDirectory, "summary.png"), true)
}

func saveGraph(measurements []store.Measurement, filename string, fold bool) (err error) {
	img := MakeGraph(measurements, fold)

	var w *os.File
	if w, err = os.Create(path.Join(filename)); err != nil {
		return err
	}

	defer func() {
		_ = w.Close()
	}()

	_, err = img.WriteTo(w)

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

	return measurements[0].Timestamp, nil
}
