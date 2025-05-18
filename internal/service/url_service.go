package service

import (
	"fmt"
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
		return "", err
	}

	shortenedURL := fmt.Sprintf("%s/%s", s.baseURL, id)
	return shortenedURL, nil
}

func (s *URLService) GetOriginalURL(id string) (string, bool) {
	return s.storage.Get(id)
}
