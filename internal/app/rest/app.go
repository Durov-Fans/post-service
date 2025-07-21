package rest

import (
	"context"
	"log"
	"net/http"
	"os"
	"post-service/internal/server/rest"
	"post-service/internal/services/post"
	"time"
)

type RestApp struct {
	server *http.Server
	port   string
}

func NewRestApp(postService *post.Post, port string) *RestApp {
	server := rest.NewServer(*postService, port)

	return &RestApp{
		server: server,
		port:   port,
	}
}

func (app *RestApp) MustRun() {
	if err := app.Run(); err != nil {
		panic(err)
	}
}

func (app *RestApp) Run() error {
	log.Printf("Rest rest listening on port %s", app.port)

	go func() {
		if err := app.server.ListenAndServe(); err != nil {
			log.Printf("Error starting http rest on port %s", app.port)
		}
	}()
	return nil
}
func (app *RestApp) Stop() {
	log.Printf("Rest rest shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	app.server.Shutdown(ctx)

	os.Exit(0)
}
