package main

import (
	"net/http"

	"github.com/darkseear/shortener/internal/handlers"
	"github.com/go-chi/chi/v5"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

// запуск сервера
func run() error {

	//router chi
	r := chi.NewRouter()

	r.Post("/", handlers.AddURL)
	r.Get("/{id}", handlers.GetURL)

	// mux := http.NewServeMux()
	// mux.HandleFunc("/", handlers.AddURL)
	// mux.HandleFunc("/{id}", handlers.GetURL)
	return http.ListenAndServe(`:8080`, r)
}
