package main

import (
	"fmt"
	"net/http"

	"github.com/coolycow/shortener/internal/config"
	"github.com/coolycow/shortener/internal/handler"
)

func main() {
	mux := http.NewServeMux()
	cfg := config.NewConfig()
	shortToOriginal := map[string]string{}
	originalToShort := map[string]string{}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handler.PostHandler(w, r, cfg, shortToOriginal, originalToShort)
		case http.MethodGet:
			handler.GetHandler(w, r, shortToOriginal)
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
