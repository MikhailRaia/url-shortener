package handler

import (
	"io"
	"net/http"
	"strings"

	"github.com/MikhailRaia/url-shortener/internal/logger"
	"github.com/MikhailRaia/url-shortener/internal/middleware"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type URLService interface {
	ShortenURL(originalURL string) (string, error)
	GetOriginalURL(id string) (string, bool)
}

type Handler struct {
	urlService URLService
}

func NewHandler(urlService URLService) *Handler {
	return &Handler{
		urlService: urlService,
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
	r.Get("/{id}", h.handleRedirect)

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
