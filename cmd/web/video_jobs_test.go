package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go-breeders/streamer"
)

func TestApplication_GetVideoJobs(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/video-jobs", nil)
	rr := httptest.NewRecorder()

	testApp.GetVideoJobs(rr, req)

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

func TestApplication_ProcessVideoJobRejectsInvalidEncoding(t *testing.T) {
	body := strings.NewReader("encoding=avi")
	req := httptest.NewRequest(http.MethodPost, "/api/video-jobs/1/process", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", "1")
	rr := httptest.NewRecorder()

	app := application{
		videoQueue:   make(chan streamer.VideoProcessingJob, 1),
		videoResults: make(chan streamer.ProcessingMessage, 1),
		videoStream:  streamer.New(make(chan streamer.VideoProcessingJob, 1), 1),
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

func TestVideoOptionsFromRequest(t *testing.T) {
	tests := []struct {
		name      string
		form      url.Values
		wantType  string
		wantError bool
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
			wantType: "hls",
		},
		{
			name:      "invalid encoding",
			form:      url.Values{"encoding": {"avi"}},
			wantError: true,
		},
		{
			name: "invalid HLS bitrate",
			form: url.Values{
				"encoding":         {"hls"},
				"segment_duration": {"10"},
				"max_rate_1080p":   {"invalid"},
				"max_rate_720p":    {"2800k"},
				"max_rate_480p":    {"1400k"},
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

			encodingType, _, err := videoOptionsFromRequest(req)
			if (err != nil) != tt.wantError {
				t.Fatalf("error = %v, wantError %t", err, tt.wantError)
			}
			if encodingType != tt.wantType {
				t.Errorf("encoding type = %q, want %q", encodingType, tt.wantType)
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
