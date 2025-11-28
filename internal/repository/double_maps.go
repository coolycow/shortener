package repository

import (
	"strconv"

	"github.com/coolycow/shortener/internal/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DoubleMapsRepository представляет репозиторий для хранения URL
type DoubleMapsRepository struct {
	shortToOriginal map[string]string
	originalToShort map[string]string
	jsonStorage     *JSONStorage
}

// NewDoubleMapsRepository NewURLRepository создает новый экземпляр URLRepository
func NewDoubleMapsRepository(filename string) *DoubleMapsRepository {
	repository := &DoubleMapsRepository{
		shortToOriginal: make(map[string]string),
		originalToShort: make(map[string]string),
	}

	jsonStorage, err := NewJSONStorage(filename)

	if err != nil {
		logger.Log.Warn("Error creating json storage", zap.Error(err))
		return repository
	}

	repository.jsonStorage = jsonStorage

	if err = repository.jsonStorage.Load(repository); err != nil {
		logger.Log.Warn("Error loading json storage", zap.Error(err))
		return repository
	}

	logger.Log.Info("Load " + filename + " successfully (" + strconv.Itoa(repository.GetSize()) + ")")

	return repository
}

// AddURL сохраняет соответствие между короткой и оригинальной ссылкой
func (r *DoubleMapsRepository) AddURL(shortURL string, originalURL string) {
	r.shortToOriginal[shortURL] = originalURL
	r.originalToShort[originalURL] = shortURL
}

// SaveURL сохраняет соответствие между короткой и оригинальной ссылкой
func (r *DoubleMapsRepository) SaveURL(shortURL string, originalURL string) {
	r.AddURL(shortURL, originalURL)

	err := r.jsonStorage.Write(ShortURL{
		UUID:        uuid.New().String(),
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	})

	if err != nil {
		logger.Log.Warn("Error saving url", zap.Error(err))
		return
	}
}

// GetOriginalURL получает оригинальный URL по короткому
func (r *DoubleMapsRepository) GetOriginalURL(shortURL string) (string, bool) {
	originalURL, exists := r.shortToOriginal[shortURL]
	return originalURL, exists
}

// GetShortURL получает короткий URL по оригинальному
func (r *DoubleMapsRepository) GetShortURL(originalURL string) (string, bool) {
	shortURL, exists := r.originalToShort[originalURL]
	return shortURL, exists
}

// IsShortURLExists проверяет, существует ли короткий URL
func (r *DoubleMapsRepository) IsShortURLExists(shortURL string) bool {
	_, exists := r.shortToOriginal[shortURL]
	return exists
}

// GetSize возвращает размер хранилища
func (r *DoubleMapsRepository) GetSize() int {
	return len(r.shortToOriginal)
}

func (r *DoubleMapsRepository) Close() error {
	if r.jsonStorage != nil {
		return r.jsonStorage.Close()
	}
	return nil
}
