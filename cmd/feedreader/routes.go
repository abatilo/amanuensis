package feedreader

import (
	"net/http"

	"github.com/PuerkitoBio/purell"
	"github.com/a-h/templ"
	"github.com/abatilo/amanuensis/cmd/feedreader/static/layouts"
	"github.com/abatilo/amanuensis/cmd/feedreader/static/pages"
	pagesfeeds "github.com/abatilo/amanuensis/cmd/feedreader/static/pages/feeds"
	"github.com/abatilo/amanuensis/internal/db"
)

func (s *Server) prepareRoutes() {
	s.mux.HandleFunc("GET /", s.index())
	s.mux.HandleFunc("GET /feeds", s.renderFeeds())
	s.mux.HandleFunc("GET /feeds/create", s.renderCreateFeed())
	s.mux.HandleFunc("POST /feeds/create", s.createFeed())
}

func (s *Server) index() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = layouts.Base(pages.Index()).Render(r.Context(), w)
	}
}

func (s *Server) renderFeeds() http.HandlerFunc {
	type feed struct {
		ID  uint   `json:"id"`
		URL string `json:"url"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var feeds []feed
		result := s.db.Find(&feeds)
		if result.Error != nil {
			s.logger.Error("failed to fetch feeds", "error", result.Error)
			http.Error(w, "failed to fetch feeds", http.StatusInternalServerError)
			return
		}

		feedRows := make([]templ.Component, len(feeds))
		for i, f := range feeds {
			feedRows[i] = pagesfeeds.FeedRow(f.ID, f.URL)
		}
		_ = layouts.Base(pagesfeeds.Index(feedRows)).Render(r.Context(), w)
	}
}

func (s *Server) renderCreateFeed() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = layouts.Base(pagesfeeds.Create()).Render(r.Context(), w)
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
