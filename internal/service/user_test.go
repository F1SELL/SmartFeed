package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"SmartFeed/internal/domain"
)

type userRepoStub struct{}

func (s *userRepoStub) GetByID(_ context.Context, userID int64) (*domain.User, error) {
	return &domain.User{ID: userID, Username: "u"}, nil
}

type subRepoStub struct {
	created bool
}

func (s *subRepoStub) Create(_ context.Context, _, _ int64) error {
	s.created = true
	return nil
}

func TestUserService_FollowSelfFails(t *testing.T) {
	subs := &subRepoStub{}
	svc := NewUserService(&userRepoStub{}, subs)

	err := svc.Follow(context.Background(), 10, 10)
	require.Error(t, err)
	require.False(t, subs.created)
}

func TestUserService_FollowOK(t *testing.T) {
	subs := &subRepoStub{}
	svc := NewUserService(&userRepoStub{}, subs)

	err := svc.Follow(context.Background(), 10, 11)
	require.NoError(t, err)
	require.True(t, subs.created)
}
