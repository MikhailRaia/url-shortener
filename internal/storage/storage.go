package storage

import (
	"errors"

	"github.com/MikhailRaia/url-shortener/internal/model"
)

var (
	ErrURLExists  = errors.New("url already exists")
	ErrURLDeleted = errors.New("url has been deleted")
)

type URLStorage interface {
	Save(originalURL string) (string, error)

	SaveWithUser(originalURL, userID string) (string, error)

	Get(id string) (string, bool)

	GetWithDeletedStatus(id string) (string, error)

	SaveBatch(items []model.BatchRequestItem) (map[string]string, error)

	SaveBatchWithUser(items []model.BatchRequestItem, userID string) (map[string]string, error)

	GetUserURLs(userID string) ([]model.UserURL, error)

	DeleteUserURLs(userID string, urlIDs []string) error
}
