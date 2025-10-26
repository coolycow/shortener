package handler

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/coolycow/shortener/internal/config"
	"github.com/coolycow/shortener/internal/service"
)

// PostHandler обрабатывает POST-запросы к серверу
func PostHandler(w http.ResponseWriter, r *http.Request, cfg *config.Config, shortToOriginal map[string]string, originalToShort map[string]string) {
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
		http.Error(w, "Empty url", http.StatusBadRequest)
		return
	}

	// Парсим URL из строки (фактически проверяем, что это действительно URL)
	validUrl, err := url.ParseRequestURI(trimBody)

	if err != nil || validUrl.Scheme == "" {
		http.Error(w, "Invalid url", http.StatusBadRequest)
		return
	}

	// Если ссылка уже есть в originalToShort, то возвращаем короткую ссылку
	if key := originalToShort[validUrl.String()]; key != "" {
		shortUrl := cfg.ServerAddress + "/" + key
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(shortUrl))
		return
	}

	// Генерируем короткую ссылку заданной в настройках длины и гарантируем её уникальность
	key := service.RandomString(cfg.RandomStringLength)
	for shortToOriginal[key] != "" {
		key = service.RandomString(cfg.RandomStringLength)
	}

	// Сохраняем короткую ссылку в карте
	validStringUrl := validUrl.String()
	shortToOriginal[key] = validStringUrl
	originalToShort[validStringUrl] = key

	// Формируем ответ
	shortUrl := cfg.ServerAddress + "/" + key
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortUrl))
}
