package config

// Config Структура для хранения конфигурации
type Config struct {
	ServerAddress         string // Адрес для запуска сервера (без протокола)
	BaseURL               string // Базовый URL для формирования коротких ссылок (с протоколом)
	RandomStringLength    int
	RandomStringMaxLength int
	MaxGenerationAttempts int
}

// NewConfig создает новый экземпляр Config с заданными значениями по умолчанию
func NewConfig() *Config {
	return &Config{
		ServerAddress:         "127.0.0.1:8080",
		BaseURL:               "http://127.0.0.1:8080",
		RandomStringLength:    6,
		RandomStringMaxLength: 100,
		MaxGenerationAttempts: 1000,
	}
}
