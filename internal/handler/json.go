package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/MikhailRaia/url-shortener/internal/storage"
)

// ShortenRequest is the JSON payload for shortening a single URL.
type ShortenRequest struct {
	URL string `json:"url"`
}

// ShortenResponse is the JSON response containing a shortened URL.
type ShortenResponse struct {
	Result string `json:"result"`
}

// HandleShortenJSON handles POST /api/shorten requests with JSON payload.
func (h *Handler) HandleShortenJSON(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	contentEncoding := r.Header.Get("Content-Encoding")

	if contentEncoding != "gzip" && !strings.Contains(contentType, "application/json") {
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

	var request ShortenRequest
	if err := json.Unmarshal(body, &request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if request.URL == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortenedURL, err := h.urlService.ShortenURL(r.Context(), request.URL)
	if err != nil {
		if errors.Is(err, storage.ErrURLExists) {
			response := ShortenResponse{
				Result: shortenedURL,
			}

			responseJSON, err := json.Marshal(response)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			w.Write(responseJSON)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := ShortenResponse{
		Result: shortenedURL,
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(responseJSON)
}
