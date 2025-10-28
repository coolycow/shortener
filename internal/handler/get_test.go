package handler

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/coolycow/shortener/internal/repository"
	"github.com/stretchr/testify/assert"
)

// setupFullRepository возвращает репозиторий с заранее сохраненными ключами и соответствующими им URL
func setupFullRepository() repository.URLRepository {
	repo := repository.NewDoubleMapsRepository()

	defaultURLs := map[string]string{
		"":          "https://mail.ru",
		"AAbbCC1":   "https://google.com",
		"AAbbCC2":   "https://rambler.com",
		"AAbbCC3":   "https://yandex.com",
		"AAbbCC3$#": "https:/vk.com",
	}

	for id, url := range defaultURLs {
		repo.SaveURL(id, url)
	}

	return repo
}

func TestGetHandler(t *testing.T) {
	type want struct {
		code     int
		location string
	}

	tests := []struct {
		name   string
		method string
		key    string
		want   want
	}{
		{
			name:   "Valid URL",
			method: "GET",
			key:    "AAbbCC1",
			want: want{
				code:     307,
				location: "https://google.com",
			},
		},
		{
			name:   "ID with params",
			method: "GET",
			key:    "AAbbCC2?param=value",
			want: want{
				code:     307,
				location: "https://rambler.com",
			},
		},
		{
			name:   "ID with special chars",
			method: "GET",
			key:    "AAbbCC3$#",
			want: want{
				code:     307,
				location: "https:/vk.com",
			},
		},
		{
			name:   "Start spaces",
			method: "GET",
			key:    "%20%20%20%20AAbbCC1",
			want: want{
				code:     307,
				location: "https://google.com",
			},
		},
		{
			name:   "End spaces",
			method: "GET",
			key:    "AAbbCC2%20%20%20%20",
			want: want{
				code:     307,
				location: "https://rambler.com",
			},
		},
		{
			name:   "Method not allowed",
			method: "POST",
			key:    "AAbbCC2",
			want: want{
				code:     405,
				location: "",
			},
		},
		{
			name:   "Empty ID",
			method: "GET",
			key:    "",
			want: want{
				code:     400,
				location: "",
			},
		},
		{
			name:   "Not found ID",
			method: "GET",
			key:    "AAbbCC4",
			want: want{
				code:     404,
				location: "",
			},
		},
		{
			name:   "Incorrect URL",
			method: "GET",
			key:    "AAbbCC4/extrapath",
			want: want{
				code:     404,
				location: "",
			},
		},
		{
			name:   "ID with spaces",
			method: "GET",
			key:    "AAbbCC4%20extrapath1%20extrapath2",
			want: want{
				code:     404,
				location: "",
			},
		},
		{
			name:   "Long ID",
			method: "GET",
			key:    strings.Repeat("C", 1000),
			want: want{
				code:     404,
				location: "",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, "/"+test.key, nil)

			w := httptest.NewRecorder()
			GetHandler(setupFullRepository())(w, request)

			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, test.want.code, result.StatusCode)
			assert.Equal(t, test.want.location, result.Header.Get("Location"))
		})
	}
}
