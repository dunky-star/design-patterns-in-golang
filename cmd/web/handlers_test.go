package main

import (
	"go-breeders/streamer"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestApplication_GetAllDogBreedsJSON(t *testing.T) {
	req, _ := http.NewRequest("GET", "/api/dog-breeds", nil)

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(testApp.GetAllDogBreedsJSON)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("wrong response code; got %d, wanted 200", rr.Code)
	}
}

func TestApplication_ProcessVideosUnavailable(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/process-videos", nil)
	rr := httptest.NewRecorder()

	app := application{}
	app.ProcessVideos(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf(
			"wrong response code; got %d, wanted %d",
			rr.Code,
			http.StatusServiceUnavailable,
		)
	}
}

func TestApplication_ProcessVideosRejectsInvalidEncoding(t *testing.T) {
	body := strings.NewReader("encoding=avi")
	req := httptest.NewRequest(http.MethodPost, "/api/process-videos", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	app := application{
		videoStream: streamer.New(make(chan streamer.VideoProcessingJob), 1),
	}
	app.ProcessVideos(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf(
			"wrong response code; got %d, wanted %d",
			rr.Code,
			http.StatusBadRequest,
		)
	}
}
