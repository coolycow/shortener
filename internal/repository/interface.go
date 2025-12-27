package repository

import (
	"context"

	"github.com/coolycow/shortener/internal/model"
)

// URLRepository определяет интерфейс для работы с хранилищем URL
type URLRepository interface {
	AddURL(ctx context.Context, userID int, originalURL string, key string) (string, bool, error)
	SaveURL(ctx context.Context, userID int, originalURL string, key string) (string, bool, error)
	SaveManyURL(ctx context.Context, userID int, URLs []model.ShortURL) error

	GetOriginalURL(ctx context.Context, userID int, key string) (string, bool)
	GetKey(ctx context.Context, userID int, originalURL string) (string, bool)
	GetManyKeys(ctx context.Context, userID int, URLs []model.ShortURL) ([]model.ShortURL, error)
	GetManyShortURLs(ctx context.Context, userID int) ([]model.ShortURL, error)

	IsKeyExists(ctx context.Context, key string) bool
	GetSize(ctx context.Context) int

	Close() error
	Ping(ctx context.Context) error
	RunMigrations() error

	GetUser(ctx context.Context, userID int) (model.User, error)
	CreateUser(ctx context.Context) (model.User, error)
}
