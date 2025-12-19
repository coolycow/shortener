package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/coolycow/shortener/internal/config"
	error2 "github.com/coolycow/shortener/internal/error"
	"github.com/coolycow/shortener/internal/logger"
	"github.com/coolycow/shortener/internal/model"
	"github.com/coolycow/shortener/internal/repository"
)

// URLService Сервис для работы в Handler
type URLService interface {
	GetOriginalURL(ctx context.Context, key string) (string, error)
	CreateShortURL(ctx context.Context, originalURL string) (string, error)
	CreateManyShortURL(ctx context.Context, URLs []model.ShortURL) error
	GetBaseURL() string
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
func (s *urlService) GetOriginalURL(ctx context.Context, key string) (string, error) {
	// Если key пустой, то возвращаем ошибку 400
	if key == "" {
		return "", error2.CustomError{
			Message:    "key is required",
			StatusCode: http.StatusBadRequest}
	}

	url, exists := s.repo.GetOriginalURL(ctx, key)

	// Если запись не найдена, то возвращаем ошибку 404
	if !exists {
		return "", error2.CustomError{
			Message:    fmt.Sprintf("key %s does not exist", key),
			StatusCode: http.StatusNotFound,
		}
	}

	return url, nil
}

// CreateShortURL создание новой пары короткой и исходной ссылки
func (s *urlService) CreateShortURL(ctx context.Context, originalURL string) (string, error) {
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

	// Сохраняем короткую ссылку в репозитории.
	// В случае дублирования originalURL в момент сохранения будет использован уже существующий ключ.
	resultKey, err := s.repo.SaveURL(ctx, originalURL, key)

	if err != nil {
		logger.Log.Error(err.Error())
		return "", error2.CustomError{
			Message:    "Internal Server Error",
			StatusCode: http.StatusInternalServerError,
		}
	}

	if key != resultKey {
		return "", error2.CustomError{
			Message:    s.cfg.BaseURL + "/" + resultKey,
			StatusCode: http.StatusConflict,
		}
	}

	return s.cfg.BaseURL + "/" + resultKey, nil
}

// CreateManyShortURL создание множества пар.
// В массиве URLs происходит замена ключей в случае дублирования исходных URL.
func (s *urlService) CreateManyShortURL(ctx context.Context, URLs []model.ShortURL) error {
	for i, u := range URLs {
		// Проверяем существование ключа для URL.
		if key, exists := s.repo.GetKey(ctx, u.OriginalURL); exists {
			URLs[i].Key = key
			continue
		}

		// Генерируем ключи для несуществующих URL.
		key, err := CreateUniqueStringForURL(
			ctx,
			s.repo,
			s.cfg.RandomStringLength,
			s.cfg.RandomStringMaxLength,
			s.cfg.RandomStringMaxGenerationAttempts)

		if err != nil {
			return error2.CustomError{
				Message:    "Internal Server Error",
				StatusCode: http.StatusInternalServerError,
			}
		}

		// Присваиваем URL её ключ.
		URLs[i].Key = key
	}

	return s.repo.SaveManyURL(ctx, URLs)
}

// GetBaseURL просто возвращает базовый URL
func (s *urlService) GetBaseURL() string {
	return s.cfg.BaseURL
}
