package main

import (
	"fmt"

	"github.com/coolycow/shortener/internal/config"
	"github.com/coolycow/shortener/internal/repository"
	"github.com/coolycow/shortener/internal/router"
)

func main() {
	cfg, err := config.InitConfig()

	if err != nil {
		panic(err)
	}

	repo := repository.NewDoubleMapsRepository()
	r := router.NewRouter(cfg, repo)

	serverAddress := cfg.GetServerAddress()
	fmt.Printf("Starting server %s\n", serverAddress)

	err = r.Run(serverAddress)

	if err != nil {

		panic(err)
	}
}
