package gzip

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/darkseear/shortener/internal/handlers"
	"github.com/darkseear/shortener/internal/services"
	"github.com/stretchr/testify/require"
)

func TestGzipCompression(t *testing.T) {

	m := services.NewMemory()
	rw := handlers.Routers("http://localhost:8080", m)
	handler := http.HandlerFunc(GzipMiddleware(handlers.Shorten(*rw)))
	srv := httptest.NewServer(handler)
	defer srv.Close()

	tests := []struct {
		name        string
		defURL      string
		statusWant  int
		requestBody string
	}{
		{
			name:        "sends_gzip",
			defURL:      "http://localhost:8080",
			statusWant:  201,
			requestBody: `{"url":"https://yandex.ru/"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			buf := bytes.NewBuffer(nil)
			zb := gzip.NewWriter(buf)
			_, err := zb.Write([]byte(tt.requestBody))
			require.NoError(t, err)
			err = zb.Close()
			require.NoError(t, err)

			r := httptest.NewRequest("POST", srv.URL, buf)
			r.RequestURI = ""
			r.Header.Set("Content-Encoding", "gzip")
			r.Header.Set("Content-Type", "application/json")
			r.Header.Set("Accept-Encoding", "")

			resp, err := http.DefaultClient.Do(r)
			require.NoError(t, err)
			require.Equal(t, tt.statusWant, resp.StatusCode)

			defer resp.Body.Close()

			_, err = io.ReadAll(resp.Body)
			require.NoError(t, err)
		})
	}

	tests_two := []struct {
		name        string
		statusWant  int
		requestBody string
	}{
		{
			name:        "accepts_gzip",
			statusWant:  201,
			requestBody: `{"url":"https://yandex.ru/"}`,
		},
	}

	for _, tt := range tests_two {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBufferString(tt.requestBody)
			r := httptest.NewRequest("POST", srv.URL, buf)
			r.RequestURI = ""
			r.Header.Set("Content-Type", "application/json")
			r.Header.Set("Accept-Encoding", "gzip")

			resp, err := http.DefaultClient.Do(r)
			require.NoError(t, err)
			require.Equal(t, tt.statusWant, resp.StatusCode)

			defer resp.Body.Close()

			zr, err := gzip.NewReader(resp.Body)
			require.NoError(t, err)

			_, err = io.ReadAll(zr)
			require.NoError(t, err)
		})
	}
}
