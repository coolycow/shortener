package main

import (
	"fmt"

	"github.com/coolycow/shortener/internal/config"
	"github.com/coolycow/shortener/internal/repository"
	"github.com/coolycow/shortener/internal/router"
)

func main() {
	cfg := config.NewConfig()
	repo := repository.NewDoubleMapsRepository()
	r := router.NewRouter(cfg, repo)

	fmt.Println("Starting server ")

	err := r.Run(cfg.ServerAddress)

	if err != nil {

		panic(err)
	}
}
