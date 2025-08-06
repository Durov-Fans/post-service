package app

import (
	"post-service/internal/app/rest"
	"post-service/internal/lib/jwt"
	"post-service/internal/services"
	"post-service/internal/storage/postgres"
)

type App struct {
	RestApp rest.RestApp
}

func New(port string, storageUrl string, jwt *jwt.JWT) *App {

	storage, err := postgres.InitDB(storageUrl)
	if err != nil {
		panic(err)
	}

	postService := services.New(storage)

	postApp := rest.New(postService, port, jwt)
	return &App{
		RestApp: *postApp,
	}
}
