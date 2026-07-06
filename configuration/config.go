package configuration

import (
	"database/sql"
	"sync"

	"go-breeders/internal/repository"
)

type Application struct {
	DB repository.Repository
}

var instance *Application
var once sync.Once
var db *sql.DB

func New(pool *sql.DB) *Application {
	db = pool
	return GetInstance()
}

func GetInstance() *Application {
	once.Do(func() {
		instance = &Application{
			DB: repository.New(db),
		}
	})
	return instance
}
