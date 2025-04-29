package logger

import (
	"net/http"
	"testing"
)

func BenchmarkWhithLogging(b *testing.B) {
	b.Run("WithLogging", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			WhithLogging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		}
	})
	b.Run("InitLogger", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			Initialize("info")
		}
	})
}
