package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"SmartFeed/internal/domain"
)

type authRepoStub struct {
	user *domain.User
}

func (s *authRepoStub) Create(_ context.Context, user *domain.User) error {
	user.ID = 1
	s.user = user
	return nil
}

func (s *authRepoStub) GetByUsername(_ context.Context, _ string) (*domain.User, error) {
	return s.user, nil
}

func TestAuthService_RegisterAndLogin(t *testing.T) {
	repo := &authRepoStub{}
	svc := NewAuthService(repo, "secret")

	registered, err := svc.Register(context.Background(), "john", "john@example.com", "123456")
	require.NoError(t, err)
	require.Equal(t, int64(1), registered.ID)
	require.NotEmpty(t, registered.PasswordHash)

	token, user, err := svc.Login(context.Background(), "john", "123456")
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.Equal(t, registered.ID, user.ID)
}
