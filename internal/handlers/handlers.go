package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand/v2"
	"net/http"
	"strings"
	"sync"

	"github.com/darkseear/shortener/internal/config"
	"github.com/darkseear/shortener/internal/logger"
	"github.com/darkseear/shortener/internal/models"
	"github.com/darkseear/shortener/internal/services"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

type Router struct {
	Handle *chi.Mux
	Store  *services.Store
	Cfg    *config.Config
	Auth   *services.AuthService
}

func Routers(cfg *config.Config, store *services.Store) *Router {

	r := Router{
		Handle: chi.NewRouter(),
		Store:  store,
		Cfg:    cfg,
		Auth:   services.NewAuthService(cfg.SecretKey),
	}

	r.Handle.Post("/", r.AddURL())
	r.Handle.Get("/{id}", r.GetURL())
	r.Handle.Post("/api/shorten", r.Shorten())
	r.Handle.Post("/api/shorten/batch", r.ShortenBatch())
	r.Handle.Get("/ping", r.PingDB())
	r.Handle.Get("/api/user/urls", r.ListURL())
	r.Handle.Delete("/api/user/urls", r.DeleteURL())

	return &r
}

type Handlers interface {
	GetUrl() http.HandlerFunc
	AddURL() http.HandlerFunc
	Shorten() http.HandlerFunc
	ShortenBatch() http.HandlerFunc
	PingDB() http.HandlerFunc
	ListURL() http.HandlerFunc
}

func readJSON(req *http.Request, v interface{}) error {
	dec := json.NewDecoder(req.Body)
	defer req.Body.Close()
	return dec.Decode(v)
}

func writeJSON(res http.ResponseWriter, status int, v interface{}) error {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(status)
	enc := json.NewEncoder(res)
	return enc.Encode(v)
}

func (r *Router) GetURL() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		userRand := fmt.Sprintf("%d", int(math.Floor(1000+math.Floor(9000*rand.Float64()))))
		userID := r.Auth.IssueCookie(res, req, userRand)
		path := strings.TrimSuffix(strings.TrimPrefix(req.URL.Path, "/"), "/")
		parts := strings.Split(path, "/")
		paramURLID := parts[0]

		if paramURLID == "" {
			http.Error(res, "Not found", http.StatusBadRequest)
			return
		}

		count, err := r.Store.GetOriginalURL(paramURLID, r.Cfg, userID)
		if err != nil && count == "GoneStatus" {
			res.WriteHeader(http.StatusGone)
			return
		} else if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		http.Redirect(res, req, count, http.StatusTemporaryRedirect)

	}
}

func (r *Router) AddURL() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		userRand := fmt.Sprintf("%d", int(math.Floor(1000+math.Floor(9000*rand.Float64()))))
		userID := r.Auth.IssueCookie(res, req, userRand)
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
		short, status := r.Store.ShortenURL(strURL, r.Cfg, userID)
		res.WriteHeader(status)
		res.Write([]byte(r.Cfg.URL + "/" + short))
	}
}

func (r *Router) Shorten() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var longJSON models.LongJSON
		userRand := fmt.Sprintf("%d", int(math.Floor(1000+math.Floor(9000*rand.Float64()))))
		userID := r.Auth.IssueCookie(res, req, userRand)

		if err := readJSON(req, &longJSON); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		longURL := longJSON.URL
		if longURL == "" {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		shortenURL, status := r.Store.ShortenURL(longURL, r.Cfg, userID)
		shortenJSON := models.ShortenJSON{Result: r.Cfg.URL + "/" + shortenURL}

		if err := writeJSON(res, status, shortenJSON); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
		}
	}
}

func (r *Router) ShortenBatch() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		userRand := fmt.Sprintf("%d", int(math.Floor(1000+math.Floor(9000*rand.Float64()))))
		userID := r.Auth.IssueCookie(res, req, userRand)
		var batchLongJSON []models.BatchLongJSON
		if err := readJSON(req, &batchLongJSON); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		var batchShortenJSON []models.BatchShortenJSON
		for _, item := range batchLongJSON {
			shortenURL, _ := r.Store.ShortenURL(item.LongJSON, r.Cfg, userID)
			batchShortenJSON = append(batchShortenJSON, models.BatchShortenJSON{
				CorrelationID: item.CorrelationID,
				ShortJSON:     r.Cfg.URL + "/" + shortenURL,
			})
		}

		if err := writeJSON(res, http.StatusCreated, batchShortenJSON); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
		}
	}
}

func (r *Router) PingDB() http.HandlerFunc {
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

func (r *Router) ListURL() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		userRand := fmt.Sprintf("%d", int(math.Floor(1000+math.Floor(9000*rand.Float64()))))
		userID := r.Auth.IssueCookie(res, req, userRand)

		if userID == "" {
			res.WriteHeader(http.StatusUnauthorized)
		}

		urls, err := r.Store.GetOriginalURLByUserID(r.Cfg, userID)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
		}
		if len(urls) == 0 {
			res.WriteHeader(http.StatusNoContent)
		}
		if err := writeJSON(res, http.StatusOK, urls); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
		}
		logger.Log.Info("User", zap.String("userID", userID))

		// // res.WriteHeader(http.StatusOK)
		// res.Write([]byte(userID))
	}
}

func (r *Router) DeleteURL() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var mut sync.Mutex
		userRand := fmt.Sprintf("%d", int(math.Floor(1000+math.Floor(9000*rand.Float64()))))
		userID := r.Auth.IssueCookie(res, req, userRand)

		if userID == "" {
			res.WriteHeader(http.StatusUnauthorized)
		}
		fmt.Println(userID)

		var urlsToDelete []string
		if err := readJSON(req, &urlsToDelete); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		fmt.Println(urlsToDelete)

		for _, urlsToDelete := range urlsToDelete {
			fmt.Println(urlsToDelete)
			mut.Lock()
			err := r.Store.DeleteURLByUserID(urlsToDelete, r.Cfg, userID)
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				logger.Log.Error("Delete error", zap.Error(err))
			}
			mut.Unlock()
		}

		logger.Log.Info("User", zap.String("Delete url is userID:", userID))
		res.WriteHeader(http.StatusAccepted)
	}
}
