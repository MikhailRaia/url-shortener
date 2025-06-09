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
)

type Storage struct {
	filePath    string
	urlMap      map[string]string
	idCounter   int
	mutex       sync.RWMutex
	fileWriteMu sync.Mutex
}

func NewStorage(filePath string) (*Storage, error) {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	storage := &Storage{
		filePath:  filePath,
		urlMap:    make(map[string]string),
		idCounter: 0,
	}

	if err := storage.loadFromFile(); err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *Storage) Save(originalURL string) (string, error) {
	id, err := generator.GenerateID(8)
	if err != nil {
		return "", err
	}

	s.mutex.Lock()
	s.idCounter++
	uuid := strconv.Itoa(s.idCounter)
	s.urlMap[id] = originalURL
	s.mutex.Unlock()

	record := model.URLRecord{
		UUID:        uuid,
		ShortURL:    id,
		OriginalURL: originalURL,
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
	return originalURL, found
}

func (s *Storage) SaveBatch(items []model.BatchRequestItem) (map[string]string, error) {
	result := make(map[string]string)

	for _, item := range items {
		id, err := generator.GenerateID(8)
		if err != nil {
			return nil, fmt.Errorf("failed to generate ID: %w", err)
		}

		s.mutex.Lock()
		s.idCounter++
		uuid := strconv.Itoa(s.idCounter)
		s.urlMap[id] = item.OriginalURL
		s.mutex.Unlock()

		record := model.URLRecord{
			UUID:        uuid,
			ShortURL:    id,
			OriginalURL: item.OriginalURL,
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
