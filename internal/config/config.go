package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	flag "github.com/spf13/pflag"
)

// Config Структура для хранения конфигурации
type Config struct {
	Host                              string
	Port                              int
	BaseURL                           string
	RandomStringLength                int
	RandomStringMaxLength             int
	RandomStringMaxGenerationAttempts int
}

// GetServerAddress возвращает полный адрес сервера для его запуска
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// InitConfig возвращает настройки и ошибку если парсинг аргументов не удался
func InitConfig() (*Config, error) {
	return InitConfigWithArgs(os.Args[1:])
}

// parseAddress отдельная функция для парсинга адреса в виде одной строки
func parseAddress(c *Config) func(value string) error {
	return func(value string) error {
		addressParts := strings.Split(value, ":")

		if len(addressParts) != 2 {
			return fmt.Errorf("incorrect server address %s", value)
		}

		c.Host = addressParts[0]
		port, err := strconv.Atoi(addressParts[1])

		if err != nil {
			return fmt.Errorf("incorrect server port %s", addressParts[1])
		}

		c.Port = port

		return nil
	}
}

// parseRandomString отдельная функция для парсинга настроек уникальной случайной строки.
// Поддерживается передача 1-3 значений в порядке: длина, максимальная длина, количество попыток генерации
func parseRandomString(c *Config) func(value string) error {
	return func(value string) error {
		values := strings.Split(value, ",")

		// Минимум должен быть передан хотя бы один параметр
		if len(values) < 1 || len(values) > 3 {
			return fmt.Errorf("incorrect string options %s", value)
		}

		length, err := strconv.Atoi(values[0])
		if err != nil {
			return fmt.Errorf("incorrect length %s", value)
		}
		c.RandomStringLength = length

		// Если передано 2 параметра или более, то можем задать максимальную длину
		if len(values) > 1 {
			maxLength, err := strconv.Atoi(values[1])
			if err != nil {
				return fmt.Errorf("incorrect max length %s", value)
			}
			c.RandomStringMaxLength = maxLength
		}

		// Если передано 3 параметра, то можем задать максимальное количество попыток генерации
		if len(values) > 2 {
			attempts, err := strconv.Atoi(values[2])
			if err != nil {
				return fmt.Errorf("incorrect attempts %s", value)
			}
			c.RandomStringMaxGenerationAttempts = attempts
		}

		return nil
	}
}

// InitConfigWithArgs инициализация с переданными аргументами
func InitConfigWithArgs(args []string) (*Config, error) {
	var config Config

	flagSet := flag.NewFlagSet("main", flag.ContinueOnError)

	flagSet.StringVarP(&config.Host, "host", "h", "127.0.0.1", "server host")
	flagSet.IntVarP(&config.Port, "port", "p", 8080, "server port")
	flagSet.StringVarP(&config.BaseURL, "base", "b", "http://127.0.0.1:8080", "base url")

	flagSet.IntVarP(&config.RandomStringLength, "random-length", "l", 6, "random string length")
	flagSet.IntVarP(&config.RandomStringMaxLength, "random-max-length", "m", 100, "random string max length")
	flagSet.IntVarP(&config.RandomStringMaxGenerationAttempts, "random-attempts", "t", 1000, "max generation attempts")

	// Определение адреса сервера в виде строки 127.0.0.1:8080
	flagSet.FuncP("address", "a", "server address", parseAddress(&config))

	// Определение параметров уникальной строки в виде одной строки
	flagSet.FuncP("string", "s", "unique random string options", parseRandomString(&config))

	err := flagSet.Parse(args)

	if err != nil {
		return nil, err
	}

	return &config, nil
}
