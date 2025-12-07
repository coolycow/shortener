package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	fileStoragePath := getDefaultStoragePath()
	databaseDSN := getDefaultDatabaseDSN()

	tests := []struct {
		name string
		args []string
		env  map[string]string
		want Config
	}{
		{
			name: "Default",
			want: Config{
				Host:                              "127.0.0.1",
				Port:                              8080,
				BaseURL:                           "http://127.0.0.1:8080",
				RandomStringLength:                6,
				RandomStringMaxLength:             255,
				RandomStringMaxGenerationAttempts: 1000,
				LogLevel:                          "info",
				FileStoragePath:                   fileStoragePath,
				DatabaseDSN:                       databaseDSN,
			},
		},
		{
			name: "Host And Port",
			args: []string{"-h", "localhost", "-p", "8082"},
			want: Config{
				Host:                              "localhost",
				Port:                              8082,
				BaseURL:                           "http://127.0.0.1:8080",
				RandomStringLength:                6,
				RandomStringMaxLength:             255,
				RandomStringMaxGenerationAttempts: 1000,
				LogLevel:                          "info",
				FileStoragePath:                   fileStoragePath,
				DatabaseDSN:                       databaseDSN,
			},
		},
		{
			name: "Address",
			args: []string{"-a", "localhost:8083"},
			want: Config{
				Host:                              "localhost",
				Port:                              8083,
				BaseURL:                           "http://127.0.0.1:8080",
				RandomStringLength:                6,
				RandomStringMaxLength:             255,
				RandomStringMaxGenerationAttempts: 1000,
				LogLevel:                          "info",
				FileStoragePath:                   fileStoragePath,
				DatabaseDSN:                       databaseDSN,
			},
		},
		{
			name: "Base URL",
			args: []string{"-b", "http://localhost:8083"},
			want: Config{
				Host:                              "127.0.0.1",
				Port:                              8080,
				BaseURL:                           "http://localhost:8083",
				RandomStringLength:                6,
				RandomStringMaxLength:             255,
				RandomStringMaxGenerationAttempts: 1000,
				LogLevel:                          "info",
				FileStoragePath:                   fileStoragePath,
				DatabaseDSN:                       databaseDSN,
			},
		},
		{
			name: "Database DSN",
			args: []string{"-d", "test"},
			want: Config{
				Host:                              "127.0.0.1",
				Port:                              8080,
				BaseURL:                           "http://127.0.0.1:8080",
				RandomStringLength:                6,
				RandomStringMaxLength:             255,
				RandomStringMaxGenerationAttempts: 1000,
				LogLevel:                          "info",
				FileStoragePath:                   fileStoragePath,
				DatabaseDSN:                       "test",
			},
		},
		{
			name: "String options 1",
			args: []string{"-s", "8"},
			want: Config{
				Host:                              "127.0.0.1",
				Port:                              8080,
				BaseURL:                           "http://127.0.0.1:8080",
				RandomStringLength:                8,
				RandomStringMaxLength:             255,
				RandomStringMaxGenerationAttempts: 1000,
				LogLevel:                          "info",
				FileStoragePath:                   fileStoragePath,
				DatabaseDSN:                       databaseDSN,
			},
		},
		{
			name: "String options 2",
			args: []string{"-s", "8,200"},
			want: Config{
				Host:                              "127.0.0.1",
				Port:                              8080,
				BaseURL:                           "http://127.0.0.1:8080",
				RandomStringLength:                8,
				RandomStringMaxLength:             200,
				RandomStringMaxGenerationAttempts: 1000,
				LogLevel:                          "info",
				FileStoragePath:                   fileStoragePath,
				DatabaseDSN:                       databaseDSN,
			},
		},
		{
			name: "String options 3",
			args: []string{"-s", "8,200,5000"},
			want: Config{
				Host:                              "127.0.0.1",
				Port:                              8080,
				BaseURL:                           "http://127.0.0.1:8080",
				RandomStringLength:                8,
				RandomStringMaxLength:             200,
				RandomStringMaxGenerationAttempts: 5000,
				LogLevel:                          "info",
				FileStoragePath:                   fileStoragePath,
				DatabaseDSN:                       databaseDSN,
			},
		},
		{
			name: "Env: all settings",
			env: map[string]string{
				"HOST":                                  "localhost",
				"PORT":                                  "8888",
				"BASE_URL":                              "https://shortener.com",
				"RANDOM_STRING_LENGTH":                  "7",
				"RANDOM_STRING_MAX_LENGTH":              "8",
				"RANDOM_STRING_MAX_GENERATION_ATTEMPTS": "9",
			},
			want: Config{
				Host:                              "localhost",
				Port:                              8888,
				BaseURL:                           "https://shortener.com",
				RandomStringLength:                7,
				RandomStringMaxLength:             8,
				RandomStringMaxGenerationAttempts: 9,
				LogLevel:                          "info",
				FileStoragePath:                   fileStoragePath,
				DatabaseDSN:                       databaseDSN,
			},
		},
		{
			name: "Env: server address and base url",
			env: map[string]string{
				"SERVER_ADDRESS": "127.0.0.2:8888",
				"BASE_URL":       "https://shortener.com",
			},
			want: Config{
				Host:                              "127.0.0.2",
				Port:                              8888,
				BaseURL:                           "https://shortener.com",
				RandomStringLength:                6,
				RandomStringMaxLength:             255,
				RandomStringMaxGenerationAttempts: 1000,
				LogLevel:                          "info",
				FileStoragePath:                   fileStoragePath,
				DatabaseDSN:                       databaseDSN,
			},
		},
		{
			name: "Settings priority",
			args: []string{"-a", "localhost:8083", "-b", "http://localhost:8083"},
			env: map[string]string{
				"SERVER_ADDRESS": "127.0.0.2:8888",
				"BASE_URL":       "https://shortener.com",
			},
			want: Config{
				Host:                              "127.0.0.2",
				Port:                              8888,
				BaseURL:                           "https://shortener.com",
				RandomStringLength:                6,
				RandomStringMaxLength:             255,
				RandomStringMaxGenerationAttempts: 1000,
				LogLevel:                          "info",
				FileStoragePath:                   fileStoragePath,
				DatabaseDSN:                       databaseDSN,
			},
		},
		{
			name: "Log level debug",
			args: []string{"-e", "debug"},
			want: Config{
				Host:                              "127.0.0.1",
				Port:                              8080,
				BaseURL:                           "http://127.0.0.1:8080",
				RandomStringLength:                6,
				RandomStringMaxLength:             255,
				RandomStringMaxGenerationAttempts: 1000,
				LogLevel:                          "debug",
				FileStoragePath:                   fileStoragePath,
				DatabaseDSN:                       databaseDSN,
			},
		},
		{
			name: "Env: log level warning",
			env: map[string]string{
				"SERVER_ADDRESS": "127.0.0.2:8888",
				"BASE_URL":       "https://shortener.com",
				"LOG_LEVEL":      "warn",
			},
			want: Config{
				Host:                              "127.0.0.2",
				Port:                              8888,
				BaseURL:                           "https://shortener.com",
				RandomStringLength:                6,
				RandomStringMaxLength:             255,
				RandomStringMaxGenerationAttempts: 1000,
				LogLevel:                          "warn",
				FileStoragePath:                   fileStoragePath,
				DatabaseDSN:                       databaseDSN,
			},
		},
		{
			name: "Env: file storage path",
			env: map[string]string{
				"FILE_STORAGE_PATH": "test.json",
			},
			want: Config{
				Host:                              "127.0.0.1",
				Port:                              8080,
				BaseURL:                           "http://127.0.0.1:8080",
				RandomStringLength:                6,
				RandomStringMaxLength:             255,
				RandomStringMaxGenerationAttempts: 1000,
				LogLevel:                          "info",
				FileStoragePath:                   "test.json",
				DatabaseDSN:                       databaseDSN,
			},
		},
		{
			name: "Env: database DSN",
			env: map[string]string{
				"DATABASE_DSN": "test",
			},
			want: Config{
				Host:                              "127.0.0.1",
				Port:                              8080,
				BaseURL:                           "http://127.0.0.1:8080",
				RandomStringLength:                6,
				RandomStringMaxLength:             255,
				RandomStringMaxGenerationAttempts: 1000,
				LogLevel:                          "info",
				FileStoragePath:                   fileStoragePath,
				DatabaseDSN:                       "test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Устанавливаем переменные окружения для теста
			for key, value := range tt.env {
				t.Setenv(key, value)
			}

			// Сначала применяем настройки из флагов
			cfg, err := InitConfigWithArgs(tt.args)
			assert.NoError(t, err)

			// Затем применяем настройки из переменных окружения
			cfg, err = initConfigWithEnv(cfg)
			assert.NoError(t, err)

			assert.Equal(t, *cfg, tt.want)
		})
	}
}
