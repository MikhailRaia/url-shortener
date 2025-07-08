package model

type URL struct {
	ID          string
	OriginalURL string
	UserID      string
}

type UserURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
