package model

// BatchRequestItem представляет элемент запроса на пакетное сокращение URL
type BatchRequestItem struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// BatchResponseItem представляет элемент ответа на пакетное сокращение URL
type BatchResponseItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
