package feedreader

import (
	"embed"
	"html/template"
	"net/http"

	"github.com/PuerkitoBio/purell"
	"github.com/abatilo/amanuensis/internal/db"
)

//go:embed static/*
var static embed.FS

func (s *Server) prepareRoutes() {
	s.mux.HandleFunc("GET /", s.index())
	s.mux.HandleFunc("GET /feeds", s.renderFeeds())
	s.mux.HandleFunc("GET /feeds/create", s.renderCreateFeed())
	s.mux.HandleFunc("POST /feeds/create", s.createFeed())
}

func (s *Server) index() http.HandlerFunc {
	logger := s.logger.With("handler", "root")

	tmpl, err := template.ParseFS(
		static,
		"static/layouts/base.html.tmpl",
		"static/pages/index.html",
	)
	if err != nil {
		logger.Error("Failed to parse template", "error", err)
		panic(err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = tmpl.ExecuteTemplate(w, "base", nil)
	}
}

func (s *Server) renderFeeds() http.HandlerFunc {
	logger := s.logger.With("handler", "renderFeeds")

	type feed struct {
		ID  uint   `json:"id"`
		URL string `json:"url"`
	}

	tmpl, err := template.ParseFS(
		static,
		"static/layouts/base.html.tmpl",
		"static/pages/feeds/index.html.tmpl",
	)
	if err != nil {
		logger.Error("Failed to parse template", "error", err)
		panic(err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var feeds []feed
		result := s.db.Find(&feeds)
		if result.Error != nil {
			s.logger.Error("failed to fetch feeds", "error", result.Error)
			http.Error(w, "failed to fetch feeds", http.StatusInternalServerError)
			return
		}

		err = tmpl.ExecuteTemplate(w, "base", feeds)
		if err != nil {
			logger.Error("Failed to render template", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s *Server) renderCreateFeed() http.HandlerFunc {
	logger := s.logger.With("handler", "renderCreateFeed")

	tmpl, err := template.ParseFS(
		static,
		"static/layouts/base.html.tmpl",
		"static/pages/feeds/create.html.tmpl",
	)
	if err != nil {
		logger.Error("Failed to parse template", "error", err)
		panic(err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		err := tmpl.ExecuteTemplate(w, "base", nil)
		if err != nil {
			logger.Error("Failed to render template", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s *Server) createFeed() http.HandlerFunc {
	logger := s.logger.With("handler", "createFeed")

	type req struct {
		URL string `json:"url"`
	}

	type feed struct {
		ID  uint   `json:"id"`
		URL string `json:"url"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			logger.Error("failed to parse form", "error", err)
			http.Error(w, "failed to parse form", http.StatusInternalServerError)
			return
		}
		req := req{
			URL: r.FormValue("url"),
		}

		sanitizedURL := purell.MustNormalizeURLString(req.URL, purell.FlagsUsuallySafeGreedy)
		if sanitizedURL == "" {
			s.logger.Error("invalid URL", "url", req.URL)
			http.Error(w, "invalid URL", http.StatusBadRequest)
			return
		}

		// Send HTTP request to the URL to check if it's a valid feed
		checkRequest, err := http.NewRequestWithContext(r.Context(), "GET", sanitizedURL, nil)
		if err != nil {
			s.logger.Error("failed to create request", "error", err)
			http.Error(w, "failed to create request", http.StatusInternalServerError)
			return
		}

		checkResponse, err := s.httpClient.Do(checkRequest)
		if err != nil {
			s.logger.Error("failed to send request", "error", err)
			http.Error(w, "failed to send request", http.StatusInternalServerError)
			return
		}

		if checkResponse.StatusCode != http.StatusOK {
			s.logger.Error("invalid status code", "status", checkResponse.StatusCode)
			http.Error(w, "invalid status code", http.StatusBadRequest)
			return
		}

		logger.Debug("creating feed", "url", sanitizedURL)

		f := &db.Feed{URL: sanitizedURL}
		result := s.db.Where(db.Feed{URL: sanitizedURL}).FirstOrCreate(f)
		if result.Error != nil {
			s.logger.Error("failed to create feed", "error", result.Error)
			http.Error(w, "failed to create feed", http.StatusInternalServerError)
			return
		}

		var feeds []feed
		result = s.db.Find(&feeds)
		if result.Error != nil {
			s.logger.Error("failed to fetch feeds", "error", result.Error)
			http.Error(w, "failed to fetch feeds", http.StatusInternalServerError)
			return
		}

		w.Header().Set("HX-Location", "/feeds")
	}
}
