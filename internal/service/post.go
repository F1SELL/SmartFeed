package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"SmartFeed/internal/domain"
)

type PostProducer interface {
	PublishPostCreated(ctx context.Context, event domain.PostCreatedEvent) error
}

type PostService struct {
	posts    postRepository
	producer PostProducer
}

type postRepository interface {
	Create(ctx context.Context, post *domain.Post) error
}

func NewPostService(posts postRepository, producer PostProducer) *PostService {
	return &PostService{posts: posts, producer: producer}
}

func (s *PostService) Create(ctx context.Context, authorID int64, content string) (*domain.Post, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, errors.New("content cannot be empty")
	}
	if len(content) > 5000 {
		return nil, errors.New("content too large")
	}

	post := &domain.Post{AuthorID: authorID, Content: content, Tags: []string{}}
	if err := s.posts.Create(ctx, post); err != nil {
		return nil, fmt.Errorf("service.post.create.store: %w", err)
	}

	event := domain.PostCreatedEvent{
		PostID:     post.ID,
		AuthorID:   post.AuthorID,
		Content:    post.Content,
		CreatedAt:  post.CreatedAt,
		OccurredAt: time.Now().UTC(),
	}
	if err := s.producer.PublishPostCreated(ctx, event); err != nil {
		return nil, fmt.Errorf("service.post.create.publish: %w", err)
	}

	return post, nil
}
