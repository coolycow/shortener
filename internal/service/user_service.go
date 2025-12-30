package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"

	"github.com/coolycow/shortener/internal/config"
	"github.com/coolycow/shortener/internal/model"
	"github.com/coolycow/shortener/internal/repository"
)

// UserService Сервис для работы в Handler
type UserService interface {
	GetUser(ctx context.Context, userID string) (model.User, error)
	CreateUser(ctx context.Context) (model.User, error)
	GetUserIDFromCookie(cookie *http.Cookie) (string, error)

	GetCookieValueByUser(user model.User) (string, error)
	GetCookieValueByUserID(userID string) (string, error)
}

// Реализация сервисного слоя
type userService struct {
	repo repository.URLRepository
	cfg  *config.Config
}

// NewUserService инициализация сервиса
func NewUserService(cfg *config.Config, repo repository.URLRepository) UserService {
	return &userService{
		repo: repo,
		cfg:  cfg,
	}
}

// generateRandom генерация случайных байт
func generateRandom(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)

	if err != nil {
		return nil, err
	}

	return b, nil
}

// GetUser получает пользователя по его ID
func (s *userService) GetUser(ctx context.Context, userID string) (model.User, error) {
	return s.repo.GetUser(ctx, userID)
}

// CreateUser создаёт нового пользователя
func (s *userService) CreateUser(ctx context.Context) (model.User, error) {
	return s.repo.CreateUser(ctx)
}

// GetUserIDFromCookie достаёт UserID из переданной куки
func (s *userService) GetUserIDFromCookie(cookie *http.Cookie) (string, error) {
	cookieValue := cookie.Value

	// Декодируем hex
	data, err := hex.DecodeString(cookieValue)
	if err != nil {
		return "", fmt.Errorf("failed to decode hex cookie value: %w", err)
	}

	if len(data) == 0 {
		return "", errors.New("invalid cookie")
	}

	key := sha256.Sum256([]byte(s.cfg.SecretKey))

	aesBlock, err := aes.NewCipher(key[:])
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(aesBlock)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM cipher: %w", err)
	}

	// создаём вектор инициализации
	nonceSize := aesGCM.NonceSize()

	// Разделяем nonce и зашифрованные данные
	nonce := data[:nonceSize]
	ciphertext := data[nonceSize:]

	// расшифровываем
	decrypted, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}

// GetCookieValueByUser возвращает значение куки для указанного user
func (s *userService) GetCookieValueByUser(user model.User) (string, error) {
	return s.GetCookieValueByUserID(user.ID)
}

// GetCookieValueByUserID возвращает значение куки для указанного userID
func (s *userService) GetCookieValueByUserID(userID string) (string, error) {
	key := sha256.Sum256([]byte(s.cfg.SecretKey))

	aesBlock, err := aes.NewCipher(key[:])
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(aesBlock)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM cipher: %w", err)
	}

	// создаём вектор инициализации
	nonce, err := generateRandom(aesGCM.NonceSize())
	if err != nil {
		return "", fmt.Errorf("failed to generate random nonce: %w", err)
	}

	dst := aesGCM.Seal(nil, nonce, []byte(userID), nil)

	// Сохраняем nonce вместе с зашифрованными данными
	result := append(nonce, dst...)

	return hex.EncodeToString(result), nil
}
