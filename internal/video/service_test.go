package video

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go-breeders/models"
)

type stubRepository struct {
	job           *models.VideoJob
	claimed       bool
	claimEncoding string
}

func (r *stubRepository) AllVideoJobs() ([]*models.VideoJob, error) {
	return []*models.VideoJob{r.job}, nil
}

func (r *stubRepository) GetVideoJobByID(_ int) (*models.VideoJob, error) {
	return r.job, nil
}

func (r *stubRepository) ClaimVideoJob(_ int, encodingType string) (bool, error) {
	r.claimEncoding = encodingType
	return r.claimed, nil
}

func (r *stubRepository) CompleteVideoJob(_ int, _ string) error {
	return nil
}

func (r *stubRepository) FailVideoJob(_ int, _ string) error {
	return nil
}

func (r *stubRepository) ResetProcessingVideoJobs() error {
	return nil
}

func TestServiceProcessQueuesClaimedJob(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(inputDir, "puppy.mp4"), []byte("video"), 0600); err != nil {
		t.Fatal(err)
	}

	repository := &stubRepository{
		job: &models.VideoJob{
			ID:            1,
			InputMediaKey: "puppy.mp4",
			Status:        "pending",
		},
		claimed: true,
	}
	service, err := New(repository, 1, inputDir, outputDir)
	if err != nil {
		t.Fatal(err)
	}
	service.started.Store(true)

	job, err := service.Process(
		context.Background(),
		1,
		ProcessOptions{EncodingType: "mp4"},
	)
	if err != nil {
		t.Fatal(err)
	}

	if job.Status != "processing" {
		t.Errorf("job status = %q, want processing", job.Status)
	}
	if repository.claimEncoding != "mp4" {
		t.Errorf("claimed encoding = %q, want mp4", repository.claimEncoding)
	}

	queuedJob := <-service.queue
	if queuedJob.Video.ID != 1 {
		t.Errorf("queued video ID = %d, want 1", queuedJob.Video.ID)
	}
	if queuedJob.Video.EncodingType != "mp4" {
		t.Errorf("queued encoding = %q, want mp4", queuedJob.Video.EncodingType)
	}
}

func TestServiceOutputFile(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()
	jobOutputDir := filepath.Join(outputDir, "mp4", "job-1")
	if err := os.MkdirAll(jobOutputDir, 0755); err != nil {
		t.Fatal(err)
	}

	expectedPath := filepath.Join(jobOutputDir, "processed.mp4")
	if err := os.WriteFile(expectedPath, []byte("video"), 0600); err != nil {
		t.Fatal(err)
	}

	repository := &stubRepository{
		job: &models.VideoJob{
			ID:              1,
			EncodingType:    "mp4",
			Status:          "completed",
			OutputReference: "mp4/job-1/processed.mp4",
		},
	}
	service, err := New(repository, 1, inputDir, outputDir)
	if err != nil {
		t.Fatal(err)
	}

	outputPath, err := service.OutputFile(1, "processed.mp4")
	if err != nil {
		t.Fatal(err)
	}
	expectedResolvedPath, err := filepath.EvalSymlinks(expectedPath)
	if err != nil {
		t.Fatal(err)
	}
	if outputPath != expectedResolvedPath {
		t.Errorf("output path = %q, want %q", outputPath, expectedResolvedPath)
	}

	jobs, err := service.Jobs()
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs[0].OutputFiles) != 1 {
		t.Fatalf("output file count = %d, want 1", len(jobs[0].OutputFiles))
	}
	if jobs[0].OutputFiles[0].Name != "processed.mp4" {
		t.Errorf("output file name = %q, want processed.mp4", jobs[0].OutputFiles[0].Name)
	}
	if jobs[0].OutputFiles[0].Size != int64(len("video")) {
		t.Errorf("output file size = %d, want %d", jobs[0].OutputFiles[0].Size, len("video"))
	}

	if _, err := service.OutputFile(1, "../secret.mp4"); !errors.Is(err, ErrOutputUnavailable) {
		t.Errorf("traversal error = %v, want ErrOutputUnavailable", err)
	}
}

func TestServiceJobsIncludesEachHLSVideoFileSize(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()
	jobOutputDir := filepath.Join(outputDir, "hls", "job-1")
	if err := os.MkdirAll(jobOutputDir, 0755); err != nil {
		t.Fatal(err)
	}

	files := map[string]string{
		"stream.m3u8":        "master",
		"stream-720p.m3u8":   "variant",
		"stream-720p0.ts":    "segment",
		"previous-output.ts": "ignored",
	}
	expectedFiles := make(map[string]int64)
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(jobOutputDir, name), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
		if strings.HasPrefix(name, "stream-") && filepath.Ext(name) == ".ts" {
			expectedFiles[name] = int64(len(content))
		}
	}

	repository := &stubRepository{
		job: &models.VideoJob{
			ID:              1,
			EncodingType:    "hls",
			Status:          "completed",
			OutputReference: "hls/job-1/stream.m3u8",
		},
	}
	service, err := New(repository, 1, inputDir, outputDir)
	if err != nil {
		t.Fatal(err)
	}

	jobs, err := service.Jobs()
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs[0].OutputFiles) != len(expectedFiles) {
		t.Fatalf(
			"HLS output file count = %d, want %d",
			len(jobs[0].OutputFiles),
			len(expectedFiles),
		)
	}
	for _, file := range jobs[0].OutputFiles {
		if file.Size != expectedFiles[file.Name] {
			t.Errorf(
				"HLS file %q size = %d, want %d",
				file.Name,
				file.Size,
				expectedFiles[file.Name],
			)
		}
	}
}

func TestValidateOptions(t *testing.T) {
	tests := []struct {
		name    string
		options ProcessOptions
		wantErr bool
	}{
		{
			name:    "mp4",
			options: ProcessOptions{EncodingType: "mp4"},
		},
		{
			name: "hls",
			options: ProcessOptions{
				EncodingType:    "hls",
				SegmentDuration: 10,
				MaxRate1080p:    "5000k",
				MaxRate720p:     "2800k",
				MaxRate480p:     "1400k",
			},
		},
		{
			name:    "unsupported encoding",
			options: ProcessOptions{EncodingType: "avi"},
			wantErr: true,
		},
		{
			name: "invalid HLS bitrate",
			options: ProcessOptions{
				EncodingType:    "hls",
				SegmentDuration: 10,
				MaxRate1080p:    "invalid",
				MaxRate720p:     "2800k",
				MaxRate480p:     "1400k",
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := validateOptions(test.options)
			if (err != nil) != test.wantErr {
				t.Fatalf("error = %v, wantError %t", err, test.wantErr)
			}
			if err != nil && !errors.Is(err, ErrInvalidOptions) {
				t.Errorf("error = %v, want ErrInvalidOptions", err)
			}
		})
	}
}

func TestResolveMediaPath(t *testing.T) {
	root := t.TempDir()
	inputPath := filepath.Join(root, "puppy.mp4")
	if err := os.WriteFile(inputPath, []byte("video"), 0600); err != nil {
		t.Fatal(err)
	}

	resolved, err := resolveMediaPath(root, "puppy.mp4")
	if err != nil {
		t.Fatal(err)
	}
	expectedPath, err := filepath.EvalSymlinks(inputPath)
	if err != nil {
		t.Fatal(err)
	}
	if resolved != expectedPath {
		t.Errorf("resolved path = %q, want %q", resolved, expectedPath)
	}

	if _, err := resolveMediaPath(root, "../secret.mp4"); err == nil {
		t.Error("expected traversal media key to be rejected")
	}
}
