package evaluation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// AnthropicClient implements LLMClient using the Anthropic API
type AnthropicClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewAnthropicClient creates a new Anthropic API client
func NewAnthropicClient() (*AnthropicClient, error) {
	apiKey := os.Getenv("ANTHROPIC_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_KEY environment variable not set")
	}

	return &AnthropicClient{
		apiKey:  apiKey,
		baseURL: "https://api.anthropic.com/v1",
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

// anthropicRequest represents a request to the Anthropic API
type anthropicRequest struct {
	Model       string                   `json:"model"`
	Messages    []anthropicMessage       `json:"messages"`
	MaxTokens   int                      `json:"max_tokens"`
	Temperature float64                  `json:"temperature,omitempty"`
	System      string                   `json:"system,omitempty"`
}

// anthropicMessage represents a message in the Anthropic API
type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// anthropicResponse represents a response from the Anthropic API
type anthropicResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model        string `json:"model"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// Complete implements LLMClient.Complete
func (c *AnthropicClient) Complete(ctx context.Context, prompt string) (string, error) {
	return c.CompleteWithOptions(ctx, prompt, LLMOptions{
		Temperature: 0.7,
		MaxTokens:   1000,
	})
}

// CompleteWithOptions implements LLMClient.CompleteWithOptions
func (c *AnthropicClient) CompleteWithOptions(ctx context.Context, prompt string, options LLMOptions) (string, error) {
	// Prepare the request
	req := anthropicRequest{
		Model: "claude-3-haiku-20240307", // Fast and cost-effective for evaluations
		Messages: []anthropicMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:   options.MaxTokens,
		Temperature: options.Temperature,
	}

	if options.SystemPrompt != "" {
		req.System = options.SystemPrompt
	}

	// Marshal the request
	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	// Send the request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Error struct {
				Type    string `json:"type"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Error.Message != "" {
			return "", fmt.Errorf("anthropic API error: %s", errorResp.Error.Message)
		}
		return "", fmt.Errorf("anthropic API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse the response
	var anthropicResp anthropicResponse
	if err := json.Unmarshal(body, &anthropicResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract the text content
	if len(anthropicResp.Content) > 0 && anthropicResp.Content[0].Type == "text" {
		return anthropicResp.Content[0].Text, nil
	}

	return "", fmt.Errorf("no text content in response")
}

// ValidateConnection implements LLMClient.ValidateConnection
func (c *AnthropicClient) ValidateConnection(ctx context.Context) error {
	// Test the connection with a simple prompt
	testPrompt := "Please respond with 'OK' to confirm the connection is working."
	response, err := c.CompleteWithOptions(ctx, testPrompt, LLMOptions{
		MaxTokens:   10,
		Temperature: 0,
	})
	if err != nil {
		return fmt.Errorf("failed to validate connection: %w", err)
	}
	
	if response == "" {
		return fmt.Errorf("empty response from Anthropic API")
	}
	
	return nil
}