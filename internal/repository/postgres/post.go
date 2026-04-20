package postgres

import (
	"context"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"

	"SmartFeed/internal/domain"
)

type PostRepository struct {
	db *pgxpool.Pool
}

func NewPostRepository(db *pgxpool.Pool) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) Create(ctx context.Context, post *domain.Post) error {
	query := `
		INSERT INTO posts (author_id, content, tags)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	if err := r.db.QueryRow(ctx, query, post.AuthorID, post.Content, post.Tags).Scan(&post.ID, &post.CreatedAt); err != nil {
		return fmt.Errorf("repository.post.create: %w", err)
	}

	return nil
}

func (r *PostRepository) GetByIDs(ctx context.Context, postIDs []int64) ([]domain.Post, error) {
	if len(postIDs) == 0 {
		return []domain.Post{}, nil
	}

	query := `
		SELECT id, author_id, content, tags, created_at
		FROM posts
		WHERE id = ANY($1)
	`

	rows, err := r.db.Query(ctx, query, postIDs)
	if err != nil {
		return nil, fmt.Errorf("repository.post.get_by_ids.query: %w", err)
	}
	defer rows.Close()

	postsMap := make(map[int64]domain.Post, len(postIDs))
	for rows.Next() {
		var p domain.Post
		if err := rows.Scan(&p.ID, &p.AuthorID, &p.Content, &p.Tags, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("repository.post.get_by_ids.scan: %w", err)
		}
		postsMap[p.ID] = p
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository.post.get_by_ids.rows: %w", err)
	}

	ordered := make([]domain.Post, 0, len(postsMap))
	for _, id := range postIDs {
		if p, ok := postsMap[id]; ok {
			ordered = append(ordered, p)
		}
	}

	return ordered, nil
}

func (r *PostRepository) ListRecentByAuthorIDs(ctx context.Context, authorIDs []int64, limit int) ([]domain.Post, error) {
	if len(authorIDs) == 0 {
		return []domain.Post{}, nil
	}
	if limit <= 0 {
		limit = 20
	}

	query := `
		SELECT id, author_id, content, tags, created_at
		FROM posts
		WHERE author_id = ANY($1)
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, authorIDs, limit)
	if err != nil {
		return nil, fmt.Errorf("repository.post.list_recent.query: %w", err)
	}
	defer rows.Close()

	posts := make([]domain.Post, 0, limit)
	for rows.Next() {
		var p domain.Post
		if err := rows.Scan(&p.ID, &p.AuthorID, &p.Content, &p.Tags, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("repository.post.list_recent.scan: %w", err)
		}
		posts = append(posts, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository.post.list_recent.rows: %w", err)
	}

	sort.SliceStable(posts, func(i, j int) bool {
		return posts[i].CreatedAt.After(posts[j].CreatedAt)
	})

	return posts, nil
}

func (r *PostRepository) UpdateTags(ctx context.Context, postID int64, tags []string) error {
	query := `UPDATE posts SET tags = $1 WHERE id = $2`
	if _, err := r.db.Exec(ctx, query, tags, postID); err != nil {
		return fmt.Errorf("repository.post.update_tags: %w", err)
	}
	return nil
}
