package service

import (
	"fmt"
	"github.com/MikhailRaia/url-shortener/internal/model"
	"github.com/MikhailRaia/url-shortener/internal/storage"
	"net/url"
)

// URLService provides business logic for creating and resolving short URLs.
type URLService struct {
	storage storage.URLStorage
	baseURL string
}

// NewURLService constructs a URLService with the given storage and base URL.
func NewURLService(storage storage.URLStorage, baseURL string) *URLService {
	return &URLService{
		storage: storage,
		baseURL: baseURL,
	}
}

// ShortenURL creates a short URL and returns its absolute form.
func (s *URLService) ShortenURL(originalURL string) (string, error) {
	id, err := s.storage.Save(originalURL)
	if err != nil {
		if err == storage.ErrURLExists && id != "" {
			shortenedURL, _ := url.JoinPath(s.baseURL, id)
			return shortenedURL, err
		}
		return "", err
	}

	shortenedURL, _ := url.JoinPath(s.baseURL, id)
	return shortenedURL, nil
}

// GetOriginalURL resolves an ID to the original URL if it exists and not deleted.
func (s *URLService) GetOriginalURL(id string) (string, bool) {
	return s.storage.Get(id)
}

// GetOriginalURLWithDeletedStatus resolves an ID and reports deletion via error.
func (s *URLService) GetOriginalURLWithDeletedStatus(id string) (string, error) {
	return s.storage.GetWithDeletedStatus(id)
}

// ShortenBatch creates short URLs for a batch of items.
func (s *URLService) ShortenBatch(items []model.BatchRequestItem) ([]model.BatchResponseItem, error) {
	idMap, err := s.storage.SaveBatch(items)
	if err != nil {
		return nil, fmt.Errorf("error saving batch: %w", err)
	}

	result := make([]model.BatchResponseItem, 0, len(items))
	for _, item := range items {
		id, ok := idMap[item.CorrelationID]
		if !ok {
			continue
		}

		shortURL := fmt.Sprintf("%s/%s", s.baseURL, id)
		result = append(result, model.BatchResponseItem{
			CorrelationID: item.CorrelationID,
			ShortURL:      shortURL,
		})
	}

	return result, nil
}

// ShortenURLWithUser creates a short URL associated with a user.
func (s *URLService) ShortenURLWithUser(originalURL, userID string) (string, error) {
	id, err := s.storage.SaveWithUser(originalURL, userID)
	if err != nil {
		if err == storage.ErrURLExists && id != "" {
			shortenedURL, _ := url.JoinPath(s.baseURL, id)
			return shortenedURL, err
		}
		return "", err
	}

	shortenedURL, _ := url.JoinPath(s.baseURL, id)
	return shortenedURL, nil
}

// ShortenBatchWithUser creates short URLs for a batch and associates them with a user.
func (s *URLService) ShortenBatchWithUser(items []model.BatchRequestItem, userID string) ([]model.BatchResponseItem, error) {
	idMap, err := s.storage.SaveBatchWithUser(items, userID)
	if err != nil {
		return nil, fmt.Errorf("error saving batch: %w", err)
	}

	result := make([]model.BatchResponseItem, 0, len(items))
	for _, item := range items {
		id, ok := idMap[item.CorrelationID]
		if !ok {
			continue
		}

		shortURL := fmt.Sprintf("%s/%s", s.baseURL, id)
		result = append(result, model.BatchResponseItem{
			CorrelationID: item.CorrelationID,
			ShortURL:      shortURL,
		})
	}

	return result, nil
}

// GetUserURLs returns all URLs belonging to a user, excluding deleted ones.
func (s *URLService) GetUserURLs(userID string) ([]model.UserURL, error) {
	urls, err := s.storage.GetUserURLs(userID)
	if err != nil {
		return nil, fmt.Errorf("error getting user URLs: %w", err)
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

// DeleteUserURLs marks user's URLs as deleted.
func (s *URLService) DeleteUserURLs(userID string, urlIDs []string) error {
	return s.storage.DeleteUserURLs(userID, urlIDs)
}
