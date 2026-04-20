package main

import (
	"testing"

	"SmartFeed/internal/config"
)

func TestClassifyTags(t *testing.T) {
	tags := classifyTags("Go + Kafka + Redis + AI")
	if len(tags) == 0 {
		t.Fatal("expected non-empty tags")
	}
}

func TestInitTagGeneratorDisabled(t *testing.T) {
	cfg := &config.Config{LLMEnabled: false}
	if got := initTagGenerator(cfg); got != nil {
		t.Fatal("expected nil generator when llm is disabled")
	}
}

func TestInitTagGeneratorNoAPIKey(t *testing.T) {
	cfg := &config.Config{LLMEnabled: true, LLMProvider: "openai", LLMModel: "gpt-4o-mini"}
	if got := initTagGenerator(cfg); got != nil {
		t.Fatal("expected nil generator when api key is empty")
	}
}
