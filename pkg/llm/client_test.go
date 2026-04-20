package llm

import "testing"

func TestParseTagsFromContentJSON(t *testing.T) {
	tags, err := parseTagsFromContent(`{"tags":["Go","Backend","Go"]}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tags) != 2 || tags[0] != "go" || tags[1] != "backend" {
		t.Fatalf("unexpected tags: %#v", tags)
	}
}

func TestParseTagsFromContentCodeFence(t *testing.T) {
	tags, err := parseTagsFromContent("```json\n{\"tags\":[\"ai\",\"llm\"]}\n```")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tags) != 2 || tags[0] != "ai" {
		t.Fatalf("unexpected tags: %#v", tags)
	}
}

func TestParseTagsFromContentCSV(t *testing.T) {
	tags, err := parseTagsFromContent("ai, backend, go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tags) != 3 {
		t.Fatalf("unexpected tags count: %#v", tags)
	}
}

func TestNewClientValidation(t *testing.T) {
	if _, err := NewClient("openai", "", "", "gpt-4o-mini", 0); err == nil {
		t.Fatal("expected api key validation error")
	}
	if _, err := NewClient("openai", "", "key", "", 0); err == nil {
		t.Fatal("expected model validation error")
	}
}
