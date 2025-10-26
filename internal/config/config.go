package config

// Структура для хранения конфигурации
type Config struct {
	ServerAddress      string // Адрес для запуска сервера (без протокола)
	BaseURL            string // Базовый URL для формирования коротких ссылок (с протоколом)
	RandomStringLength int
}

// NewConfig создает новый экземпляр Config с заданными значениями по умолчанию
func NewConfig() *Config {
	return &Config{
		ServerAddress:      "127.0.0.1:8080",
		BaseURL:            "http://127.0.0.1:8080",
		RandomStringLength: 6,
	}
}
