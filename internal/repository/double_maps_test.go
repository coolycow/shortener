package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/coolycow/shortener/internal/model"
)

// uniqueKey генерирует ключ для итерации бенчмарка (в тесте не импортируем service — цикл).
func uniqueKey(i int) string {
	return fmt.Sprintf("k%08d", i)
}

func BenchmarkDoubleMapsRepository_SaveURL(b *testing.B) {
	ctx := context.Background()
	repo := NewDoubleMapsRepository("")
	defer repo.Close()
	userID := "bench-user"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := uniqueKey(i)
		_, _, _ = repo.SaveURL(ctx, userID, fmt.Sprintf("https://example.com/%d", i), key)
	}
}

func BenchmarkDoubleMapsRepository_GetShortURL(b *testing.B) {
	ctx := context.Background()
	repo := NewDoubleMapsRepository("")
	defer repo.Close()
	userID := "bench-user"
	key := uniqueKey(0)
	_, _, _ = repo.SaveURL(ctx, userID, "https://example.com/unique", key)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.GetShortURL(ctx, key)
	}
}

func BenchmarkDoubleMapsRepository_AddURL(b *testing.B) {
	ctx := context.Background()
	repo := NewDoubleMapsRepository("")
	defer repo.Close()
	userID := "bench-user"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := uniqueKey(i)
		_, _, _ = repo.AddURL(ctx, userID, fmt.Sprintf("https://example.com/%d", i), key)
	}
}

func BenchmarkDoubleMapsRepository_GetManyKeys(b *testing.B) {
	ctx := context.Background()
	repo := NewDoubleMapsRepository("")
	defer repo.Close()
	userID := "bench-user"
	urls := make([]model.ShortURL, 10)
	for i := range urls {
		key := uniqueKey(i)
		_, _, _ = repo.SaveURL(ctx, userID, fmt.Sprintf("https://example.com/%d", i), key)
		urls[i] = model.ShortURL{CorrelationID: fmt.Sprintf("c%d", i), OriginalURL: fmt.Sprintf("https://example.com/%d", i)}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.GetManyKeys(ctx, userID, urls)
	}
}

func BenchmarkDoubleMapsRepository_GetManyShortURLs(b *testing.B) {
	ctx := context.Background()
	repo := NewDoubleMapsRepository("")
	defer repo.Close()
	userID := "bench-user"
	for i := 0; i < 50; i++ {
		key := uniqueKey(i)
		_, _, _ = repo.SaveURL(ctx, userID, fmt.Sprintf("https://example.com/%d", i), key)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.GetManyShortURLs(ctx, userID)
	}
}
