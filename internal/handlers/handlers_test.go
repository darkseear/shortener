package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/darkseear/shortener/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetURL(t *testing.T) {

	tests := []struct {
		name    string
		url     string
		want    int
		request string
	}{
		{
			name:    "test#1",
			url:     "https://www.yandex.ru",
			want:    307,
			request: "/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			minURL := storage.RandStringBytes(8)
			storage.MyMap[tt.url] = minURL
			fmt.Println(tt.request + minURL)
			request := httptest.NewRequest(http.MethodGet, tt.request+minURL, nil)

			w := httptest.NewRecorder()
			h := http.HandlerFunc(GetURL)

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

			w := httptest.NewRecorder()
			h := http.HandlerFunc(AddURL)

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
