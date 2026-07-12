package configuration

import (
	"database/sql"
	"sync"

	"go-breeders/adapters"
	"go-breeders/internal/repository"
)

type Application struct {
	DB         repository.Repository
	CatService *adapters.RemoteService
}

var instance *Application
var once sync.Once
var db *sql.DB
var catService *adapters.RemoteService

func New(pool *sql.DB, cs *adapters.RemoteService) *Application {
	db = pool
	catService = cs
	return GetInstance()
}

func GetInstance() *Application {
	once.Do(func() {
		instance = &Application{
			DB:         repository.New(db),
			CatService: catService,
		}
	})
	return instance
}
