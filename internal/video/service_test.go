package video

import (
	"context"
	"errors"
	"os"
	"path/filepath"
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
