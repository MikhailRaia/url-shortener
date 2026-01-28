package handler

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MikhailRaia/url-shortener/internal/config"
	"github.com/MikhailRaia/url-shortener/internal/middleware"
	"github.com/MikhailRaia/url-shortener/internal/model"
	"github.com/go-chi/chi/v5"
)

type MockGzipURLService struct{}

func (m *MockGzipURLService) ShortenURL(ctx context.Context, originalURL string) (string, error) {
	return "http://localhost:8080/abc123", nil
}

func (m *MockGzipURLService) ShortenURLWithUser(ctx context.Context, originalURL, userID string) (string, error) {
	return "http://localhost:8080/abc123", nil
}

func (m *MockGzipURLService) GetOriginalURL(ctx context.Context, id string) (string, bool) {
	if id == "abc123" {
		return "https://example.com", true
	}
	return "", false
}

func (m *MockGzipURLService) ShortenBatch(ctx context.Context, items []model.BatchRequestItem) ([]model.BatchResponseItem, error) {
	result := make([]model.BatchResponseItem, 0, len(items))
	for _, item := range items {
		result = append(result, model.BatchResponseItem{
			CorrelationID: item.CorrelationID,
			ShortURL:      "http://localhost:8080/batch" + item.CorrelationID,
		})
	}
	return result, nil
}

func (m *MockGzipURLService) ShortenBatchWithUser(ctx context.Context, items []model.BatchRequestItem, userID string) ([]model.BatchResponseItem, error) {
	result := make([]model.BatchResponseItem, 0, len(items))
	for _, item := range items {
		result = append(result, model.BatchResponseItem{
			CorrelationID: item.CorrelationID,
			ShortURL:      "http://localhost:8080/batch" + item.CorrelationID,
		})
	}
	return result, nil
}

func (m *MockGzipURLService) GetUserURLs(ctx context.Context, userID string) ([]model.UserURL, error) {
	return []model.UserURL{}, nil
}

func (m *MockGzipURLService) GetOriginalURLWithDeletedStatus(ctx context.Context, id string) (string, error) {
	if id == "abc123" {
		return "https://example.com", nil
	}
	return "", nil
}

func (m *MockGzipURLService) DeleteUserURLs(userID string, urlIDs []string) error {
	return nil
}

func (m *MockGzipURLService) GetStats(ctx context.Context) (int, int, error) {
	return 0, 0, nil
}

func TestGzipCompression(t *testing.T) {
	h := NewHandler(&MockGzipURLService{}, nil, &config.Config{})

	r := chi.NewRouter()
	r.Use(middleware.GzipReader)
	r.Use(middleware.GzipMiddleware)
	r.Post("/api/shorten", h.HandleShortenJSON)

	reqBody := ShortenRequest{URL: "https://example.com"}
	reqBodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")

	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, rec.Code)
	}

	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("Expected Content-Encoding to be gzip, got %s", rec.Header().Get("Content-Encoding"))
	}

	reader, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer reader.Close()

	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to read gzipped response: %v", err)
	}

	var response ShortenResponse
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Result != "http://localhost:8080/abc123" {
		t.Errorf("Expected result to be %s, got %s", "http://localhost:8080/abc123", response.Result)
	}
}

func TestGzipDecompression(t *testing.T) {
	h := NewHandler(&MockGzipURLService{}, nil, &config.Config{})

	r := chi.NewRouter()
	r.Use(middleware.GzipReader)
	r.Use(middleware.GzipMiddleware)
	r.Post("/api/shorten", h.HandleShortenJSON)

	reqBody := ShortenRequest{URL: "https://example.com"}
	reqBodyBytes, _ := json.Marshal(reqBody)

	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	_, err := gzWriter.Write(reqBodyBytes)
	if err != nil {
		t.Fatalf("Failed to write to gzip writer: %v", err)
	}
	gzWriter.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/shorten", &buf)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, rec.Code)
	}

	var response ShortenResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Result != "http://localhost:8080/abc123" {
		t.Errorf("Expected result to be %s, got %s", "http://localhost:8080/abc123", response.Result)
	}
}

func TestTextPlainGzipCompression(t *testing.T) {
	h := NewHandler(&MockGzipURLService{}, nil, &config.Config{})

	r := chi.NewRouter()
	r.Use(middleware.GzipReader)
	r.Use(middleware.GzipMiddleware)
	r.Post("/", h.handleShorten)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com"))
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Accept-Encoding", "gzip")

	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, rec.Code)
	}

	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("Expected Content-Encoding to be gzip, got %s", rec.Header().Get("Content-Encoding"))
	}

	reader, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer reader.Close()

	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to read gzipped response: %v", err)
	}

	if string(body) != "http://localhost:8080/abc123" {
		t.Errorf("Expected result to be %s, got %s", "http://localhost:8080/abc123", string(body))
	}
}
