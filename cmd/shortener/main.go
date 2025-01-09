package main

import (
	"net/http"

	"github.com/darkseear/shortener/internal/handlers"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

// запуск сервера
func run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.AddURL)
	mux.HandleFunc("/{id}", handlers.GetURL)
	return http.ListenAndServe(`:8080`, mux)
}
