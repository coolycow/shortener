package repository

import "context"

// URLRepository определяет интерфейс для работы с хранилищем URL
type URLRepository interface {
	AddURL(ctx context.Context, shortURL, originalURL string) error
	SaveURL(ctx context.Context, shortURL, originalURL string) error
	GetOriginalURL(ctx context.Context, shortURL string) (string, bool)
	GetShortURL(ctx context.Context, originalURL string) (string, bool)
	IsShortURLExists(ctx context.Context, shortURL string) bool
	GetSize(ctx context.Context) int
	Close() error
}
