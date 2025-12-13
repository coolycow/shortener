package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/coolycow/shortener/internal/logger"
	"github.com/coolycow/shortener/internal/model"
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
func (r *PostgresRepository) AddURL(ctx context.Context, originalURL string, key string) (string, error) {
	return r.SaveURL(ctx, originalURL, key)
}

// SaveURL сохраняет соответствие между короткой и оригинальной ссылкой
func (r *PostgresRepository) SaveURL(ctx context.Context, originalURL string, key string) (string, error) {
	if (key == "") || (originalURL == "") {
		return "", errors.New("key or originalURL is empty")
	}

	if len(key) > 255 {
		return "", errors.New("key is too long")
	}

	// Учитываем, что может произойти дублирование URL, поэтому мы возвращаем ключ который реально был использован.
	row := r.db.QueryRowContext(ctx,
		"INSERT INTO urls (url, key) VALUES ($1, $2) ON CONFLICT (url) DO UPDATE SET key = urls.key RETURNING key",
		originalURL, key)

	var resultKey string
	err := row.Scan(&resultKey)

	if err != nil {
		return "", err
	}

	return resultKey, nil
}

// SaveManyURL сохраняет множество пар короткой и оригинальной ссылок
// В массиве URLs происходит замена ключей в случае дублирования исходных URL.
func (r *PostgresRepository) SaveManyURL(ctx context.Context, URLs []model.ShortURL) error {
	tx, err := r.db.BeginTx(ctx, nil)

	if err != nil {
		return err
	}

	// Подготавливаем запрос
	stmt, err := tx.PrepareContext(ctx,
		"INSERT INTO urls (url, key) VALUES ($1, $2) ON CONFLICT (url) DO UPDATE SET key = urls.key RETURNING key")

	if err != nil {
		return err
	}

	defer stmt.Close()

	// В рамках транзакции проводим сохранение пар URL и ключа
	// Если в момент сохранения произойдет дублирование URL, то мы это учтём.
	for i, u := range URLs {
		if (u.Key == "") || (u.OriginalURL == "") {
			tx.Rollback()
			return errors.New("key or originalURL is empty")
		}

		var resultKey string
		row := stmt.QueryRowContext(ctx, u.OriginalURL, u.Key)
		err = row.Scan(&resultKey)

		if err != nil {
			logger.Log.Warn("Error saving url", zap.Error(err))
			tx.Rollback()
			return err
		}

		URLs[i].Key = resultKey
	}

	return tx.Commit()
}

// GetOriginalURL получает оригинальный URL по ключу
func (r *PostgresRepository) GetOriginalURL(ctx context.Context, key string) (string, bool) {
	row := r.db.QueryRowContext(ctx, "select url from urls where key = $1", key)

	var url string
	err := row.Scan(&url)

	if err != nil {
		return "", false
	}

	return url, true
}

// GetKey получает ключ по оригинальному URL
func (r *PostgresRepository) GetKey(ctx context.Context, originalURL string) (string, bool) {
	row := r.db.QueryRowContext(ctx, "select key from urls where url = $1", originalURL)

	var key string
	err := row.Scan(&key)

	if err != nil {
		return "", false
	}

	return key, true
}

// IsKeyExists проверяет, существует ли ключ
func (r *PostgresRepository) IsKeyExists(ctx context.Context, key string) bool {
	row := r.db.QueryRowContext(ctx, "select count(*) from urls where key = $1", key)

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
