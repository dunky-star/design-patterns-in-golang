package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go-breeders/internal/video"
	"go-breeders/models"
	"go-breeders/pets"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	jsonxmltool "github.com/dunky-star/go-json-xml-tool"
)

type videoProcessingService interface {
	Jobs() ([]*models.VideoJob, error)
	OutputFile(int, string) (string, error)
	Process(context.Context, int, video.ProcessOptions) (*models.VideoJob, error)
}

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

func (app *application) VideoProcessing(w http.ResponseWriter, r *http.Request) {
	app.render(w, "video-processing.page.gohtml", nil)
}

func (app *application) GetVideoJobs(w http.ResponseWriter, r *http.Request) {
	k := jsonxmltool.NewKit()

	if app.videoService == nil {
		_ = k.ErrorJSON(w, errors.New("video processing service is not available"), http.StatusServiceUnavailable)
		return
	}

	jobs, err := app.videoService.Jobs()
	if err != nil {
		log.Printf("Error getting video jobs: %v", err)
		_ = k.ErrorJSON(w, errors.New("unable to load video jobs"), http.StatusInternalServerError)
		return
	}

	_ = k.WriteJSON(w, http.StatusOK, jobs)
}

func (app *application) GetProcessedVideo(w http.ResponseWriter, r *http.Request) {
	if app.videoService == nil {
		http.Error(w, "Video processing service is not available", http.StatusServiceUnavailable)
		return
	}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.Error(w, "Invalid video job ID", http.StatusBadRequest)
		return
	}

	outputPath, err := app.videoService.OutputFile(id, r.PathValue("file"))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, video.ErrOutputUnavailable) {
			http.NotFound(w, r)
			return
		}

		log.Printf("Error loading output for video job %d: %v", id, err)
		http.Error(w, "Unable to load processed video", http.StatusInternalServerError)
		return
	}

	switch strings.ToLower(filepath.Ext(outputPath)) {
	case ".m3u8":
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	case ".ts":
		w.Header().Set("Content-Type", "video/mp2t")
	}
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeFile(w, r, outputPath)
}

func (app *application) ProcessVideoJob(w http.ResponseWriter, r *http.Request) {
	k := jsonxmltool.NewKit()

	if app.videoService == nil {
		_ = k.ErrorJSON(w, errors.New("video processing service is not available"), http.StatusServiceUnavailable)
		return
	}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		_ = k.ErrorJSON(w, errors.New("invalid video job id"), http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		_ = k.ErrorJSON(w, errors.New("invalid processing options"), http.StatusBadRequest)
		return
	}

	options, err := videoProcessOptionsFromRequest(r)
	if err != nil {
		_ = k.ErrorJSON(w, err, http.StatusBadRequest)
		return
	}

	job, err := app.videoService.Process(r.Context(), id, options)
	if err != nil {
		switch {
		case errors.Is(err, video.ErrNotStarted):
			_ = k.ErrorJSON(w, errors.New("video processing service is not available"), http.StatusServiceUnavailable)
		case errors.Is(err, video.ErrInvalidOptions):
			_ = k.ErrorJSON(w, err, http.StatusBadRequest)
		case errors.Is(err, sql.ErrNoRows):
			_ = k.ErrorJSON(w, errors.New("video job not found"), http.StatusNotFound)
		case errors.Is(err, video.ErrAlreadyProcessing):
			_ = k.ErrorJSON(w, err, http.StatusConflict)
		case errors.Is(err, video.ErrInputUnavailable):
			_ = k.ErrorJSON(w, err, http.StatusUnprocessableEntity)
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			_ = k.ErrorJSON(w, errors.New("video job was not queued"), http.StatusRequestTimeout)
		default:
			log.Printf("Error processing video job %d: %v", id, err)
			_ = k.ErrorJSON(w, errors.New("unable to process video job"), http.StatusInternalServerError)
		}
		return
	}

	_ = k.WriteJSON(w, http.StatusAccepted, job)
}

func videoProcessOptionsFromRequest(r *http.Request) (video.ProcessOptions, error) {
	options := video.ProcessOptions{
		EncodingType: r.FormValue("encoding"),
		RenameOutput: r.FormValue("rename_output") == "true",
		MaxRate1080p: r.FormValue("max_rate_1080p"),
		MaxRate720p:  r.FormValue("max_rate_720p"),
		MaxRate480p:  r.FormValue("max_rate_480p"),
	}

	if options.EncodingType == "hls" {
		segmentDuration, err := strconv.Atoi(r.FormValue("segment_duration"))
		if err != nil {
			return video.ProcessOptions{}, errors.New("segment duration must be an integer")
		}
		options.SegmentDuration = segmentDuration
	}

	return options, nil
}
