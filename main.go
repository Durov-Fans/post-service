package main

import (
	"log"
	"os"
	"os/signal"
	"post-service/internal/app"
	"post-service/internal/config"
	"syscall"
)

func main() {
	cfg := config.MustLoad()

	log.Printf("Config: %v", cfg)

	application := app.New(cfg.Server.Port, cfg.DatabaseUrl, cfg.JWT.Secret)

	application.RestApp.MustRun()

	stop := make(chan os.Signal, 1)

	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	sign := <-stop

	log.Printf("Signal", sign.String())

	application.RestApp.Stop()

	log.Printf("Shutting down")
}
