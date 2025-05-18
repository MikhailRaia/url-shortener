package storage

type URLStorage interface {
	Save(originalURL string) (string, error)

	Get(id string) (string, bool)
}
