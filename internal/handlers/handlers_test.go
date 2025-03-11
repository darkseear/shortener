package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/darkseear/shortener/internal/config"
	"github.com/darkseear/shortener/internal/logger"
	"github.com/darkseear/shortener/internal/storage"
)

func TestGetURL(t *testing.T) {

	type testConfig struct {
		config *config.Config
	}

	lc := testConfig{
		config: &config.Config{
			Address:     "localhost:8080",
			URL:         "http://localhost:8080",
			LogLevel:    "info",
			MemoryFile:  "memory.log",
			DatabaseDSN: "",
		},
	}

	// config := config.New()
	store, err := storage.New(lc.config)
	if err != nil {
		logger.Log.Error("Error created store")
	}

	tests := []struct {
		name    string
		url     string
		want    int
		request string
		userID  string
	}{
		{
			name:    "test#1",
			url:     "https://www.yandex.ru",
			want:    307,
			request: "/",
			userID:  "122",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := *Routers(lc.config, store)
			minURL, status := store.ShortenURL(tt.url, tt.userID)
			logger.Log.Info("Status", zap.Int("status", status))
			request := httptest.NewRequest(http.MethodGet, tt.request+minURL, nil)
			w := httptest.NewRecorder()
			h := logger.WhithLogging(r.GetURL())

			h(w, request)
			result := w.Result()

			assert.Equal(t, tt.want, result.StatusCode)
			assert.Equal(t, tt.url, result.Header.Get("Location"))

			newURL, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)
			assert.NotNil(t, newURL)
		})
	}
}

func TestAddURL(t *testing.T) {
	type testConfig struct {
		config *config.Config
	}

	lc := testConfig{
		config: &config.Config{
			Address:     "localhost:8080",
			URL:         "http://localhost:8080",
			LogLevel:    "info",
			MemoryFile:  "memory.log",
			DatabaseDSN: "",
		},
	}
	// config := config.New()
	store, err := storage.New(lc.config)
	if err != nil {
		logger.Log.Error("Error created store")
	}
	type want struct {
		contentType string
		statusCode  int
	}
	tests := []struct {
		name     string
		urlPlain string
		request  string
		want     want
	}{
		{
			name:     "addurl_test#1",
			urlPlain: "https://www.yandex.ru",
			want: want{
				contentType: "text/plain",
				statusCode:  201,
			},
			request: "/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.request, strings.NewReader(tt.urlPlain))

			r := *Routers(lc.config, store)
			w := httptest.NewRecorder()
			h := logger.WhithLogging(r.AddURL())

			h(w, request)

			result := w.Result()
			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			newURL, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)
			assert.NotNil(t, newURL)
		})
	}
}
