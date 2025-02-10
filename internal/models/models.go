package models

type ShortenJSON struct {
	Result string `json:"result"`
}

type LongJSON struct {
	URL string `json:"url"`
}

type MemoryFile struct {
	ShortURL string `json:"shortURL"`
	LongURL  string `json:"longURL"`
}

type BatchLongJSON struct {
	CorrelationID string `json:"correlation_id"`
	LongJSON      string `json:"original_url"`
}
type BatchShortenJSON struct {
	CorrelationID string `json:"correlation_id"`
	ShortJSON     string `json:"short_url"`
}
