package pets

import (
	"encoding/json"
	"testing"
)

func TestPetJSONOmitsEmptyOptionalFields(t *testing.T) {
	pet := Pet{
		Species: "dog",
		Breed:   "Labrador",
	}

	encoded, err := json.Marshal(pet)
	if err != nil {
		t.Fatal(err)
	}

	const expected = `{"species":"dog","breed":"Labrador"}`
	if string(encoded) != expected {
		t.Errorf("encoded pet = %s, want %s", encoded, expected)
	}
}
