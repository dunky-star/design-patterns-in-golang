package main

import (
	"net/http"
	"time"
)

const handlerTimeout = 60 * time.Second

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", app.ShowHome)

	return recoverMiddleware(timeoutMiddleware(mux, handlerTimeout))
}
