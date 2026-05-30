package main

import (
	"fmt"
	"net/http"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Handle the request
		fmt.Fprint(w, "Hello, this is the beginning!")
	})

	return mux
}
