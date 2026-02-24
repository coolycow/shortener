package audit

import (
	"time"

	"github.com/coolycow/shortener/internal/model"
)

// Константы типа действия в событии аудита.
const (
	ActionShorten = "shorten" // создание короткой ссылки
	ActionFollow  = "follow" // переход по короткой ссылке
)

// Receiver — приёмник событий аудита (файл, HTTP и т.д.).
type Receiver interface {
	Send(event *model.Audit) error
}

// Notifier рассылает события аудита всем зарегистрированным приёмникам.
type Notifier struct {
	receivers []Receiver
}

// Notify отправляет событие во все приёмники (ошибки приёмников игнорируются).
func (n *Notifier) Notify(event *model.Audit) {
	for _, r := range n.receivers {
		_ = r.Send(event) // по заданию можно не блокировать из-за ошибки приёмника
	}
}

// NewEvent создаёт событие аудита с текущим временем (action: ActionShorten или ActionFollow).
func NewEvent(action, userID, originalURL string) *model.Audit {
	return &model.Audit{
		TS:     int(time.Now().Unix()),
		Action: action,
		UserID: userID,
		URL:    originalURL,
	}
}

// NewNotifier создаёт Notifier: при auditFile != "" — запись в файл, при auditURL != "" — отправка на URL.
func NewNotifier(auditFile, auditURL string) *Notifier {
	n := &Notifier{receivers: nil}

	if auditFile != "" {
		n.receivers = append(n.receivers, NewFileReceiver(auditFile))
	}

	if auditURL != "" {
		n.receivers = append(n.receivers, NewURLReceiver(auditURL))
	}

	return n
}
