package feedreader

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/PuerkitoBio/purell"
	"github.com/abatilo/amanuensis/internal/db"
	"gorm.io/gorm"
)

func (s *Server) prepareRoutes() {
	s.mux.HandleFunc("GET /feeds", s.listFeeds())
	s.mux.HandleFunc("GET /feeds/{id}", s.getFeed())
	s.mux.HandleFunc("POST /feeds", s.createFeed())
}

func (s *Server) listFeeds() http.HandlerFunc {
	type feed struct {
		ID  uint   `json:"id"`
		URL string `json:"url"`
	}

	type listFeedsResponse struct {
		Feeds []feed `json:"feeds"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var feeds []feed
		result := s.db.Find(&feeds)
		if result.Error != nil {
			s.logger.Error("failed to fetch feeds", "error", result.Error)
			http.Error(w, "failed to fetch feeds", http.StatusInternalServerError)
			return
		}

		resp := listFeedsResponse{Feeds: feeds}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			s.logger.Error("failed to encode response", "error", err)
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

func (s *Server) getFeed() http.HandlerFunc {
	logger := s.logger.With("handler", "getFeed")

	type feed struct {
		ID  uint   `json:"id"`
		URL string `json:"url"`
	}

	type getFeedResponse struct {
		ID  uint   `json:"id"`
		URL string `json:"url"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			logger.Error("missing id")
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}

		var feed feed
		result := s.db.First(&feed, id)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				http.Error(w, "feed not found", http.StatusNotFound)
				return
			}

			s.logger.Error("failed to fetch feed", "error", result.Error)
			http.Error(w, "failed to fetch feed", http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(getFeedResponse{ID: feed.ID, URL: feed.URL}); err != nil {
			s.logger.Error("failed to encode response", "error", err)
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

func (s *Server) createFeed() http.HandlerFunc {
	logger := s.logger.With("handler", "createFeed")

	type createFeedRequest struct {
		URL string `json:"url"`
	}
	type createFeedResponse struct {
		ID uint `json:"id"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var req createFeedRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.logger.Error("failed to decode request", "error", err)
			http.Error(w, "failed to decode request", http.StatusBadRequest)
			return
		}

		sanitizedURL := purell.MustNormalizeURLString(req.URL, purell.FlagsAllGreedy)

		if sanitizedURL == "" {
			s.logger.Error("invalid URL", "url", req.URL)
			http.Error(w, "invalid URL", http.StatusBadRequest)
			return
		}

		logger.Debug("creating feed", "url", sanitizedURL)

		feed := &db.Feed{URL: sanitizedURL}
		result := s.db.Where(db.Feed{URL: sanitizedURL}).FirstOrCreate(feed)
		if result.Error != nil {
			s.logger.Error("failed to create feed", "error", result.Error)
			http.Error(w, "failed to create feed", http.StatusInternalServerError)
			return
		}

		resp := createFeedResponse{ID: feed.ID}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			s.logger.Error("failed to encode response", "error", err)
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}
