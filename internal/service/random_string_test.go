package service

import (
	"context"
	"testing"

	"github.com/coolycow/shortener/internal/repository"
)

func BenchmarkGenerate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = generate(8)
	}
}

func BenchmarkCreateUniqueStringForURL(b *testing.B) {
	ctx := context.Background()
	repo := repository.NewDoubleMapsRepository("")

	for i := 0; i < b.N; i++ {
		_, _ = CreateUniqueStringForURL(ctx, repo, 8, 16, 100)
	}
}
