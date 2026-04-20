package domain

import (
	"time"
)

// Role определяет права доступа пользователя (RBAC)
type Role string

const (
	RoleUser      Role = "user"
	RoleModerator Role = "moderator"
	RoleAdmin     Role = "admin"
)

// User — доменная модель пользователя
type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

// Post — доменная модель поста
type Post struct {
	ID        int64     `json:"id"`
	AuthorID  int64     `json:"author_id"`
	Content   string    `json:"content"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
}

// Subscription — доменная модель подписки
type Subscription struct {
	FollowerID int64     `json:"follower_id"`
	FolloweeID int64     `json:"followee_id"`
	CreatedAt  time.Time `json:"created_at"`
}

// PostCreatedEvent — доменная модель события создания поста
type PostCreatedEvent struct {
	PostID     int64     `json:"post_id"`
	AuthorID   int64     `json:"author_id"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
	OccurredAt time.Time `json:"occurred_at"`
}
