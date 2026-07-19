package repository

import "go-breeders/models"

func (m *testRepository) AllVideoJobs() ([]*models.VideoJob, error) {
	return []*models.VideoJob{
		{ID: 1, InputMediaKey: "puppy1.mp4", EncodingType: "mp4", Status: "pending"},
		{ID: 2, InputMediaKey: "puppy2.mp4", EncodingType: "mp4", Status: "pending"},
	}, nil
}

func (m *testRepository) GetVideoJobByID(id int) (*models.VideoJob, error) {
	return &models.VideoJob{
		ID:            id,
		InputMediaKey: "puppy1.mp4",
		EncodingType:  "mp4",
		Status:        "pending",
	}, nil
}

func (m *testRepository) ClaimVideoJob(id int, encodingType string) (bool, error) {
	return true, nil
}

func (m *testRepository) CompleteVideoJob(id int, outputReference string) error {
	return nil
}

func (m *testRepository) FailVideoJob(id int, errorMessage string) error {
	return nil
}

func (m *testRepository) ResetProcessingVideoJobs() error {
	return nil
}
