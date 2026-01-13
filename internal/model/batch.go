package model

// BatchRequestItem describes a single URL to shorten in batch.
type BatchRequestItem struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// BatchResponseItem contains the correlation ID and resulting short URL.
type BatchResponseItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// generate:reset
type ResetTestStruct struct {
	IntField    int
	StringField string
	BoolField   bool
	SliceField  []string
	MapField    map[string]string
	PointerStr  *string
	PointerInt  *int
}
