package html

import (
	"github.com/clambin/solaredge-monitor/internal/web/handlers/arguments"
	"log/slog"
	"net/http"
	"net/url"
	"text/template"
	"time"
)

type PlotHandler struct {
	Logger *slog.Logger
}

var tmpl = template.Must(template.ParseFS(html, "templates/plot.html"))

func (h PlotHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	args, err := arguments.Parse(req)
	if err != nil {
		h.Logger.Error("failed to determine start/stop parameters", "err", err)
		http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
		return
	}

	if args.Stop.IsZero() {
		args.Stop = time.Now()
	}
	values := make(url.Values)
	if args.Fold {
		values.Add("fold", "true")
	}
	values.Add("start", args.Start.Format(time.RFC3339))
	values.Add("stop", args.Stop.Format(time.RFC3339))

	data := struct {
		PlotType string
		Args     string
	}{
		PlotType: req.PathValue("plotType"),
		Args:     values.Encode(),
	}

	//	w.WriteHeader(http.StatusOK)
	if err = tmpl.Execute(w, data); err != nil {
		h.Logger.Error("failed to generate page", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
