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
	"go-breeders/streamer"

	"github.com/joho/godotenv"
)

const port = ":4000"

type application struct {
	templateMap  map[string]*template.Template
	config       appConfig
	App          *configuration.Application
	videoQueue   chan streamer.VideoProcessingJob
	videoStream  *streamer.VideoDispatcher
	videoResults chan streamer.ProcessingMessage
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

	videoQueue := make(chan streamer.VideoProcessingJob, numWorkers)
	defer close(videoQueue)
	videoResults := make(chan streamer.ProcessingMessage, numWorkers)

	// Load .env file into environment (no-op if file missing)
	_ = godotenv.Load()

	app := application{
		templateMap:  make(map[string]*template.Template),
		videoQueue:   videoQueue,
		videoResults: videoResults,
	}

	flag.BoolVar(&app.config.useCache, "cache", false, "Use template cache")
	flag.StringVar(&app.config.db.dsn, "dsn", os.Getenv("DSN"), "MySQL data source name")
	flag.StringVar(&app.config.media.inputDir, "media-input", "./input", "Video input directory")
	flag.StringVar(&app.config.media.outputDir, "media-output", "./output", "Video output directory")
	flag.Parse()

	conn, err := driver.OpenDB(app.config.db.dsn)
	if err != nil {
		log.Println(err)
		log.Fatal(err)
	}
	defer conn.Close()

	xmlBackend := &adapters.XMLBackend{}
	xmlAdapter := &adapters.RemoteService{Remote: xmlBackend}

	app.App = configuration.New(conn, xmlAdapter)
	if err := app.App.DB.ResetProcessingVideoJobs(); err != nil {
		log.Println(err)
		log.Fatal(err)
	}

	wp := streamer.New(videoQueue, numWorkers)
	wp.Run()
	app.videoStream = wp
	go app.listenToVideoResults()

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
