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
	GetOriginalURL(ctx context.Context, userID string, key string) (string, error)
	GetManyShortURLs(ctx context.Context, userID string) ([]model.ShortURL, error)

	CreateShortURL(ctx context.Context, userID string, originalURL string) (string, error)
	CreateManyShortURL(ctx context.Context, userID string, URLs []model.ShortURL) error

	GetBaseURL() string
	PingRepository(ctx context.Context) error
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

// GetOriginalURL возвращает исходную ссылку пользователя по короткому ключу
func (s *urlService) GetOriginalURL(ctx context.Context, userID string, key string) (string, error) {
	// Если key пустой, то возвращаем ошибку 400
	if key == "" {
		return "", error2.CustomError{
			Message:    "key is required",
			StatusCode: http.StatusBadRequest}
	}

	url, exists := s.repo.GetOriginalURL(ctx, userID, key)

	// Если запись не найдена, то возвращаем ошибку 404
	if !exists {
		return "", error2.CustomError{
			Message:    fmt.Sprintf("key %s does not exist", key),
			StatusCode: http.StatusNotFound,
		}
	}

	return url, nil
}

// GetManyOriginalURL возвращает все ссылки когда-либо сокращенные пользователем
func (s *urlService) GetManyShortURLs(ctx context.Context, userID string) ([]model.ShortURL, error) {
	urls, err := s.repo.GetManyShortURLs(ctx, userID)

	if err != nil {
		return []model.ShortURL{}, error2.CustomError{
			Message:    "Internal Server Error",
			StatusCode: http.StatusInternalServerError,
		}
	}

	return urls, nil
}

// CreateShortURL создание новой пары короткой и исходной ссылки для пользователя
func (s *urlService) CreateShortURL(ctx context.Context, userID string, originalURL string) (string, error) {
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
	// hasConflict показывает было ли реальное добавление или ссылка уже была в репозитории
	resultKey, hasConflict, err := s.repo.SaveURL(ctx, userID, originalURL, key)

	if err != nil {
		logger.Log.Error(err.Error())
		return "", error2.CustomError{
			Message:    "Internal Server Error",
			StatusCode: http.StatusInternalServerError,
		}
	}

	if hasConflict {
		return "", error2.CustomError{
			Message:    s.cfg.BaseURL + "/" + resultKey,
			StatusCode: http.StatusConflict,
		}
	}

	return s.cfg.BaseURL + "/" + resultKey, nil
}

// CreateManyShortURL создание множества пар для пользователя.
// В массиве URLs происходит замена ключей в случае дублирования исходных URL.
func (s *urlService) CreateManyShortURL(ctx context.Context, userID string, URLs []model.ShortURL) error {
	// Получаем все уже существующие ключи для переданного массива ShortURL
	exists, err := s.repo.GetManyKeys(ctx, userID, URLs)

	if err != nil {
		return error2.CustomError{
			Message:    "Internal Server Error",
			StatusCode: http.StatusInternalServerError,
		}
	}

	// Создаем map с ключом в виде OriginalURL для быстрого поиска существующих URL
	existingURLs := make(map[string]string, len(exists))

	for _, existing := range exists {
		existingURLs[existing.OriginalURL] = existing.Key
	}

	// Разделяем URL на существующие и новые в виде массива newURLs
	var newURLs []model.ShortURL
	for i, u := range URLs {
		if key, found := existingURLs[u.OriginalURL]; found {
			URLs[i].Key = key
		} else {
			newURLs = append(newURLs, u)
		}
	}

	for i := range newURLs {
		// Генерируем ключ для URL.
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

		// Находим индекс в исходном массиве и обновляем ключ
		for j := range URLs {
			if URLs[j].OriginalURL == newURLs[i].OriginalURL && URLs[j].Key == "" {
				URLs[j].Key = key
				break
			}
		}
	}

	return s.repo.SaveManyURL(ctx, userID, URLs)
}

// GetBaseURL просто возвращает базовый URL
func (s *urlService) GetBaseURL() string {
	return s.cfg.BaseURL
}

// PingRepository проверяет доступность репозитория
func (s *urlService) PingRepository(ctx context.Context) error {
	return s.repo.Ping(ctx)
}
