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
	"github.com/rs/zerolog/log"
)

type URLService interface {
	ShortenURL(originalURL string) (string, error)
	ShortenURLWithUser(originalURL, userID string) (string, error)
	GetOriginalURL(id string) (string, bool)
	GetOriginalURLWithDeletedStatus(id string) (string, error)
	ShortenBatch(items []model.BatchRequestItem) ([]model.BatchResponseItem, error)
	ShortenBatchWithUser(items []model.BatchRequestItem, userID string) ([]model.BatchResponseItem, error)
	GetUserURLs(userID string) ([]model.UserURL, error)
	DeleteUserURLs(userID string, urlIDs []string) error
}

type DBPinger interface {
	Ping(ctx context.Context) error
}

type DeleteWorker interface {
	Submit(userID string, urlIDs []string) error
}

type Handler struct {
	urlService   URLService
	dbPinger     DBPinger
	deleteWorker DeleteWorker
}

func NewHandler(urlService URLService, dbPinger DBPinger) *Handler {
	return &Handler{
		urlService: urlService,
		dbPinger:   dbPinger,
	}
}

func NewHandlerWithDeleteWorker(urlService URLService, dbPinger DBPinger, deleteWorker DeleteWorker) *Handler {
	return &Handler{
		urlService:   urlService,
		dbPinger:     dbPinger,
		deleteWorker: deleteWorker,
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

func (h *Handler) RegisterRoutesWithAuth(authMiddleware *middleware.AuthMiddleware) http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)

	r.Use(logger.RequestLogger)

	r.Use(middleware.GzipReader)
	r.Use(middleware.GzipMiddleware)
	r.Use(authMiddleware.AuthenticateUser)

	r.Post("/", h.handleShortenWithAuth)
	r.Post("/api/shorten", h.HandleShortenJSONWithAuth)
	r.Post("/api/shorten/batch", h.handleShortenBatchWithAuth)
	r.Get("/{id}", h.handleRedirect)
	r.Get("/ping", h.handlePing)

	r.Get("/api/user/urls", h.handleGetUserURLs)
	r.Delete("/api/user/urls", h.handleDeleteUserURLs)

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

		log.Error().Err(err).Msg("Failed to shorten URL")
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

	originalURL, err := h.urlService.GetOriginalURLWithDeletedStatus(id)
	if err != nil {
		if errors.Is(err, storage.ErrURLDeleted) {
			w.WriteHeader(http.StatusGone)
			return
		}
		log.Error().Err(err).Msg("Failed to get original URL")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if originalURL == "" {
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
		log.Error().Err(err).Msg("Failed to ping database")
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
		log.Error().Err(err).Msg("Failed to shorten batch URLs")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(result)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal batch response")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(response)
}

func (h *Handler) handleGetUserURLs(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		log.Debug().Msg("No userID found in context")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	log.Debug().Str("userID", userID).Msg("Found userID in context")

	urls, err := h.urlService.GetUserURLs(userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user URLs")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(urls) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	response, err := json.Marshal(urls)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal user URLs response")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func (h *Handler) handleShortenWithAuth(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

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

	originalURL := strings.TrimSpace(string(body))
	if originalURL == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortenedURL, err := h.urlService.ShortenURLWithUser(originalURL, userID)
	if err != nil {
		if errors.Is(err, storage.ErrURLExists) {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(shortenedURL))
			return
		}
		log.Error().Err(err).Msg("Failed to shorten URL with user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortenedURL))
}

func (h *Handler) HandleShortenJSONWithAuth(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var request ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if request.URL == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortenedURL, err := h.urlService.ShortenURLWithUser(request.URL, userID)
	if err != nil {
		if errors.Is(err, storage.ErrURLExists) {
			response := ShortenResponse{Result: shortenedURL}
			jsonResponse, _ := json.Marshal(response)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			w.Write(jsonResponse)
			return
		}
		log.Error().Err(err).Msg("Failed to shorten JSON URL with user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := ShortenResponse{Result: shortenedURL}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal shorten response")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonResponse)
}

func (h *Handler) handleShortenBatchWithAuth(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var items []model.BatchRequestItem
	if err := json.NewDecoder(r.Body).Decode(&items); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(items) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	result, err := h.urlService.ShortenBatchWithUser(items, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to shorten batch URLs with user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(result)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal batch response with user")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(response)
}

func (h *Handler) handleDeleteUserURLs(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var urlIDs []string
	if err := json.NewDecoder(r.Body).Decode(&urlIDs); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(urlIDs) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Отправляем запрос на удаление в воркер-пул
	if h.deleteWorker != nil {
		if err := h.deleteWorker.Submit(userID, urlIDs); err != nil {
			log.Error().Err(err).Msg("Failed to submit delete request to worker pool")
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		log.Debug().
			Str("userID", userID).
			Int("urlCount", len(urlIDs)).
			Msg("Delete request submitted to worker pool")
	} else {
		go func() {
			if err := h.urlService.DeleteUserURLs(userID, urlIDs); err != nil {
				log.Error().Err(err).Msg("Failed to delete user URLs")
			}
		}()
		log.Warn().Msg("DeleteWorker not configured, using fallback goroutine")
	}

	w.WriteHeader(http.StatusAccepted)
}
