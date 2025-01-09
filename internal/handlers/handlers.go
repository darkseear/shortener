package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/darkseear/shortener/internal/storage"
	"github.com/go-chi/chi/v5"
)

type Router struct {
	Handle *chi.Mux
	URL    string
}

func Routers(url string) *Router {

	r := Router{
		Handle: chi.NewRouter(),
		URL:    url,
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

		// paramURLID := req.PathValue("id")
		// fmt.Println(req.URL)

		path := strings.TrimSuffix(strings.TrimPrefix(req.URL.Path, "/"), "/")
		parts := strings.Split(path, "/")
		paramURLID := parts[0]
		if paramURLID == "" {
			fmt.Println("tut")
			http.Error(res, "Not found", http.StatusBadRequest)
			return
		}

		s := storage.MyMap

		for key, value := range s {
			if paramURLID == value {
				http.Redirect(res, req,
					key,
					http.StatusTemporaryRedirect)
				return
			}
		}
	}
}

func AddURL(r Router) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			//StatusBadRequest  400
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
		minURL := storage.RandStringBytes(8)
		if strURL == "" {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		s := storage.MyMap
		res.Header().Set("Content-Type", "text/plain")

		res.WriteHeader(http.StatusCreated)

		if s[strURL] == "" {
			s[strURL] = minURL
			res.Write([]byte(r.URL + "/" + minURL))
		} else {
			res.Write([]byte(r.URL + "/" + s[strURL]))
		}
	}
}
