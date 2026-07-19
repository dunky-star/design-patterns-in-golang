package streamer

import (
	"errors"
	"os"
	"testing"
)

// testProcessor satisfies Encoder and always returns no error.
var testProcessor Processor

// testProcessorFailing satisfies Encoder and always returns an error.
var testProcessorFailing Processor

// testNotifyChan receives encode results during tests.
var testNotifyChan chan ProcessingMessage

func TestMain(m *testing.M) {
	var successfulEncoder testEncoder
	testProcessor = Processor{
		Engine: &successfulEncoder,
	}

	var failingEncoder testEncoderFailing
	testProcessorFailing = Processor{
		Engine: &failingEncoder,
	}

	testNotifyChan = make(chan ProcessingMessage, 10)

	os.Exit(m.Run())
}

// testEncoder simulates successful encoding.
type testEncoder struct{}

func (te *testEncoder) EncodeToMP4(v *Video, baseFileName string) error {
	return nil
}

func (te *testEncoder) EncodeToHLS(v *Video, baseFileName string) error {
	return nil
}

// testEncoderFailing simulates failed encoding.
type testEncoderFailing struct{}

func (tef *testEncoderFailing) EncodeToMP4(v *Video, baseFileName string) error {
	return errors.New("some error")
}

func (tef *testEncoderFailing) EncodeToHLS(v *Video, baseFileName string) error {
	return errors.New("some error")
}
