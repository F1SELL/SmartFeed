package service

import (
	"context"
	"errors"
	"fmt"

	"SmartFeed/internal/domain"
)

type UserService struct {
	users         userRepository
	subscriptions subscriptionRepository
}

type userRepository interface {
	GetByID(ctx context.Context, userID int64) (*domain.User, error)
}

type subscriptionRepository interface {
	Create(ctx context.Context, followerID, followeeID int64) error
}

func NewUserService(users userRepository, subscriptions subscriptionRepository) *UserService {
	return &UserService{users: users, subscriptions: subscriptions}
}

func (s *UserService) GetMe(ctx context.Context, userID int64) (*domain.User, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("service.user.get_me: %w", err)
	}
	return user, nil
}

func (s *UserService) Follow(ctx context.Context, followerID, followeeID int64) error {
	if followerID == followeeID {
		return errors.New("cannot follow yourself")
	}
	if _, err := s.users.GetByID(ctx, followeeID); err != nil {
		return fmt.Errorf("service.user.follow.followee_not_found: %w", err)
	}
	if err := s.subscriptions.Create(ctx, followerID, followeeID); err != nil {
		return fmt.Errorf("service.user.follow.create_subscription: %w", err)
	}
	return nil
}
