package models

// ShortenJSON - структура для хранения короткой ссылки.
type ShortenJSON struct {
	Result string `json:"result"`
}

// LongJSON - структура для хранения длинной ссылки.
type LongJSON struct {
	URL string `json:"url"`
}

// MemoryFile - структура для хранения короткой и длинной ссылки в памяти.
type MemoryFile struct {
	ShortURL string `json:"shortURL"`
	LongURL  string `json:"longURL"`
}

// BatchLongJSON - структура для хранения длинной ссылки в батче.
type BatchLongJSON struct {
	CorrelationID string `json:"correlation_id"`
	LongJSON      string `json:"original_url"`
}

// BatchShortenJSON - структура для хранения короткой ссылки в батче.
type BatchShortenJSON struct {
	CorrelationID string `json:"correlation_id"`
	ShortJSON     string `json:"short_url"`
}

// URLPair - структура для хранения короткой и длинной ссылки.
// Используется для передачи данных между клиентом и сервером.
type URLPair struct {
	ShortURL string `json:"short_url"`
	LongURL  string `json:"original_url"`
}

// URLPairBatch - структура для хранения флага удвления, номера пользователя, короткой и длинной ссылки в батче для бд.
type DBUrlShorten struct {
	ShortURL    string `json:"short_url"`
	LongURL     string `json:"original_url"`
	UserID      string `json:"user_id"`
	DeletedFlag bool   `json:"is_deleted"`
}

// Stats - структура для хранения статистики по сокращенным ссылкам.
type Stats struct {
	URLs  int `json:"urls"`
	Users int `json:"users"`
}
