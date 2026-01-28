package memory

import (
	"fmt"
	"github.com/MikhailRaia/url-shortener/internal/generator"
	"github.com/MikhailRaia/url-shortener/internal/model"
	"github.com/MikhailRaia/url-shortener/internal/storage"
	"sync"
)

// Storage implements in-memory URLStorage for testing and development.
type Storage struct {
	urlMap     map[string]string
	userURLs   map[string][]model.URL
	deletedMap map[string]bool
	mutex      sync.RWMutex
}

// NewStorage creates a new in-memory storage instance.
func NewStorage() *Storage {
	return &Storage{
		urlMap:     make(map[string]string),
		userURLs:   make(map[string][]model.URL),
		deletedMap: make(map[string]bool),
	}
}

// Save stores a new URL and returns its generated short ID.
func (s *Storage) Save(originalURL string) (string, error) {
	id, err := generator.GenerateID(8)
	if err != nil {
		return "", err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.urlMap[id] = originalURL
	return id, nil
}

// Get retrieves the original URL for a given short ID.
func (s *Storage) Get(id string) (string, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	originalURL, found := s.urlMap[id]
	if !found {
		return "", false
	}

	if s.deletedMap[id] {
		return "", false
	}

	return originalURL, true
}

// GetWithDeletedStatus retrieves the original URL and checks if it has been deleted.
func (s *Storage) GetWithDeletedStatus(id string) (string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	originalURL, found := s.urlMap[id]
	if !found {
		return "", nil
	}

	if s.deletedMap[id] {
		return "", storage.ErrURLDeleted
	}

	return originalURL, nil
}

// SaveBatch stores multiple URLs and returns a map of correlation IDs to short IDs.
func (s *Storage) SaveBatch(items []model.BatchRequestItem) (map[string]string, error) {
	result := make(map[string]string)

	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, item := range items {
		id, err := generator.GenerateID(8)
		if err != nil {
			return nil, fmt.Errorf("failed to generate ID: %w", err)
		}

		s.urlMap[id] = item.OriginalURL
		result[item.CorrelationID] = id
	}

	return result, nil
}

// SaveWithUser stores a new URL associated with a user and returns its generated short ID.
func (s *Storage) SaveWithUser(originalURL, userID string) (string, error) {
	id, err := generator.GenerateID(8)
	if err != nil {
		return "", err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.urlMap[id] = originalURL

	url := model.URL{
		ID:          id,
		OriginalURL: originalURL,
		UserID:      userID,
	}
	s.userURLs[userID] = append(s.userURLs[userID], url)

	return id, nil
}

// SaveBatchWithUser stores multiple URLs associated with a user and returns a map of correlation IDs to short IDs.
func (s *Storage) SaveBatchWithUser(items []model.BatchRequestItem, userID string) (map[string]string, error) {
	result := make(map[string]string)

	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, item := range items {
		id, err := generator.GenerateID(8)
		if err != nil {
			return nil, fmt.Errorf("failed to generate ID: %w", err)
		}

		s.urlMap[id] = item.OriginalURL
		result[item.CorrelationID] = id

		url := model.URL{
			ID:          id,
			OriginalURL: item.OriginalURL,
			UserID:      userID,
		}
		s.userURLs[userID] = append(s.userURLs[userID], url)
	}

	return result, nil
}

// GetUserURLs retrieves all non-deleted URLs associated with a user.
func (s *Storage) GetUserURLs(userID string) ([]model.UserURL, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	urls, exists := s.userURLs[userID]
	if !exists {
		return []model.UserURL{}, nil
	}

	var result []model.UserURL
	for _, url := range urls {
		if !s.deletedMap[url.ID] {
			result = append(result, model.UserURL{
				ShortURL:    url.ID,
				OriginalURL: url.OriginalURL,
			})
		}
	}

	return result, nil
}

// DeleteUserURLs marks specified URLs as deleted for a user.
func (s *Storage) DeleteUserURLs(userID string, urlIDs []string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	userURLs, exists := s.userURLs[userID]
	if !exists {
		return nil
	}

	// Create a set of URLs that belong to the user
	userURLSet := make(map[string]bool)
	for _, url := range userURLs {
		userURLSet[url.ID] = true
	}

	// Mark URLs as deleted only if they belong to the user
	for _, urlID := range urlIDs {
		if userURLSet[urlID] {
			s.deletedMap[urlID] = true
		}
	}

	return nil
}

// GetStats returns total number of URLs and users.
func (s *Storage) GetStats() (int, int, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return len(s.urlMap), len(s.userURLs), nil
}
