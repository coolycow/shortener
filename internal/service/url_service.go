// Package service содержит бизнес-логику сервиса коротких ссылок и пользователей.
package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"sync"

	"github.com/coolycow/shortener/internal/config"
	httpError "github.com/coolycow/shortener/internal/error"
	"github.com/coolycow/shortener/internal/logger"
	"github.com/coolycow/shortener/internal/model"
	"github.com/coolycow/shortener/internal/repository"
	"go.uber.org/zap"
)

// URLService — интерфейс бизнес-логики: создание и получение коротких ссылок, пинг хранилища, удаление.
type URLService interface {
	GetShortURL(ctx context.Context, key string) (model.ShortURL, error)
	GetManyShortURLs(ctx context.Context, userID string) ([]model.ShortURL, error)

	CreateShortURL(ctx context.Context, userID string, originalURL string) (string, error)
	CreateManyShortURL(ctx context.Context, userID string, URLs []model.ShortURL) error

	GetBaseURL() string
	PingRepository(ctx context.Context) error

	DeleteManyURLs(ctx context.Context, userID string, keys []string) error
}

// Реализация сервисного слоя
type urlService struct {
	repo repository.URLRepository
	cfg  *config.Config
}

// NewURLService создаёт реализацию URLService для использования в хендлерах.
func NewURLService(cfg *config.Config, repo repository.URLRepository) URLService {
	return &urlService{
		repo: repo,
		cfg:  cfg,
	}
}

// GetShortURL возвращает исходную ссылку пользователя по короткому ключу
func (s *urlService) GetShortURL(ctx context.Context, key string) (model.ShortURL, error) {
	// Если key пустой, то возвращаем ошибку 400
	if key == "" {
		return model.ShortURL{}, httpError.CustomError{
			Message:    "key is required",
			StatusCode: http.StatusBadRequest}
	}

	url, exists := s.repo.GetShortURL(ctx, key)

	// Если запись не найдена, то возвращаем ошибку 404
	if !exists {
		return model.ShortURL{}, httpError.CustomError{
			Message:    fmt.Sprintf("key %s does not exist", key),
			StatusCode: http.StatusNotFound,
		}
	}

	return url, nil
}

// GetManyShortURLs возвращает все ссылки когда-либо сокращенные пользователем
func (s *urlService) GetManyShortURLs(ctx context.Context, userID string) ([]model.ShortURL, error) {
	urls, err := s.repo.GetManyShortURLs(ctx, userID)

	if err != nil {
		return []model.ShortURL{}, httpError.CustomError{
			Message:    http.StatusText(http.StatusInternalServerError),
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
		return "", httpError.CustomError{
			Message:    http.StatusText(http.StatusInternalServerError),
			StatusCode: http.StatusInternalServerError,
		}
	}

	// Сохраняем короткую ссылку в репозитории.
	// В случае дублирования originalURL в момент сохранения будет использован уже существующий ключ.
	// hasConflict показывает было ли реальное добавление или ссылка уже была в репозитории
	resultKey, hasConflict, err := s.repo.SaveURL(ctx, userID, originalURL, key)

	if err != nil {
		logger.Log.Error(err.Error())
		return "", httpError.CustomError{
			Message:    http.StatusText(http.StatusInternalServerError),
			StatusCode: http.StatusInternalServerError,
		}
	}

	if hasConflict {
		return "", httpError.CustomError{
			Message:    buildFullURL(s.cfg.BaseURL, resultKey),
			StatusCode: http.StatusConflict,
		}
	}

	// Оптимизированный вариант
	return buildFullURL(s.cfg.BaseURL, resultKey), nil
}

// CreateManyShortURL создание множества пар для пользователя.
// В массиве URLs происходит замена ключей в случае дублирования исходных URL.
func (s *urlService) CreateManyShortURL(ctx context.Context, userID string, URLs []model.ShortURL) error {
	// Получаем все уже существующие ключи для переданного массива ShortURL
	exists, err := s.repo.GetManyKeys(ctx, userID, URLs)

	if err != nil {
		return httpError.CustomError{
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		}
	}

	// Создаем map с ключом в виде OriginalURL для быстрого поиска существующих URL
	existingURLs := make(map[string]string, len(exists))

	for _, existing := range exists {
		existingURLs[existing.OriginalURL] = existing.Key
	}

	// Разделяем URL на существующие и новые в виде массива newURLs
	// Оптимизированный вариант
	newURLs := make([]model.ShortURL, 0, len(URLs))

	for i, u := range URLs {
		if key, found := existingURLs[u.OriginalURL]; found {
			URLs[i].Key = key
		} else {
			newURLs = append(newURLs, u)
		}
	}

	// Оптимизированный вариант
	urlIndexByOriginal := make(map[string]int, len(URLs))
	for idx := range URLs {
		urlIndexByOriginal[URLs[idx].OriginalURL] = idx
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
			return httpError.CustomError{
				Message:    http.StatusText(http.StatusInternalServerError),
				StatusCode: http.StatusInternalServerError,
			}
		}

		// Оптимизированный вариант
		idx := urlIndexByOriginal[newURLs[i].OriginalURL]
		URLs[idx].Key = key
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

// DeleteManyURLs удаление множества URL
func (s *urlService) DeleteManyURLs(ctx context.Context, userID string, keys []string) error {
	// Если массив пустой, то просто возвращаем пустой результат
	if len(keys) == 0 {
		return nil
	}

	// Логируем начало выполнения операции удаления
	logger.Log.Info("Starting deletion of URLs", zap.String("userID", userID), zap.Int("count", len(keys)))

	// Размер порции для батчинга
	const batchSize = 100

	// Создаем канал для задач удаления
	size := len(keys)/batchSize + 1
	jobs := make(chan []string, size)

	// Создаем канал для результатов с использованием паттерна fan-in
	results := make(chan error, size)

	// Количество воркеров
	numWorkers := runtime.NumCPU()

	// Запускаем воркеры, которые будут обрабатывать порции ключей
	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for batch := range jobs {
				logger.Log.Info("Processing batch for deletion", zap.Strings("keys", batch))

				// Передаем порцию ключей в репозиторий для удаления
				err := s.repo.DeleteManyURLs(ctx, userID, batch)

				if err != nil {
					logger.Log.Error("Error deleting URLs batch", zap.Error(err))
				} else {
					logger.Log.Info("Successfully deleted batch", zap.Strings("keys", batch))
				}

				results <- err
			}
		}()
	}

	// Добавляем еще одну WaitGroup для отправителя
	var senderWg sync.WaitGroup

	// Перед запуском горутины-отправителя увеличиваем счетчик
	senderWg.Add(1)

	// Разбиваем массив ключей на порции и отправляем в канал jobs
	go func() {
		defer close(jobs)

		defer senderWg.Done() // Уменьшаем счетчик при завершении

		for i := 0; i < len(keys); i += batchSize {
			end := i + batchSize
			if end > len(keys) {
				end = len(keys)
			}

			batch := make([]string, end-i)
			copy(batch, keys[i:end])
			jobs <- batch
		}
	}()

	// Ждем завершения всех воркеров и закрываем канал результатов
	go func() {
		wg.Wait()
		close(results)
	}()

	// Собираем результаты и проверяем на наличие ошибок
	var hasErrors bool
	for err := range results {
		if err != nil {
			// Логируем ошибку, но не прерываем процесс
			logger.Log.Error("Error deleting URLs batch", zap.Error(err))
			hasErrors = true
		}
	}

	// Ждем завершения отправителя и всех воркеров
	senderWg.Wait()
	wg.Wait()

	if hasErrors {
		return errors.New("some URLs deletion failed")
	}

	logger.Log.Info("All deletion tasks completed")
	return nil
}

// buildFullURL собирает BaseURL + "/" + key без лишних аллокаций.
func buildFullURL(baseURL, key string) string {
	n := len(baseURL) + 1 + len(key)
	var sb strings.Builder
	sb.Grow(n)
	sb.WriteString(baseURL)
	sb.WriteByte('/')
	sb.WriteString(key)
	return sb.String()
}
