package service

import (
	"errors"
	"log"
	"math/rand"
	"strings"

	"github.com/coolycow/shortener/internal/repository"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// Генерация строки заданной длины на основе заданного набора символов
func generate(length int) string {
	// strings.Builder для эффективной конкатенации строк
	var sb strings.Builder
	sb.Grow(length)

	// Генерация случайных символов
	for range length {
		sb.WriteByte(charset[rand.Intn(len(charset))])
	}

	return sb.String()
}

// CreateUniqueStringForURL генерирует уникальную строку с учетом проверки в репозитории
func CreateUniqueStringForURL(
	repo repository.URLRepository,
	length int,
	maxLength int,
	attempts int) (string, error) {
	i := 0
	key := generate(length)

	for repo.IsShortURLExists(key) {
		i += 1
		key = generate(length)

		if i >= attempts {
			length += 1
			i = 0
			log.Printf("Increased unique random string length: %d", length)
		}

		if length > maxLength {
			log.Printf("Length (%d) exceeds maximum length: (%d)", length, maxLength)
			return "", errors.New("length exceeds maximum length")
		}
	}
	return key, nil
}
