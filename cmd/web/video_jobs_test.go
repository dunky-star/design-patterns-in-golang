package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go-breeders/internal/video"
	"go-breeders/models"
)

type stubVideoProcessingService struct {
	jobs       []*models.VideoJob
	processErr error
	outputPath string
	outputErr  error
}

func (s *stubVideoProcessingService) Jobs() ([]*models.VideoJob, error) {
	return s.jobs, nil
}

func (s *stubVideoProcessingService) OutputFile(_ int, _ string) (string, error) {
	return s.outputPath, s.outputErr
}

func (s *stubVideoProcessingService) Process(
	_ context.Context,
	_ int,
	_ video.ProcessOptions,
) (*models.VideoJob, error) {
	if s.processErr != nil {
		return nil, s.processErr
	}
	return &models.VideoJob{ID: 1, Status: "processing"}, nil
}

func TestApplication_GetVideoJobs(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/video-jobs", nil)
	rr := httptest.NewRecorder()
	app := application{
		videoService: &stubVideoProcessingService{
			jobs: []*models.VideoJob{{ID: 1, Status: "pending"}},
		},
	}

	app.GetVideoJobs(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("wrong response code; got %d, wanted %d", rr.Code, http.StatusOK)
	}
}

func TestApplication_ProcessVideoJobUnavailable(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/video-jobs/1/process", nil)
	rr := httptest.NewRecorder()

	app := application{}
	app.ProcessVideoJob(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf(
			"wrong response code; got %d, wanted %d",
			rr.Code,
			http.StatusServiceUnavailable,
		)
	}
}

func TestApplication_GetProcessedVideo(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "processed.mp4")
	if err := os.WriteFile(outputPath, []byte("processed video"), 0600); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/video-jobs/1/output/processed.mp4",
		nil,
	)
	req.SetPathValue("id", "1")
	req.SetPathValue("file", "processed.mp4")
	rr := httptest.NewRecorder()
	app := application{
		videoService: &stubVideoProcessingService{outputPath: outputPath},
	}

	app.GetProcessedVideo(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("wrong response code; got %d, wanted %d", rr.Code, http.StatusOK)
	}
	if rr.Body.String() != "processed video" {
		t.Errorf("response body = %q, want processed video", rr.Body.String())
	}
}

func TestApplication_ProcessVideoJobRejectsInvalidEncoding(t *testing.T) {
	body := strings.NewReader("encoding=avi")
	req := httptest.NewRequest(http.MethodPost, "/api/video-jobs/1/process", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", "1")
	rr := httptest.NewRecorder()

	app := application{
		videoService: &stubVideoProcessingService{
			processErr: video.ErrInvalidOptions,
		},
	}
	app.ProcessVideoJob(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf(
			"wrong response code; got %d, wanted %d",
			rr.Code,
			http.StatusBadRequest,
		)
	}
}

func TestVideoProcessOptionsFromRequest(t *testing.T) {
	tests := []struct {
		name                string
		form                url.Values
		wantType            string
		wantSegmentDuration int
		wantError           bool
	}{
		{
			name:     "mp4",
			form:     url.Values{"encoding": {"mp4"}},
			wantType: "mp4",
		},
		{
			name: "hls",
			form: url.Values{
				"encoding":         {"hls"},
				"segment_duration": {"10"},
				"max_rate_1080p":   {"5000k"},
				"max_rate_720p":    {"2800k"},
				"max_rate_480p":    {"1400k"},
			},
			wantType:            "hls",
			wantSegmentDuration: 10,
		},
		{
			name: "invalid HLS segment duration",
			form: url.Values{
				"encoding":         {"hls"},
				"segment_duration": {"invalid"},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(
				http.MethodPost,
				"/api/video-jobs/1/process",
				strings.NewReader(tt.form.Encode()),
			)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			if err := req.ParseForm(); err != nil {
				t.Fatal(err)
			}

			options, err := videoProcessOptionsFromRequest(req)
			if (err != nil) != tt.wantError {
				t.Fatalf("error = %v, wantError %t", err, tt.wantError)
			}
			if options.EncodingType != tt.wantType {
				t.Errorf("encoding type = %q, want %q", options.EncodingType, tt.wantType)
			}
			if options.SegmentDuration != tt.wantSegmentDuration {
				t.Errorf(
					"segment duration = %d, want %d",
					options.SegmentDuration,
					tt.wantSegmentDuration,
				)
			}
		})
	}
}
