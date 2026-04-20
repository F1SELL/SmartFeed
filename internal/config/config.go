package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	CoreHTTPPort int      `env:"CORE_HTTP_PORT" env-default:"8080"`
	PGDSN        string   `env:"PG_DSN" env-required:"true"`
	RedisAddr    string   `env:"REDIS_ADDR" env-required:"true"`
	RedisPass    string   `env:"REDIS_PASSWORD"`
	RedisDB      int      `env:"REDIS_DB" env-default:"0"`
	KafkaBrokers []string `env:"KAFKA_BROKERS" env-required:"true" env-separator:","`
	TopicPost    string   `env:"KAFKA_TOPIC_POST_CREATED" env-default:"post_created"`
	GroupFeed    string   `env:"KAFKA_GROUP_FEED" env-default:"feed-worker"`
	GroupAI      string   `env:"KAFKA_GROUP_AI" env-default:"ai-worker"`
	JWTSecret    string   `env:"JWT_SECRET" env-required:"true"`

	LLMEnabled    bool   `env:"LLM_ENABLED" env-default:"false"`
	LLMProvider   string `env:"LLM_PROVIDER" env-default:"openai"`
	LLMBaseURL    string `env:"LLM_BASE_URL"`
	LLMAPIKey     string `env:"LLM_API_KEY"`
	LLMModel      string `env:"LLM_MODEL" env-default:"gpt-4o-mini"`
	LLMTimeoutSec int    `env:"LLM_TIMEOUT_SEC" env-default:"15"`
}

// Load читает переменные окружения из .env файла и/или системы
func Load() *Config {
	var cfg Config

	if err := cleanenv.ReadConfig(".env", &cfg); err == nil {
		return &cfg
	}
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("config error: %v", err)
	}

	return &cfg
}
