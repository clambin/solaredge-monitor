package server

import (
	"net/http"
)

func (server *Server) main(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
