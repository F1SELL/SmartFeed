package kafka

import (
	"context"
	"testing"
)

func TestEnsureTopicEmptyBrokers(t *testing.T) {
	if err := EnsureTopic(context.Background(), nil, "post_created", 1, 1); err == nil {
		t.Fatal("expected error for empty brokers")
	}
}
