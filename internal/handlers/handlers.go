package handlers

import (
	"io"
	"net/http"
	"strings"

	"github.com/darkseear/shortener/internal/storage"
	"github.com/go-chi/chi/v5"
)

type Router struct {
	Handle *chi.Mux
	URL    string
	Memory storage.URLService
}

func Routers(url string, m storage.URLService) *Router {

	r := Router{
		Handle: chi.NewRouter(),
		URL:    url,
		Memory: m,
	}

	r.Handle.Post("/", AddURL(r))
	r.Handle.Get("/{id}", GetURL(r))

	return &r
}

func GetURL(r Router) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			//StatusBadRequest  400
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		path := strings.TrimSuffix(strings.TrimPrefix(req.URL.Path, "/"), "/")
		parts := strings.Split(path, "/")
		paramURLID := parts[0]

		if paramURLID == "" {
			http.Error(res, "Not found", http.StatusBadRequest)
			return
		}

		count, err := r.Memory.GetOriginalURL(paramURLID)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		http.Redirect(res, req, count, http.StatusTemporaryRedirect)

	}
}

func AddURL(r Router) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		defer req.Body.Close()

		strURL := string(body)
		if strURL == "" {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusCreated)
		res.Write([]byte(r.URL + "/" + r.Memory.ShortenURL(strURL)))
	}
}
