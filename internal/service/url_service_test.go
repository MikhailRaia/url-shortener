package service

import (
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

func (m *mockStorage) GetWithDeletedStatus(id string) (string, bool, error) {
	if m.getWithDeletedStatusFunc != nil {
		return m.getWithDeletedStatusFunc(id)
	}
	return "", false, nil
}

func (m *mockStorage) DeleteUserURLs(userID string, urlIDs []string) error {
	if m.deleteUserURLsFunc != nil {
		return m.deleteUserURLsFunc(userID, urlIDs)
	}
	return nil
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
			got, err := service.ShortenURL(tt.originalURL)

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
			gotURL, gotFound := service.GetOriginalURL(tt.id)

			if gotFound != tt.wantFound {
				t.Errorf("URLService.GetOriginalURL() found = %v, want %v", gotFound, tt.wantFound)
			}

			if gotURL != tt.wantURL {
				t.Errorf("URLService.GetOriginalURL() = %v, want %v", gotURL, tt.wantURL)
			}
		})
	}
}
