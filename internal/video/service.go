package video

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"

	"go-breeders/models"
	"go-breeders/streamer"
)

var (
	ErrNotStarted        = errors.New("video processing service is not started")
	ErrInvalidOptions    = errors.New("invalid video processing options")
	ErrAlreadyProcessing = errors.New("video job is already processing")
	ErrInputUnavailable  = errors.New("input media is not available")
	ErrOutputUnavailable = errors.New("processed video is not available")
)

var bitratePattern = regexp.MustCompile(`^[1-9][0-9]*[kKmM]?$`)

type Repository interface {
	AllVideoJobs() ([]*models.VideoJob, error)
	GetVideoJobByID(id int) (*models.VideoJob, error)
	ClaimVideoJob(id int, encodingType string) (bool, error)
	CompleteVideoJob(id int, outputReference string) error
	FailVideoJob(id int, errorMessage string) error
	ResetProcessingVideoJobs() error
}

type ProcessOptions struct {
	EncodingType    string
	RenameOutput    bool
	SegmentDuration int
	MaxRate1080p    string
	MaxRate720p     string
	MaxRate480p     string
}

type Service struct {
	repository Repository
	inputDir   string
	outputDir  string
	queue      chan streamer.VideoProcessingJob
	results    chan streamer.ProcessingMessage
	dispatcher *streamer.VideoDispatcher
	started    atomic.Bool
}

func New(repository Repository, workers int, inputDir, outputDir string) (*Service, error) {
	if repository == nil {
		return nil, errors.New("video repository is required")
	}
	if workers < 1 {
		return nil, errors.New("video worker count must be positive")
	}
	if inputDir == "" || outputDir == "" {
		return nil, errors.New("video input and output directories are required")
	}

	queue := make(chan streamer.VideoProcessingJob, workers)
	results := make(chan streamer.ProcessingMessage, workers)

	return &Service{
		repository: repository,
		inputDir:   inputDir,
		outputDir:  outputDir,
		queue:      queue,
		results:    results,
		dispatcher: streamer.New(queue, workers),
	}, nil
}

func (s *Service) Start() error {
	if !s.started.CompareAndSwap(false, true) {
		return nil
	}

	if err := s.repository.ResetProcessingVideoJobs(); err != nil {
		s.started.Store(false)
		return fmt.Errorf("reset interrupted video jobs: %w", err)
	}

	s.dispatcher.Run()
	go s.listenToResults()

	return nil
}

func (s *Service) Jobs() ([]*models.VideoJob, error) {
	jobs, err := s.repository.AllVideoJobs()
	if err != nil {
		return nil, err
	}

	for _, job := range jobs {
		if job.Status != "completed" || job.OutputReference == "" {
			continue
		}

		size, err := s.outputSize(job)
		if err == nil {
			job.OutputSize = size
		}
	}

	return jobs, nil
}

// OutputFile returns a completed job's requested output file after ensuring it
// remains inside that job's output directory.
func (s *Service) OutputFile(id int, name string) (string, error) {
	if id < 1 || name == "" {
		return "", ErrOutputUnavailable
	}

	job, err := s.repository.GetVideoJobByID(id)
	if err != nil {
		return "", err
	}
	if job.Status != "completed" || job.OutputReference == "" {
		return "", ErrOutputUnavailable
	}

	jobOutputDir, _, err := s.jobOutput(job)
	if err != nil {
		return "", err
	}
	outputPath, err := resolveMediaPath(jobOutputDir, name)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrOutputUnavailable, err)
	}

	return outputPath, nil
}

func (s *Service) outputSize(job *models.VideoJob) (int64, error) {
	jobOutputDir, referenceName, err := s.jobOutput(job)
	if err != nil {
		return 0, err
	}

	if job.EncodingType != "hls" {
		outputPath, err := resolveMediaPath(jobOutputDir, referenceName)
		if err != nil {
			return 0, err
		}
		info, err := os.Stat(outputPath)
		if err != nil {
			return 0, err
		}
		return info.Size(), nil
	}

	baseName := strings.TrimSuffix(referenceName, filepath.Ext(referenceName))
	entries, err := os.ReadDir(jobOutputDir)
	if err != nil {
		return 0, err
	}

	var size int64
	for _, entry := range entries {
		if entry.Name() != referenceName && !strings.HasPrefix(entry.Name(), baseName+"-") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			return 0, err
		}
		if info.Mode().IsRegular() {
			size += info.Size()
		}
	}
	if size == 0 {
		return 0, ErrOutputUnavailable
	}

	return size, nil
}

func (s *Service) jobOutput(job *models.VideoJob) (string, string, error) {
	if job.EncodingType != "mp4" && job.EncodingType != "hls" {
		return "", "", ErrOutputUnavailable
	}

	jobReferenceRoot := filepath.ToSlash(filepath.Join(
		job.EncodingType,
		fmt.Sprintf("job-%d", job.ID),
	))
	cleanReference := filepath.ToSlash(filepath.Clean(job.OutputReference))
	if !strings.HasPrefix(cleanReference, jobReferenceRoot+"/") {
		return "", "", ErrOutputUnavailable
	}

	referenceName := strings.TrimPrefix(cleanReference, jobReferenceRoot+"/")
	if referenceName == "" || strings.Contains(referenceName, "/") {
		return "", "", ErrOutputUnavailable
	}

	jobOutputDir := filepath.Join(s.outputDir, filepath.FromSlash(jobReferenceRoot))
	return jobOutputDir, referenceName, nil
}

func (s *Service) Process(ctx context.Context, id int, options ProcessOptions) (*models.VideoJob, error) {
	if !s.started.Load() {
		return nil, ErrNotStarted
	}
	if id < 1 {
		return nil, fmt.Errorf("%w: invalid video job id", ErrInvalidOptions)
	}

	streamOptions, err := validateOptions(options)
	if err != nil {
		return nil, err
	}

	job, err := s.repository.GetVideoJobByID(id)
	if err != nil {
		return nil, err
	}

	inputPath, err := resolveMediaPath(s.inputDir, job.InputMediaKey)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInputUnavailable, err)
	}

	outputDir := filepath.Join(
		s.outputDir,
		options.EncodingType,
		fmt.Sprintf("job-%d", job.ID),
	)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("create video output directory: %w", err)
	}

	claimed, err := s.repository.ClaimVideoJob(id, options.EncodingType)
	if err != nil {
		return nil, err
	}
	if !claimed {
		return nil, ErrAlreadyProcessing
	}

	encodedVideo := s.dispatcher.NewVideo(
		job.ID,
		inputPath,
		outputDir,
		options.EncodingType,
		s.results,
		streamOptions,
	)

	select {
	case s.queue <- streamer.VideoProcessingJob{Video: encodedVideo}:
		job.EncodingType = options.EncodingType
		job.Status = "processing"
		job.OutputReference = ""
		job.ErrorMessage = ""
		return job, nil
	case <-ctx.Done():
		if err := s.repository.FailVideoJob(id, "Request cancelled before job was queued"); err != nil {
			log.Printf("Error failing cancelled video job %d: %v", id, err)
		}
		return nil, ctx.Err()
	}
}

func (s *Service) listenToResults() {
	for result := range s.results {
		job, err := s.repository.GetVideoJobByID(result.ID)
		if err != nil {
			log.Printf("Error loading completed video job %d: %v", result.ID, err)
			continue
		}

		if !result.Successful {
			if err := s.repository.FailVideoJob(result.ID, result.Message); err != nil {
				log.Printf("Error persisting failed video job %d: %v", result.ID, err)
			}
			continue
		}

		outputReference := filepath.ToSlash(filepath.Join(
			job.EncodingType,
			fmt.Sprintf("job-%d", job.ID),
			result.OutputFile,
		))
		if err := s.repository.CompleteVideoJob(result.ID, outputReference); err != nil {
			log.Printf("Error completing video job %d: %v", result.ID, err)
		}
	}
}

func validateOptions(options ProcessOptions) (*streamer.VideoOptions, error) {
	streamOptions := &streamer.VideoOptions{
		RenameOutput: options.RenameOutput,
	}

	switch options.EncodingType {
	case "mp4":
		return streamOptions, nil
	case "hls":
		if options.SegmentDuration < 1 {
			return nil, fmt.Errorf("%w: segment duration must be a positive integer", ErrInvalidOptions)
		}

		for _, bitrate := range []string{
			options.MaxRate1080p,
			options.MaxRate720p,
			options.MaxRate480p,
		} {
			if !bitratePattern.MatchString(bitrate) {
				return nil, fmt.Errorf(
					"%w: HLS bitrates must be positive values such as 5000k",
					ErrInvalidOptions,
				)
			}
		}

		streamOptions.SegmentDuration = options.SegmentDuration
		streamOptions.MaxRate1080p = options.MaxRate1080p
		streamOptions.MaxRate720p = options.MaxRate720p
		streamOptions.MaxRate480p = options.MaxRate480p
		return streamOptions, nil
	default:
		return nil, fmt.Errorf("%w: encoding must be mp4 or hls", ErrInvalidOptions)
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
