package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MikhailRaia/url-shortener/internal/model"
	"github.com/MikhailRaia/url-shortener/internal/storage"
)

type mockURLServiceForStats struct {
	stats *storage.Stats
	err   error
}

func (m *mockURLServiceForStats) ShortenURL(ctx context.Context, originalURL string) (string, error) {
	return "", nil
}

func (m *mockURLServiceForStats) ShortenURLWithUser(ctx context.Context, originalURL, userID string) (string, error) {
	return "", nil
}

func (m *mockURLServiceForStats) GetOriginalURL(ctx context.Context, id string) (string, bool) {
	return "", false
}

func (m *mockURLServiceForStats) GetOriginalURLWithDeletedStatus(ctx context.Context, id string) (string, error) {
	return "", nil
}

func (m *mockURLServiceForStats) ShortenBatch(ctx context.Context, items []model.BatchRequestItem) ([]model.BatchResponseItem, error) {
	return nil, nil
}

func (m *mockURLServiceForStats) ShortenBatchWithUser(ctx context.Context, items []model.BatchRequestItem, userID string) ([]model.BatchResponseItem, error) {
	return nil, nil
}

func (m *mockURLServiceForStats) GetUserURLs(ctx context.Context, userID string) ([]model.UserURL, error) {
	return nil, nil
}

func (m *mockURLServiceForStats) DeleteUserURLs(userID string, urlIDs []string) error {
	return nil
}

func (m *mockURLServiceForStats) GetStats() (*storage.Stats, error) {
	return m.stats, m.err
}

func TestHandleStats(t *testing.T) {
	tests := []struct {
		name           string
		trustedSubnet  string
		xRealIP        string
		stats          *storage.Stats
		statsErr       error
		expectedStatus int
		expectedBody   *storage.Stats
	}{
		{
			name:           "Empty trusted subnet returns 403",
			trustedSubnet:  "",
			xRealIP:        "192.168.1.10",
			stats:          &storage.Stats{URLs: 10, Users: 5},
			expectedStatus: http.StatusForbidden,
			expectedBody:   nil,
		},
		{
			name:           "Missing X-Real-IP returns 403",
			trustedSubnet:  "192.168.1.0/24",
			xRealIP:        "",
			stats:          &storage.Stats{URLs: 10, Users: 5},
			expectedStatus: http.StatusForbidden,
			expectedBody:   nil,
		},
		{
			name:           "IP not in trusted subnet returns 403",
			trustedSubnet:  "192.168.1.0/24",
			xRealIP:        "192.168.2.10",
			stats:          &storage.Stats{URLs: 10, Users: 5},
			expectedStatus: http.StatusForbidden,
			expectedBody:   nil,
		},
		{
			name:           "IP in trusted subnet returns stats",
			trustedSubnet:  "192.168.1.0/24",
			xRealIP:        "192.168.1.10",
			stats:          &storage.Stats{URLs: 10, Users: 5},
			expectedStatus: http.StatusOK,
			expectedBody:   &storage.Stats{URLs: 10, Users: 5},
		},
		{
			name:           "Loopback IP in loopback subnet returns stats",
			trustedSubnet:  "127.0.0.0/8",
			xRealIP:        "127.0.0.1",
			stats:          &storage.Stats{URLs: 20, Users: 3},
			expectedStatus: http.StatusOK,
			expectedBody:   &storage.Stats{URLs: 20, Users: 3},
		},
		{
			name:           "Service error returns 500",
			trustedSubnet:  "192.168.1.0/24",
			xRealIP:        "192.168.1.10",
			stats:          nil,
			statsErr:       storage.ErrURLExists,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   nil,
		},
		{
			name:           "Zero stats returns correct response",
			trustedSubnet:  "192.168.1.0/24",
			xRealIP:        "192.168.1.10",
			stats:          &storage.Stats{URLs: 0, Users: 0},
			expectedStatus: http.StatusOK,
			expectedBody:   &storage.Stats{URLs: 0, Users: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockURLServiceForStats{
				stats: tt.stats,
				err:   tt.statsErr,
			}

			handler := &Handler{
				urlService:    mockService,
				trustedSubnet: tt.trustedSubnet,
			}

			req := httptest.NewRequest("GET", "/api/internal/stats", nil)
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}

			w := httptest.NewRecorder()
			handler.handleStats(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("handleStats() status = %v, want %v", w.Code, tt.expectedStatus)
			}

			if tt.expectedBody != nil {
				var result storage.Stats
				if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
					t.Errorf("Failed to decode response: %v", err)
					return
				}

				if result.URLs != tt.expectedBody.URLs || result.Users != tt.expectedBody.Users {
					t.Errorf("handleStats() body = %+v, want %+v", result, tt.expectedBody)
				}

				if contentType := w.Header().Get("Content-Type"); contentType != "application/json" {
					t.Errorf("handleStats() Content-Type = %v, want application/json", contentType)
				}
			}
		})
	}
}
