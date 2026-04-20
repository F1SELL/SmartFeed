package kafka

import (
	"context"
	"testing"
	"time"

	"SmartFeed/internal/domain"
)

func TestProducerPublishError(t *testing.T) {
	p := NewPostEventProducer([]string{"127.0.0.1:1"}, "post_created")
	defer p.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := p.PublishPostCreated(ctx, domain.PostCreatedEvent{PostID: 1, AuthorID: 1, Content: "x"})
	if err == nil {
		t.Fatal("expected publish error")
	}
}

func TestNewPostEventReader(t *testing.T) {
	r := NewPostEventReader([]string{"127.0.0.1:1"}, "post_created", "g1")
	if r == nil {
		t.Fatal("reader is nil")
	}
	_ = r.Close()
}
