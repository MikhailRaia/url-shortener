package file

import (
	"os"
	"testing"

	"github.com/MikhailRaia/url-shortener/internal/model"
)

func TestGetStats(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_storage_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	tests := []struct {
		name          string
		setupFunc     func(*Storage)
		expectedURLs  int
		expectedUsers int
	}{
		{
			name: "Empty storage",
			setupFunc: func(s *Storage) {
			},
			expectedURLs:  0,
			expectedUsers: 0,
		},
		{
			name: "Single URL without user",
			setupFunc: func(s *Storage) {
				s.Save("https://example.com")
			},
			expectedURLs:  1,
			expectedUsers: 0,
		},
		{
			name: "Multiple URLs without users",
			setupFunc: func(s *Storage) {
				s.Save("https://example.com")
				s.Save("https://google.com")
				s.Save("https://github.com")
			},
			expectedURLs:  3,
			expectedUsers: 0,
		},
		{
			name: "Single URL with one user",
			setupFunc: func(s *Storage) {
				s.SaveWithUser("https://example.com", "user1")
			},
			expectedURLs:  1,
			expectedUsers: 1,
		},
		{
			name: "Multiple URLs with one user",
			setupFunc: func(s *Storage) {
				s.SaveWithUser("https://example.com", "user1")
				s.SaveWithUser("https://google.com", "user1")
			},
			expectedURLs:  2,
			expectedUsers: 1,
		},
		{
			name: "Multiple URLs with multiple users",
			setupFunc: func(s *Storage) {
				s.SaveWithUser("https://example.com", "user1")
				s.SaveWithUser("https://google.com", "user1")
				s.SaveWithUser("https://github.com", "user2")
				s.SaveWithUser("https://rust.com", "user3")
			},
			expectedURLs:  4,
			expectedUsers: 3,
		},
		{
			name: "Mixed URLs (with and without users)",
			setupFunc: func(s *Storage) {
				s.Save("https://example.com")
				s.SaveWithUser("https://google.com", "user1")
				s.Save("https://github.com")
				s.SaveWithUser("https://rust.com", "user2")
			},
			expectedURLs:  4,
			expectedUsers: 2,
		},
		{
			name: "Deleted URLs not counted",
			setupFunc: func(s *Storage) {
				id1, _ := s.SaveWithUser("https://example.com", "user1")
				id2, _ := s.SaveWithUser("https://google.com", "user1")
				s.SaveWithUser("https://github.com", "user2")
				s.DeleteUserURLs("user1", []string{id1, id2})
			},
			expectedURLs:  1,
			expectedUsers: 1,
		},
		{
			name: "All URLs deleted",
			setupFunc: func(s *Storage) {
				id1, _ := s.SaveWithUser("https://example.com", "user1")
				id2, _ := s.SaveWithUser("https://google.com", "user1")
				s.DeleteUserURLs("user1", []string{id1, id2})
			},
			expectedURLs:  0,
			expectedUsers: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Remove(tmpFile.Name())

			storage, err := NewStorage(tmpFile.Name())
			if err != nil {
				t.Fatalf("Failed to create storage: %v", err)
			}

			tt.setupFunc(storage)

			stats, err := storage.GetStats()
			if err != nil {
				t.Fatalf("GetStats() error = %v", err)
			}

			if stats.URLs != tt.expectedURLs {
				t.Errorf("GetStats() URLs = %v, want %v", stats.URLs, tt.expectedURLs)
			}

			if stats.Users != tt.expectedUsers {
				t.Errorf("GetStats() Users = %v, want %v", stats.Users, tt.expectedUsers)
			}
		})
	}
}

func TestGetStatsBatch(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_storage_batch_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	storage, err := NewStorage(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	items1 := []model.BatchRequestItem{
		{CorrelationID: "1", OriginalURL: "https://example.com"},
		{CorrelationID: "2", OriginalURL: "https://google.com"},
		{CorrelationID: "3", OriginalURL: "https://github.com"},
	}

	items2 := []model.BatchRequestItem{
		{CorrelationID: "4", OriginalURL: "https://rust.com"},
		{CorrelationID: "5", OriginalURL: "https://golang.com"},
		{CorrelationID: "6", OriginalURL: "https://python.com"},
	}

	storage.SaveBatchWithUser(items1, "user1")
	storage.SaveBatchWithUser(items2, "user2")

	stats, err := storage.GetStats()
	if err != nil {
		t.Fatalf("GetStats() error = %v", err)
	}

	if stats.URLs != 6 {
		t.Errorf("GetStats() URLs = %v, want 6", stats.URLs)
	}

	if stats.Users != 2 {
		t.Errorf("GetStats() Users = %v, want 2", stats.Users)
	}
}
