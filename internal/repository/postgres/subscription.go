package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SubscriptionRepository struct {
	db *pgxpool.Pool
}

func NewSubscriptionRepository(db *pgxpool.Pool) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) Create(ctx context.Context, followerID, followeeID int64) error {
	query := `
		INSERT INTO subscriptions (follower_id, followee_id)
		VALUES ($1, $2)
		ON CONFLICT (follower_id, followee_id) DO NOTHING
	`
	if _, err := r.db.Exec(ctx, query, followerID, followeeID); err != nil {
		return fmt.Errorf("repository.subscription.create: %w", err)
	}
	return nil
}

func (r *SubscriptionRepository) ListFollowerIDs(ctx context.Context, followeeID int64) ([]int64, error) {
	query := `
		SELECT follower_id
		FROM subscriptions
		WHERE followee_id = $1
	`

	rows, err := r.db.Query(ctx, query, followeeID)
	if err != nil {
		return nil, fmt.Errorf("repository.subscription.list_followers.query: %w", err)
	}
	defer rows.Close()

	ids := make([]int64, 0, 128)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("repository.subscription.list_followers.scan: %w", err)
		}
		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository.subscription.list_followers.rows: %w", err)
	}

	return ids, nil
}

func (r *SubscriptionRepository) ListFolloweeIDs(ctx context.Context, followerID int64) ([]int64, error) {
	query := `
		SELECT followee_id
		FROM subscriptions
		WHERE follower_id = $1
	`

	rows, err := r.db.Query(ctx, query, followerID)
	if err != nil {
		return nil, fmt.Errorf("repository.subscription.list_followees.query: %w", err)
	}
	defer rows.Close()

	ids := make([]int64, 0, 128)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("repository.subscription.list_followees.scan: %w", err)
		}
		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository.subscription.list_followees.rows: %w", err)
	}

	return ids, nil
}
