package webserver

import (
	"net/http"
)

type mux struct {
	*http.ServeMux
	notFound http.Handler
}

func newMux() *mux {
	hm := http.NewServeMux()
	return &mux{ServeMux: hm}
}

func (m *mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "*" {
		if r.ProtoAtLeast(1, 1) {
			w.Header().Set("Connection", "close")
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	h, pattern := m.Handler(r)
	if pattern == "" {
		m.notFound.ServeHTTP(w, r)
	} else {
		h.ServeHTTP(w, r)
	}
}

type Server struct {
	*http.Server
}

func NewServer() *Server {
	srv := http.Server{
		Handler: newMux(),
	}
	s := Server{Server: &srv}
	return &s
}

func (s *Server) Start() {

}

func (s *Server) Shutdown() {

}
