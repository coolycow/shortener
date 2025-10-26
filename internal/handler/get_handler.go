package handler

import (
	"net/http"
	"strings"
)

// GetHandler Обрабатываем GET-запросы к серверу.
func GetHandler(w http.ResponseWriter, r *http.Request, urls map[string]string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем id из URL
	id := strings.TrimPrefix(r.URL.Path, "/")

	// Проверяем, что id не пустой
	if id == "" {
		http.Error(w, "Empty id", http.StatusBadRequest)
		return
	}

	// Получаем исходный URL по id
	rawUrl := urls[id]

	// Проверяем, что полученное значение не пустое
	if rawUrl == "" {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	// Формируем ответ
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Location", rawUrl)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
