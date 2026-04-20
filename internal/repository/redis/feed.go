package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type FeedRepository struct {
	client *goredis.Client
}

func NewFeedRepository(client *goredis.Client) *FeedRepository {
	return &FeedRepository{client: client}
}

func (r *FeedRepository) AddPostToFeed(ctx context.Context, userID, postID int64, createdAt time.Time) error {
	key := feedKey(userID)
	member := strconv.FormatInt(postID, 10)
	score := float64(createdAt.UnixNano())

	if err := r.client.ZAdd(ctx, key, goredis.Z{Score: score, Member: member}).Err(); err != nil {
		return fmt.Errorf("repository.feed.add: %w", err)
	}
	return nil
}

func (r *FeedRepository) GetPostIDs(ctx context.Context, userID int64, limit, offset int64) ([]int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	idsStr, err := r.client.ZRevRange(ctx, feedKey(userID), offset, offset+limit-1).Result()
	if err != nil {
		return nil, fmt.Errorf("repository.feed.get: %w", err)
	}

	ids := make([]int64, 0, len(idsStr))
	for _, v := range idsStr {
		parsed, parseErr := strconv.ParseInt(v, 10, 64)
		if parseErr != nil {
			continue
		}
		ids = append(ids, parsed)
	}

	return ids, nil
}

func feedKey(userID int64) string {
	return "feed:user:" + strconv.FormatInt(userID, 10)
}
