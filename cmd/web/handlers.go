package main

import (
	"database/sql"
	"errors"
	"fmt"
	"go-breeders/models"
	"go-breeders/pets"
	"go-breeders/streamer"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	jsonxmltool "github.com/dunky-star/go-json-xml-tool"
)

var bitratePattern = regexp.MustCompile(`^[1-9][0-9]*[kKmM]?$`)

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

	jobs, err := app.App.DB.AllVideoJobs()
	if err != nil {
		log.Printf("Error getting video jobs: %v", err)
		_ = k.ErrorJSON(w, errors.New("unable to load video jobs"), http.StatusInternalServerError)
		return
	}

	_ = k.WriteJSON(w, http.StatusOK, jobs)
}

func (app *application) ProcessVideoJob(w http.ResponseWriter, r *http.Request) {
	k := jsonxmltool.NewKit()

	if app.videoStream == nil || app.videoQueue == nil || app.videoResults == nil {
		_ = k.ErrorJSON(w, errors.New("video worker pool is not available"), http.StatusServiceUnavailable)
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

	encodingType, options, err := videoOptionsFromRequest(r)
	if err != nil {
		_ = k.ErrorJSON(w, err, http.StatusBadRequest)
		return
	}

	job, err := app.App.DB.GetVideoJobByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_ = k.ErrorJSON(w, errors.New("video job not found"), http.StatusNotFound)
			return
		}
		log.Printf("Error getting video job %d: %v", id, err)
		_ = k.ErrorJSON(w, errors.New("unable to load video job"), http.StatusInternalServerError)
		return
	}

	inputPath, err := resolveMediaPath(app.config.media.inputDir, job.InputMediaKey)
	if err != nil {
		log.Printf("Invalid input media for video job %d: %v", id, err)
		_ = k.ErrorJSON(w, errors.New("input media is not available"), http.StatusUnprocessableEntity)
		return
	}

	outputDir := filepath.Join(
		app.config.media.outputDir,
		encodingType,
		fmt.Sprintf("job-%d", job.ID),
	)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Printf("Error creating output directory for video job %d: %v", id, err)
		_ = k.ErrorJSON(w, errors.New("unable to prepare video output"), http.StatusInternalServerError)
		return
	}

	claimed, err := app.App.DB.ClaimVideoJob(id, encodingType)
	if err != nil {
		log.Printf("Error claiming video job %d: %v", id, err)
		_ = k.ErrorJSON(w, errors.New("unable to start video job"), http.StatusInternalServerError)
		return
	}
	if !claimed {
		_ = k.ErrorJSON(w, errors.New("video job is already processing"), http.StatusConflict)
		return
	}

	video := app.videoStream.NewVideo(
		job.ID,
		inputPath,
		outputDir,
		encodingType,
		app.videoResults,
		options,
	)

	select {
	case app.videoQueue <- streamer.VideoProcessingJob{Video: video}:
		job.EncodingType = encodingType
		job.Status = "processing"
		job.OutputReference = ""
		job.ErrorMessage = ""
		_ = k.WriteJSON(w, http.StatusAccepted, job)
	case <-r.Context().Done():
		if err := app.App.DB.FailVideoJob(id, "Request cancelled before job was queued"); err != nil {
			log.Printf("Error failing cancelled video job %d: %v", id, err)
		}
		_ = k.ErrorJSON(w, errors.New("video job was not queued"), http.StatusRequestTimeout)
	}
}

func (app *application) listenToVideoResults() {
	for result := range app.videoResults {
		job, err := app.App.DB.GetVideoJobByID(result.ID)
		if err != nil {
			log.Printf("Error loading completed video job %d: %v", result.ID, err)
			continue
		}

		if !result.Successful {
			if err := app.App.DB.FailVideoJob(result.ID, result.Message); err != nil {
				log.Printf("Error persisting failed video job %d: %v", result.ID, err)
			}
			continue
		}

		outputReference := filepath.ToSlash(filepath.Join(
			job.EncodingType,
			fmt.Sprintf("job-%d", job.ID),
			result.OutputFile,
		))
		if err := app.App.DB.CompleteVideoJob(result.ID, outputReference); err != nil {
			log.Printf("Error completing video job %d: %v", result.ID, err)
		}
	}
}

func videoOptionsFromRequest(r *http.Request) (string, *streamer.VideoOptions, error) {
	encodingType := r.FormValue("encoding")
	options := &streamer.VideoOptions{
		RenameOutput: r.FormValue("rename_output") == "true",
	}

	switch encodingType {
	case "mp4":
		return encodingType, options, nil
	case "hls":
		segmentDuration, err := strconv.Atoi(r.FormValue("segment_duration"))
		if err != nil || segmentDuration < 1 {
			return "", nil, errors.New("segment duration must be a positive integer")
		}

		options.SegmentDuration = segmentDuration
		options.MaxRate1080p = r.FormValue("max_rate_1080p")
		options.MaxRate720p = r.FormValue("max_rate_720p")
		options.MaxRate480p = r.FormValue("max_rate_480p")

		for _, bitrate := range []string{
			options.MaxRate1080p,
			options.MaxRate720p,
			options.MaxRate480p,
		} {
			if !bitratePattern.MatchString(bitrate) {
				return "", nil, errors.New("HLS bitrates must be positive values such as 5000k")
			}
		}

		return encodingType, options, nil
	default:
		return "", nil, errors.New("encoding must be mp4 or hls")
	}
}

func resolveMediaPath(root, mediaKey string) (string, error) {
	if mediaKey == "" || filepath.IsAbs(mediaKey) {
		return "", errors.New("invalid media key")
	}

	rootPath, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	rootPath, err = filepath.EvalSymlinks(rootPath)
	if err != nil {
		return "", err
	}

	inputPath, err := filepath.Abs(filepath.Join(rootPath, filepath.Clean(mediaKey)))
	if err != nil {
		return "", err
	}
	inputPath, err = filepath.EvalSymlinks(inputPath)
	if err != nil {
		return "", err
	}

	relativePath, err := filepath.Rel(rootPath, inputPath)
	if err != nil {
		return "", err
	}
	if relativePath == ".." || strings.HasPrefix(relativePath, ".."+string(filepath.Separator)) {
		return "", errors.New("media key escapes input directory")
	}

	info, err := os.Stat(inputPath)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", errors.New("media key does not reference a regular file")
	}

	return inputPath, nil
}
