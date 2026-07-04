package main

import (
	"fmt"
	"go-breeders/pets"
	"net/http"

	"github.com/dunky-star/go-json-xml-tool"
)

func (app *application) ShowHome(w http.ResponseWriter, r *http.Request) {
	app.render(w, "home.page.gohtml", nil)
}

func (app *application) ShowPage(w http.ResponseWriter, r *http.Request) {
	page := r.PathValue("page")
	app.render(w, fmt.Sprintf("%s.page.gohtml", page), nil)
}

func (app *application) CreateDogFromFactory(w http.ResponseWriter, r *http.Request) {
	k := jsonxmltool.NewKit()
	_ = k.WriteJSON(w, http.StatusOK, pets.NewPet("dog"))
}

func (app *application) CreateCatFromFactory(w http.ResponseWriter, r *http.Request) {
	k := jsonxmltool.NewKit()
	_ = k.WriteJSON(w, http.StatusOK, pets.NewPet("cat"))
}

func (app *application) TestPatterns(w http.ResponseWriter, r *http.Request) {
	app.render(w, "test.page.gohtml", nil)
}

func (app *application) CreateDogFromAbstractFactory(w http.ResponseWriter, r *http.Request) {
	k := jsonxmltool.NewKit()
	dog, err := pets.NewPetFromAbstractFactory("dog")
	if err != nil {
		_ = k.ErrorJSON(w, err, http.StatusBadRequest)
		return
	}
	_ = k.WriteJSON(w, http.StatusOK, dog)
}

func (app *application) CreateCatFromAbstractFactory(w http.ResponseWriter, r *http.Request) {
	k := jsonxmltool.NewKit()
	cat, err := pets.NewPetFromAbstractFactory("cat")
	if err != nil {
		_ = k.ErrorJSON(w, err, http.StatusBadRequest)
		return
	}
	_ = k.WriteJSON(w, http.StatusOK, cat)
}

func (app *application) GetAllDogBreedsJSON(w http.ResponseWriter, r *http.Request) {
	k := jsonxmltool.NewKit()
	dogBreeds, err := app.DB.AllDogBreeds()
	if err != nil {
		_ = k.ErrorJSON(w, err, http.StatusBadRequest)
		return
	}

	_ = k.WriteJSON(w, http.StatusOK, dogBreeds)
}
