package file

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/MikhailRaia/url-shortener/internal/generator"
	"github.com/MikhailRaia/url-shortener/internal/model"
	"github.com/MikhailRaia/url-shortener/internal/storage"
)

type Storage struct {
	filePath      string
	urlMap        map[string]string
	reverseURLMap map[string]string
	userURLs      map[string][]model.URL
	deletedMap    map[string]bool
	idCounter     int
	mutex         sync.RWMutex
	fileWriteMu   sync.Mutex
}

func NewStorage(filePath string) (*Storage, error) {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	storage := &Storage{
		filePath:      filePath,
		urlMap:        make(map[string]string),
		reverseURLMap: make(map[string]string),
		userURLs:      make(map[string][]model.URL),
		deletedMap:    make(map[string]bool),
		idCounter:     0,
	}

	if err := storage.loadFromFile(); err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *Storage) Save(originalURL string) (string, error) {
	s.mutex.RLock()
	existingID, exists := s.reverseURLMap[originalURL]
	s.mutex.RUnlock()

	if exists {
		return existingID, storage.ErrURLExists
	}

	id, err := generator.GenerateID(8)
	if err != nil {
		return "", err
	}

	s.mutex.Lock()
	if existingID, exists := s.reverseURLMap[originalURL]; exists {
		s.mutex.Unlock()
		return existingID, storage.ErrURLExists
	}

	s.idCounter++
	uuid := strconv.Itoa(s.idCounter)
	s.urlMap[id] = originalURL
	s.reverseURLMap[originalURL] = id
	s.mutex.Unlock()

	record := model.URLRecord{
		UUID:        uuid,
		ShortURL:    id,
		OriginalURL: originalURL,
		UserID:      "",
		IsDeleted:   false,
	}

	if err := s.saveRecordToFile(record); err != nil {
		return "", err
	}

	return id, nil
}

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

func (s *Storage) GetWithDeletedStatus(id string) (string, bool, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	originalURL, found := s.urlMap[id]
	if !found {
		return "", false, nil
	}

	if s.deletedMap[id] {
		return "", false, storage.ErrURLDeleted
	}

	return originalURL, true, nil
}

func (s *Storage) SaveBatch(items []model.BatchRequestItem) (map[string]string, error) {
	result := make(map[string]string)

	for _, item := range items {
		s.mutex.RLock()
		existingID, exists := s.reverseURLMap[item.OriginalURL]
		s.mutex.RUnlock()

		if exists {
			result[item.CorrelationID] = existingID
			continue
		}

		id, err := generator.GenerateID(8)
		if err != nil {
			return nil, fmt.Errorf("failed to generate ID: %w", err)
		}

		s.mutex.Lock()
		if existingID, exists := s.reverseURLMap[item.OriginalURL]; exists {
			s.mutex.Unlock()
			result[item.CorrelationID] = existingID
			continue
		}

		s.idCounter++
		uuid := strconv.Itoa(s.idCounter)
		s.urlMap[id] = item.OriginalURL
		s.reverseURLMap[item.OriginalURL] = id
		s.mutex.Unlock()

		record := model.URLRecord{
			UUID:        uuid,
			ShortURL:    id,
			OriginalURL: item.OriginalURL,
			UserID:      "",
			IsDeleted:   false,
		}

		if err := s.saveRecordToFile(record); err != nil {
			return nil, fmt.Errorf("failed to save record to file: %w", err)
		}

		result[item.CorrelationID] = id
	}

	return result, nil
}

func (s *Storage) loadFromFile() error {
	file, err := os.OpenFile(s.filePath, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	maxID := 0

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var record model.URLRecord
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			return fmt.Errorf("failed to unmarshal record: %w", err)
		}

		s.urlMap[record.ShortURL] = record.OriginalURL
		s.reverseURLMap[record.OriginalURL] = record.ShortURL
		s.deletedMap[record.ShortURL] = record.IsDeleted

		if record.UserID != "" {
			url := model.URL{
				ID:          record.ShortURL,
				OriginalURL: record.OriginalURL,
				UserID:      record.UserID,
			}
			s.userURLs[record.UserID] = append(s.userURLs[record.UserID], url)
		}

		if id, err := strconv.Atoi(record.UUID); err == nil && id > maxID {
			maxID = id
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	s.idCounter = maxID
	return nil
}

func (s *Storage) saveRecordToFile(record model.URLRecord) error {
	s.fileWriteMu.Lock()
	defer s.fileWriteMu.Unlock()

	file, err := os.OpenFile(s.filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file for writing: %w", err)
	}
	defer file.Close()

	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal record: %w", err)
	}

	if _, err := file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}

func (s *Storage) SaveWithUser(originalURL, userID string) (string, error) {
	s.mutex.RLock()
	existingID, exists := s.reverseURLMap[originalURL]
	s.mutex.RUnlock()

	if exists {
		return existingID, storage.ErrURLExists
	}

	id, err := generator.GenerateID(8)
	if err != nil {
		return "", err
	}

	s.mutex.Lock()
	if existingID, exists := s.reverseURLMap[originalURL]; exists {
		s.mutex.Unlock()
		return existingID, storage.ErrURLExists
	}

	s.idCounter++
	uuid := strconv.Itoa(s.idCounter)
	s.urlMap[id] = originalURL
	s.reverseURLMap[originalURL] = id

	url := model.URL{
		ID:          id,
		OriginalURL: originalURL,
		UserID:      userID,
	}
	s.userURLs[userID] = append(s.userURLs[userID], url)
	s.mutex.Unlock()

	record := model.URLRecord{
		UUID:        uuid,
		ShortURL:    id,
		OriginalURL: originalURL,
		UserID:      userID,
		IsDeleted:   false,
	}

	if err := s.saveRecordToFile(record); err != nil {
		return "", err
	}

	return id, nil
}

func (s *Storage) SaveBatchWithUser(items []model.BatchRequestItem, userID string) (map[string]string, error) {
	result := make(map[string]string)

	for _, item := range items {
		s.mutex.RLock()
		existingID, exists := s.reverseURLMap[item.OriginalURL]
		s.mutex.RUnlock()

		if exists {
			result[item.CorrelationID] = existingID
			continue
		}

		id, err := generator.GenerateID(8)
		if err != nil {
			return nil, fmt.Errorf("failed to generate ID: %w", err)
		}

		s.mutex.Lock()
		if existingID, exists := s.reverseURLMap[item.OriginalURL]; exists {
			s.mutex.Unlock()
			result[item.CorrelationID] = existingID
			continue
		}

		s.idCounter++
		uuid := strconv.Itoa(s.idCounter)
		s.urlMap[id] = item.OriginalURL
		s.reverseURLMap[item.OriginalURL] = id

		url := model.URL{
			ID:          id,
			OriginalURL: item.OriginalURL,
			UserID:      userID,
		}
		s.userURLs[userID] = append(s.userURLs[userID], url)
		s.mutex.Unlock()

		record := model.URLRecord{
			UUID:        uuid,
			ShortURL:    id,
			OriginalURL: item.OriginalURL,
			UserID:      userID,
			IsDeleted:   false,
		}

		if err := s.saveRecordToFile(record); err != nil {
			return nil, fmt.Errorf("failed to save record to file: %w", err)
		}

		result[item.CorrelationID] = id
	}

	return result, nil
}

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

func (s *Storage) DeleteUserURLs(userID string, urlIDs []string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	userURLs, exists := s.userURLs[userID]
	if !exists {
		return nil
	}

	userURLSet := make(map[string]bool)
	for _, url := range userURLs {
		userURLSet[url.ID] = true
	}

	for _, urlID := range urlIDs {
		if userURLSet[urlID] && !s.deletedMap[urlID] {
			s.deletedMap[urlID] = true

			s.idCounter++
			uuid := strconv.Itoa(s.idCounter)
			record := model.URLRecord{
				UUID:        uuid,
				ShortURL:    urlID,
				OriginalURL: s.urlMap[urlID],
				UserID:      userID,
				IsDeleted:   true,
			}

			if err := s.saveRecordToFile(record); err != nil {
				return fmt.Errorf("failed to save deletion record: %w", err)
			}
		}
	}

	return nil
}
