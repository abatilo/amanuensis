package feedreader

import (
	"fmt"
	"net/http"
)

func (s *Server) prepareRoutes() {
	s.mux.HandleFunc("GET /feeds", s.listFeeds())
	s.mux.HandleFunc("GET /feeds/{id}", s.getFeed())
}

func (s *Server) listFeeds() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, feeds!")
	}
}

func (s *Server) getFeed() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		fmt.Fprintln(w, "Hello, feed!", id)
	}
}
