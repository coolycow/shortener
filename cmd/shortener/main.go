package main

import (
	"fmt"
	"net/http"

	"github.com/coolycow/shortener/internal/config"
	"github.com/coolycow/shortener/internal/handler"
	"github.com/coolycow/shortener/internal/repository"
)

func main() {
	mux := http.NewServeMux()
	cfg := config.NewConfig()
	repo := repository.NewDoubleMapsRepository()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handler.PostHandler(cfg, repo)(w, r)
		case http.MethodGet:
			handler.GetHandler(repo)(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	fmt.Println("Starting server ")

	err := http.ListenAndServe(cfg.ServerAddress, mux)

	if err != nil {
		panic(err)
	}
}
