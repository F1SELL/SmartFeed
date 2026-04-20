package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"SmartFeed/internal/config"
	"SmartFeed/internal/repository/postgres"
	redisrepo "SmartFeed/internal/repository/redis"
	"SmartFeed/internal/service"
	"SmartFeed/internal/transport/http"
	"SmartFeed/internal/transport/http/handlers"
	"SmartFeed/pkg/kafka"
	pgpkg "SmartFeed/pkg/postgres"
	// @title SmartFeed API
	// @version 1.0
	// @description API Server for SmartFeed application.
	// @host localhost:8080
	// @BasePath /
	// @securityDefinitions.apikey BearerAuth
	// @in header
	// @name Authorization
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

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

	userRepo := postgres.NewUserRepository(pgPool)
	postRepo := postgres.NewPostRepository(pgPool)
	subRepo := postgres.NewSubscriptionRepository(pgPool)
	feedRepo := redisrepo.NewFeedRepository(redisClient)
	if err := kafka.EnsureTopic(ctx, cfg.KafkaBrokers, cfg.TopicPost, 1, 1); err != nil {
		log.Fatalf("kafka topic init failed: %v", err)
	}
	producer := kafka.NewPostEventProducer(cfg.KafkaBrokers, cfg.TopicPost)
	defer producer.Close()

	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	userService := service.NewUserService(userRepo, subRepo)
	postService := service.NewPostService(postRepo, producer)
	feedService := service.NewFeedService(feedRepo, postRepo)

	h := httptransport.Handlers{
		Auth: handlers.NewAuthHandler(authService),
		User: handlers.NewUserHandler(userService),
		Post: handlers.NewPostHandler(postService),
		Feed: handlers.NewFeedHandler(feedService),
	}

	router := httptransport.NewRouter(h, cfg.JWTSecret)
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.CoreHTTPPort),
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("core api listening on %s", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http server failed: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("http shutdown error: %v", err)
	}
	log.Println("core api stopped")
}
