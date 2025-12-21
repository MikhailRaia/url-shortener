package model

// URL represents a stored shortened URL with owner information.
type URL struct {
	ID          string
	OriginalURL string
	UserID      string
}

// UserURL is the external representation returned in API responses.
type UserURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
