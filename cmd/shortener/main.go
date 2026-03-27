package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/coolycow/shortener/internal/config"
	"github.com/coolycow/shortener/internal/logger"
	"github.com/coolycow/shortener/internal/observer/audit"
	"github.com/coolycow/shortener/internal/repository"
	"github.com/coolycow/shortener/internal/router"
	"go.uber.org/zap"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func orNA(s string) string {
	if s == "" {
		return "N/A"
	}
	return s
}

func main() {
	// Вывод информации о сборке при старте
	fmt.Fprintf(os.Stdout, "Build version: %s\n", orNA(buildVersion))
	fmt.Fprintf(os.Stdout, "Build date: %s\n", orNA(buildDate))
	fmt.Fprintf(os.Stdout, "Build commit: %s\n", orNA(buildCommit))

	// Инициализируем настройки (приоритет: окружение, флаги, дефолт)
	cfg, err := config.InitConfig()

	if err != nil {
		log.Fatalf("Failed to initialize configuration: %v", err)
	}

	// Выводим настройки в консоль для наглядности
	cfg.PrintConfig()

	// Инициализируем логер
	if err = logger.Initialize(cfg.LogLevel); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Инициализируем репозиторий
	var repo repository.URLRepository
	if cfg.DatabaseDSN != "" {
		repo, err = repository.NewPostgresRepository(cfg.DatabaseDSN)

		if err != nil {
			log.Fatalf("Failed to initialize postgres repository: %v", err)
		}

		logger.Log.Info("Initialized postgres repository", zap.String("dsn", cfg.DatabaseDSN))

		if cfg.RunMigrations {
			if err = repo.RunMigrations(); err != nil {
				log.Fatalf("Failed to run migrations: %v", err)
			}
			logger.Log.Info("Run migrations succeeded")
			return
		}
	} else {
		repo = repository.NewDoubleMapsRepository(cfg.FileStoragePath)
		logger.Log.Info("Initialized doublemaps repository", zap.String("path", cfg.FileStoragePath))
	}

	// В конце работы приложения необходимо правильно закрыть хранилище.
	defer func() {
		if err = repo.Close(); err != nil {
			logger.Log.Error("Error closing repository", zap.Error(err))
		}
	}()

	// Инициализируем нотифайер аудита (файл и/или URL из конфига; если оба пустые — приёмников не будет)
	auditNotifier := audit.NewNotifier(cfg.AuditFile, cfg.AuditURL)

	// Инициализируем роутер
	r := router.NewRouter(cfg, repo, auditNotifier)

	// Получаем адрес сервера из настроек и запускаем сервер
	serverAddress := cfg.GetServerAddress()
	logger.Log.Info("Running server ", zap.String("address", serverAddress))

	// Если включено HTTPS, то запускаем сервер с HTTPS
	if cfg.EnableHTTPS {
		srv := &http.Server{
			Addr:    serverAddress,
			Handler: r,
		}

		err = srv.ListenAndServeTLS(cfg.TLSCertFile, cfg.TLSKeyFile)
	} else {
		// Если не включено HTTPS, то запускаем сервер с HTTP
		err = r.Run(serverAddress)
	}

	// Если сервер не стартовал - фатальная ошибка
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
