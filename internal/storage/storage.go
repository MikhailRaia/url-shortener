package storage

import (
	"errors"

	"github.com/MikhailRaia/url-shortener/internal/model"
)

var ErrURLExists = errors.New("url already exists")

type URLStorage interface {
	// Save сохраняет один URL и возвращает его идентификатор
	Save(originalURL string) (string, error)

	// Get возвращает оригинальный URL по идентификатору
	Get(id string) (string, bool)

	// SaveBatch сохраняет множество URL-ов и возвращает их идентификаторы
	// map[correlation_id]shortURLID
	SaveBatch(items []model.BatchRequestItem) (map[string]string, error)
}
