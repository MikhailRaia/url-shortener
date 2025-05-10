package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
		{
			name:          "Invalid method",
			requestURL:    "/",
			requestMethod: http.MethodGet,
			requestBody:   "https://example.com",
			contentType:   "text/plain",
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

			handler.handleRequest(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("handler.handleRequest() status = %v, want %v", rr.Code, tt.wantStatus)
			}

			if tt.wantBody != "" && strings.TrimSpace(rr.Body.String()) != tt.wantBody {
				t.Errorf("handler.handleRequest() body = %v, want %v", rr.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestHandler_handleRedirect(t *testing.T) {
	tests := []struct {
		name          string
		requestURL    string
		requestMethod string
		mockOrigURL   string
		mockFound     bool
		wantStatus    int
		wantLocation  string
	}{
		{
			name:          "Valid redirect",
			requestURL:    "/abc123",
			requestMethod: http.MethodGet,
			mockOrigURL:   "https://example.com",
			mockFound:     true,
			wantStatus:    http.StatusTemporaryRedirect,
			wantLocation:  "https://example.com",
		},
		{
			name:          "ID not found",
			requestURL:    "/nonexistent",
			requestMethod: http.MethodGet,
			mockOrigURL:   "",
			mockFound:     false,
			wantStatus:    http.StatusBadRequest,
			wantLocation:  "",
		},
		{
			name:          "Invalid method",
			requestURL:    "/abc123",
			requestMethod: http.MethodPost,
			wantStatus:    http.StatusBadRequest,
			wantLocation:  "",
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

			req := httptest.NewRequest(tt.requestMethod, tt.requestURL, nil)

			rr := httptest.NewRecorder()

			handler.handleRequest(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("handler.handleRequest() status = %v, want %v", rr.Code, tt.wantStatus)
			}

			if tt.wantLocation != "" {
				location := rr.Header().Get("Location")
				if location != tt.wantLocation {
					t.Errorf("handler.handleRequest() Location = %v, want %v", location, tt.wantLocation)
				}
			}
		})
	}
}

func TestHandler_RegisterRoutes(t *testing.T) {
	mockService := &mockURLService{}
	handler := NewHandler(mockService)

	mux := handler.RegisterRoutes()
	if mux == nil {
		t.Error("handler.RegisterRoutes() returned nil")
	}
}
