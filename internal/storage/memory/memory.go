package memory

import (
	"github.com/MikhailRaia/url-shortener/internal/generator"
	"sync"
)

type Storage struct {
	urlMap map[string]string
	mutex  sync.RWMutex
}

func NewStorage() *Storage {
	return &Storage{
		urlMap: make(map[string]string),
	}
}

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

func (s *Storage) Get(id string) (string, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	originalURL, found := s.urlMap[id]
	return originalURL, found
}
