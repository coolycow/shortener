package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	flag "github.com/spf13/pflag"
)

// Config Структура для хранения конфигурации, задаются соответствия ENV
type Config struct {
	Host                              string `env:"HOST"`
	Port                              int    `env:"PORT"`
	BaseURL                           string `env:"BASE_URL"`
	RandomStringLength                int    `env:"RANDOM_STRING_LENGTH"`
	RandomStringMaxLength             int    `env:"RANDOM_STRING_MAX_LENGTH"`
	RandomStringMaxGenerationAttempts int    `env:"RANDOM_STRING_MAX_GENERATION_ATTEMPTS"`
	LogLevel                          string `env:"LOG_LEVEL"`
}

// GetServerAddress возвращает полный адрес сервера для его запуска
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// PrintConfig выводит настройки в консоль
func (c *Config) PrintConfig() {
	fmt.Printf("Host: %s\n", c.Host)
	fmt.Printf("Port: %d\n", c.Port)
	fmt.Printf("BaseURL: %s\n", c.BaseURL)
	fmt.Printf("RandomStringLength: %d\n", c.RandomStringLength)
	fmt.Printf("RandomStringMaxLength: %d\n", c.RandomStringMaxLength)
	fmt.Printf("RandomStringMaxGenerationAttempts: %d\n", c.RandomStringMaxGenerationAttempts)
	fmt.Printf("LogLevel: %s\n", c.LogLevel)
}

// InitConfig возвращает настройки и ошибку если парсинг аргументов не удался
func InitConfig() (*Config, error) {
	// Инициализируем настройки из флагов, также будут заданы значения по умолчанию
	config, err := InitConfigWithArgs(os.Args[1:])

	if err != nil {
		return nil, err
	}

	// Инициализируем настройки из переменных окружения если таковые указаны
	config, err = initConfigWithEnv(config)

	if err != nil {
		return nil, err
	}

	return config, nil
}

// initConfigWithEnv получение настроек из переменных окружения
func initConfigWithEnv(config *Config) (*Config, error) {
	if host := os.Getenv("HOST"); host != "" {
		config.Host = host
	}

	if port := os.Getenv("PORT"); port != "" {
		p, err := strconv.Atoi(port)

		if err != nil {
			return nil, fmt.Errorf("invalid env port %s", port)
		}

		config.Port = p
	}

	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		config.BaseURL = baseURL
	}

	if length := os.Getenv("RANDOM_STRING_LENGTH"); length != "" {
		randomStringLength, err := strconv.Atoi(length)

		if err != nil {
			return nil, fmt.Errorf("invalid env random string length %s", length)
		}

		config.RandomStringLength = randomStringLength
	}

	if maxLength := os.Getenv("RANDOM_STRING_MAX_LENGTH"); maxLength != "" {
		randomStringMaxLength, err := strconv.Atoi(maxLength)

		if err != nil {
			return nil, fmt.Errorf("invalid env random string max length %s", maxLength)
		}

		config.RandomStringMaxLength = randomStringMaxLength
	}

	if maxAttempts := os.Getenv("RANDOM_STRING_MAX_GENERATION_ATTEMPTS"); maxAttempts != "" {
		randomStringMaxGenerationAttempts, err := strconv.Atoi(maxAttempts)

		if err != nil {
			return nil, fmt.Errorf("invalid env random string max generation attempts %s", maxAttempts)
		}

		config.RandomStringMaxGenerationAttempts = randomStringMaxGenerationAttempts
	}

	if serverAddress := os.Getenv("SERVER_ADDRESS"); serverAddress != "" {
		host, port, err := splitServerAddress(serverAddress)

		if err != nil {
			return nil, fmt.Errorf("invalid env server address %s", serverAddress)
		}

		config.Host = host
		config.Port = port
	}

	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		config.LogLevel = logLevel
	}

	return config, nil
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

	flagSet.StringVarP(&config.LogLevel, "log-level", "e", "info", "log level")

	// Определение адреса сервера в виде строки 127.0.0.1:8080
	flagSet.FuncP("address", "a", "server address", parseAddress(&config))

	// Определение параметров уникальной строки в виде одной строки
	flagSet.FuncP("string", "s", "unique random string options", parseRandomString(&config))

	err := flagSet.Parse(args)

	if err != nil {
		return nil, err
	}

	// Возвращаем адрес переменной config
	return &config, nil
}

// parseAddress отдельная функция для парсинга адреса в виде одной строки
func parseAddress(c *Config) func(value string) error {
	return func(value string) error {
		host, port, err := splitServerAddress(value)

		if err != nil {
			return err
		}

		c.Host = host
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

// splitServerAddress общая функция парсинга хоста и порта из строки адреса
func splitServerAddress(address string) (string, int, error) {
	addressParts := strings.Split(address, ":")

	if len(addressParts) != 2 {
		return ``, 0, fmt.Errorf("incorrect server address %s", address)
	}

	host := addressParts[0]
	port, err := strconv.Atoi(addressParts[1])

	if err != nil {
		return ``, 0, fmt.Errorf("incorrect server port %s", addressParts[1])
	}

	return host, port, nil
}
