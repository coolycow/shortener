package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"

	"github.com/coolycow/shortener/internal/config"
	"github.com/coolycow/shortener/internal/model"
	"github.com/coolycow/shortener/internal/repository"
)

// UserService Сервис для работы в Handler
type UserService interface {
	CreateUser(ctx context.Context) (model.User, error)
	GetUserIDFromCookie(cookie *http.Cookie) (int, error)
	GetCookieValueByUserID(userID int) (string, error)
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

// CreateUser создаёт нового пользователя
func (s *userService) CreateUser(ctx context.Context) (model.User, error) {
	return s.repo.CreateUser(ctx)
}

// GetUserIDFromCookie достаёт UserID из переданной куки
func (s *userService) GetUserIDFromCookie(cookie *http.Cookie) (int, error) {
	cookieValue := cookie.Value

	// Декодируем hex
	data, err := hex.DecodeString(cookieValue)
	if err != nil {
		return 0, err
	}

	key := sha256.Sum256([]byte(s.cfg.SecretKey))

	aesBlock, err := aes.NewCipher(key[:])
	if err != nil {
		return 0, err
	}

	aesGCM, err := cipher.NewGCM(aesBlock)
	if err != nil {
		return 0, err
	}

	// создаём вектор инициализации
	nonceSize := aesGCM.NonceSize()

	// Разделяем nonce и зашифрованные данные
	nonce := data[:nonceSize]
	ciphertext := data[nonceSize:]

	// расшифровываем
	decrypted, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return 0, err
	}

	uID, err := strconv.Atoi(string(decrypted))
	if err != nil {
		return 0, err
	}

	return uID, nil
}

// GetCookieValueByUserID возвращает значение куки для указанного userID
func (s *userService) GetCookieValueByUserID(userID int) (string, error) {
	key := sha256.Sum256([]byte(s.cfg.SecretKey))

	aesBlock, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(aesBlock)
	if err != nil {
		return "", err
	}

	// создаём вектор инициализации
	nonce, err := generateRandom(aesGCM.NonceSize())
	if err != nil {
		return "", err
	}

	dst := aesGCM.Seal(nil, nonce, []byte(strconv.Itoa(userID)), nil)

	// Сохраняем nonce вместе с зашифрованными данными
	result := append(nonce, dst...)

	return hex.EncodeToString(result), nil
}
