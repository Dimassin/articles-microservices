package ports

import (
	"blog/internal/domain"
	"context"
)

type ArticleRepository interface {
	Create(ctx context.Context, article *domain.Article) error

	FindByID(ctx context.Context, id string) (*domain.Article, error)

	FindAll(ctx context.Context, limit, offset int) ([]*domain.Article, error)

	Update(ctx context.Context, id string, req *domain.UpdateArticleRequest) error

	Delete(ctx context.Context, id string) error
}
