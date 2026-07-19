package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"go-breeders/adapters"
	"go-breeders/configuration"
	"go-breeders/internal/driver"
	"go-breeders/internal/video"

	"github.com/joho/godotenv"
)

const port = ":4000"

type application struct {
	templateMap  map[string]*template.Template
	config       appConfig
	App          *configuration.Application
	videoService videoProcessingService
}

type appConfig struct {
	useCache bool
	db       struct {
		dsn string
	}
	media struct {
		inputDir  string
		outputDir string
	}
}

func main() {

	const numWorkers = 4

	// Load .env file into environment (no-op if file missing)
	_ = godotenv.Load()

	app := application{
		templateMap: make(map[string]*template.Template),
	}

	flag.BoolVar(&app.config.useCache, "cache", false, "Use template cache")
	flag.StringVar(&app.config.db.dsn, "dsn", os.Getenv("DSN"), "MySQL data source name")
	flag.StringVar(&app.config.media.inputDir, "media-input", "./input", "Video input directory")
	flag.StringVar(&app.config.media.outputDir, "media-output", "./output", "Video output directory")
	flag.Parse()

	conn, err := driver.OpenDB(app.config.db.dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	xmlBackend := &adapters.XMLBackend{}
	xmlAdapter := &adapters.RemoteService{Remote: xmlBackend}

	app.App = configuration.New(conn, xmlAdapter)

	videoService, err := video.New(
		app.App.DB,
		numWorkers,
		app.config.media.inputDir,
		app.config.media.outputDir,
	)
	if err != nil {
		log.Fatal(err)
	}
	if err := videoService.Start(); err != nil {
		log.Fatal(err)
	}
	app.videoService = videoService

	fmt.Println("Starting server on port", port)

	srv := &http.Server{
		Addr:              port,
		Handler:           app.routes(),
		IdleTimeout:       30 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      20 * time.Second,
		ReadHeaderTimeout: 30 * time.Second,
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
