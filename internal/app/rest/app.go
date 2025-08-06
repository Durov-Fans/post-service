package rest

import (
	"github.com/rs/cors"
	"log"
	"net/http"
	"os"
	"post-service/internal/lib/jwt"
	"post-service/internal/server/rest"
	"post-service/internal/services"
)

type RestApp struct {
	server *rest.ServerApi
	port   string
}

func New(postService *services.Post, port string, jwt *jwt.JWT) *RestApp {
	server := rest.New(*postService, port, jwt)

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

	r := app.server.ConfigureRoutes()
	http.Handle("/", cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: false,
	}).Handler(r))

	go func() {
		if err := http.ListenAndServe(app.port, nil); err != nil {
			log.Printf("Rest rest listening on port %s", app.port)
		}
	}()
	return nil
}
func (app *RestApp) Stop() {
	log.Printf("Rest rest shutting down")

	os.Exit(0)
}
