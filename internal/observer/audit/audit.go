package audit

import (
	"time"

	"github.com/coolycow/shortener/internal/model"
)

// Действия для события аудита
const (
	ActionShorten = "shorten"
	ActionFollow  = "follow"
)

// Receiver — интерфейс приёмника аудита (наблюдатель).
// Реализации: запись в файл, отправка на URL.
type Receiver interface {
	Send(event *model.Audit) error
}

// Notifier — субъект: хранит список приёмников и рассылает им события.
type Notifier struct {
	receivers []Receiver
}

// Notify отправляет событие во все зарегистрированные приёмники.
func (n *Notifier) Notify(event *model.Audit) {
	for _, r := range n.receivers {
		_ = r.Send(event) // по заданию можно не блокировать из-за ошибки приёмника
	}
}

// NewEvent создаёт событие аудита с текущим timestamp.
func NewEvent(action, userID, originalURL string) *model.Audit {
	return &model.Audit{
		TS:     int(time.Now().Unix()),
		Action: action,
		UserID: userID,
		URL:    originalURL,
	}
}

// NewNotifier создаёт Notifier и добавляет приёмники по конфигу.
// auditFile != "" — добавляется приёмник в файл, auditURL != "" — на удалённый сервер.
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
