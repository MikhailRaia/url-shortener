package handler

import (
	"io"
	"net/http"
	"strings"
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
	mux := http.NewServeMux()

	mux.HandleFunc("/", h.handleRequest)

	return mux
}

func (h *Handler) handleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost && r.URL.Path == "/" {
		h.handleShorten(w, r)
		return
	}

	if r.Method == http.MethodGet && r.URL.Path != "/" {
		h.handleRedirect(w, r)
		return
	}

	w.WriteHeader(http.StatusBadRequest)
}

func (h *Handler) handleShorten(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/plain") && contentType != "" {
		w.WriteHeader(http.StatusBadRequest)
		return
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
	id := strings.TrimPrefix(r.URL.Path, "/")

	originalURL, found := h.urlService.GetOriginalURL(id)
	if !found {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
