package main

import (
	"net/http"
	"time"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	mux.HandleFunc("GET /test-patterns", app.TestPatterns)
	mux.HandleFunc("GET /api/dog-from-factory", app.CreateDogFromFactory)
	mux.HandleFunc("GET /api/cat-from-factory", app.CreateCatFromFactory)
	mux.HandleFunc("GET /api/dog-from-abstract-factory", app.CreateDogFromAbstractFactory)
	mux.HandleFunc("GET /api/cat-from-abstract-factory", app.CreateCatFromAbstractFactory)

	mux.HandleFunc("GET /", app.ShowHome)
	mux.HandleFunc("GET /{page}", app.ShowPage)

	return recoverMiddleware(timeoutMiddleware(mux, 60*time.Second))
}
