package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", okorok)
	mux.HandleFunc("/{id}", okorok)
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}

func okorok(res http.ResponseWriter, req *http.Request) {

	if req.Method == http.MethodGet {
		//get
		if req.PathValue("id") == "EwHXdJfB" {
			fmt.Println("Parametr ID:", req.PathValue("id"))
			http.Redirect(res, req,
				"https://practicum.yandex.ru/",
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
		fmt.Println("Bodyparam:", textString)
		if textString == "https://practicum.yandex.ru/" {
			res.WriteHeader(http.StatusCreated)
			res.Write([]byte("http://localhost:8080/EwHXdJfB"))
		} else {
			res.WriteHeader(http.StatusBadRequest)
		}
	} else {
		//StatusBadRequest  400
		res.WriteHeader(http.StatusBadRequest)
	}
}
