package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	model      string
	provider   string
}

func NewClient(provider, baseURL, apiKey, model string, timeout time.Duration) (*Client, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("llm api key is empty")
	}
	if strings.TrimSpace(model) == "" {
		return nil, fmt.Errorf("llm model is empty")
	}
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider == "" {
		provider = "openai"
	}
	if strings.TrimSpace(baseURL) == "" {
		switch provider {
		case "gigachat":
			baseURL = "https://gigachat.devices.sberbank.ru/api/v1"
		default:
			baseURL = "https://api.openai.com/v1"
		}
	}
	if timeout <= 0 {
		timeout = 15 * time.Second
	}

	return &Client{
		httpClient: &http.Client{Timeout: timeout},
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		model:      model,
		provider:   provider,
	}, nil
}

type chatCompletionRequest struct {
	Model       string        `json:"model"`
	Temperature float64       `json:"temperature"`
	Messages    []chatMessage `json:"messages"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

func (c *Client) GenerateTags(ctx context.Context, text string) ([]string, error) {
	prompt := buildPrompt(text)
	reqBody := chatCompletionRequest{
		Model:       c.model,
		Temperature: 0.2,
		Messages: []chatMessage{
			{Role: "system", Content: "You are a content moderation and tagging assistant. Respond only with JSON."},
			{Role: "user", Content: prompt},
		},
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("llm marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("llm create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("llm send request: %w", err)
	}
	defer httpResp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(httpResp.Body, 1<<20))
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("llm bad status %d: %s", httpResp.StatusCode, string(body))
	}

	var resp chatCompletionResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("llm decode response: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("llm empty choices")
	}

	tags, err := parseTagsFromContent(resp.Choices[0].Message.Content)
	if err != nil {
		return nil, err
	}
	if len(tags) == 0 {
		return nil, fmt.Errorf("llm returned no tags")
	}
	return tags, nil
}

func buildPrompt(text string) string {
	return "Analyze this post and return JSON exactly in format {\"tags\":[\"tag1\",\"tag2\"]}. " +
		"Use 1 to 5 short lowercase tags. Post: " + text
}

func parseTagsFromContent(content string) ([]string, error) {
	trimmed := strings.TrimSpace(content)
	trimmed = strings.TrimPrefix(trimmed, "```json")
	trimmed = strings.TrimPrefix(trimmed, "```")
	trimmed = strings.TrimSuffix(trimmed, "```")
	trimmed = strings.TrimSpace(trimmed)

	type jsonTags struct {
		Tags []string `json:"tags"`
	}
	var payload jsonTags
	if err := json.Unmarshal([]byte(trimmed), &payload); err == nil {
		return normalizeTags(payload.Tags), nil
	}

	var direct []string
	if err := json.Unmarshal([]byte(trimmed), &direct); err == nil {
		return normalizeTags(direct), nil
	}

	parts := strings.Split(trimmed, ",")
	if len(parts) > 0 {
		tags := make([]string, 0, len(parts))
		for _, p := range parts {
			tags = append(tags, p)
		}
		norm := normalizeTags(tags)
		if len(norm) > 0 {
			return norm, nil
		}
	}

	return nil, fmt.Errorf("llm response is not parseable as tags")
}

func normalizeTags(tags []string) []string {
	seen := make(map[string]struct{}, len(tags))
	out := make([]string, 0, len(tags))
	for _, t := range tags {
		norm := strings.ToLower(strings.TrimSpace(strings.Trim(t, `"'`)))
		if norm == "" {
			continue
		}
		if _, ok := seen[norm]; ok {
			continue
		}
		seen[norm] = struct{}{}
		out = append(out, norm)
		if len(out) == 5 {
			break
		}
	}
	return out
}
