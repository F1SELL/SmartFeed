package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"SmartFeed/internal/config"
	"SmartFeed/internal/domain"
	"SmartFeed/internal/repository/postgres"
	"SmartFeed/pkg/kafka"
	"SmartFeed/pkg/llm"
	pgpkg "SmartFeed/pkg/postgres"
)

type tagGenerator interface {
	GenerateTags(ctx context.Context, text string) ([]string, error)
}

func main() {
	cfg := config.Load()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pgPool, err := pgpkg.New(ctx, cfg.PGDSN)
	if err != nil {
		log.Fatalf("postgres init failed: %v", err)
	}
	defer pgPool.Close()

	postRepo := postgres.NewPostRepository(pgPool)
	if err := kafka.EnsureTopic(ctx, cfg.KafkaBrokers, cfg.TopicPost, 1, 1); err != nil {
		log.Fatalf("kafka topic init failed: %v", err)
	}
	reader := kafka.NewPostEventReader(cfg.KafkaBrokers, cfg.TopicPost, cfg.GroupAI)
	defer reader.Close()

	generator := initTagGenerator(cfg)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-stop
		cancel()
	}()

	log.Println("ai worker started")
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

		tags := classifyTags(event.Content)
		if generator != nil {
			llmTags, err := generator.GenerateTags(ctx, event.Content)
			if err != nil {
				log.Printf("llm tags failed post=%d err=%v; fallback=rule-based", event.PostID, err)
			} else if len(llmTags) > 0 {
				tags = llmTags
			}
		}

		if err := postRepo.UpdateTags(ctx, event.PostID, tags); err != nil {
			log.Printf("update tags failed post=%d err=%v", event.PostID, err)
			continue
		}

		if err := reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("commit message error: %v", err)
		}
	}

	log.Println("ai worker stopped")
}

func initTagGenerator(cfg *config.Config) tagGenerator {
	if !cfg.LLMEnabled {
		log.Println("llm disabled, using rule-based tagging")
		return nil
	}
	if strings.TrimSpace(cfg.LLMAPIKey) == "" {
		log.Println("llm api key is empty, using rule-based tagging")
		return nil
	}

	client, err := llm.NewClient(
		cfg.LLMProvider,
		cfg.LLMBaseURL,
		cfg.LLMAPIKey,
		cfg.LLMModel,
		time.Duration(cfg.LLMTimeoutSec)*time.Second,
	)
	if err != nil {
		log.Printf("llm init failed: %v; using rule-based tagging", err)
		return nil
	}

	log.Printf("llm enabled provider=%s model=%s", cfg.LLMProvider, cfg.LLMModel)
	return client
}

func classifyTags(text string) []string {
	lower := strings.ToLower(text)
	tags := make([]string, 0, 3)

	if strings.Contains(lower, "go") || strings.Contains(lower, "golang") {
		tags = append(tags, "golang")
	}
	if strings.Contains(lower, "kafka") || strings.Contains(lower, "redis") || strings.Contains(lower, "postgres") {
		tags = append(tags, "backend")
	}
	if strings.Contains(lower, "ai") || strings.Contains(lower, "llm") {
		tags = append(tags, "ai")
	}
	if len(tags) == 0 {
		tags = append(tags, "general")
	}

	return tags
}
