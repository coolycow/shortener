package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/coolycow/shortener/internal/config"
	"github.com/coolycow/shortener/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostHandler(t *testing.T) {
	type want struct {
		code        int
		contentType string
		body        string // Добавляем поле для проверки тела ответа
	}

	tests := []struct {
		name        string
		method      string
		contentType string
		body        string
		want        want
	}{
		{
			name:        "Create new",
			method:      http.MethodPost,
			body:        "https://example.com",
			contentType: "text/plain",
			want: want{
				code:        201,
				contentType: "text/plain",
				body:        "",
			},
		},
		{
			name:        "URL without domain",
			method:      http.MethodPost,
			body:        "https://example",
			contentType: "text/plain",
			want: want{
				code:        201,
				contentType: "text/plain",
				body:        "",
			},
		},
		{
			name:        "Content type with charset",
			method:      http.MethodPost,
			body:        "https://example.com",
			contentType: "text/plain; charset=utf-8",
			want: want{
				code:        201,
				contentType: "text/plain",
				body:        "",
			},
		},
		{
			name:        "URL with params",
			method:      http.MethodPost,
			body:        "https://example.com?a=1,2,3&b=true&c=filter[abd]",
			contentType: "text/plain; charset=utf-8",
			want: want{
				code:        201,
				contentType: "text/plain",
				body:        "",
			},
		},
		{
			name:        "Cyrillic URL",
			method:      http.MethodPost,
			body:        "https://пример.рф/путь?параметр=значение",
			contentType: "text/plain; charset=utf-8",
			want: want{
				code:        201,
				contentType: "text/plain",
				body:        "",
			},
		},
		{
			name:        "Invalid method",
			method:      http.MethodGet,
			body:        "https://example.com",
			contentType: "text/plain",
			want: want{
				code:        405,
				contentType: "text/plain",
				body:        "Method not allowed\n",
			},
		},
		{
			name:        "Invalid content type",
			method:      http.MethodPost,
			body:        "https://example.com",
			contentType: "application/json",
			want: want{
				code:        415,
				contentType: "text/plain",
				body:        "Content type not allowed\n",
			},
		},
		{
			name:        "Empty body",
			method:      http.MethodPost,
			body:        "",
			contentType: "text/plain",
			want: want{
				code:        400,
				contentType: "text/plain",
				body:        "Empty body\n",
			},
		},
		{
			name:        "Empty URL",
			method:      http.MethodPost,
			body:        "           ",
			contentType: "text/plain",
			want: want{
				code:        400,
				contentType: "text/plain",
				body:        "Empty URL\n",
			},
		},
		{
			name:        "URL without protocol",
			method:      http.MethodPost,
			body:        "invalidurl.ru",
			contentType: "text/plain",
			want: want{
				code:        400,
				contentType: "text/plain",
				body:        "Invalid URL\n",
			},
		},
	}

	cfg := config.NewConfig()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := repository.NewDoubleMapsRepository()

			w := httptest.NewRecorder()
			request := httptest.NewRequest(test.method, "/", strings.NewReader(test.body))
			request.Header.Set("Content-Type", test.contentType)

			PostHandler(cfg, repo)(w, request)

			result := w.Result()

			// Проверяем, что код ответа и тип контента соответствуют ожиданиям
			assert.Equal(t, test.want.code, result.StatusCode)

			defer result.Body.Close()
			resultBody, err := io.ReadAll(result.Body)

			resultString := string(resultBody)

			require.NoError(t, err)

			// Проверяем, что тело ответа соответствует ожиданиям
			if test.want.code == 201 {
				key := strings.TrimPrefix(resultString, cfg.BaseURL+"/")

				// Получаем из репозитория оригинальную ссылку по ключу короткой ссылки
				originalURL, _ := repo.GetOriginalURL(key)

				// Парсим URL из строки, чтобы корректно сравнивать кириллические адреса
				originalParsedURL, _ := url.Parse(originalURL)
				bodyParsedURL, _ := url.Parse(test.body)

				// Проверяем, что длина репозитория увеличилась на 1
				assert.Equal(t, repo.GetSize(), 1)

				// Проверяем, что оригинальная ссылка и запроса и ссылка из репозитория совпадают
				assert.Equal(t, originalParsedURL.String(), bodyParsedURL.String())

				// Проверяем, что ссылка из ответа соответствует схеме
				assert.Equal(t, resultString, cfg.BaseURL+"/"+key)
			} else if test.want.body != "" {
				assert.Equal(t, test.want.body, resultString)
			}
		})
	}
}
