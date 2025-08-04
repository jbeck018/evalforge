package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
)

// Configuration loaded from environment
type Config struct {
	OpenAIPort       int
	AnthropicPort    int
	ResponseDelayMin int
	ResponseDelayMax int
	ErrorRate        float32
	TimeoutRate      float32
	LogLevel         string
}

// OpenAI compatible structures
type OpenAIRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float32   `json:"temperature,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Anthropic compatible structures
type AnthropicRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []Message `json:"messages"`
	System    string    `json:"system,omitempty"`
}

type AnthropicResponse struct {
	ID           string              `json:"id"`
	Type         string              `json:"type"`
	Role         string              `json:"role"`
	Content      []AnthropicContent  `json:"content"`
	Model        string              `json:"model"`
	StopReason   string              `json:"stop_reason"`
	StopSequence interface{}         `json:"stop_sequence"`
	Usage        AnthropicUsage      `json:"usage"`
}

type AnthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type AnthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// Configuration for different response behaviors
type MockBehavior struct {
	ErrorScenarios []ErrorScenario `json:"error_scenarios"`
	ResponseTypes  []ResponseType  `json:"response_types"`
}

type ErrorScenario struct {
	TriggerPhrase string `json:"trigger_phrase"`
	ErrorType     string `json:"error_type"`
	ErrorMessage  string `json:"error_message"`
}

type ResponseType struct {
	Pattern     string   `json:"pattern"`
	Templates   []string `json:"templates"`
	TokenRange  [2]int   `json:"token_range"`
	LatencyMs   [2]int   `json:"latency_ms"`
}

var (
	config   Config
	behavior MockBehavior
	logger   *logrus.Logger
)

func init() {
	logger = logrus.New()
	rand.Seed(time.Now().UnixNano())
	
	// Load configuration from environment
	config = Config{
		OpenAIPort:       getEnvInt("OPENAI_PORT", 8080),
		AnthropicPort:    getEnvInt("ANTHROPIC_PORT", 8081),
		ResponseDelayMin: getEnvInt("RESPONSE_DELAY_MIN_MS", 100),
		ResponseDelayMax: getEnvInt("RESPONSE_DELAY_MAX_MS", 500),
		ErrorRate:        getEnvFloat("ERROR_RATE", 0.01),
		TimeoutRate:      getEnvFloat("TIMEOUT_RATE", 0.005),
		LogLevel:         getEnvString("LOG_LEVEL", "info"),
	}
	
	// Set log level
	level, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)
	
	// Initialize default behavior
	behavior = MockBehavior{
		ErrorScenarios: []ErrorScenario{
			{
				TriggerPhrase: "cause error",
				ErrorType:     "invalid_request_error",
				ErrorMessage:  "Mock error triggered by test phrase",
			},
			{
				TriggerPhrase: "rate limit",
				ErrorType:     "rate_limit_exceeded",
				ErrorMessage:  "Rate limit exceeded",
			},
		},
		ResponseTypes: []ResponseType{
			{
				Pattern:    ".*code.*|.*programming.*",
				Templates:  []string{"Here's a code solution:\n\n```\n%s\n```", "The code approach would be:\n\n%s"},
				TokenRange: [2]int{50, 200},
				LatencyMs:  [2]int{200, 800},
			},
			{
				Pattern:    ".*explain.*|.*what.*|.*how.*",
				Templates:  []string{"Let me explain: %s", "The explanation is: %s", "To clarify: %s"},
				TokenRange: [2]int{30, 150},
				LatencyMs:  [2]int{100, 400},
			},
		},
	}
}

func main() {
	logger.Infof("Starting Mock LLM services on ports %d (OpenAI) and %d (Anthropic)", 
		config.OpenAIPort, config.AnthropicPort)
	
	// Setup OpenAI-compatible server
	go func() {
		r := mux.NewRouter()
		
		// CORS middleware
		c := cors.New(cors.Options{
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{"*"},
		})
		
		// OpenAI endpoints
		r.HandleFunc("/v1/chat/completions", handleOpenAI).Methods("POST")
		r.HandleFunc("/v1/models", handleModels).Methods("GET")
		r.HandleFunc("/health", handleHealth).Methods("GET")
		r.HandleFunc("/config", handleGetConfig).Methods("GET")
		r.HandleFunc("/config", handleSetConfig).Methods("POST")
		
		handler := c.Handler(r)
		logger.Infof("OpenAI-compatible server listening on :%d", config.OpenAIPort)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", config.OpenAIPort), handler); err != nil {
			logger.Fatalf("OpenAI server failed: %v", err)
		}
	}()
	
	// Setup Anthropic-compatible server
	r := mux.NewRouter()
	
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})
	
	// Anthropic endpoints
	r.HandleFunc("/v1/messages", handleAnthropic).Methods("POST")
	r.HandleFunc("/health", handleHealth).Methods("GET")
	r.HandleFunc("/config", handleGetConfig).Methods("GET")
	r.HandleFunc("/config", handleSetConfig).Methods("POST")
	
	handler := c.Handler(r)
	logger.Infof("Anthropic-compatible server listening on :%d", config.AnthropicPort)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", config.AnthropicPort), handler); err != nil {
		logger.Fatalf("Anthropic server failed: %v", err)
	}
}

func handleOpenAI(w http.ResponseWriter, r *http.Request) {
	var req OpenAIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	logger.Debugf("OpenAI request: model=%s, messages=%d", req.Model, len(req.Messages))
	
	// Simulate processing delay
	simulateLatency()
	
	// Check for error simulation
	if shouldSimulateError(req.Messages) {
		simulateError(w)
		return
	}
	
	// Generate response
	prompt := extractPrompt(req.Messages)
	responseText := generateResponse(prompt, req.Model)
	
	// Calculate token usage
	promptTokens := estimateTokens(prompt)
	completionTokens := estimateTokens(responseText)
	
	response := OpenAIResponse{
		ID:      generateID("chatcmpl"),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   req.Model,
		Choices: []Choice{
			{
				Index: 0,
				Message: Message{
					Role:    "assistant",
					Content: responseText,
				},
				FinishReason: "stop",
			},
		},
		Usage: Usage{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      promptTokens + completionTokens,
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleAnthropic(w http.ResponseWriter, r *http.Request) {
	var req AnthropicRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	logger.Debugf("Anthropic request: model=%s, messages=%d", req.Model, len(req.Messages))
	
	// Simulate processing delay
	simulateLatency()
	
	// Check for error simulation
	if shouldSimulateError(req.Messages) {
		simulateError(w)
		return
	}
	
	// Generate response
	prompt := extractPrompt(req.Messages)
	responseText := generateResponse(prompt, req.Model)
	
	// Calculate token usage
	inputTokens := estimateTokens(prompt)
	outputTokens := estimateTokens(responseText)
	
	response := AnthropicResponse{
		ID:   generateID("msg"),
		Type: "message",
		Role: "assistant",
		Content: []AnthropicContent{
			{
				Type: "text",
				Text: responseText,
			},
		},
		Model:        req.Model,
		StopReason:   "end_turn",
		StopSequence: nil,
		Usage: AnthropicUsage{
			InputTokens:  inputTokens,
			OutputTokens: outputTokens,
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleModels(w http.ResponseWriter, r *http.Request) {
	models := map[string]interface{}{
		"object": "list",
		"data": []map[string]interface{}{
			{
				"id":       "gpt-4",
				"object":   "model",
				"created":  1677610602,
				"owned_by": "openai",
			},
			{
				"id":       "gpt-3.5-turbo",
				"object":   "model",
				"created":  1677610602,
				"owned_by": "openai",
			},
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
		"config": map[string]interface{}{
			"error_rate":    config.ErrorRate,
			"timeout_rate":  config.TimeoutRate,
			"latency_range": [2]int{config.ResponseDelayMin, config.ResponseDelayMax},
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleGetConfig(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"config":   config,
		"behavior": behavior,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleSetConfig(w http.ResponseWriter, r *http.Request) {
	var newBehavior MockBehavior
	if err := json.NewDecoder(r.Body).Decode(&newBehavior); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	behavior = newBehavior
	logger.Infof("Updated mock behavior configuration")
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func simulateLatency() {
	if config.ResponseDelayMax > config.ResponseDelayMin {
		delay := config.ResponseDelayMin + rand.Intn(config.ResponseDelayMax-config.ResponseDelayMin)
		time.Sleep(time.Duration(delay) * time.Millisecond)
	}
}

func shouldSimulateError(messages []Message) bool {
	if rand.Float32() < config.ErrorRate {
		return true
	}
	
	// Check for specific error triggers
	prompt := extractPrompt(messages)
	for _, scenario := range behavior.ErrorScenarios {
		if strings.Contains(strings.ToLower(prompt), strings.ToLower(scenario.TriggerPhrase)) {
			return true
		}
	}
	
	return false
}

func simulateError(w http.ResponseWriter) {
	errorTypes := []int{
		http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadRequest,
		http.StatusUnauthorized,
	}
	
	statusCode := errorTypes[rand.Intn(len(errorTypes))]
	
	errorResponse := map[string]interface{}{
		"error": map[string]interface{}{
			"message": "Mock error for testing",
			"type":    "mock_error",
			"code":    statusCode,
		},
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errorResponse)
}

func extractPrompt(messages []Message) string {
	var parts []string
	for _, msg := range messages {
		parts = append(parts, msg.Content)
	}
	return strings.Join(parts, " ")
}

func generateResponse(prompt, model string) string {
	// Create deterministic but varied responses based on prompt hash
	hash := md5.Sum([]byte(prompt + model))
	hashStr := hex.EncodeToString(hash[:])
	
	// Response templates based on prompt analysis
	templates := []string{
		"Based on your request, I would recommend: %s. This approach takes into account the key factors you mentioned.",
		"After analyzing your prompt, here's my response: %s. Let me know if you need any clarification.",
		"The solution to your question involves: %s. This should address your main concerns effectively.",
		"Here's my analysis: %s. I've considered multiple perspectives to provide a comprehensive answer.",
		"To address your inquiry: %s. This approach balances practicality with best practices.",
	}
	
	// Select template based on hash
	templateIdx := int(hash[0]) % len(templates)
	
	// Generate response content based on prompt characteristics
	var content string
	promptLower := strings.ToLower(prompt)
	
	switch {
	case strings.Contains(promptLower, "code") || strings.Contains(promptLower, "programming"):
		content = generateCodeResponse(hashStr)
	case strings.Contains(promptLower, "explain") || strings.Contains(promptLower, "what"):
		content = generateExplanationResponse(hashStr)
	case strings.Contains(promptLower, "list") || strings.Contains(promptLower, "steps"):
		content = generateListResponse(hashStr)
	default:
		content = generateGeneralResponse(hashStr)
	}
	
	return fmt.Sprintf(templates[templateIdx], content)
}

func generateCodeResponse(seed string) string {
	codeExamples := []string{
		"```go\nfunc main() {\n    fmt.Println(\"Hello, World!\")\n}\n```",
		"```python\ndef process_data(data):\n    return [x * 2 for x in data]\n```",
		"```javascript\nconst result = data.map(item => item.value * 2);\n```",
		"```sql\nSELECT * FROM users WHERE created_at > NOW() - INTERVAL '7 days';\n```",
	}
	
	hash := md5.Sum([]byte(seed))
	idx := int(hash[0]) % len(codeExamples)
	return codeExamples[idx]
}

func generateExplanationResponse(seed string) string {
	explanations := []string{
		"The fundamental concept here involves understanding how different components interact with each other",
		"This process works by leveraging established patterns and principles in the domain",
		"The key insight is that we need to balance performance, maintainability, and scalability",
		"The approach centers around creating efficient data flows and minimizing bottlenecks",
	}
	
	hash := md5.Sum([]byte(seed))
	idx := int(hash[0]) % len(explanations)
	return explanations[idx]
}

func generateListResponse(seed string) string {
	lists := []string{
		"1. First, analyze the requirements\n2. Design the architecture\n3. Implement core functionality\n4. Test thoroughly\n5. Deploy and monitor",
		"• Gather necessary data\n• Process and validate inputs\n• Apply business logic\n• Generate output\n• Handle edge cases",
		"→ Define clear objectives\n→ Identify key stakeholders\n→ Create implementation timeline\n→ Execute with monitoring\n→ Iterate based on feedback",
	}
	
	hash := md5.Sum([]byte(seed))
	idx := int(hash[0]) % len(lists)
	return lists[idx]
}

func generateGeneralResponse(seed string) string {
	responses := []string{
		"a comprehensive approach that considers multiple factors and provides practical solutions",
		"a balanced strategy that addresses both immediate needs and long-term objectives",
		"an efficient methodology that optimizes for performance while maintaining clarity",
		"a robust framework that handles various scenarios and edge cases effectively",
	}
	
	hash := md5.Sum([]byte(seed))
	idx := int(hash[0]) % len(responses)
	return responses[idx]
}

func estimateTokens(text string) int {
	// Rough token estimation: ~4 characters per token
	return len(text) / 4
}

func generateID(prefix string) string {
	timestamp := time.Now().UnixNano()
	random := rand.Intn(1000000)
	return fmt.Sprintf("%s-%d-%d", prefix, timestamp, random)
}

// Utility functions for environment variables
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float32) float32 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 32); err == nil {
			return float32(floatValue)
		}
	}
	return defaultValue
}

func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}