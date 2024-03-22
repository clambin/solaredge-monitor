package html

import (
	"github.com/clambin/solaredge-monitor/internal/web/handlers/arguments"
	"log/slog"
	"net/http"
	"net/url"
	"text/template"
	"time"
)

type ReportHandler struct {
	Logger *slog.Logger
}

type Data struct {
	PlotTypes []string
	FoldTypes []string
	Args      string
}

func (h ReportHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
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
	values.Add("start", args.Start.Format(time.RFC3339))
	values.Add("stop", args.Stop.Format(time.RFC3339))

	tmpl := template.Must(template.ParseFS(html, "templates/report.html"))
	data := Data{
		PlotTypes: []string{"scatter", "heatmap"},
		FoldTypes: []string{"false", "true"},
		Args:      values.Encode(),
	}

	//	w.WriteHeader(http.StatusOK)
	if err = tmpl.Execute(w, data); err != nil {
		h.Logger.Error("failed to generate page", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
