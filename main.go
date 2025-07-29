package main

import (
	"log"
	"os"
	"os/signal"
	"post-service/internal/app"
	"post-service/internal/config"
	"post-service/internal/lib/jwt"
	"syscall"
)

func main() {
	cfg := config.MustLoad()

	log.Printf("Config: %v", cfg)

	jwt := jwt.NewJWT(cfg.JWT.Secret)

	application := app.New(cfg.Server.Port, cfg.DatabaseUrl, jwt)

	application.RestApp.MustRun()

	stop := make(chan os.Signal, 1)

	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	sign := <-stop

	log.Printf("Signal", sign.String())

	application.RestApp.Stop()

	log.Printf("Shutting down")
}
