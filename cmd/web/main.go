package main

import (
	"fmt"
	"log"
	"net/http"
)

const port = ":4000"

type application struct {
}

func main() {

	// app := application{}

	fmt.Println("Starting server on port", port)

	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Panic(err)
	}

}
