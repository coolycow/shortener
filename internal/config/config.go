package config

// Структура для хранения конфигурации
type Config struct {
	ServerAddress      string
	RandomStringLength int
}

// NewConfig создает новый экземпляр Config с заданными значениями по умолчанию
func NewConfig() *Config {
	return &Config{
		ServerAddress:      "localhost:8080",
		RandomStringLength: 6,
	}
}
