package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"SmartFeed/internal/config"
	"SmartFeed/internal/domain"
	"SmartFeed/internal/repository/postgres"
	redisrepo "SmartFeed/internal/repository/redis"
	"SmartFeed/pkg/kafka"
	pgpkg "SmartFeed/pkg/postgres"
)

func main() {
	cfg := config.Load()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pgPool, err := pgpkg.New(ctx, cfg.PGDSN)
	if err != nil {
		log.Fatalf("postgres init failed: %v", err)
	}
	defer pgPool.Close()

	redisClient := goredis.NewClient(&goredis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPass,
		DB:       cfg.RedisDB,
	})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("redis init failed: %v", err)
	}
	defer redisClient.Close()

	subRepo := postgres.NewSubscriptionRepository(pgPool)
	feedRepo := redisrepo.NewFeedRepository(redisClient)
	if err := kafka.EnsureTopic(ctx, cfg.KafkaBrokers, cfg.TopicPost, 1, 1); err != nil {
		log.Fatalf("kafka topic init failed: %v", err)
	}
	reader := kafka.NewPostEventReader(cfg.KafkaBrokers, cfg.TopicPost, cfg.GroupFeed)
	defer reader.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-stop
		cancel()
	}()

	log.Println("feed worker started")
	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				break
			}
			log.Printf("fetch message error: %v", err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		var event domain.PostCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("bad event payload: %v", err)
			_ = reader.CommitMessages(ctx, msg)
			continue
		}

		followers, err := subRepo.ListFollowerIDs(ctx, event.AuthorID)
		if err != nil {
			log.Printf("followers fetch error: %v", err)
			continue
		}
		followers = append(followers, event.AuthorID)

		for _, followerID := range followers {
			if err := feedRepo.AddPostToFeed(ctx, followerID, event.PostID, event.CreatedAt); err != nil {
				log.Printf("feed update failed user=%d post=%d err=%v", followerID, event.PostID, err)
			}
		}

		if err := reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("commit message error: %v", err)
		}
	}

	log.Println("feed worker stopped")
}
