package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/coolycow/shortener/internal/config"
	error2 "github.com/coolycow/shortener/internal/error"
	"github.com/coolycow/shortener/internal/logger"
	"github.com/coolycow/shortener/internal/repository"
)

// URLService Сервис для работы в Handler
type URLService interface {
	GetOriginalURL(ctx context.Context, shortURL string) (string, error)
	CreateShortURL(ctx context.Context, originalURL string) (string, error)
}

// Реализация сервисного слоя
type urlService struct {
	repo repository.URLRepository
	cfg  *config.Config
}

// NewURLService инициализация сервиса
func NewURLService(cfg *config.Config, repo repository.URLRepository) URLService {
	return &urlService{
		repo: repo,
		cfg:  cfg,
	}
}

// GetOriginalURL возвращает исходную ссылку по короткому ключу
func (s *urlService) GetOriginalURL(ctx context.Context, shortURL string) (string, error) {
	// Если shortURL пустой, то возвращаем ошибку 400
	if shortURL == "" {
		return "", error2.CustomError{
			Message:    "key is required",
			StatusCode: http.StatusBadRequest}
	}

	url, exists := s.repo.GetOriginalURL(ctx, shortURL)

	// Если запись не найдена, то возвращаем ошибку 404
	if !exists {
		return "", error2.CustomError{
			Message:    fmt.Sprintf("key %s does not exist", shortURL),
			StatusCode: http.StatusNotFound,
		}
	}

	return url, nil
}

// CreateShortURL создание новой пары короткой и исходной ссылки
func (s *urlService) CreateShortURL(ctx context.Context, originalURL string) (string, error) {
	// Проверяем существование оригинальной ссылки в репозитории
	if key, exists := s.repo.GetShortURL(ctx, originalURL); exists {
		return s.cfg.BaseURL + "/" + key, nil
	}

	// Генерируем короткую ссылку заданной в настройках длины и гарантируем её уникальность
	key, err := CreateUniqueStringForURL(
		ctx,
		s.repo,
		s.cfg.RandomStringLength,
		s.cfg.RandomStringMaxLength,
		s.cfg.RandomStringMaxGenerationAttempts)

	// Произошла ошибка генерации
	if err != nil {
		return "", error2.CustomError{
			Message:    "Internal Server Error",
			StatusCode: http.StatusInternalServerError,
		}
	}

	// Сохраняем короткую ссылку в репозитории
	err = s.repo.SaveURL(ctx, key, originalURL)

	if err != nil {
		logger.Log.Error(err.Error())
		return "", error2.CustomError{
			Message:    "Internal Server Error",
			StatusCode: http.StatusInternalServerError,
		}
	}

	return s.cfg.BaseURL + "/" + key, nil
}
