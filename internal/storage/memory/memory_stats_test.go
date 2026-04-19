package memory

import (
	"testing"

	"github.com/MikhailRaia/url-shortener/internal/model"
)

func TestGetStats(t *testing.T) {
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
		{
			name: "User with deleted and non-deleted URLs",
			setupFunc: func(s *Storage) {
				id1, _ := s.SaveWithUser("https://example.com", "user1")
				_, _ = s.SaveWithUser("https://google.com", "user1")
				s.SaveWithUser("https://github.com", "user1")
				s.DeleteUserURLs("user1", []string{id1})
			},
			expectedURLs:  2,
			expectedUsers: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewStorage()
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
	storage := NewStorage()

	items := []model.BatchRequestItem{
		{CorrelationID: "1", OriginalURL: "https://example.com"},
		{CorrelationID: "2", OriginalURL: "https://google.com"},
		{CorrelationID: "3", OriginalURL: "https://github.com"},
	}

	storage.SaveBatchWithUser(items, "user1")
	storage.SaveBatchWithUser(items, "user2")

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
