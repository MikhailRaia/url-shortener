package storage

import (
	"errors"

	"github.com/MikhailRaia/url-shortener/internal/model"
)

var ErrURLExists = errors.New("url already exists")

type URLStorage interface {
	Save(originalURL string) (string, error)

	SaveWithUser(originalURL, userID string) (string, error)

	Get(id string) (string, bool)

	SaveBatch(items []model.BatchRequestItem) (map[string]string, error)

	SaveBatchWithUser(items []model.BatchRequestItem, userID string) (map[string]string, error)

	GetUserURLs(userID string) ([]model.UserURL, error)
}
