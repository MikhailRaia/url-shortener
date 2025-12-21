package storage

import (
	"errors"

	"github.com/MikhailRaia/url-shortener/internal/model"
)

var (
	// ErrURLExists indicates the submitted URL has already been shortened.
	ErrURLExists = errors.New("url already exists")
	// ErrURLDeleted indicates the short URL was deleted by the user.
	ErrURLDeleted = errors.New("url has been deleted")
)

// URLStorage defines persistence operations for shortened URLs.
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
