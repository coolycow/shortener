package audit

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/coolycow/shortener/internal/model"
)

// URLReceiver отправляет события аудита на удалённый сервер методом POST.
type URLReceiver struct {
	url string
}

// NewURLReceiver создаёт приёмник на удалённый URL.
func NewURLReceiver(auditURL string) *URLReceiver {
	return &URLReceiver{url: auditURL}
}

// Send отправляет событие POST-запросом с JSON-телом.
func (u *URLReceiver) Send(event *model.Audit) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	resp, err := http.Post(u.url, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return err // или обернуть статус в свою ошибку
	}
	return nil
}
