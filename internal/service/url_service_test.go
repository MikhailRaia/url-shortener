package service

import (
	"context"
	"errors"
	"testing"

	"github.com/MikhailRaia/url-shortener/internal/model"
)

type mockStorage struct {
	saveFunc                 func(originalURL string) (string, error)
	saveWithUserFunc         func(originalURL, userID string) (string, error)
	getFunc                  func(id string) (string, bool)
	getWithDeletedStatusFunc func(id string) (string, bool, error)
	saveBatchFunc            func(items []model.BatchRequestItem) (map[string]string, error)
	saveBatchWithUserFunc    func(items []model.BatchRequestItem, userID string) (map[string]string, error)
	getUserURLsFunc          func(userID string) ([]model.UserURL, error)
	deleteUserURLsFunc       func(userID string, urlIDs []string) error
	getStatsFunc             func() (int, int, error)
}

func (m *mockStorage) Save(originalURL string) (string, error) {
	return m.saveFunc(originalURL)
}

func (m *mockStorage) SaveWithUser(originalURL, userID string) (string, error) {
	if m.saveWithUserFunc != nil {
		return m.saveWithUserFunc(originalURL, userID)
	}
	return "", nil
}

func (m *mockStorage) Get(id string) (string, bool) {
	return m.getFunc(id)
}

func (m *mockStorage) SaveBatch(items []model.BatchRequestItem) (map[string]string, error) {
	if m.saveBatchFunc != nil {
		return m.saveBatchFunc(items)
	}
	return make(map[string]string), nil
}

func (m *mockStorage) SaveBatchWithUser(items []model.BatchRequestItem, userID string) (map[string]string, error) {
	if m.saveBatchWithUserFunc != nil {
		return m.saveBatchWithUserFunc(items, userID)
	}
	return make(map[string]string), nil
}

func (m *mockStorage) GetUserURLs(userID string) ([]model.UserURL, error) {
	if m.getUserURLsFunc != nil {
		return m.getUserURLsFunc(userID)
	}
	return []model.UserURL{}, nil
}

func (m *mockStorage) GetWithDeletedStatus(id string) (string, error) {
	if m.getWithDeletedStatusFunc != nil {
		str, _, err := m.getWithDeletedStatusFunc(id)
		return str, err
	}
	return "", nil
}

func (m *mockStorage) DeleteUserURLs(userID string, urlIDs []string) error {
	if m.deleteUserURLsFunc != nil {
		return m.deleteUserURLsFunc(userID, urlIDs)
	}
	return nil
}

func (m *mockStorage) GetStats() (int, int, error) {
	if m.getStatsFunc != nil {
		return m.getStatsFunc()
	}
	return 0, 0, nil
}

func TestURLService_ShortenURL(t *testing.T) {
	baseURL := "http://localhost:8080"

	tests := []struct {
		name        string
		originalURL string
		mockID      string
		mockErr     error
		want        string
		wantErr     bool
	}{
		{
			name:        "Successful shortening",
			originalURL: "https://example.com",
			mockID:      "abc123",
			mockErr:     nil,
			want:        "http://localhost:8080/abc123",
			wantErr:     false,
		},
		{
			name:        "Storage error",
			originalURL: "https://example.com",
			mockID:      "",
			mockErr:     errors.New("storage error"),
			want:        "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := &mockStorage{
				saveFunc: func(originalURL string) (string, error) {
					return tt.mockID, tt.mockErr
				},
			}

			service := NewURLService(mockStorage, baseURL)
			got, err := service.ShortenURL(context.Background(), tt.originalURL)

			if (err != nil) != tt.wantErr {
				t.Errorf("URLService.ShortenURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("URLService.ShortenURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestURLService_GetOriginalURL(t *testing.T) {
	baseURL := "http://localhost:8080"

	tests := []struct {
		name      string
		id        string
		mockURL   string
		mockFound bool
		wantURL   string
		wantFound bool
	}{
		{
			name:      "URL found",
			id:        "abc123",
			mockURL:   "https://example.com",
			mockFound: true,
			wantURL:   "https://example.com",
			wantFound: true,
		},
		{
			name:      "URL not found",
			id:        "nonexistent",
			mockURL:   "",
			mockFound: false,
			wantURL:   "",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := &mockStorage{
				getFunc: func(id string) (string, bool) {
					if id == tt.id {
						return tt.mockURL, tt.mockFound
					}
					return "", false
				},
			}

			service := NewURLService(mockStorage, baseURL)
			gotURL, gotFound := service.GetOriginalURL(context.Background(), tt.id)

			if gotFound != tt.wantFound {
				t.Errorf("URLService.GetOriginalURL() found = %v, want %v", gotFound, tt.wantFound)
			}

			if gotURL != tt.wantURL {
				t.Errorf("URLService.GetOriginalURL() = %v, want %v", gotURL, tt.wantURL)
			}
		})
	}
}

func BenchmarkURLService_ShortenURL(b *testing.B) {
	baseURL := "http://localhost:8080"
	mockStorage := &mockStorage{
		saveFunc: func(originalURL string) (string, error) {
			return "abc123", nil
		},
	}
	service := NewURLService(mockStorage, baseURL)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.ShortenURL(context.Background(), "https://example.com/very/long/url/path")
	}
}

func BenchmarkURLService_GetOriginalURL(b *testing.B) {
	baseURL := "http://localhost:8080"
	mockStorage := &mockStorage{
		getFunc: func(id string) (string, bool) {
			return "https://example.com", true
		},
	}
	service := NewURLService(mockStorage, baseURL)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.GetOriginalURL(context.Background(), "abc123")
	}
}

func BenchmarkURLService_ShortenBatch(b *testing.B) {
	baseURL := "http://localhost:8080"
	mockStorage := &mockStorage{
		saveBatchFunc: func(items []model.BatchRequestItem) (map[string]string, error) {
			result := make(map[string]string)
			for i := range items {
				result[items[i].CorrelationID] = "abc" + string(rune(i))
			}
			return result, nil
		},
	}
	service := NewURLService(mockStorage, baseURL)

	items := []model.BatchRequestItem{
		{CorrelationID: "1", OriginalURL: "https://example.com/1"},
		{CorrelationID: "2", OriginalURL: "https://example.com/2"},
		{CorrelationID: "3", OriginalURL: "https://example.com/3"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.ShortenBatch(context.Background(), items)
	}
}
