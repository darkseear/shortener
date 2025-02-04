package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/darkseear/shortener/internal/logger"
	"github.com/darkseear/shortener/internal/models"
	"github.com/darkseear/shortener/internal/storage"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Router struct {
	Handle   *chi.Mux
	URL      string
	Memory   storage.URLService
	FileName string
	DDB      string
}

func Routers(url string, m storage.URLService, fileName string, DDB string) *Router {

	r := Router{
		Handle:   chi.NewRouter(),
		URL:      url,
		Memory:   m,
		FileName: fileName,
		DDB:      DDB,
	}

	// logging := logger.WhithLogging

	r.Handle.Post("/", AddURL(r))
	r.Handle.Get("/{id}", GetURL(r))
	r.Handle.Post("/api/shorten", Shorten(r))
	r.Handle.Get("/ping", PingDB(r))

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
		res.Write([]byte(r.URL + "/" + r.Memory.ShortenURL(strURL, r.FileName)))
	}
}

func Shorten(r Router) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		var buf bytes.Buffer
		var shortenJSON models.ShortenJSON
		var longJSON models.LongJSON
		//read body request
		_, err := buf.ReadFrom(req.Body)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		//deserial JSON
		if err = json.Unmarshal(buf.Bytes(), &longJSON); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		longURL := longJSON.URL

		if longURL == "" {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		shortenURL := r.Memory.ShortenURL(longURL, r.FileName)

		shortenJSON.Result = r.URL + "/" + shortenURL
		resp, err := json.Marshal(shortenJSON)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		defer req.Body.Close()

		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusCreated)
		// json
		res.Write(resp)
	}
}

func PingDB(r Router) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		fmt.Println(r.DDB)

		db, errSql := sql.Open("pgx", r.DDB)
		if errSql != nil {
			logger.Log.Error(errSql.Error())
			res.WriteHeader(http.StatusInternalServerError)
		}

		if errPing := db.Ping(); errPing != nil {
			logger.Log.Error(errPing.Error())
			res.WriteHeader(http.StatusInternalServerError)
		}

		defer db.Close()
		res.WriteHeader(http.StatusOK)
	}
}
