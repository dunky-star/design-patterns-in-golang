package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

const port = ":4000"

type application struct {
}

func main() {

	app := application{}

	fmt.Println("Starting server on port", port)

	srv := &http.Server{
		Addr:              port,
		Handler:           app.routes(),
		IdleTimeout:       30 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      20 * time.Second,
		ReadHeaderTimeout: 30 * time.Second,
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}
