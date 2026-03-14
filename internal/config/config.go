// Package config предоставляет загрузку и хранение конфигурации приложения.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
	FileStoragePath                   string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN                       string `env:"DATABASE_DSN"`
	RunMigrations                     bool   `env:"RUN_MIGRATIONS"`
	SecretKey                         string `env:"SECRET_KEY"`
	AuditFile                         string `env:"AUDIT_FILE"`
	AuditURL                          string `env:"AUDIT_URL"`
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
	fmt.Printf("FileStoragePath: %s\n", c.FileStoragePath)
	fmt.Printf("DatabaseDSN: %s\n", c.DatabaseDSN)
	fmt.Printf("RunMigrations: %t\n", c.RunMigrations)
	fmt.Printf("SecretKey: %s\n", c.SecretKey)
	fmt.Printf("AuditFile: %s\n", c.AuditFile)
	fmt.Printf("AuditURL: %s\n", c.AuditURL)
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

	// Проверяем параметры генерации случайных строк.
	// Параметры должны иметь корректные значения, чтобы избежать ошибок при генерации и сохранении.
	var errs []error
	if config.RandomStringLength <= 0 || config.RandomStringLength > 255 {
		errs = append(errs, errors.New("random string length must be between 1 and 255"))
	}

	if config.RandomStringMaxLength <= 0 || config.RandomStringMaxLength > 255 {
		errs = append(errs, errors.New("random string max length must be between 1 and 255"))
	}

	if config.RandomStringMaxGenerationAttempts <= 0 || config.RandomStringMaxGenerationAttempts > 1000 {
		errs = append(errs, errors.New("random string max generation attempts must be between 1 and 1000"))
	}

	if config.RandomStringMaxLength < config.RandomStringLength {
		errs = append(errs, errors.New("random string max length must be greater than or equal to random string length"))
	}

	return config, errors.Join(errs...)
}

// initConfigWithEnv получение настроек из переменных окружения
func initConfigWithEnv(config *Config) (*Config, error) {
	if host, present := os.LookupEnv("HOST"); present {
		config.Host = host
	}

	if err := parseIntFromEnv(config, "PORT",
		func(c *Config, v int) { c.Port = v }); err != nil {
		return nil, err
	}

	if baseURL, present := os.LookupEnv("BASE_URL"); present {
		config.BaseURL = baseURL
	}

	if secretKey, present := os.LookupEnv("SECRET_KEY"); present {
		config.SecretKey = secretKey
	}

	if err := parseIntFromEnv(config, "RANDOM_STRING_LENGTH",
		func(c *Config, v int) { c.RandomStringLength = v }); err != nil {
		return nil, err
	}

	if err := parseIntFromEnv(config, "RANDOM_STRING_MAX_LENGTH",
		func(c *Config, v int) { c.RandomStringMaxLength = v }); err != nil {
		return nil, err
	}

	if err := parseIntFromEnv(config, "RANDOM_STRING_MAX_GENERATION_ATTEMPTS",
		func(c *Config, v int) { c.RandomStringMaxGenerationAttempts = v }); err != nil {
		return nil, err
	}

	if serverAddress, present := os.LookupEnv("SERVER_ADDRESS"); present {
		host, port, err := splitServerAddress(serverAddress)

		if err != nil {
			return nil, fmt.Errorf("invalid env server address %s", serverAddress)
		}

		config.Host = host
		config.Port = port
	}

	if logLevel, present := os.LookupEnv("LOG_LEVEL"); present {
		config.LogLevel = logLevel
	}

	if fileStoragePath, present := os.LookupEnv("FILE_STORAGE_PATH"); present {
		config.FileStoragePath = fileStoragePath
	}

	if databaseDSN, present := os.LookupEnv("DATABASE_DSN"); present {
		config.DatabaseDSN = databaseDSN
	}

	if runMigrations, present := os.LookupEnv("RUN_MIGRATIONS"); present {
		config.RunMigrations, _ = strconv.ParseBool(runMigrations)
	}

	if auditFile, present := os.LookupEnv("AUDIT_FILE"); present {
		config.AuditFile = auditFile
	}

	if auditURL, present := os.LookupEnv("AUDIT_URL"); present {
		config.AuditURL = auditURL
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
	flagSet.IntVarP(&config.RandomStringMaxLength, "random-max-length", "m", 255, "random string max length")
	flagSet.IntVarP(&config.RandomStringMaxGenerationAttempts, "random-attempts", "t", 1000, "max generation attempts")

	flagSet.StringVarP(&config.LogLevel, "log-level", "e", "info", "log level")

	flagSet.StringVarP(&config.FileStoragePath, "file-storage-path", "f", getDefaultStoragePath(), "file storage path")

	flagSet.StringVarP(&config.DatabaseDSN, "database-dsn", "d", getDefaultDatabaseDSN(), "database DSN")
	flagSet.BoolVarP(&config.RunMigrations, "run-migrations", "r", false, "run migrations")
	flagSet.StringVarP(&config.SecretKey, "secret-key", "k", getDefaultSecretKey(), "secret key")

	flagSet.StringVarP(&config.AuditFile, "audit-file", "z", "", "audit file")
	flagSet.StringVarP(&config.AuditURL, "audit-url", "u", "", "audit url")

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

// parseIntFromEnv парсит int-значение из переменной окружения и устанавливает его в поле конфигурации
func parseIntFromEnv(config *Config, envKey string, setter func(*Config, int)) error {
	if value, present := os.LookupEnv(envKey); present {
		intValue, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid env %s %s", envKey, value)
		}
		setter(config, intValue)
	}
	return nil
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

// getDefaultStoragePath Получаем директорию, где находится исполняемый файл
func getDefaultStoragePath() string {
	exe, err := os.Executable()

	if err != nil {
		return "urls.json"
	}

	return filepath.Join(filepath.Dir(exe), "urls.json")
}

// getDefaultDatabaseDSN Стандартные настройки подключения к БД
func getDefaultDatabaseDSN() string {
	return ""
}

// getDefaultSecretKey секретный ключ по умолчанию
func getDefaultSecretKey() string {
	return "shortener_secret_key"
}
