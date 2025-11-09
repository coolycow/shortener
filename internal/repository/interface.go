package repository

// URLRepository определяет интерфейс для работы с хранилищем URL
type URLRepository interface {
	SaveURL(shortURL, originalURL string)
	GetOriginalURL(shortURL string) (string, bool)
	GetShortURL(originalURL string) (string, bool)
	IsShortURLExists(shortURL string) bool
	GetSize() int
}
