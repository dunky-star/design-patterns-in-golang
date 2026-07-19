package main

import (
	"errors"
	"fmt"
	"go-breeders/models"
	"go-breeders/pets"
	"go-breeders/streamer"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	jsonxmltool "github.com/dunky-star/go-json-xml-tool"
)

func (app *application) ShowHome(w http.ResponseWriter, r *http.Request) {
	app.render(w, "home.page.gohtml", nil)
}

func (app *application) ShowPage(w http.ResponseWriter, r *http.Request) {
	page := r.PathValue("page")
	app.render(w, fmt.Sprintf("%s.page.gohtml", page), nil)
}

func (app *application) DogOfMonth(w http.ResponseWriter, r *http.Request) {
	// Get the breed
	breed, err := app.App.DB.GetBreedByName("German Shepherd Dog")
	if err != nil || breed == nil {
		if err != nil {
			log.Printf("Error getting breed for dog of the month: %v", err)
		}
		app.render(w, "dog-of-month.page.gohtml", &templateData{Data: map[string]any{
			"error": "Dog of the Month is not available right now.",
		}})
		return
	}

	// Get the dog of the month from database
	dom, err := app.App.DB.GetDogOfMonthByID(1)
	if err != nil || dom == nil {
		if err != nil {
			log.Printf("Error getting dog of the month: %v", err)
		}
		app.render(w, "dog-of-month.page.gohtml", &templateData{Data: map[string]any{
			"error": "Dog of the Month is not available right now.",
		}})
		return
	}

	layout := "2006-01-02"
	dob, _ := time.Parse(layout, "2023-11-01")

	// Create dog and decorate it
	dog := models.DogOfMonth{
		Dog: &models.Dog{
			ID:               1,
			DogName:          "Sam",
			BreedID:          breed.ID,
			Color:            "Black & Tan",
			DateOfBirth:      dob,
			SpayedOrNeutered: false,
			Description:      "Sam is a very good boy.",
			Weight:           20,
			Breed:            *breed,
		},
		Video: dom.Video,
		Image: dom.Image,
	}

	// Serve the web page
	data := make(map[string]any)
	data["dog"] = dog

	app.render(w, "dog-of-month.page.gohtml", &templateData{Data: data})
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

func (app *application) WorkerPool(w http.ResponseWriter, r *http.Request) {
	app.render(w, "worker-pool.page.gohtml", nil)
}

func (app *application) ProcessVideos(w http.ResponseWriter, r *http.Request) {
	k := jsonxmltool.NewKit()

	if app.videoStream == nil {
		_ = k.ErrorJSON(w, errors.New("video worker pool is not available"), http.StatusServiceUnavailable)
		return
	}

	if err := r.ParseForm(); err != nil {
		_ = k.ErrorJSON(w, err, http.StatusBadRequest)
		return
	}

	encodingType := r.FormValue("encoding")
	options := &streamer.VideoOptions{
		RenameOutput: r.FormValue("rename_output") == "true",
	}

	switch encodingType {
	case "mp4":
	case "hls":
		segmentDuration, err := strconv.Atoi(r.FormValue("segment_duration"))
		if err != nil || segmentDuration < 1 {
			_ = k.ErrorJSON(w, errors.New("segment duration must be a positive integer"), http.StatusBadRequest)
			return
		}

		options.SegmentDuration = segmentDuration
		options.MaxRate1080p = r.FormValue("max_rate_1080p")
		options.MaxRate720p = r.FormValue("max_rate_720p")
		options.MaxRate480p = r.FormValue("max_rate_480p")

		if options.MaxRate1080p == "" || options.MaxRate720p == "" || options.MaxRate480p == "" {
			_ = k.ErrorJSON(w, errors.New("all HLS maximum bitrates are required"), http.StatusBadRequest)
			return
		}
	default:
		_ = k.ErrorJSON(w, errors.New("encoding must be mp4 or hls"), http.StatusBadRequest)
		return
	}

	outputDir := fmt.Sprintf("./output/%s", encodingType)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		_ = k.ErrorJSON(w, err, http.StatusInternalServerError)
		return
	}

	inputs := []string{
		"./input/puppy1.mp4",
		"./input/puppy2.mp4",
	}

	notifyChan := make(chan streamer.ProcessingMessage, len(inputs))

	for i, input := range inputs {
		video := app.videoStream.NewVideo(
			i+1,
			input,
			outputDir,
			encodingType,
			notifyChan,
			options,
		)
		app.videoQueue <- streamer.VideoProcessingJob{Video: video}
	}

	results := make([]streamer.ProcessingMessage, 0, len(inputs))
	for range inputs {
		select {
		case result := <-notifyChan:
			results = append(results, result)
		case <-r.Context().Done():
			_ = k.ErrorJSON(w, r.Context().Err(), http.StatusRequestTimeout)
			return
		}
	}

	_ = k.WriteJSON(w, http.StatusOK, results)
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
	k := jsonxmltool.NewKit()

	// Get species from URL itself.
	species := r.PathValue("species")

	// Get breed from the URL.
	b := r.PathValue("breed")
	breed, _ := url.QueryUnescape(b)

	// Create a pet from abstract factory
	pet, err := pets.NewPetWithBreedFromAbstractFactory(species, breed)
	if err != nil {
		_ = k.ErrorJSON(w, err, http.StatusBadRequest)
		return
	}

	// Write the result as JSON.
	_ = k.WriteJSON(w, http.StatusOK, pet)
}
