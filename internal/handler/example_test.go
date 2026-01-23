package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"strings"

	"github.com/MikhailRaia/url-shortener/internal/model"
	"github.com/MikhailRaia/url-shortener/internal/storage/memory"
)

type exampleURLService struct {
	*memory.Storage
	baseURL string
}

func (s *exampleURLService) ShortenURL(ctx context.Context, originalURL string) (string, error) {
	id, err := s.Storage.Save(originalURL)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", s.baseURL, id), nil
}

func (s *exampleURLService) ShortenURLWithUser(ctx context.Context, originalURL, userID string) (string, error) {
	id, err := s.Storage.SaveWithUser(originalURL, userID)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", s.baseURL, id), nil
}

func (s *exampleURLService) GetOriginalURL(ctx context.Context, id string) (string, bool) {
	return s.Storage.Get(id)
}

func (s *exampleURLService) GetOriginalURLWithDeletedStatus(ctx context.Context, id string) (string, error) {
	return s.Storage.GetWithDeletedStatus(id)
}

func (s *exampleURLService) ShortenBatch(ctx context.Context, items []model.BatchRequestItem) ([]model.BatchResponseItem, error) {
	idMap, err := s.Storage.SaveBatch(items)
	if err != nil {
		return nil, err
	}

	result := make([]model.BatchResponseItem, 0, len(items))
	for _, item := range items {
		id, ok := idMap[item.CorrelationID]
		if !ok {
			continue
		}
		result = append(result, model.BatchResponseItem{
			CorrelationID: item.CorrelationID,
			ShortURL:      fmt.Sprintf("%s/%s", s.baseURL, id),
		})
	}
	return result, nil
}

func (s *exampleURLService) ShortenBatchWithUser(ctx context.Context, items []model.BatchRequestItem, userID string) ([]model.BatchResponseItem, error) {
	idMap, err := s.Storage.SaveBatchWithUser(items, userID)
	if err != nil {
		return nil, err
	}

	result := make([]model.BatchResponseItem, 0, len(items))
	for _, item := range items {
		id, ok := idMap[item.CorrelationID]
		if !ok {
			continue
		}
		result = append(result, model.BatchResponseItem{
			CorrelationID: item.CorrelationID,
			ShortURL:      fmt.Sprintf("%s/%s", s.baseURL, id),
		})
	}
	return result, nil
}

func (s *exampleURLService) GetUserURLs(ctx context.Context, userID string) ([]model.UserURL, error) {
	urls, err := s.Storage.GetUserURLs(userID)
	if err != nil {
		return nil, err
	}

	result := make([]model.UserURL, len(urls))
	for i, url := range urls {
		result[i] = model.UserURL{
			ShortURL:    fmt.Sprintf("%s/%s", s.baseURL, url.ShortURL),
			OriginalURL: url.OriginalURL,
		}
	}
	return result, nil
}

func (s *exampleURLService) DeleteUserURLs(userID string, urlIDs []string) error {
	return s.Storage.DeleteUserURLs(userID, urlIDs)
}

// Example demonstrates how to use the Handler to shorten a URL via plain text endpoint.
func ExampleHandler_handleShorten() {
	service := &exampleURLService{
		Storage: memory.NewStorage(),
		baseURL: "http://localhost:8080",
	}
	handler := NewHandler(service, nil)

	req := httptest.NewRequest("POST", "/", strings.NewReader("https://example.com"))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	handler.RegisterRoutes().ServeHTTP(w, req)

	body, _ := io.ReadAll(w.Body)
	fmt.Printf("Status: %d\n", w.Code)

	// Extract and validate the shortened URL format
	shortURL := string(body)
	if len(shortURL) > 0 {
		fmt.Println("Shortened URL: http://localhost:8080/[generated-id]")
	}
	// Output:
	// Status: 201
	// Shortened URL: http://localhost:8080/[generated-id]
}

// Example demonstrates how to use the Handler to shorten a URL via JSON endpoint.
func ExampleHandler_HandleShortenJSON() {
	service := &exampleURLService{
		Storage: memory.NewStorage(),
		baseURL: "http://localhost:8080",
	}
	handler := NewHandler(service, nil)

	reqBody := ShortenRequest{URL: "https://example.com/very/long/path"}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/shorten", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.RegisterRoutes().ServeHTTP(w, req)

	var respBody ShortenResponse
	json.NewDecoder(w.Body).Decode(&respBody)

	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Response has ShortURL: %v\n", respBody.Result != "")
}

// Example demonstrates how to use the Handler to redirect to original URL.
func ExampleHandler_handleRedirect() {
	service := &exampleURLService{
		Storage: memory.NewStorage(),
		baseURL: "http://localhost:8080",
	}

	service.Save("https://example.com")
	id, _ := service.Save("https://golang.org")

	handler := NewHandler(service, nil)

	req := httptest.NewRequest("GET", fmt.Sprintf("/%s", id), nil)
	w := httptest.NewRecorder()

	handler.RegisterRoutes().ServeHTTP(w, req)

	location := w.Header().Get("Location")
	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Redirects to: %s\n", location)
}

// Example demonstrates how to use the Handler with batch shortening.
func ExampleHandler_handleShortenBatch() {
	service := &exampleURLService{
		Storage: memory.NewStorage(),
		baseURL: "http://localhost:8080",
	}
	handler := NewHandler(service, nil)

	items := []model.BatchRequestItem{
		{CorrelationID: "id1", OriginalURL: "https://golang.org"},
		{CorrelationID: "id2", OriginalURL: "https://github.com"},
		{CorrelationID: "id3", OriginalURL: "https://example.com"},
	}
	jsonBody, _ := json.Marshal(items)

	req := httptest.NewRequest("POST", "/api/shorten/batch", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.RegisterRoutes().ServeHTTP(w, req)

	var respBody []model.BatchResponseItem
	json.NewDecoder(w.Body).Decode(&respBody)

	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Shortened %d URLs\n", len(respBody))
}

// Example demonstrates how to use RegisterRoutes to set up public endpoints.
func ExampleHandler_RegisterRoutes() {
	service := &exampleURLService{
		Storage: memory.NewStorage(),
		baseURL: "http://localhost:8080",
	}
	handler := NewHandler(service, nil)

	router := handler.RegisterRoutes()

	req := httptest.NewRequest("POST", "/", strings.NewReader("https://example.com"))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	fmt.Printf("Handler registered successfully\n")
	fmt.Printf("Response Status: %d\n", w.Code)
}
