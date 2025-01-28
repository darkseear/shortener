package main

import (
	"net/http"

	"github.com/darkseear/shortener/internal/config"
	"github.com/darkseear/shortener/internal/handlers"
	"github.com/darkseear/shortener/internal/services"
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

	m := services.NewMemory()
	//router chi
	r := handlers.Routers(config.URL, m).Handle

	return http.ListenAndServe(address, r)
}
