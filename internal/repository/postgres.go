package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/coolycow/shortener/internal/logger"
	"go.uber.org/zap"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// PostgresRepository представляет репозиторий для хранения URL
type PostgresRepository struct {
	db *sql.DB
}

func runMigrations(db *sql.DB) error {
	// Создаем экземпляр драйвера для PostgreSQL
	driver, err := pgx.WithInstance(db, &pgx.Config{})
	if err != nil {
		return err
	}

	// Указываем путь к директории с миграциями
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"pgx",
		driver,
	)
	if err != nil {
		return err
	}

	// Применяем миграции
	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

// NewPostgresRepository создает новый экземпляр URLRepository
func NewPostgresRepository(DSN string) (*PostgresRepository, error) {
	db, err := sql.Open("pgx", DSN)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		err = db.Close()
		return nil, err
	}

	if err = runMigrations(db); err != nil {
		err = db.Close()
		return nil, err
	}

	return &PostgresRepository{db: db}, nil
}

// AddURL сохраняет соответствие между короткой и оригинальной ссылкой
func (r *PostgresRepository) AddURL(ctx context.Context, shortURL string, originalURL string) error {
	return r.SaveURL(ctx, shortURL, originalURL)
}

// SaveURL сохраняет соответствие между короткой и оригинальной ссылкой
func (r *PostgresRepository) SaveURL(ctx context.Context, shortURL string, originalURL string) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO urls (url, key) VALUES ($1, $2)", originalURL, shortURL)

	if err != nil {
		return err
	}

	return nil
}

// GetOriginalURL получает оригинальный URL по короткому
func (r *PostgresRepository) GetOriginalURL(ctx context.Context, shortURL string) (string, bool) {
	row := r.db.QueryRowContext(ctx, "select url from urls where key = $1", shortURL)

	var url string
	err := row.Scan(&url)

	if err != nil {
		return "", false
	}

	return url, true
}

// GetShortURL получает короткий URL по оригинальному
func (r *PostgresRepository) GetShortURL(ctx context.Context, originalURL string) (string, bool) {
	row := r.db.QueryRowContext(ctx, "select key from urls where url = $1", originalURL)

	var key string
	err := row.Scan(&key)

	if err != nil {
		return "", false
	}

	return key, true
}

// IsShortURLExists проверяет, существует ли короткий URL
func (r *PostgresRepository) IsShortURLExists(ctx context.Context, shortURL string) bool {
	row := r.db.QueryRowContext(ctx, "select count(*) from urls where key = $1", shortURL)

	var count int64
	err := row.Scan(&count)

	if err != nil {
		return false
	}

	return count > 0
}

// GetSize возвращает размер хранилища
func (r *PostgresRepository) GetSize(ctx context.Context) int {
	row := r.db.QueryRowContext(ctx, "select count(*) from urls")

	count := 0
	err := row.Scan(&count)

	if err != nil {
		logger.Log.Error("Error scanning database", zap.Error(err))
	}

	return count
}

// Close закрывает хранилище
func (r *PostgresRepository) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}
