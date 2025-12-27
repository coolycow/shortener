package repository

import (
	"context"
	"errors"
	"math/rand"
	"strconv"
	"sync"

	"github.com/coolycow/shortener/internal/logger"
	"github.com/coolycow/shortener/internal/model"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DoubleMapsRepository представляет репозиторий для хранения URL
type DoubleMapsRepository struct {
	keyToOriginal map[string]string
	originalToKey map[string]string
	mutex         sync.RWMutex
	jsonStorage   *JSONStorage
}

func (r *DoubleMapsRepository) RunMigrations() error {
	return nil
}

// NewDoubleMapsRepository NewURLRepository создает новый экземпляр URLRepository
func NewDoubleMapsRepository(filename string) *DoubleMapsRepository {
	repository := &DoubleMapsRepository{
		keyToOriginal: make(map[string]string),
		originalToKey: make(map[string]string),
	}

	jsonStorage, err := NewJSONStorage(filename)

	if err != nil {
		logger.Log.Warn("Error creating json storage", zap.Error(err))
		return repository
	}

	repository.jsonStorage = jsonStorage

	if err = repository.jsonStorage.Load(repository); err != nil {
		logger.Log.Fatal("Error loading json storage", zap.Error(err))
		return repository
	}

	logger.Log.Info("Load " + filename + " successfully (" +
		strconv.Itoa(repository.GetSize(context.Background())) + ")")

	return repository
}

// AddURL сохраняет соответствие между короткой и оригинальной ссылкой
func (r *DoubleMapsRepository) AddURL(_ context.Context, userID int, originalURL string, key string) (string, bool, error) {
	if (key == "") || (originalURL == "") {
		return "", false, errors.New("key or originalURL is empty")
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Проверяем существование originalURL и если она есть, то возвращаем существующий ключ
	if existingKey, ok := r.originalToKey[originalURL]; ok {
		return existingKey, false, nil
	}

	r.originalToKey[originalURL] = key
	r.keyToOriginal[key] = originalURL

	return key, true, nil

}

// SaveURL сохраняет соответствие между короткой и оригинальной ссылкой
func (r *DoubleMapsRepository) SaveURL(_ context.Context, userID int, originalURL string, key string) (string, bool, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Проверяем существование URL и возвращаем её ключ
	if existingKey, exists := r.originalToKey[originalURL]; exists {
		return existingKey, true, nil
	}

	// Проверяем существование ключа и возвращаем ошибку
	if _, exists := r.keyToOriginal[key]; exists {
		return "", false, errors.New("key already exists")
	}

	// Создаём пары
	r.originalToKey[originalURL] = key
	r.keyToOriginal[key] = originalURL

	// Если есть файловое хранилище, то дублируем в него
	if r.jsonStorage != nil {
		err := r.jsonStorage.Write(model.ShortURL{
			CorrelationID: uuid.New().String(),
			Key:           key,
			OriginalURL:   originalURL,
		})

		if err != nil {
			logger.Log.Warn("Error saving url", zap.Error(err))
			return "", false, err
		}
	}

	return key, false, nil
}

// SaveManyURL сохраняет множество пар короткой и оригинальной ссылок
// В массиве URLs происходит замена ключей в случае дублирования исходных URL.
func (r *DoubleMapsRepository) SaveManyURL(ctx context.Context, userID int, URLs []model.ShortURL) error {
	for i, u := range URLs {
		resultKey, _, err := r.AddURL(ctx, userID, u.OriginalURL, u.Key)

		if err != nil {
			logger.Log.Warn("Error saving url", zap.Error(err))
			return err
		}

		URLs[i].Key = resultKey

		if r.jsonStorage != nil {
			err = r.jsonStorage.Write(URLs[i])

			if err != nil {
				logger.Log.Warn("Error saving url", zap.Error(err))
				return err
			}
		}
	}

	return nil
}

// GetOriginalURL получает оригинальный URL по ключу
func (r *DoubleMapsRepository) GetOriginalURL(_ context.Context, key string) (string, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	originalURL, exists := r.keyToOriginal[key]
	return originalURL, exists
}

// GetKey получает ключ по оригинальному URL
func (r *DoubleMapsRepository) GetKey(_ context.Context, userID int, originalURL string) (string, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	shortURL, exists := r.originalToKey[originalURL]
	return shortURL, exists
}

// GetManyKeys получает массив найденных ShortURL по массиву исходных ShortURL
func (r *DoubleMapsRepository) GetManyKeys(ctx context.Context, userID int, URLs []model.ShortURL) ([]model.ShortURL, error) {
	// Если массив пустой, то просто возвращаем пустой результат
	if len(URLs) == 0 {
		return []model.ShortURL{}, nil
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var result []model.ShortURL
	for _, u := range URLs {
		if shortURL, exists := r.originalToKey[u.OriginalURL]; exists {
			result = append(result, model.ShortURL{
				CorrelationID: u.CorrelationID,
				OriginalURL:   u.OriginalURL,
				Key:           shortURL,
			})
		}
	}

	return result, nil
}

// GetManyShortURLs возвращает все сокращенные URL
func (r *DoubleMapsRepository) GetManyShortURLs(ctx context.Context, userID int) ([]model.ShortURL, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var result []model.ShortURL
	for u, k := range r.originalToKey {
		result = append(result, model.ShortURL{
			CorrelationID: "",
			OriginalURL:   u,
			Key:           k,
		})
	}

	return result, nil
}

// IsKeyExists проверяет, существует ли ключ
func (r *DoubleMapsRepository) IsKeyExists(_ context.Context, key string) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	_, exists := r.keyToOriginal[key]
	return exists
}

// GetSize возвращает размер хранилища
func (r *DoubleMapsRepository) GetSize(_ context.Context) int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return len(r.keyToOriginal)
}

// Close закрывает хранилище
func (r *DoubleMapsRepository) Close() error {
	if r.jsonStorage != nil {
		return r.jsonStorage.Close()
	}
	return nil
}

// Ping проверяет доступность хранилища
func (r *DoubleMapsRepository) Ping(_ context.Context) error {
	return nil
}

// GetUser возвращает пользователя по его ID
func (r *DoubleMapsRepository) GetUser(_ context.Context, userID int) (model.User, error) {
	return model.User{ID: userID}, nil
}

func (r *DoubleMapsRepository) CreateUser(_ context.Context) (model.User, error) {
	return model.User{ID: rand.Int()}, nil
}
