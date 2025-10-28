package handler

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/coolycow/shortener/internal/config"
	"github.com/coolycow/shortener/internal/repository"
	"github.com/coolycow/shortener/internal/service"
)

// PostHandler обрабатывает POST-запросы к серверу
func PostHandler(cfg *config.Config, repo repository.URLRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, что пришёл POST-запрос
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Проверяем, что тип контента - text/plain
		// Учитываем, что "Content-Type" может содержать и другие значения, например, charset=utf-8
		contentType := r.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "text/plain") {
			http.Error(w, "Content type not allowed", http.StatusUnsupportedMediaType)
			return
		}

		// Читаем тело запроса
		body, err := io.ReadAll(r.Body)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Проверяем, что не пришла пустота
		if len(body) == 0 {
			http.Error(w, "Empty body", http.StatusBadRequest)
			return
		}

		// Извлекаем строку из тела запроса и проверяем, что она не пуста
		trimBody := strings.TrimSpace(string(body))

		if trimBody == "" {
			http.Error(w, "Empty URL", http.StatusBadRequest)
			return
		}

		// Парсим URL из строки (фактически проверяем, что это действительно URL)
		validURL, err := url.ParseRequestURI(trimBody)

		if err != nil || validURL.Scheme == "" {
			http.Error(w, "Invalid URL", http.StatusBadRequest)
			return
		}

		// Если ссылка уже есть в репозитории, то возвращаем короткую ссылку
		validStringURL := validURL.String()
		if key, exists := repo.GetShortURL(validStringURL); exists {
			shortURL := cfg.BaseURL + "/" + key
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(shortURL))
			return
		}

		// Генерируем короткую ссылку заданной в настройках длины и гарантируем её уникальность
		key := service.RandomString(cfg.RandomStringLength)
		for repo.IsShortURLExists(key) {
			key = service.RandomString(cfg.RandomStringLength)
		}

		// Сохраняем короткую ссылку в репозитории
		repo.SaveURL(key, validStringURL)

		// Формируем ответ
		shortURL := cfg.BaseURL + "/" + key
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shortURL))
	}
}
