package streamer

import (
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name       string
		jobQueue   chan VideoProcessingJob
		maxWorkers int
	}{
		{
			name:       "one worker",
			jobQueue:   make(chan VideoProcessingJob, 10),
			maxWorkers: 1,
		},
		{
			name:       "three workers",
			jobQueue:   make(chan VideoProcessingJob, 10),
			maxWorkers: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.jobQueue, tt.maxWorkers)

			if got.maxWorkers != tt.maxWorkers {
				t.Errorf("New() maxWorkers = %d, want %d", got.maxWorkers, tt.maxWorkers)
			}

			if reflect.ValueOf(got.jobQueue).Kind() != reflect.Chan {
				t.Fatal("jobQueue is not a channel")
			}

			channelType := reflect.ValueOf(got.jobQueue).Type().Elem()
			if channelType.Name() != "VideoProcessingJob" {
				t.Errorf("jobQueue element = %s, want VideoProcessingJob", channelType.Name())
			}
		})
	}
}

func TestNewVideo(t *testing.T) {
	tests := []struct {
		name string
		id   int
		enc  string
		ops  *VideoOptions
	}{
		{name: "mp4", id: 1, enc: "mp4", ops: &VideoOptions{RenameOutput: false}},
		{name: "mp4 rename", id: 1, enc: "mp4", ops: &VideoOptions{RenameOutput: true}},
		{name: "hls", id: 1, enc: "hls", ops: &VideoOptions{RenameOutput: false}},
		{name: "hls rename", id: 1, enc: "hls", ops: &VideoOptions{RenameOutput: true}},
		{name: "nil options", id: 1, enc: "mp4", ops: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wp := New(make(chan VideoProcessingJob), 1)
			video := wp.NewVideo(
				tt.id,
				"./a/b.mp4",
				"./testdata/output",
				tt.enc,
				testNotifyChan,
				tt.ops,
			)

			if video.Options == nil {
				t.Fatal("NewVideo() returned nil options")
			}

			if tt.ops != nil && video.Options.RenameOutput != tt.ops.RenameOutput {
				t.Errorf(
					"RenameOutput = %t, want %t",
					video.Options.RenameOutput,
					tt.ops.RenameOutput,
				)
			}
		})
	}
}

func TestEncode(t *testing.T) {
	tests := []struct {
		name           string
		id             int
		enc            string
		ops            *VideoOptions
		useFailEncoder bool
		expectSuccess  bool
	}{
		{name: "mp4", id: 1, enc: "mp4", ops: &VideoOptions{}, expectSuccess: true},
		{name: "mp4 without options", id: 2, enc: "mp4", ops: nil, expectSuccess: true},
		{name: "mp4 failure", id: 3, enc: "mp4", ops: &VideoOptions{}, useFailEncoder: true},
		{name: "hls", id: 4, enc: "hls", ops: &VideoOptions{RenameOutput: true}, expectSuccess: true},
		{name: "hls failure", id: 5, enc: "hls", ops: &VideoOptions{RenameOutput: true}, useFailEncoder: true},
		{name: "invalid encoding type", id: 6, enc: "fish", ops: &VideoOptions{}, expectSuccess: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wp := New(make(chan VideoProcessingJob), 1)
			video := wp.NewVideo(
				tt.id,
				"./testdata/input.mp4",
				"./testdata/output",
				tt.enc,
				testNotifyChan,
				tt.ops,
			)

			if tt.useFailEncoder {
				video.Encoder = testProcessorFailing
			} else {
				video.Encoder = testProcessor
			}

			video.encode()

			result := <-testNotifyChan
			if result.Successful != tt.expectSuccess {
				t.Errorf(
					"Successful = %t, want %t",
					result.Successful,
					tt.expectSuccess,
				)
			}
		})
	}
}

func TestPool(t *testing.T) {
	tests := []struct {
		name string
		enc  string
		ops  *VideoOptions
	}{
		{name: "mp4", enc: "mp4", ops: &VideoOptions{RenameOutput: false}},
		{name: "mp4 rename", enc: "mp4", ops: &VideoOptions{RenameOutput: true}},
		{name: "hls", enc: "hls", ops: &VideoOptions{RenameOutput: false}},
		{name: "hls rename", enc: "hls", ops: &VideoOptions{RenameOutput: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			videoQueue := make(chan VideoProcessingJob, 10)
			wp := New(videoQueue, 3)
			wp.Processor = testProcessor
			wp.Run()

			video := wp.NewVideo(
				1,
				"./a/b.mp4",
				"./testdata/output",
				tt.enc,
				testNotifyChan,
				tt.ops,
			)

			videoQueue <- VideoProcessingJob{Video: video}

			result := <-testNotifyChan
			if !result.Successful {
				t.Errorf("%s: encoding failed: %s", tt.name, result.Message)
			}
		})
	}
}
