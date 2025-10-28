package handler

import (
	"net/http"
	"strings"

	"github.com/coolycow/shortener/internal/repository"
)

// GetHandler Обрабатываем GET-запросы к серверу.
func GetHandler(repo repository.URLRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Получаем id из URL
		id := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/"))

		// Проверяем, что id не пустой
		if id == "" {
			http.Error(w, "Empty id", http.StatusBadRequest)
			return
		}

		// Получаем исходный URL по id из репозитория
		rawURL, exists := repo.GetOriginalURL(id)

		// Проверяем, что URL существует
		if !exists {
			http.Error(w, "URL not found", http.StatusNotFound)
			return
		}

		// Формируем ответ
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Location", rawURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}
