package main

import (
	"fmt"
	"log"

	"github.com/coolycow/shortener/internal/config"
	"github.com/coolycow/shortener/internal/repository"
	"github.com/coolycow/shortener/internal/router"
)

func main() {
	// Инициализируем настройки (приоритет: окружение, флаги, дефолт)
	cfg, err := config.InitConfig()

	if err != nil {
		log.Fatalf("Failed to initialize configuration: %v", err)
	}

	// Выводим настройки в консоль для наглядности
	cfg.PrintConfig()

	// Инициализируем репозиторий и роутер
	repo := repository.NewDoubleMapsRepository()
	r := router.NewRouter(cfg, repo)

	// Получаем адрес сервера из настроек и запускаем сервер
	serverAddress := cfg.GetServerAddress()
	fmt.Printf("Starting server %s\n", serverAddress)

	err = r.Run(serverAddress)

	// Если сервер не стартовал - фатальная ошибка
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
