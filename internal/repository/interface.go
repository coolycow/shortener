package repository

import (
	"context"

	"github.com/coolycow/shortener/internal/model"
)

// URLRepository определяет интерфейс для работы с хранилищем URL
type URLRepository interface {
	AddURL(ctx context.Context, originalURL string, key string) (string, bool, error)
	SaveURL(ctx context.Context, originalURL string, key string) (string, bool, error)
	SaveManyURL(ctx context.Context, URLs []model.ShortURL) error
	GetOriginalURL(ctx context.Context, key string) (string, bool)
	GetKey(ctx context.Context, originalURL string) (string, bool)
	GetManyKeys(ctx context.Context, URLs []model.ShortURL) ([]model.ShortURL, error)
	IsKeyExists(ctx context.Context, key string) bool
	GetSize(ctx context.Context) int
	Close() error
	Ping(ctx context.Context) error
	RunMigrations() error
}
