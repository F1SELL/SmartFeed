package service

import (
	"context"
	"fmt"

	"SmartFeed/internal/domain"
)

type FeedService struct {
	feedCache feedCacheRepository
	posts     feedPostRepository
}

type feedCacheRepository interface {
	GetPostIDs(ctx context.Context, userID int64, limit, offset int64) ([]int64, error)
}

type feedPostRepository interface {
	GetByIDs(ctx context.Context, postIDs []int64) ([]domain.Post, error)
}

func NewFeedService(feedCache feedCacheRepository, posts feedPostRepository) *FeedService {
	return &FeedService{feedCache: feedCache, posts: posts}
}

func (s *FeedService) GetFeed(ctx context.Context, userID int64, limit, offset int64) ([]domain.Post, error) {
	postIDs, err := s.feedCache.GetPostIDs(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("service.feed.get.ids: %w", err)
	}
	posts, err := s.posts.GetByIDs(ctx, postIDs)
	if err != nil {
		return nil, fmt.Errorf("service.feed.get.posts: %w", err)
	}
	return posts, nil
}
