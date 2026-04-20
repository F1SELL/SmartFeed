package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"SmartFeed/internal/domain"
	"SmartFeed/pkg/hash"
	jwtpkg "SmartFeed/pkg/jwt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type AuthService struct {
	users     authUserRepository
	jwtSecret string
}

type authUserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
}

func NewAuthService(users authUserRepository, jwtSecret string) *AuthService {
	return &AuthService{users: users, jwtSecret: jwtSecret}
}

func (s *AuthService) Register(ctx context.Context, username, email, password string) (*domain.User, error) {
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(strings.ToLower(email))
	if username == "" || email == "" || len(password) < 6 {
		return nil, errors.New("username/email required and password must be at least 6 chars")
	}

	passwordHash, err := hash.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("service.auth.register.hash: %w", err)
	}

	user := &domain.User{
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         domain.RoleUser,
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("service.auth.register.create: %w", err)
	}
	return user, nil
}

func (s *AuthService) Login(ctx context.Context, username, password string) (string, *domain.User, error) {
	user, err := s.users.GetByUsername(ctx, strings.TrimSpace(username))
	if err != nil {
		return "", nil, ErrInvalidCredentials
	}
	if !hash.CheckPassword(password, user.PasswordHash) {
		return "", nil, ErrInvalidCredentials
	}

	token, err := jwtpkg.GenerateToken(user.ID, string(user.Role), s.jwtSecret)
	if err != nil {
		return "", nil, fmt.Errorf("service.auth.login.generate_token: %w", err)
	}

	return token, user, nil
}
