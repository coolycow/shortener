package repository

import (
	"context"
	"errors"
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
func (r *DoubleMapsRepository) AddURL(_ context.Context, originalURL string, key string) (string, error) {
	if (key == "") || (originalURL == "") {
		return "", errors.New("key or originalURL is empty")
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Проверяем существование originalURL и если она есть, то возвращаем существующий ключ
	if existingKey, ok := r.originalToKey[originalURL]; ok {
		return existingKey, nil
	}

	r.originalToKey[originalURL] = key
	r.keyToOriginal[key] = originalURL

	return key, nil

}

// SaveURL сохраняет соответствие между короткой и оригинальной ссылкой
func (r *DoubleMapsRepository) SaveURL(ctx context.Context, originalURL string, key string) (string, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Проверяем существование URL и возвращаем её ключ
	if existingKey, exists := r.originalToKey[originalURL]; exists {
		return existingKey, nil
	}

	// Проверяем существование ключа и возвращаем ошибку
	if _, exists := r.keyToOriginal[key]; exists {
		return "", errors.New("key already exists")
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
			return "", err
		}
	}

	return key, nil
}

// SaveManyURL сохраняет множество пар короткой и оригинальной ссылок
// В массиве URLs происходит замена ключей в случае дублирования исходных URL.
func (r *DoubleMapsRepository) SaveManyURL(ctx context.Context, URLs []model.ShortURL) error {
	for i, u := range URLs {
		resultKey, err := r.AddURL(ctx, u.OriginalURL, u.Key)

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
func (r *DoubleMapsRepository) GetKey(_ context.Context, originalURL string) (string, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	shortURL, exists := r.originalToKey[originalURL]
	return shortURL, exists
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
