// Package config предоставляет загрузку и хранение конфигурации приложения.
package config

import (
	"encoding/json"
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
	Host                              string `env:"HOST" json:"host,omitempty"`
	Port                              int    `env:"PORT" json:"port,omitempty"`
	BaseURL                           string `env:"BASE_URL" json:"base_url,omitempty"`
	RandomStringLength                int    `env:"RANDOM_STRING_LENGTH" json:"random_string_length,omitempty"`
	RandomStringMaxLength             int    `env:"RANDOM_STRING_MAX_LENGTH" json:"random_string_max_length,omitempty"`
	RandomStringMaxGenerationAttempts int    `env:"RANDOM_STRING_MAX_GENERATION_ATTEMPTS" json:"random_string_max_generation_attempts,omitempty"`
	LogLevel                          string `env:"LOG_LEVEL" json:"log_level,omitempty"`
	FileStoragePath                   string `env:"FILE_STORAGE_PATH" json:"file_storage_path,omitempty"`
	DatabaseDSN                       string `env:"DATABASE_DSN" json:"database_dsn,omitempty"`
	RunMigrations                     bool   `env:"RUN_MIGRATIONS" json:"run_migrations,omitempty"`
	SecretKey                         string `env:"SECRET_KEY" json:"secret_key,omitempty"`
	AuditFile                         string `env:"AUDIT_FILE" json:"audit_file,omitempty"`
	AuditURL                          string `env:"AUDIT_URL" json:"audit_url,omitempty"`
	EnableHTTPS                       bool   `env:"ENABLE_HTTPS" json:"enable_https,omitempty"`
	TLSCertFile                       string `env:"TLS_CERT_FILE" json:"tls_cert_file,omitempty"`
	TLSKeyFile                        string `env:"TLS_KEY_FILE" json:"tls_key_file,omitempty"`
	Config                            string `env:"CONFIG" json:"config,omitempty"`
}

// fileConfig — JSON-файл; указатели задают поля, явно присутствующие в файле.
// Ключ "config" в файле не разбираем (путь к файлу только из -c / CONFIG).
type fileConfig struct {
	ServerAddress                     *string `json:"server_address"`
	Host                              *string `json:"host"`
	Port                              *int    `json:"port"`
	BaseURL                           *string `json:"base_url"`
	RandomStringLength                *int    `json:"random_string_length"`
	RandomStringMaxLength             *int    `json:"random_string_max_length"`
	RandomStringMaxGenerationAttempts *int    `json:"random_string_max_generation_attempts"`
	LogLevel                          *string `json:"log_level"`
	FileStoragePath                   *string `json:"file_storage_path"`
	DatabaseDSN                       *string `json:"database_dsn"`
	RunMigrations                     *bool   `json:"run_migrations"`
	SecretKey                         *string `json:"secret_key"`
	AuditFile                         *string `json:"audit_file"`
	AuditURL                          *string `json:"audit_url"`
	EnableHTTPS                       *bool   `json:"enable_https"`
	TLSCertFile                       *string `json:"tls_cert_file"`
	TLSKeyFile                        *string `json:"tls_key_file"`
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
	fmt.Printf("EnableHTTPS: %t\n", c.EnableHTTPS)
	fmt.Printf("TLSCertFile: %s\n", c.TLSCertFile)
	fmt.Printf("TLSKeyFile: %s\n", c.TLSKeyFile)
	fmt.Printf("Config: %s\n", c.Config)
}

// InitConfig возвращает настройки и ошибку если парсинг аргументов не удался.
// Порядок приоритета: значения по умолчанию → JSON-файл → переменные окружения → флаги.
func InitConfig() (*Config, error) {
	args := os.Args[1:]

	// Получаем настройки из флагов
	flagCfg, fs, err := parseFlags(args)
	if err != nil {
		return nil, err
	}

	// Получаем настройки из переменных окружения
	configPath := strings.TrimSpace(flagCfg.Config)
	if !fs.Changed("config") {
		if v, ok := os.LookupEnv("CONFIG"); ok {
			configPath = strings.TrimSpace(v)
		}
	}

	// Получаем настройки из файла конфигурации
	cfg := defaultConfig()
	_, configEnvSet := os.LookupEnv("CONFIG")
	explicitConfigPath := fs.Changed("config") || configEnvSet

	if configPath != "" {
		if err := mergeConfigFromFile(&cfg, configPath); err != nil {
			if errors.Is(err, os.ErrNotExist) && !explicitConfigPath {
				cfg = defaultConfig()
			} else {
				return nil, err
			}
		}
	}

	cfg.Config = configPath

	// Применяем настройки из переменных окружения
	if _, err := applyEnvToConfig(&cfg, fs.Changed("config")); err != nil {
		return nil, err
	}

	applyExplicitFlags(&cfg, flagCfg, fs)

	// Проверяем настройки на корректность
	var errs []error
	if cfg.RandomStringLength <= 0 || cfg.RandomStringLength > 255 {
		errs = append(errs, errors.New("random string length must be between 1 and 255"))
	}

	if cfg.RandomStringMaxLength <= 0 || cfg.RandomStringMaxLength > 255 {
		errs = append(errs, errors.New("random string max length must be between 1 and 255"))
	}

	if cfg.RandomStringMaxGenerationAttempts <= 0 || cfg.RandomStringMaxGenerationAttempts > 1000 {
		errs = append(errs, errors.New("random string max generation attempts must be between 1 and 1000"))
	}

	if cfg.RandomStringMaxLength < cfg.RandomStringLength {
		errs = append(errs, errors.New("random string max length must be greater than or equal to random string length"))
	}

	return &cfg, errors.Join(errs...)
}

// initConfigWithEnv получение настроек из переменных окружения.
func initConfigWithEnv(config *Config) (*Config, error) {
	return applyEnvToConfig(config, false)
}

// applyEnvToConfig применяет переменные окружения. Если skipConfigFromEnv, CONFIG не трогаем
// (путь к файлу задан явно флагом -c).
func applyEnvToConfig(config *Config, skipConfigFromEnv bool) (*Config, error) {
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

	if enableHTTPS, present := os.LookupEnv("ENABLE_HTTPS"); present {
		config.EnableHTTPS, _ = strconv.ParseBool(enableHTTPS)
	}

	if certFile, present := os.LookupEnv("TLS_CERT_FILE"); present {
		config.TLSCertFile = certFile
	}

	if keyFile, present := os.LookupEnv("TLS_KEY_FILE"); present {
		config.TLSKeyFile = keyFile
	}

	if !skipConfigFromEnv {
		if configFile, present := os.LookupEnv("CONFIG"); present {
			config.Config = strings.TrimSpace(configFile)
		}
	}

	return config, nil
}

// InitConfigWithArgs инициализация с переданными аргументами (только флаги; как раньше для тестов).
func InitConfigWithArgs(args []string) (*Config, error) {
	cfg, _, err := parseFlags(args)
	return cfg, err
}

func parseFlags(args []string) (*Config, *flag.FlagSet, error) {
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

	flagSet.BoolVarP(&config.EnableHTTPS, "enable-https", "s", false, "enable HTTPS")
	flagSet.StringVar(&config.TLSCertFile, "tls-cert-file", getDefaultTLSCertFile(), "TLS certificate file (PEM), for HTTPS")
	flagSet.StringVar(&config.TLSKeyFile, "tls-key-file", getDefaultTLSKeyFile(), "TLS private key file (PEM), for HTTPS")

	flagSet.FuncP("address", "a", "server address", parseAddress(&config))

	flagSet.FuncP("string", "x", "unique random string options", parseRandomString(&config))

	flagSet.StringVarP(&config.Config, "config", "c", getDefaultConfigFile(), "config file")

	err := flagSet.Parse(args)
	if err != nil {
		return nil, nil, err
	}

	return &config, flagSet, nil
}

func defaultConfig() Config {
	return Config{
		Host:                              "127.0.0.1",
		Port:                              8080,
		BaseURL:                           "http://127.0.0.1:8080",
		RandomStringLength:                6,
		RandomStringMaxLength:             255,
		RandomStringMaxGenerationAttempts: 1000,
		LogLevel:                          "info",
		FileStoragePath:                   getDefaultStoragePath(),
		DatabaseDSN:                       getDefaultDatabaseDSN(),
		RunMigrations:                     false,
		SecretKey:                         getDefaultSecretKey(),
		AuditFile:                         "",
		AuditURL:                          "",
		EnableHTTPS:                       false,
		TLSCertFile:                       getDefaultTLSCertFile(),
		TLSKeyFile:                        getDefaultTLSKeyFile(),
		Config:                            "",
	}
}

func mergeConfigFromFile(cfg *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	var fc fileConfig
	if err := json.Unmarshal(data, &fc); err != nil {
		return fmt.Errorf("invalid config file %s: %w", path, err)
	}

	if fc.ServerAddress != nil && strings.TrimSpace(*fc.ServerAddress) != "" {
		h, p, err := splitServerAddress(strings.TrimSpace(*fc.ServerAddress))
		if err != nil {
			return fmt.Errorf("invalid server_address in config file: %w", err)
		}
		cfg.Host = h
		cfg.Port = p
	} else {
		if fc.Host != nil {
			cfg.Host = *fc.Host
		}
		if fc.Port != nil {
			cfg.Port = *fc.Port
		}
	}

	if fc.BaseURL != nil {
		cfg.BaseURL = *fc.BaseURL
	}
	if fc.RandomStringLength != nil {
		cfg.RandomStringLength = *fc.RandomStringLength
	}
	if fc.RandomStringMaxLength != nil {
		cfg.RandomStringMaxLength = *fc.RandomStringMaxLength
	}
	if fc.RandomStringMaxGenerationAttempts != nil {
		cfg.RandomStringMaxGenerationAttempts = *fc.RandomStringMaxGenerationAttempts
	}
	if fc.LogLevel != nil {
		cfg.LogLevel = *fc.LogLevel
	}
	if fc.FileStoragePath != nil {
		cfg.FileStoragePath = *fc.FileStoragePath
	}
	if fc.DatabaseDSN != nil {
		cfg.DatabaseDSN = *fc.DatabaseDSN
	}
	if fc.RunMigrations != nil {
		cfg.RunMigrations = *fc.RunMigrations
	}
	if fc.SecretKey != nil {
		cfg.SecretKey = *fc.SecretKey
	}
	if fc.AuditFile != nil {
		cfg.AuditFile = *fc.AuditFile
	}
	if fc.AuditURL != nil {
		cfg.AuditURL = *fc.AuditURL
	}
	if fc.EnableHTTPS != nil {
		cfg.EnableHTTPS = *fc.EnableHTTPS
	}
	if fc.TLSCertFile != nil {
		cfg.TLSCertFile = *fc.TLSCertFile
	}
	if fc.TLSKeyFile != nil {
		cfg.TLSKeyFile = *fc.TLSKeyFile
	}

	return nil
}

func applyExplicitFlags(dst *Config, src *Config, fs *flag.FlagSet) {
	if fs.Changed("host") {
		dst.Host = src.Host
	}
	if fs.Changed("port") {
		dst.Port = src.Port
	}
	if fs.Changed("base") {
		dst.BaseURL = src.BaseURL
	}
	if fs.Changed("random-length") {
		dst.RandomStringLength = src.RandomStringLength
	}
	if fs.Changed("random-max-length") {
		dst.RandomStringMaxLength = src.RandomStringMaxLength
	}
	if fs.Changed("random-attempts") {
		dst.RandomStringMaxGenerationAttempts = src.RandomStringMaxGenerationAttempts
	}
	if fs.Changed("log-level") {
		dst.LogLevel = src.LogLevel
	}
	if fs.Changed("file-storage-path") {
		dst.FileStoragePath = src.FileStoragePath
	}
	if fs.Changed("database-dsn") {
		dst.DatabaseDSN = src.DatabaseDSN
	}
	if fs.Changed("run-migrations") {
		dst.RunMigrations = src.RunMigrations
	}
	if fs.Changed("secret-key") {
		dst.SecretKey = src.SecretKey
	}
	if fs.Changed("audit-file") {
		dst.AuditFile = src.AuditFile
	}
	if fs.Changed("audit-url") {
		dst.AuditURL = src.AuditURL
	}
	if fs.Changed("enable-https") {
		dst.EnableHTTPS = src.EnableHTTPS
	}
	if fs.Changed("tls-cert-file") {
		dst.TLSCertFile = src.TLSCertFile
	}
	if fs.Changed("tls-key-file") {
		dst.TLSKeyFile = src.TLSKeyFile
	}
	if fs.Changed("address") {
		dst.Host = src.Host
		dst.Port = src.Port
	}
	if fs.Changed("string") {
		dst.RandomStringLength = src.RandomStringLength
		dst.RandomStringMaxLength = src.RandomStringMaxLength
		dst.RandomStringMaxGenerationAttempts = src.RandomStringMaxGenerationAttempts
	}
	if fs.Changed("config") {
		dst.Config = strings.TrimSpace(src.Config)
	}
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

		if len(values) < 1 || len(values) > 3 {
			return fmt.Errorf("incorrect string options %s", value)
		}

		length, err := strconv.Atoi(values[0])
		if err != nil {
			return fmt.Errorf("incorrect length %s", value)
		}
		c.RandomStringLength = length

		if len(values) > 1 {
			maxLength, err := strconv.Atoi(values[1])
			if err != nil {
				return fmt.Errorf("incorrect max length %s", value)
			}
			c.RandomStringMaxLength = maxLength
		}

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

// getDefaultTLSCertFile файл сертификата по умолчанию
func getDefaultTLSCertFile() string {
	return "server.crt"
}

// getDefaultTLSKeyFile файл ключа по умолчанию
func getDefaultTLSKeyFile() string {
	return "server.key"
}

// getDefaultConfigFile файл конфигурации по умолчанию
func getDefaultConfigFile() string {
	return "config.json"
}
