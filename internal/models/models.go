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
