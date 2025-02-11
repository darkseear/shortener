package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/darkseear/shortener/internal/config"
	"github.com/darkseear/shortener/internal/logger"
	"github.com/darkseear/shortener/internal/models"
	"github.com/darkseear/shortener/internal/services"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Router struct {
	Handle *chi.Mux
	Store  *services.Store
	Cfg    *config.Config
}

func Routers(cfg *config.Config, store *services.Store) *Router {

	r := Router{
		Handle: chi.NewRouter(),
		Store:  store,
		Cfg:    cfg,
	}

	// logging := logger.WhithLogging

	r.Handle.Post("/", AddURL(r))
	r.Handle.Get("/{id}", GetURL(r))
	r.Handle.Post("/api/shorten", Shorten(r))
	r.Handle.Post("/api/shorten/batch", ShortenBatch(r))
	r.Handle.Get("/ping", PingDB(r))

	return &r
}

func GetURL(r Router) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		path := strings.TrimSuffix(strings.TrimPrefix(req.URL.Path, "/"), "/")
		parts := strings.Split(path, "/")
		paramURLID := parts[0]

		if paramURLID == "" {
			http.Error(res, "Not found", http.StatusBadRequest)
			return
		}

		count, err := r.Store.GetOriginalURL(paramURLID, r.Cfg)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		http.Redirect(res, req, count, http.StatusTemporaryRedirect)

	}
}

func AddURL(r Router) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
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
		short, status := r.Store.ShortenURL(strURL, r.Cfg)
		res.WriteHeader(status)
		res.Write([]byte(r.Cfg.URL + "/" + short))
	}
}

func Shorten(r Router) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
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

		shortenURL, status := r.Store.ShortenURL(longURL, r.Cfg)

		shortenJSON.Result = r.Cfg.URL + "/" + shortenURL
		resp, err := json.Marshal(shortenJSON)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		defer req.Body.Close()

		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(status)
		// json
		res.Write(resp)
	}
}

func ShortenBatch(r Router) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		// var buf bytes.Buffer
		var BatchShortenJSON []models.BatchShortenJSON
		var BatchLongJSON []models.BatchLongJSON

		dec := json.NewDecoder(req.Body)
		if err := dec.Decode(&BatchLongJSON); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		batchLongJSON := BatchLongJSON
		fmt.Println(batchLongJSON)

		for key := range batchLongJSON {
			cID := batchLongJSON[key].CorrelationID
			lJSON := batchLongJSON[key].LongJSON
			fmt.Println(cID)
			fmt.Println(lJSON)
			shortenURL, _ := r.Store.ShortenURL(lJSON, r.Cfg)
			BatchShortenJSON = append(BatchShortenJSON, models.BatchShortenJSON{CorrelationID: cID, ShortJSON: r.Cfg.URL + "/" + shortenURL})
		}

		fmt.Println(BatchShortenJSON)

		enc := json.NewEncoder(res)

		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusCreated)
		logger.Log.Info("Status 201")

		if err := enc.Encode(BatchShortenJSON); err != nil {
			logger.Log.Debug("error encoding batchShortenJson")
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		defer req.Body.Close()
	}
}

func PingDB(r Router) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		db, errSQL := sql.Open("pgx", r.Cfg.DatabaseDSN)
		if errSQL != nil {
			logger.Log.Error(errSQL.Error())
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
