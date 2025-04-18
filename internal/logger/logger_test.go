package logger

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func BenchWhithLogging(b *testing.B) {

	var h http.Handler
	var want http.HandlerFunc

	h = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	want = http.HandlerFunc(func(h http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Log the request
			// Call the original handler
			h.ServeHTTP(w, r)
			// Log the response
		}
	}(h))

	b.Run("bench", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if got := WhithLogging(h); !reflect.DeepEqual(got, want) {
				b.Errorf("WhithLogging() = %v, want %v", got, want)
			}
			fmt.Println("test")
		}
	})

}
