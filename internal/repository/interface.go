package repository

import (
	"context"

	"github.com/coolycow/shortener/internal/model"
)

// URLRepository — интерфейс хранилища: сохранение/получение коротких URL, пользователи, пинг, миграции.
type URLRepository interface {
	AddURL(ctx context.Context, userID string, originalURL string, key string) (string, bool, error)
	SaveURL(ctx context.Context, userID string, originalURL string, key string) (string, bool, error)
	SaveManyURL(ctx context.Context, userID string, URLs []model.ShortURL) error

	GetShortURL(ctx context.Context, key string) (model.ShortURL, bool)
	GetKey(ctx context.Context, userID string, originalURL string) (string, bool)
	GetManyKeys(ctx context.Context, userID string, URLs []model.ShortURL) ([]model.ShortURL, error)
	GetManyShortURLs(ctx context.Context, userID string) ([]model.ShortURL, error)

	IsKeyExists(ctx context.Context, key string) bool
	GetSize(ctx context.Context) int

	Close() error
	Ping(ctx context.Context) error
	RunMigrations() error

	GetUser(ctx context.Context, userID string) (model.User, error)
	CreateUser(ctx context.Context) (model.User, error)

	DeleteManyURLs(ctx context.Context, userID string, keys []string) error
}
