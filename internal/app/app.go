package app

import (
	"post-service/internal/app/rest"
	"post-service/internal/services/post"
	"post-service/internal/storage/postgres"
)

type App struct {
	RestApp rest.RestApp
}

func New(port string, storageUrl string) *App {

	storage, err := postgres.InitDB(storageUrl)
	if err != nil {
		panic(err)
	}

	postService := post.New(storage)

	postApp := rest.NewRestApp(postService, port)

	return &App{
		RestApp: *postApp,
	}
}
