package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want Config
	}{
		{
			name: "Default",
			want: Config{
				Host:                              "127.0.0.1",
				Port:                              8080,
				BaseURL:                           "http://127.0.0.1:8080",
				RandomStringLength:                6,
				RandomStringMaxLength:             100,
				RandomStringMaxGenerationAttempts: 1000,
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
				RandomStringMaxLength:             100,
				RandomStringMaxGenerationAttempts: 1000,
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
				RandomStringMaxLength:             100,
				RandomStringMaxGenerationAttempts: 1000,
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
				RandomStringMaxLength:             100,
				RandomStringMaxGenerationAttempts: 1000,
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
				RandomStringMaxLength:             100,
				RandomStringMaxGenerationAttempts: 1000,
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
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := InitConfigWithArgs(tt.args)

			assert.NoError(t, err)
			assert.Equal(t, *cfg, tt.want)
		})
	}
}
