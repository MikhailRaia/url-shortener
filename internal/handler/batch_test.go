package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MikhailRaia/url-shortener/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockBatchURLService struct{}

func (m *MockBatchURLService) ShortenURL(ctx context.Context, originalURL string) (string, error) {
	return "http://localhost:8080/abc123", nil
}

func (m *MockBatchURLService) ShortenURLWithUser(ctx context.Context, originalURL, userID string) (string, error) {
	return "http://localhost:8080/abc123", nil
}

func (m *MockBatchURLService) GetOriginalURL(ctx context.Context, id string) (string, bool) {
	if id == "abc123" {
		return "https://example.com", true
	}
	return "", false
}

func (m *MockBatchURLService) ShortenBatch(ctx context.Context, items []model.BatchRequestItem) ([]model.BatchResponseItem, error) {
	result := make([]model.BatchResponseItem, 0, len(items))
	for _, item := range items {
		result = append(result, model.BatchResponseItem{
			CorrelationID: item.CorrelationID,
			ShortURL:      "http://localhost:8080/batch" + item.CorrelationID,
		})
	}
	return result, nil
}

func (m *MockBatchURLService) ShortenBatchWithUser(ctx context.Context, items []model.BatchRequestItem, userID string) ([]model.BatchResponseItem, error) {
	result := make([]model.BatchResponseItem, 0, len(items))
	for _, item := range items {
		result = append(result, model.BatchResponseItem{
			CorrelationID: item.CorrelationID,
			ShortURL:      "http://localhost:8080/batch" + item.CorrelationID,
		})
	}
	return result, nil
}

func (m *MockBatchURLService) GetUserURLs(ctx context.Context, userID string) ([]model.UserURL, error) {
	return []model.UserURL{}, nil
}

func (m *MockBatchURLService) GetOriginalURLWithDeletedStatus(ctx context.Context, id string) (string, error) {
	if id == "abc123" {
		return "https://example.com", nil
	}
	return "", nil
}

func (m *MockBatchURLService) DeleteUserURLs(userID string, urlIDs []string) error {
	return nil
}

func TestHandleShortenBatch(t *testing.T) {
	h := NewHandler(&MockBatchURLService{}, nil)

	r := chi.NewRouter()
	r.Post("/api/shorten/batch", h.handleShortenBatch)

	items := []model.BatchRequestItem{
		{
			CorrelationID: "1",
			OriginalURL:   "https://example.com",
		},
		{
			CorrelationID: "2",
			OriginalURL:   "https://example.org",
		},
	}

	body, err := json.Marshal(items)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var response []model.BatchResponseItem
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Len(t, response, 2)
	assert.Equal(t, "1", response[0].CorrelationID)
	assert.Equal(t, "http://localhost:8080/batch1", response[0].ShortURL)
	assert.Equal(t, "2", response[1].CorrelationID)
	assert.Equal(t, "http://localhost:8080/batch2", response[1].ShortURL)
}

func TestHandleShortenBatchInvalidJSON(t *testing.T) {
	h := NewHandler(&MockBatchURLService{}, nil)

	r := chi.NewRouter()
	r.Post("/api/shorten/batch", h.handleShortenBatch)

	invalidJSON := `[{"correlation_id": "1", "original_url": "https://example.com"},`

	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewBufferString(invalidJSON))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandleShortenBatchEmptyRequest(t *testing.T) {
	h := NewHandler(&MockBatchURLService{}, nil)

	r := chi.NewRouter()
	r.Post("/api/shorten/batch", h.handleShortenBatch)

	emptyArray := "[]"

	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewBufferString(emptyArray))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
