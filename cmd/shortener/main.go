package main

import (
	"fmt"
	"log"

	"github.com/coolycow/shortener/internal/config"
	"github.com/coolycow/shortener/internal/repository"
	"github.com/coolycow/shortener/internal/router"
)

func main() {
	cfg, err := config.InitConfig()

	if err != nil {
		log.Fatalf("Failed to initialize configuration: %v", err)
	}

	repo := repository.NewDoubleMapsRepository()
	r := router.NewRouter(cfg, repo)

	serverAddress := cfg.GetServerAddress()
	fmt.Printf("Starting server %s\n", serverAddress)

	err = r.Run(serverAddress)

	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
