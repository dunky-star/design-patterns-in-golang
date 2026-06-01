package main

import (
	"net/http"
	"time"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", app.ShowHome)
	mux.HandleFunc("GET /{page}", app.ShowPage)

	return recoverMiddleware(timeoutMiddleware(mux, 60*time.Second))
}
