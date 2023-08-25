package handlers

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

type ReportsHandler struct {
	Logger *slog.Logger
}

func (h ReportsHandler) Handle(w http.ResponseWriter, req *http.Request) {
	args, err := parseArguments(req)
	if err != nil {
		h.Logger.Error("failed to determine start/stop parameters", "err", err)
		http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
		return
	}

	values := make(url.Values)
	if !args.start.IsZero() {
		values.Add("start", args.start.Format(time.RFC3339))
	}
	if !args.stop.IsZero() {
		values.Add("stop", args.stop.Format(time.RFC3339))
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/html")
	h.generateResponse(w, values)
}

func (h ReportsHandler) generateResponse(w io.Writer, args url.Values) {
	_, _ = w.Write([]byte(`<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
    <title>Report</title>
  </head>
  <body>
`))

	for index, plot := range []string{"scatter", "heatmap"} {
		if index > 0 {
			_, _ = w.Write([]byte(`  <p>
`))
		}

		for _, fold := range []bool{false, true} {
			target := fmt.Sprintf("/plot/%s?fold=%v&%s", plot, fold, args.Encode())
			section := fmt.Sprintf(`  <a href="%s">
    <img src="%s" alt="scatter" style="width:400px;height:400px;">
  </a>
`, target, target)
			_, _ = w.Write([]byte(section))
		}
	}

	_, _ = w.Write([]byte(`  </body>
</html>
`))
}
