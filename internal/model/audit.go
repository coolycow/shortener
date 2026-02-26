package model

// Audit — событие аудита (создание/переход по ссылке) для отправки в файл или на URL.
type Audit struct {
	TS     int    `json:"ts"`                // unix timestamp события
	Action string `json:"action"`            // действие: shorten (создание) или follow (прохождение по ссылке)
	UserID string `json:"user_id,omitempty"` // идентификатор пользователя, если есть
	URL    string `json:"url"`               // оригинальный (не сокращенный) URL
}
