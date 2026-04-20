package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"SmartFeed/internal/domain"
	"SmartFeed/internal/service"
	mw "SmartFeed/internal/transport/http/middleware"
	jwtpkg "SmartFeed/pkg/jwt"
)

type authRepoForHandler struct {
	user *domain.User
}

func (s *authRepoForHandler) Create(_ context.Context, user *domain.User) error {
	if s.user != nil {
		return errors.New("exists")
	}
	user.ID = 10
	s.user = user
	return nil
}

func (s *authRepoForHandler) GetByUsername(_ context.Context, username string) (*domain.User, error) {
	if s.user == nil || s.user.Username != username {
		return nil, errors.New("not found")
	}
	return s.user, nil
}

type postRepoForHandler struct {
	err error
}

func (s *postRepoForHandler) Create(_ context.Context, post *domain.Post) error {
	if s.err != nil {
		return s.err
	}
	post.ID = 1
	post.CreatedAt = time.Now().UTC()
	return nil
}

type producerForHandler struct {
	err error
}

func (s *producerForHandler) PublishPostCreated(_ context.Context, _ domain.PostCreatedEvent) error {
	return s.err
}

type feedCacheForHandler struct {
	ids []int64
	err error
}

func (s *feedCacheForHandler) GetPostIDs(_ context.Context, _ int64, _, _ int64) ([]int64, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.ids, nil
}

type feedPostsForHandler struct {
	posts []domain.Post
	err   error
}

func (s *feedPostsForHandler) GetByIDs(_ context.Context, _ []int64) ([]domain.Post, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.posts, nil
}

type userRepoForHandler struct {
	err error
}

func (s *userRepoForHandler) GetByID(_ context.Context, userID int64) (*domain.User, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &domain.User{ID: userID, Username: "u"}, nil
}

type subRepoForHandler struct {
	err error
}

func (s *subRepoForHandler) Create(_ context.Context, _, _ int64) error {
	return s.err
}

func authRequest(method, path string, body []byte) *http.Request {
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	token, _ := jwtpkg.GenerateToken(11, "user", "secret")
	req.Header.Set("Authorization", "Bearer "+token)
	return req
}

func withAuth(handler http.HandlerFunc) http.Handler {
	return mw.Auth("secret")(handler)
}

func TestHelpersAndParseInt(t *testing.T) {
	rr := httptest.NewRecorder()
	writeJSON(rr, http.StatusCreated, map[string]string{"ok": "1"})
	if rr.Code != http.StatusCreated {
		t.Fatalf("unexpected status: %d", rr.Code)
	}

	rr2 := httptest.NewRecorder()
	writeError(rr2, http.StatusBadRequest, "bad")
	if rr2.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", rr2.Code)
	}

	if parseInt64("", 20) != 20 || parseInt64("abc", 20) != 20 || parseInt64("11", 20) != 11 {
		t.Fatal("parseInt64 fallback logic failed")
	}
}

func TestAuthHandlerRegisterAndLogin(t *testing.T) {
	repo := &authRepoForHandler{}
	h := NewAuthHandler(service.NewAuthService(repo, "secret"))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(`{"username":"john","email":"john@example.com","password":"123456"}`))
	h.Register(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}

	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"username":"john","password":"123456"}`))
	h.Login(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr2.Code)
	}

	rr3 := httptest.NewRecorder()
	req3 := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"username":"john","password":"bad"}`))
	h.Login(rr3, req3)
	if rr3.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr3.Code)
	}
}

func TestPostHandlerCreate(t *testing.T) {
	h := NewPostHandler(service.NewPostService(&postRepoForHandler{}, &producerForHandler{}))
	wrapped := withAuth(h.Create)

	rr := httptest.NewRecorder()
	req := authRequest(http.MethodPost, "/api/v1/posts", []byte(`{"content":"hello"}`))
	wrapped.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}

	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/posts", bytes.NewBufferString(`bad`))
	wrapped.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d", rr2.Code)
	}
}

func TestFeedHandlerList(t *testing.T) {
	feedSvc := service.NewFeedService(&feedCacheForHandler{ids: []int64{1}}, &feedPostsForHandler{posts: []domain.Post{{ID: 1}}})
	h := NewFeedHandler(feedSvc)
	wrapped := withAuth(h.List)

	rr := httptest.NewRecorder()
	req := authRequest(http.MethodGet, "/api/v1/feed?limit=10&offset=0", nil)
	wrapped.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var out map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
}

func TestUserHandlerMeAndFollow(t *testing.T) {
	h := NewUserHandler(service.NewUserService(&userRepoForHandler{}, &subRepoForHandler{}))
	meWrapped := withAuth(h.Me)

	rr := httptest.NewRecorder()
	req := authRequest(http.MethodGet, "/api/v1/users/me", nil)
	meWrapped.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	followWrapped := withAuth(h.Follow)
	rr2 := httptest.NewRecorder()
	req2 := authRequest(http.MethodPost, "/api/v1/users/follow/12", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "12")
	req2 = req2.WithContext(context.WithValue(req2.Context(), chi.RouteCtxKey, rctx))
	followWrapped.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr2.Code)
	}
}
