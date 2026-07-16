package repository

import "go-breeders/models"

func (m *testRepository) AllDogBreeds() ([]*models.DogBreed, error) {
	return nil, nil
}

func (m *testRepository) GetBreedByName(b string) (*models.DogBreed, error) {
	return nil, nil
}

func (m *testRepository) GetDogOfMonthByID(id int) (*models.DogOfMonth, error) {
	return nil, nil
}
