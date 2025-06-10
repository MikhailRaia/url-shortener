package service

import (
	"fmt"
	"github.com/MikhailRaia/url-shortener/internal/model"
	"github.com/MikhailRaia/url-shortener/internal/storage"
)

type URLService struct {
	storage storage.URLStorage
	baseURL string
}

func NewURLService(storage storage.URLStorage, baseURL string) *URLService {
	return &URLService{
		storage: storage,
		baseURL: baseURL,
	}
}

func (s *URLService) ShortenURL(originalURL string) (string, error) {
	id, err := s.storage.Save(originalURL)
	if err != nil {
		if err == storage.ErrURLExists && id != "" {
			shortenedURL := fmt.Sprintf("%s/%s", s.baseURL, id)
			return shortenedURL, err
		}
		return "", err
	}

	shortenedURL := fmt.Sprintf("%s/%s", s.baseURL, id)
	return shortenedURL, nil
}

func (s *URLService) GetOriginalURL(id string) (string, bool) {
	return s.storage.Get(id)
}

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
