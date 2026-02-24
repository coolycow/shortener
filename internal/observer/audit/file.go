package audit

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/coolycow/shortener/internal/model"
)

// FileReceiver записывает события аудита в файл (каждое событие — новая строка).
type FileReceiver struct {
	path string
	mu   sync.Mutex
}

// NewFileReceiver создаёт приёмник в файл.
func NewFileReceiver(path string) *FileReceiver {
	return &FileReceiver{path: path}
}

// Send добавляет событие в конец файла в виде одной строки JSON.
func (f *FileReceiver) Send(event *model.Audit) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(f.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(append(data, '\n'))
	return err
}
