package main

import (
	"net/http"
	"time"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	// display test page
	mux.HandleFunc("GET /test-patterns", app.TestPatterns)

	// factory routes
	mux.HandleFunc("GET /api/dog-from-factory", app.CreateDogFromFactory)
	mux.HandleFunc("GET /api/cat-from-factory", app.CreateCatFromFactory)
	mux.HandleFunc("GET /api/dog-from-abstract-factory", app.CreateDogFromAbstractFactory)
	mux.HandleFunc("GET /api/cat-from-abstract-factory", app.CreateCatFromAbstractFactory)

	// builder routes
	mux.HandleFunc("GET /api/dog-from-builder", app.CreateDogWithBuilder)
	mux.HandleFunc("GET /api/cat-from-builder", app.CreateCatWithBuilder)

	mux.HandleFunc("GET /api/dog-breeds", app.GetAllDogBreedsJSON)

	mux.HandleFunc("GET /api/animal-from-abstract-factory/{species}/{breed}", app.AnimalFromAbstractFactory)

	mux.HandleFunc("GET /", app.ShowHome)
	mux.HandleFunc("GET /{page}", app.ShowPage)

	return recoverMiddleware(timeoutMiddleware(mux, 60*time.Second))
}
