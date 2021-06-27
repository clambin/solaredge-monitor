package reports

import (
	"bytes"
	"github.com/clambin/solaredge-monitor/store"
	"gonum.org/v1/plot/vg/vgimg"
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
	buf := new(bytes.Buffer)

	if measurements, err = server.db.Get(start, stop); err == nil {
		var graph *vgimg.PngCanvas
		if graph, err = MakeGraph(measurements, true); err == nil {
			_, err = graph.WriteTo(buf)
		}
	}

	return buf.Bytes(), err
}

func (server *Server) TimeSeries(start, stop time.Time) (image []byte, err error) {
	var measurements []store.Measurement
	buf := new(bytes.Buffer)

	if measurements, err = server.db.Get(start, stop); err == nil {
		var graph *vgimg.PngCanvas
		if graph, err = MakeGraph(measurements, false); err == nil {
			_, err = graph.WriteTo(buf)
		}
	}

	return buf.Bytes(), err
}
