package repository

// DoubleMapsRepository представляет репозиторий для хранения URL
type DoubleMapsRepository struct {
	shortToOriginal map[string]string
	originalToShort map[string]string
}

// NewDoubleMapsRepository NewURLRepository создает новый экземпляр URLRepository
func NewDoubleMapsRepository() *DoubleMapsRepository {
	return &DoubleMapsRepository{
		shortToOriginal: make(map[string]string),
		originalToShort: make(map[string]string),
	}
}

// SaveURL сохраняет соответствие между короткой и оригинальной ссылкой
func (r *DoubleMapsRepository) SaveURL(shortURL string, originalURL string) {
	r.shortToOriginal[shortURL] = originalURL
	r.originalToShort[originalURL] = shortURL
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
