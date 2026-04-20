package httptransport

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"SmartFeed/internal/domain"
	"SmartFeed/internal/service"
	"SmartFeed/internal/transport/http/handlers"
)

type authRepoStub struct {
	user *domain.User
}

func (s *authRepoStub) Create(_ context.Context, user *domain.User) error {
	user.ID = 1
	s.user = user
	return nil
}

func (s *authRepoStub) GetByUsername(_ context.Context, username string) (*domain.User, error) {
	if s.user == nil || s.user.Username != username {
		return nil, errors.New("not found")
	}
	return s.user, nil
}

type userRepoStub struct{}

func (s *userRepoStub) GetByID(_ context.Context, userID int64) (*domain.User, error) {
	return &domain.User{ID: userID, Username: "u"}, nil
}

type subRepoStub struct{}

func (s *subRepoStub) Create(_ context.Context, _, _ int64) error { return nil }

type postRepoStub struct{}

func (s *postRepoStub) Create(_ context.Context, p *domain.Post) error {
	p.ID = 1
	p.CreatedAt = time.Now().UTC()
	return nil
}

type producerStub struct{}

func (s *producerStub) PublishPostCreated(_ context.Context, _ domain.PostCreatedEvent) error {
	return nil
}

type feedCacheStub struct{}

func (s *feedCacheStub) GetPostIDs(_ context.Context, _ int64, _, _ int64) ([]int64, error) {
	return []int64{}, nil
}

type feedPostStub struct{}

func (s *feedPostStub) GetByIDs(_ context.Context, _ []int64) ([]domain.Post, error) {
	return []domain.Post{}, nil
}

func makeHandlers() Handlers {
	authSvc := service.NewAuthService(&authRepoStub{}, "secret")
	userSvc := service.NewUserService(&userRepoStub{}, &subRepoStub{})
	postSvc := service.NewPostService(&postRepoStub{}, &producerStub{})
	feedSvc := service.NewFeedService(&feedCacheStub{}, &feedPostStub{})

	return Handlers{
		Auth: handlers.NewAuthHandler(authSvc),
		User: handlers.NewUserHandler(userSvc),
		Post: handlers.NewPostHandler(postSvc),
		Feed: handlers.NewFeedHandler(feedSvc),
	}
}

func TestRouterHealthAndMetrics(t *testing.T) {
	r := NewRouter(makeHandlers(), "secret")

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr2 := httptest.NewRecorder()
	r.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr2.Code)
	}
}

func TestRouterPrivateUnauthorized(t *testing.T) {
	r := NewRouter(makeHandlers(), "secret")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestRouterAuthRegister(t *testing.T) {
	r := NewRouter(makeHandlers(), "secret")
	body := bytes.NewBufferString(`{"username":"john","email":"john@example.com","password":"123456"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}
}
