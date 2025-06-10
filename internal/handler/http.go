package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/MikhailRaia/url-shortener/internal/logger"
	"github.com/MikhailRaia/url-shortener/internal/middleware"
	"github.com/MikhailRaia/url-shortener/internal/model"
	"github.com/MikhailRaia/url-shortener/internal/storage"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type URLService interface {
	ShortenURL(originalURL string) (string, error)
	GetOriginalURL(id string) (string, bool)
	ShortenBatch(items []model.BatchRequestItem) ([]model.BatchResponseItem, error)
}

type DBPinger interface {
	Ping(ctx context.Context) error
}

type Handler struct {
	urlService URLService
	dbPinger   DBPinger
}

func NewHandler(urlService URLService, dbPinger DBPinger) *Handler {
	return &Handler{
		urlService: urlService,
		dbPinger:   dbPinger,
	}
}

func (h *Handler) RegisterRoutes() http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)

	r.Use(logger.RequestLogger)

	r.Use(middleware.GzipReader)
	r.Use(middleware.GzipMiddleware)

	r.Post("/", h.handleShorten)
	r.Post("/api/shorten", h.HandleShortenJSON)
	r.Post("/api/shorten/batch", h.handleShortenBatch)
	r.Get("/{id}", h.handleRedirect)
	r.Get("/ping", h.handlePing)

	return r
}

func (h *Handler) handleShorten(w http.ResponseWriter, r *http.Request) {
	contentEncoding := r.Header.Get("Content-Encoding")

	if contentEncoding != "gzip" {
		contentType := r.Header.Get("Content-Type")
		if !strings.Contains(contentType, "text/plain") && contentType != "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(r.Body)

	originalURL := strings.TrimSpace(string(body))
	if originalURL == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortenedURL, err := h.urlService.ShortenURL(originalURL)
	if err != nil {
		if errors.Is(err, storage.ErrURLExists) {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(shortenedURL))
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortenedURL))
}

func (h *Handler) handleRedirect(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	originalURL, found := h.urlService.GetOriginalURL(id)
	if !found {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *Handler) handlePing(w http.ResponseWriter, r *http.Request) {
	if h.dbPinger == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err := h.dbPinger.Ping(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) handleShortenBatch(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var items []model.BatchRequestItem
	if err := json.Unmarshal(body, &items); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(items) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result, err := h.urlService.ShortenBatch(items)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(response)
}
