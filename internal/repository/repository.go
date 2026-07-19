package repository

import (
	"database/sql"

	"go-breeders/models"
)

// Repository is the database repository. Anything that implements
// this interface must implement all the methods included here.
type Repository interface {
	AllDogBreeds() ([]*models.DogBreed, error)
	GetBreedByName(b string) (*models.DogBreed, error)
	GetDogOfMonthByID(id int) (*models.DogOfMonth, error)
	AllVideoJobs() ([]*models.VideoJob, error)
	GetVideoJobByID(id int) (*models.VideoJob, error)
	ClaimVideoJob(id int, encodingType string) (bool, error)
	CompleteVideoJob(id int, outputReference string) error
	FailVideoJob(id int, errorMessage string) error
	ResetProcessingVideoJobs() error
}

// mysqlRepository is a simple wrapper for the *sql.DB type. This is
// used to return a MySQL/MariaDB repository.
type mysqlRepository struct {
	DB *sql.DB
}

// newMysqlRepository is a convenience factory method to return a new mysqlRepository.
func newMysqlRepository(conn *sql.DB) Repository {
	return &mysqlRepository{
		DB: conn,
	}
}

type testRepository struct {
	DB *sql.DB
}

func newTestRepository(conn *sql.DB) Repository {
	return &testRepository{
		DB: nil,
	}
}

// New returns the appropriate Repository implementation for the given connection.
func New(conn *sql.DB) Repository {
	if conn != nil {
		return newMysqlRepository(conn)
	}

	return newTestRepository(nil)
}
