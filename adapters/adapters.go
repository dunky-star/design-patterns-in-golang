package adapters

import (
	"encoding/json"
	"encoding/xml"
	"go-breeders/models"
	"net/http"
)

// CatBreedsInterface is simply our target interface, which defines all the methods that
// any type which implements this interface must have.
type CatBreedsInterface interface {
	GetAllCatBreeds() ([]*models.CatBreed, error)
	GetCatBreedByName(b string) (*models.CatBreed, error)
}

// RemoteService is the Adaptor type. It embeds a DataInterface interface
// (which is critical to the pattern).
type RemoteService struct {
	Remote CatBreedsInterface
}

// GetAllBreeds is the function on RemoteService which lets us
// call any adaptor which implements the DataInterface type.
func (rs *RemoteService) GetAllBreeds() ([]*models.CatBreed, error) {
	return rs.Remote.GetAllCatBreeds()
}

// JSONBackend is the JSON adaptee, which needs to satisfy the CatBreedsInterface by
// have a GetAllCatsBreeds method.
type JSONBackend struct{}

// GetAllCatBreeds is necessary so that JSONBackend satisfies the CatBreedsInterface requirements.
func (jb *JSONBackend) GetAllCatBreeds() ([]*models.CatBreed, error) {
	resp, err := http.Get("http://localhost:8081/api/cat-breeds/all/json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var breeds []*models.CatBreed
	if err := json.NewDecoder(resp.Body).Decode(&breeds); err != nil {
		return nil, err
	}

	return breeds, nil
}

func (jb *JSONBackend) GetCatBreedByName(b string) (*models.CatBreed, error) {
	resp, err := http.Get("http://localhost:8081/api/cat-breeds/" + b + "/json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var breed models.CatBreed
	if err := json.NewDecoder(resp.Body).Decode(&breed); err != nil {
		return nil, err
	}

	return &breed, nil
}

type XMLBackend struct{}

func (xb *XMLBackend) GetAllCatBreeds() ([]*models.CatBreed, error) {
	resp, err := http.Get("http://localhost:8081/api/cat-breeds/all/xml")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	type catBreeds struct {
		XMLName struct{}           `xml:"cat-breeds"`
		Breeds  []*models.CatBreed `xml:"cat-breed"`
	}

	var breeds catBreeds
	if err := xml.NewDecoder(resp.Body).Decode(&breeds); err != nil {
		return nil, err
	}

	return breeds.Breeds, nil
}

func (xb *XMLBackend) GetCatBreedByName(b string) (*models.CatBreed, error) {
	resp, err := http.Get("http://localhost:8081/api/cat-breeds/" + b + "/json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var breed models.CatBreed
	if err := xml.NewDecoder(resp.Body).Decode(&breed); err != nil {
		return nil, err
	}

	return &breed, nil
}

type TestBackend struct{}

func (tb *TestBackend) GetAllCatBreeds() ([]*models.CatBreed, error) {
	breeds := []*models.CatBreed{
		{ID: 1, Breed: "Tomcat", Details: "Some details"},
	}

	return breeds, nil
}

func (tb *TestBackend) GetCatBreedByName(b string) (*models.CatBreed, error) {
	return nil, nil
}
