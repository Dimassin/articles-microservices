package postgres

import (
	"blog/internal/domain"
	"context"
	"database/sql"
	"errors"
	"time"

	_ "github.com/lib/pq"
)

type ArticleRepository struct {
	db *sql.DB
}

func NewArticleRepository(db *sql.DB) *ArticleRepository {
	return &ArticleRepository{
		db: db,
	}
}

func (r *ArticleRepository) Create(ctx context.Context, article *domain.Article) error {
	query := `
		INSERT INTO articles (id, title, content, author_id, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query, article.Title, article.Content, article.AuthorID).
		Scan(&article.ID, &article.CreatedAt, &article.UpdatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (r *ArticleRepository) FindByID(ctx context.Context, id string) (*domain.Article, error) {
	article := &domain.Article{}

	query := `SELECT id, title, content, author_id, created_at, updated_at FROM articles WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&article.ID, &article.Title, &article.Content, &article.AuthorID, &article.CreatedAt, &article.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrArticleNotFound
	}
	if err != nil {
		return article, err
	}

	return article, nil
}

func (r *ArticleRepository) FindAll(ctx context.Context, limit, offset int) ([]*domain.Article, error) {
	query := `SELECT id, title, content, author_id, created_at, updated_at 
              FROM articles 
              ORDER BY created_at DESC 
              LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []*domain.Article

	for rows.Next() {
		article := &domain.Article{}
		err := rows.Scan(
			&article.ID,
			&article.Title,
			&article.Content,
			&article.AuthorID,
			&article.CreatedAt,
			&article.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		articles = append(articles, article)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return articles, nil
}

func (r *ArticleRepository) Update(ctx context.Context, id string, req *domain.UpdateArticleRequest) error {
	query := `
        UPDATE articles 
        SET title = $1, content = $2, updated_at = NOW()
        WHERE id = $3
        RETURNING updated_at
    `

	var updatedAt time.Time
	err := r.db.QueryRowContext(ctx, query,
		req.Title,
		req.Content,
		id,
	).Scan(&updatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrArticleNotFound
		}
		return err
	}

	return nil
}

func (r *ArticleRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM articles WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrArticleNotFound
	}

	return nil
}
