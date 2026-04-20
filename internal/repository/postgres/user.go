package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"SmartFeed/internal/domain"
)

// ErrUserNotFound можно вынести в доменные ошибки позже, но пока оставим здесь для наглядности
var ErrUserNotFound = errors.New("user not found")

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// Create сохраняет нового пользователя и возвращает его с заполненным ID и CreatedAt
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (username, email, password_hash, role) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id, created_at
	`

	err := r.db.QueryRow(ctx, query, user.Username, user.Email, user.PasswordHash, user.Role).
		Scan(&user.ID, &user.CreatedAt)

	if err != nil {
		return fmt.Errorf("repository.User.Create: %w", err)
	}

	return nil
}

// GetByUsername ищет пользователя по логину
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, role, created_at 
		FROM users 
		WHERE username = $1
	`

	var user domain.User
	err := r.db.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("repository.User.GetByUsername: %w", err)
	}

	return &user, nil
}

// GetByID ищет пользователя по ID
func (r *UserRepository) GetByID(ctx context.Context, userID int64) (*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, role, created_at 
		FROM users 
		WHERE id = $1
	`

	var user domain.User
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("repository.User.GetByID: %w", err)
	}

	return &user, nil
}
