package service

import (
	"github.com/MikhailRaia/url-shortener/internal/storage"
	"path"
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

	shortenedURL := path.Join(s.baseURL, id)
	return shortenedURL, nil
}

func (s *URLService) GetOriginalURL(id string) (string, bool) {
	return s.storage.Get(id)
}
