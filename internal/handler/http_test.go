package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

type mockURLService struct {
	shortenURLFunc     func(originalURL string) (string, error)
	getOriginalURLFunc func(id string) (string, bool)
}

func (m *mockURLService) ShortenURL(originalURL string) (string, error) {
	return m.shortenURLFunc(originalURL)
}

func (m *mockURLService) GetOriginalURL(id string) (string, bool) {
	return m.getOriginalURLFunc(id)
}

func TestHandler_handleShorten(t *testing.T) {
	tests := []struct {
		name           string
		requestURL     string
		requestMethod  string
		requestBody    string
		contentType    string
		mockShortenURL string
		mockShortenErr error
		wantStatus     int
		wantBody       string
	}{
		{
			name:           "Valid request",
			requestURL:     "/",
			requestMethod:  http.MethodPost,
			requestBody:    "https://example.com",
			contentType:    "text/plain",
			mockShortenURL: "http://localhost:8080/abc123",
			mockShortenErr: nil,
			wantStatus:     http.StatusCreated,
			wantBody:       "http://localhost:8080/abc123",
		},
		{
			name:          "Empty URL",
			requestURL:    "/",
			requestMethod: http.MethodPost,
			requestBody:   "",
			contentType:   "text/plain",
			wantStatus:    http.StatusBadRequest,
			wantBody:      "",
		},
		{
			name:          "Invalid content type",
			requestURL:    "/",
			requestMethod: http.MethodPost,
			requestBody:   "https://example.com",
			contentType:   "application/json",
			wantStatus:    http.StatusBadRequest,
			wantBody:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockURLService{
				shortenURLFunc: func(originalURL string) (string, error) {
					return tt.mockShortenURL, tt.mockShortenErr
				},
			}

			handler := NewHandler(mockService)

			req := httptest.NewRequest(tt.requestMethod, tt.requestURL, bytes.NewBufferString(tt.requestBody))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			rr := httptest.NewRecorder()

			handler.handleShorten(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("handler.handleShorten() status = %v, want %v", rr.Code, tt.wantStatus)
			}

			if tt.wantBody != "" && strings.TrimSpace(rr.Body.String()) != tt.wantBody {
				t.Errorf("handler.handleShorten() body = %v, want %v", rr.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestHandler_handleRedirect(t *testing.T) {
	tests := []struct {
		name         string
		urlID        string
		mockOrigURL  string
		mockFound    bool
		wantStatus   int
		wantLocation string
	}{
		{
			name:         "Valid redirect",
			urlID:        "abc123",
			mockOrigURL:  "https://example.com",
			mockFound:    true,
			wantStatus:   http.StatusTemporaryRedirect,
			wantLocation: "https://example.com",
		},
		{
			name:         "ID not found",
			urlID:        "nonexistent",
			mockOrigURL:  "",
			mockFound:    false,
			wantStatus:   http.StatusBadRequest,
			wantLocation: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockURLService{
				getOriginalURLFunc: func(id string) (string, bool) {
					return tt.mockOrigURL, tt.mockFound
				},
			}

			handler := NewHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/"+tt.urlID, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.urlID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()

			handler.handleRedirect(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("handler.handleRedirect() status = %v, want %v", rr.Code, tt.wantStatus)
			}

			if tt.wantLocation != "" {
				location := rr.Header().Get("Location")
				if location != tt.wantLocation {
					t.Errorf("handler.handleRedirect() Location = %v, want %v", location, tt.wantLocation)
				}
			}
		})
	}
}

func TestHandler_RegisterRoutes(t *testing.T) {
	mockService := &mockURLService{}
	handler := NewHandler(mockService)

	router := handler.RegisterRoutes()
	if router == nil {
		t.Error("handler.RegisterRoutes() returned nil")
	}

	chiRouter, ok := router.(*chi.Mux)
	if !ok {
		t.Error("handler.RegisterRoutes() did not return a chi.Mux")
	}

	if chiRouter == nil {
		t.Error("Failed to create Chi router")
	}
}
