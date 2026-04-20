package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"SmartFeed/internal/domain"
)

type postRepoStub struct {
	err error
}

func (s *postRepoStub) Create(_ context.Context, post *domain.Post) error {
	if s.err != nil {
		return s.err
	}
	post.ID = 777
	post.CreatedAt = time.Unix(1700000000, 0).UTC()
	return nil
}

type producerStub struct {
	err       error
	published bool
}

func (s *producerStub) PublishPostCreated(_ context.Context, _ domain.PostCreatedEvent) error {
	s.published = true
	return s.err
}

func TestPostService_CreateValidation(t *testing.T) {
	svc := NewPostService(&postRepoStub{}, &producerStub{})

	if _, err := svc.Create(context.Background(), 1, "   "); err == nil {
		t.Fatal("expected empty content error")
	}
	if _, err := svc.Create(context.Background(), 1, strings.Repeat("a", 5001)); err == nil {
		t.Fatal("expected too large content error")
	}
}

func TestPostService_CreateStoreError(t *testing.T) {
	svc := NewPostService(&postRepoStub{err: errors.New("db down")}, &producerStub{})
	if _, err := svc.Create(context.Background(), 1, "hello"); err == nil {
		t.Fatal("expected store error")
	}
}

func TestPostService_CreatePublishError(t *testing.T) {
	svc := NewPostService(&postRepoStub{}, &producerStub{err: errors.New("kafka down")})
	if _, err := svc.Create(context.Background(), 1, "hello"); err == nil {
		t.Fatal("expected publish error")
	}
}

func TestPostService_CreateOK(t *testing.T) {
	producer := &producerStub{}
	svc := NewPostService(&postRepoStub{}, producer)

	post, err := svc.Create(context.Background(), 2, "  hello world  ")
	if err != nil {
		t.Fatalf("unexpected create error: %v", err)
	}
	if post.ID == 0 {
		t.Fatal("expected post id to be assigned")
	}
	if post.Content != "hello world" {
		t.Fatalf("unexpected trimmed content: %q", post.Content)
	}
	if !producer.published {
		t.Fatal("expected event to be published")
	}
}
