package main

import (
	"fmt"
	"go-breeders/pets"
	"net/http"
	"net/url"

	jsonxmltool "github.com/dunky-star/go-json-xml-tool"
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
	dogBreeds, err := app.App.DB.AllDogBreeds()
	if err != nil {
		_ = k.ErrorJSON(w, err, http.StatusBadRequest)
		return
	}

	_ = k.WriteJSON(w, http.StatusOK, dogBreeds)
}

func (app *application) CreateDogWithBuilder(w http.ResponseWriter, r *http.Request) {
	k := jsonxmltool.NewKit()
	p, err := pets.NewPetBuilder().
		SetSpecies("dog").
		SetBreed("Labrador").
		SetMinWeight(30).
		SetMaxWeight(40).
		SetWeight(35).
		SetDescription("A friendly and outgoing dog").
		SetLifeSpan(10).
		SetGeographicOrigin("United States").
		SetColor("Brown").
		SetAge(5).
		SetAgeEstimated(true).
		Build()

	if err != nil {
		_ = k.ErrorJSON(w, err, http.StatusBadRequest)
		return
	}
	_ = k.WriteJSON(w, http.StatusOK, p)
}

// CreateCatWithBuilder creates a cat using a (fluent) builder pattern
func (app *application) CreateCatWithBuilder(w http.ResponseWriter, r *http.Request) {
	k := jsonxmltool.NewKit()

	// create our dog using the builder pattern.
	p, err := pets.NewPetBuilder().
		SetSpecies("cat").
		SetBreed("felis silvestris catus").
		SetWeight(4).
		SetDescription("A beautiful house cat.").
		SetGeographicOrigin("Canada").
		SetColor("Calico").
		SetAge(1).
		SetAgeEstimated(true).
		Build()

	if err != nil {
		_ = k.ErrorJSON(w, err, http.StatusBadRequest)
		return
	}

	_ = k.WriteJSON(w, http.StatusOK, p)
}

func (app *application) AnimalFromAbstractFactory(w http.ResponseWriter, r *http.Request) {
	// Setup toolbox

	// Get species from URL itself.
	species := r.PathValue("species")

	// Get breed from the URL.
	b := r.PathValue("breed")
	breed, _ := url.QueryUnescape(b)

	fmt.Println("Species:", species, "Breed:", breed)

	// Create a pet from abstract factory

	// Write the result as JSON.
}
