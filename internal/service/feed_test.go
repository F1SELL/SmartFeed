package service

import (
	"context"
	"errors"
	"testing"

	"SmartFeed/internal/domain"
)

type feedCacheStub struct {
	ids []int64
	err error
}

func (s *feedCacheStub) GetPostIDs(_ context.Context, _ int64, _, _ int64) ([]int64, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.ids, nil
}

type feedPostStub struct {
	posts []domain.Post
	err   error
}

func (s *feedPostStub) GetByIDs(_ context.Context, _ []int64) ([]domain.Post, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.posts, nil
}

func TestFeedService_GetFeedCacheError(t *testing.T) {
	svc := NewFeedService(&feedCacheStub{err: errors.New("redis fail")}, &feedPostStub{})
	if _, err := svc.GetFeed(context.Background(), 1, 20, 0); err == nil {
		t.Fatal("expected cache error")
	}
}

func TestFeedService_GetFeedPostsError(t *testing.T) {
	svc := NewFeedService(&feedCacheStub{ids: []int64{1}}, &feedPostStub{err: errors.New("db fail")})
	if _, err := svc.GetFeed(context.Background(), 1, 20, 0); err == nil {
		t.Fatal("expected posts error")
	}
}

func TestFeedService_GetFeedOK(t *testing.T) {
	expected := []domain.Post{{ID: 1, Content: "ok"}}
	svc := NewFeedService(&feedCacheStub{ids: []int64{1}}, &feedPostStub{posts: expected})

	posts, err := svc.GetFeed(context.Background(), 1, 20, 0)
	if err != nil {
		t.Fatalf("unexpected get feed error: %v", err)
	}
	if len(posts) != 1 || posts[0].ID != 1 {
		t.Fatalf("unexpected posts: %+v", posts)
	}
}
