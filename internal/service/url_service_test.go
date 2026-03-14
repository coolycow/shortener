package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/coolycow/shortener/internal/config"
	"github.com/coolycow/shortener/internal/repository"
)

func benchConfig() *config.Config {
	return &config.Config{
		BaseURL:                           "http://127.0.0.1:8080",
		RandomStringLength:                6,
		RandomStringMaxLength:             255,
		RandomStringMaxGenerationAttempts: 1000,
	}
}

func BenchmarkURLService_CreateShortURL(b *testing.B) {
	ctx := context.Background()
	cfg := benchConfig()
	repo := repository.NewDoubleMapsRepository("")
	defer func() { _ = repo.Close() }()
	svc := NewURLService(cfg, repo)
	userID := "bench-user"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.CreateShortURL(ctx, userID, fmt.Sprintf("https://example.com/%d", i))
	}
}

func BenchmarkURLService_GetShortURL(b *testing.B) {
	ctx := context.Background()
	cfg := benchConfig()
	repo := repository.NewDoubleMapsRepository("")
	defer func() { _ = repo.Close() }()
	svc := NewURLService(cfg, repo)
	userID := "bench-user"
	shortURL, _ := svc.CreateShortURL(ctx, userID, "https://example.com/unique")
	key := shortURL[len(cfg.BaseURL)+1:]

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.GetShortURL(ctx, key)
	}
}

func BenchmarkURLService_GetManyShortURLs(b *testing.B) {
	ctx := context.Background()
	cfg := benchConfig()
	repo := repository.NewDoubleMapsRepository("")
	defer func() { _ = repo.Close() }()
	svc := NewURLService(cfg, repo)
	userID := "bench-user"
	for i := 0; i < 100; i++ {
		_, _ = svc.CreateShortURL(ctx, userID, "https://example.com/"+string(rune('a'+i%26)))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.GetManyShortURLs(ctx, userID)
	}
}
