package reports

import (
	"bytes"
	"github.com/clambin/solaredge-monitor/store"
	"time"
)

type Server struct {
	db store.DB
}

func New(db store.DB) *Server {
	return &Server{
		db: db,
	}
}

func (server *Server) Summary(start, stop time.Time) (image []byte, err error) {
	var measurements []store.Measurement

	if measurements, err = server.db.Get(start, stop); err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	_, err = MakeGraph(measurements, true).WriteTo(buf)
	return buf.Bytes(), err
}

func (server *Server) TimeSeries(start, stop time.Time) (image []byte, err error) {
	var measurements []store.Measurement

	if measurements, err = server.db.Get(start, stop); err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	_, err = MakeGraph(measurements, false).WriteTo(buf)
	return buf.Bytes(), err
}
