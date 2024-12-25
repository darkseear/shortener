package main

import (
	"io"
	"log"
	"math/rand"
	"net/http"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

// запуск сервера
func run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", okorok)
	mux.HandleFunc("/{id}", okorok)
	return http.ListenAndServe(`:8080`, mux)
}

//

// map
var myMap = make(map[string]string)

//

// короткая строка
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

//

// обработчик get post
func okorok(res http.ResponseWriter, req *http.Request) {

	if req.Method == http.MethodGet {
		//get
		paramURL := req.PathValue("id")
		metka := false
		origURL := ""
		for key, value := range myMap {
			if paramURL == value {
				metka = true
				origURL = key
			}
		}
		if metka {
			http.Redirect(res, req,
				origURL,
				http.StatusTemporaryRedirect)
		} else {
			res.WriteHeader(http.StatusBadRequest)
		}
	} else if req.Method == http.MethodPost {
		//post
		text, err := io.ReadAll(req.Body)
		if err != nil {
			log.Fatal(err)
		}
		textString := string(text)
		str := RandStringBytes(8)
		metka := true
		if textString != "" {
			for key := range myMap {
				if textString == key {
					metka = false
				}
			}
			res.WriteHeader(http.StatusCreated)
			if metka {
				myMap[textString] = str
				res.Write([]byte("http://localhost:8080/" + str))
			} else {
				res.Write([]byte("http://localhost:8080/" + myMap[textString]))
			}
		} else {
			res.WriteHeader(http.StatusBadRequest)
		}

	} else {
		//StatusBadRequest  400
		res.WriteHeader(http.StatusBadRequest)
	}
}
