package main

import (
	"net/http"

	"github.com/darkseear/shortener/internal/config"
	"github.com/darkseear/shortener/internal/handlers"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

// запуск сервера
func run() error {

	//config
	config := config.New()
	address := config.Address

	//router chi
	r := handlers.Routers(config.URL).Handle

	return http.ListenAndServe(address, r)
}
