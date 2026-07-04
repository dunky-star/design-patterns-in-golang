package main

import (
	"go-breeders/internal/repository"
	"os"
	"testing"
)

var testApp application

func TestMain(m *testing.M) {
	testApp = application{
		DB: repository.New(nil),
	}

	os.Exit(m.Run())
}
