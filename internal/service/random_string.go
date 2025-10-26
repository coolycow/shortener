package service

import (
	"math/rand"
	"strings"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandomString(length int) string {
	// strings.Builder для эффективной конкатенации строк
	var sb strings.Builder
	sb.Grow(length)

	// Генерация случайных символов
	for range length {
		sb.WriteByte(charset[rand.Intn(len(charset))])
	}

	return sb.String()
}
