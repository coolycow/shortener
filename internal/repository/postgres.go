package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

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

// checkTableExists проверяет, существует ли таблица
func (r *PostgresRepository) checkTableExists(tableName string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = $1
		)`, tableName).Scan(&exists)
	return exists, err
}

func (r *PostgresRepository) RunMigrations() error {
	logger.Log.Info("Running migrations")

	// Создаем экземпляр драйвера для PostgreSQL
	driver, err := pgx.WithInstance(r.db, &pgx.Config{})
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
		closeErr := db.Close()
		if closeErr != nil {
			log.Printf("Error closing database: %v", closeErr)
		}
		return nil, err
	}

	repo := &PostgresRepository{db: db}

	// Проверяем, существует ли таблица urls
	tableExists, err := repo.checkTableExists("urls")
	if err != nil {
		return nil, err
	}

	if !tableExists {
		err = repo.RunMigrations()

		if err != nil {
			return nil, err
		}
	}

	return repo, nil
}

// AddURL сохраняет соответствие между короткой и оригинальной ссылкой
func (r *PostgresRepository) AddURL(ctx context.Context, userID string, originalURL string, key string) (string, bool, error) {
	return r.SaveURL(ctx, userID, originalURL, key)
}

// SaveURL сохраняет соответствие между короткой и оригинальной ссылкой.
// Возвращает реальный ключ, логический признак ошибки вставки (дублируется исходная URL), ошибку работы.
func (r *PostgresRepository) SaveURL(ctx context.Context, userID string, originalURL string, key string) (string, bool, error) {
	if (key == "") || (originalURL == "") {
		return "", false, errors.New("key or originalURL is empty")
	}

	if len(key) > 255 {
		return "", false, errors.New("key is too long")
	}

	// Учитываем, что может произойти дублирование URL, поэтому мы возвращаем ключ который реально был использован.
	// Дополнительно вернётся значение is_insert, которое будет равно true если это была новая вставка.
	row := r.db.QueryRowContext(ctx,
		"INSERT INTO urls (url, key, user_id) VALUES ($1, $2, $3) ON CONFLICT (url, user_id) DO UPDATE SET key = urls.key, deleted_at = NULL RETURNING key, (xmax = 0) as is_insert",
		originalURL, key, userID)

	var resultKey string
	var isInsert bool
	err := row.Scan(&resultKey, &isInsert)

	if err != nil {
		return "", false, err
	}

	return resultKey, !isInsert, nil
}

// SaveManyURL сохраняет множество пар короткой и оригинальной ссылок
// В массиве URLs происходит замена ключей в случае дублирования исходных URL.
func (r *PostgresRepository) SaveManyURL(ctx context.Context, userID string, URLs []model.ShortURL) error {
	tx, err := r.db.BeginTx(ctx, nil)

	if err != nil {
		return err
	}

	// Подготавливаем запрос
	stmt, err := tx.PrepareContext(ctx,
		"INSERT INTO urls (url, key, user_id) VALUES ($1, $2, $3) ON CONFLICT (url, user_id) DO UPDATE SET key = urls.key, deleted_at = NULL RETURNING key")

	if err != nil {
		return err
	}

	defer func() { _ = stmt.Close() }()

	// В рамках транзакции проводим сохранение пар URL и ключа
	// Если в момент сохранения произойдет дублирование URL, то мы это учтём.
	for i, u := range URLs {
		if (u.Key == "") || (u.OriginalURL == "") {
			_ = tx.Rollback()
			log.Println(u.Key)
			log.Println(u.OriginalURL)
			return errors.New("key or originalURL is empty")
		}

		var resultKey string
		row := stmt.QueryRowContext(ctx, u.OriginalURL, u.Key, userID)
		err = row.Scan(&resultKey)

		if err != nil {
			logger.Log.Warn("Error saving url", zap.Error(err))
			_ = tx.Rollback()
			return err
		}

		URLs[i].Key = resultKey
	}

	return tx.Commit()
}

// GetShortURL получает оригинальный URL по ключу
func (r *PostgresRepository) GetShortURL(ctx context.Context, key string) (model.ShortURL, bool) {
	row := r.db.QueryRowContext(ctx, "select id, user_id, url, deleted_at from urls where key = $1", key)

	var id string
	var userID string
	var url string
	var deletedAt *time.Time
	err := row.Scan(&id, &userID, &url, &deletedAt)

	if err != nil {
		fmt.Println(err.Error())
		return model.ShortURL{}, false
	}

	return model.ShortURL{
		CorrelationID: id,
		UserID:        userID,
		Key:           key,
		OriginalURL:   url,
		DeletedAt:     deletedAt,
	}, true
}

// GetKey получает ключ по оригинальному URL
func (r *PostgresRepository) GetKey(ctx context.Context, userID string, originalURL string) (string, bool) {
	row := r.db.QueryRowContext(ctx, "select key from urls where user_id = $1 AND url = $2", userID, originalURL)

	var key string
	err := row.Scan(&key)

	if err != nil {
		return "", false
	}

	return key, true
}

// GetManyKeys получает массив найденных ShortURL по массиву исходных ShortURL
func (r *PostgresRepository) GetManyKeys(ctx context.Context, userID string, URLs []model.ShortURL) ([]model.ShortURL, error) {
	// Если массив пустой, то просто возвращаем пустой результат
	if len(URLs) == 0 {
		return []model.ShortURL{}, nil
	}

	// Извлекаем только OriginalURL из входного массива структур
	originalURLs := make([]string, 0, len(URLs))
	for _, u := range URLs {
		if u.OriginalURL != "" {
			originalURLs = append(originalURLs, u.OriginalURL)
		}
	}

	if len(originalURLs) == 0 {
		return []model.ShortURL{}, nil
	}

	// Подготавливаем данные для запроса
	placeholders := make([]string, 0, len(originalURLs))
	args := make([]interface{}, 0, len(originalURLs)+1) // +1 для userID

	// Добавляем userID как первый параметр
	args = append(args, userID)

	for i, u := range originalURLs {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+2)) // Начинаем с $2, так как $1 - это userID
		args = append(args, u)
	}

	query := fmt.Sprintf("SELECT url, key FROM urls WHERE user_id = $1 AND url IN (%s)", strings.Join(placeholders, ","))
	rows, err := r.db.QueryContext(ctx, query, args...)

	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var result []model.ShortURL

	// Проходим по строкам и собираем ShortURL
	for rows.Next() {
		var url, key string
		if scanErr := rows.Scan(&url, &key); scanErr != nil {
			return nil, scanErr
		}

		// Ищем соответствующий CorrelationID в исходном массиве
		var correlationID string
		for _, originalURL := range URLs {
			if originalURL.OriginalURL == url {
				correlationID = originalURL.CorrelationID
				break
			}
		}

		result = append(result, model.ShortURL{
			CorrelationID: correlationID,
			OriginalURL:   url,
			Key:           key,
		})
	}

	// Если произошла ошибка - возвращаем её
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// GetManyShortURLs возвращает все сокращенные URL
func (r *PostgresRepository) GetManyShortURLs(ctx context.Context, userID string) ([]model.ShortURL, error) {
	var result []model.ShortURL

	rows, err := r.db.QueryContext(ctx, "select id, url, key from urls where user_id = $1", userID)

	if err != nil {
		return nil, err
	}

	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var id string
		var url string
		var key string
		if scanErr := rows.Scan(&id, &url, &key); scanErr != nil {
			return nil, scanErr
		}

		result = append(result, model.ShortURL{
			CorrelationID: id,
			OriginalURL:   url,
			Key:           key,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
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

// GetUsersCount возвращает число записей в таблице users.
func (r *PostgresRepository) GetUsersCount(ctx context.Context) int {
	row := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users")

	var count int64
	if err := row.Scan(&count); err != nil {
		logger.Log.Error("Error scanning users count", zap.Error(err))
		return 0
	}

	return int(count)
}

// Close закрывает хранилище
func (r *PostgresRepository) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

// Ping проверяет доступность хранилища
func (r *PostgresRepository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

// GetUser возвращает пользователя по его ID
func (r *PostgresRepository) GetUser(ctx context.Context, userID string) (model.User, error) {
	row := r.db.QueryRowContext(ctx, "select id from users where id = $1", userID)

	var ID string
	err := row.Scan(&ID)

	if err != nil {
		return model.User{}, err
	}

	return model.User{ID: ID}, nil
}

func (r *PostgresRepository) CreateUser(ctx context.Context) (model.User, error) {
	row := r.db.QueryRowContext(ctx, "INSERT INTO users DEFAULT VALUES RETURNING id")

	var ID string
	err := row.Scan(&ID)

	if err != nil {
		return model.User{}, err
	}

	fmt.Println("CREATE USER WITH ID = ", ID)

	return model.User{ID: ID}, nil
}

// DeleteManyURLs удаление множества URL
func (r *PostgresRepository) DeleteManyURLs(ctx context.Context, userID string, keys []string) error {
	// Если массив пустой, то просто возвращаем пустой результат
	if len(keys) == 0 {
		return nil
	}

	// Начинаем транзакцию
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Подготавливаем данные для запроса
	placeholders := make([]string, 0, len(keys))
	args := make([]interface{}, 0, len(keys)+1)

	// Добавляем userID как первый параметр
	args = append(args, userID)

	for i, k := range keys {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+2))
		args = append(args, k)
	}

	query := fmt.Sprintf("UPDATE urls SET deleted_at = CURRENT_TIMESTAMP WHERE user_id = $1 AND deleted_at IS NULL AND key IN (%s)", strings.Join(placeholders, ","))
	_, err = tx.ExecContext(ctx, query, args...)

	if err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
