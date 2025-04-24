package gzip

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// СompressWriter реализует интерфейс http.ResponseWriter и позволяет прозрачно для сервера.
// Cжимать передаваемые данные и выставлять правильные HTTP-заголовки.
type СompressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

// NewСompressWriter создает новый экземпляр СompressWriter.
// Он принимает http.ResponseWriter и создает новый gzip.Writer, связанный с ним.
func NewСompressWriter(w http.ResponseWriter) *СompressWriter {
	return &СompressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

// Header возвращает заголовки ответа.
func (c *СompressWriter) Header() http.Header {
	return c.w.Header()
}

// Write записывает данные в gzip.Writer и отправляет их в http.ResponseWriter.
// Возвращает количество записанных байт и ошибку, если она произошла.
func (c *СompressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

// / WriteHeader устанавливает код состояния ответа и добавляет заголовок Content-Encoding.
func (c *СompressWriter) WriteHeader(statusCode int) {
	if statusCode < 308 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

// Close закрывает gzip.Writer и досылает все данные из буфера.
func (c *СompressWriter) Close() error {
	return c.zw.Close()
}

// CompressReader реализует интерфейс io.ReadCloser и позволяет прозрачно для сервера
// декомпрессировать получаемые от клиента данные.
type CompressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

// NewCompressReader создает новый экземпляр CompressReader.
// Он принимает io.ReadCloser и создает новый gzip.Reader, связанный с ним.
// Возвращает указатель на CompressReader и ошибку, если она произошла.
func NewCompressReader(r io.ReadCloser) (*CompressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &CompressReader{
		r:  r,
		zr: zr,
	}, nil
}

// Read читает данные из gzip.Reader и записывает их в буфер p.
// Возвращает количество прочитанных байт и ошибку, если она произошла.
func (c CompressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

// / Close закрывает gzip.Reader и оригинальный io.ReadCloser.
// Возвращает ошибку, если она произошла.
func (c *CompressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

// GzipMiddleware - обертка для http.Handler, которая добавляет поддержку сжатия и декомпрессии данных в формате gzip.
// Она проверяет, поддерживает ли клиент сжатие данных и отправляет их в сжатом виде, если это возможно.
func GzipMiddleware(h http.Handler) http.HandlerFunc {
	gzipFn := func(w http.ResponseWriter, r *http.Request) {
		ow := w
		// проверяем, что клиент умеет получать от сервера сжатые данные в формате gzip
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		// поддержка content-type
		contentType := r.Header.Get("Content-Type")
		supportApplicationJSON := strings.Contains(contentType, "application/json")
		supportTextHTML := strings.Contains(contentType, "text/html")

		if supportsGzip && (supportApplicationJSON || supportTextHTML) {
			cw := NewСompressWriter(w)
			ow = cw
			defer cw.Close()
		}

		// проверяем, что клиент отправил серверу сжатые данные в формате gzip
		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")

		if sendsGzip {
			// оборачиваем тело запроса в io.Reader с поддержкой декомпрессии
			cr, err := NewCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			// меняем тело запроса на новое
			r.Body = cr
			defer cr.Close()
		}

		// передаём управление хендлеру
		h.ServeHTTP(ow, r)
	}
	return gzipFn
}
