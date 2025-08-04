package rest

import (
	"log"
	"net/http"
	"os"
	"post-service/internal/server/rest"
	"post-service/internal/services/post"

	"github.com/rs/cors"
)

type RestApp struct {
	server *rest.ServerApi
	port   string
}

func NewRestApp(postService *post.Post, port string, secret string) *RestApp {
	server := rest.NewServer(*postService, port, secret)

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
		if err := http.ListenAndServe(":3000", nil); err != nil {
			log.Printf("Rest rest listening on port %s", app.port)
		}
	}()
	return nil
}
func (app *RestApp) Stop() {
	log.Printf("Rest rest shutting down")

	os.Exit(0)
}
