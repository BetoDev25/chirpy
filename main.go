package main

import (
	"net/http"
	"log"
)

func main() {
	serve := http.NewServeMux()

	server := &http.Server{
		Addr:    ":8080",
		Handler: serve,
	}

	serve.Handle("/", http.FileServer(http.Dir(".")))

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
