package usecase

import (
	"blog/internal/domain"
	"blog/internal/ports"
	"context"
)

type ArticleUsecase struct {
	articleRepo    ports.ArticleRepository
	tokenValidator ports.TokenValidator
}

func NewArticleUsecase(
	repo ports.ArticleRepository,
	validator ports.TokenValidator,
) *ArticleUsecase {
	return &ArticleUsecase{
		articleRepo:    repo,
		tokenValidator: validator,
	}
}

func (uc *ArticleUsecase) CreateArticle(ctx context.Context, token string, req *domain.CreateArticleRequest) (*domain.ArticleResponse, error) {
	authorID, err := uc.tokenValidator.Validate(ctx, token)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	article := &domain.Article{
		Title:    req.Title,
		Content:  req.Content,
		AuthorID: authorID,
	}

	err = uc.articleRepo.Create(ctx, article)
	if err != nil {
		return nil, err
	}

	return &domain.ArticleResponse{
		ID:        article.ID,
		Title:     article.Title,
		Content:   article.Content,
		AuthorID:  article.AuthorID,
		CreatedAt: article.CreatedAt,
		UpdatedAt: article.UpdatedAt,
	}, nil
}

func (uc *ArticleUsecase) GetArticle(ctx context.Context, id string) (*domain.ArticleResponse, error) {
	article, err := uc.articleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &domain.ArticleResponse{
		ID:        article.ID,
		Title:     article.Title,
		Content:   article.Content,
		AuthorID:  article.AuthorID,
		CreatedAt: article.CreatedAt,
		UpdatedAt: article.UpdatedAt,
	}, nil
}

func (uc *ArticleUsecase) ListArticles(ctx context.Context, limit, offset int) ([]*domain.ArticleResponse, error) {
	articles, err := uc.articleRepo.FindAll(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	responses := make([]*domain.ArticleResponse, 0, len(articles))

	for _, article := range articles {
		responses = append(responses, &domain.ArticleResponse{
			ID:        article.ID,
			Title:     article.Title,
			Content:   article.Content,
			AuthorID:  article.AuthorID,
			CreatedAt: article.CreatedAt,
			UpdatedAt: article.UpdatedAt,
		})
	}

	return responses, nil
}

func (uc *ArticleUsecase) UpdateArticle(ctx context.Context, token, id string, req *domain.UpdateArticleRequest) (*domain.ArticleResponse, error) {
	authorID, err := uc.tokenValidator.Validate(ctx, token)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	existingArticle, err := uc.articleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if existingArticle.AuthorID != authorID {
		return nil, domain.ErrForbidden
	}

	err = uc.articleRepo.Update(ctx, id, req)
	if err != nil {
		return nil, err
	}

	updatedArticle, err := uc.articleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &domain.ArticleResponse{
		ID:        updatedArticle.ID,
		Title:     updatedArticle.Title,
		Content:   updatedArticle.Content,
		AuthorID:  updatedArticle.AuthorID,
		CreatedAt: updatedArticle.CreatedAt,
		UpdatedAt: updatedArticle.UpdatedAt,
	}, nil
}

func (uc *ArticleUsecase) DeleteArticle(ctx context.Context, token, id string) error {
	authorID, err := uc.tokenValidator.Validate(ctx, token)
	if err != nil {
		return domain.ErrInvalidToken
	}

	existingArticle, err := uc.articleRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if existingArticle.AuthorID != authorID {
		return domain.ErrForbidden
	}

	err = uc.articleRepo.Delete(ctx, id)
	if err != nil {
		return err
	}

	return nil
}
