package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/google/uuid"
)

var ctx = context.Background()

// Configuration for data generation
type Config struct {
	ClickHouseURL string
	ProjectCount  int
	EventsPerDay  int
	DaysBack      int
	Verbose       bool
}

// Event represents a single trace event
type Event struct {
	Timestamp        time.Time
	ProjectID        string
	TraceID          string
	SpanID           string
	ParentSpanID     string
	EventType        string
	OperationName    string
	Status           string
	Model            string
	Provider         string
	DurationMs       uint32
	TokensUsed       uint32
	InputTokens      uint32
	OutputTokens     uint32
	CostCents        uint32
	InputCostCents   uint32
	OutputCostCents  uint32
	RelevanceScore   float32
	AccuracyScore    float32
	SafetyScore      float32
	OverallScore     float32
	Prompt           string
	Response         string
	ErrorType        string
	ErrorMessage     string
	Metadata         string
	UserID           string
	SessionID        string
	Region           string
	Environment      string
}

// Predefined data for realistic generation
var (
	projects = []string{
		"p1p1p1p1-p1p1-p1p1-p1p1-p1p1p1p1p1p1", // Customer Support Chatbot
		"p2p2p2p2-p2p2-p2p2-p2p2-p2p2p2p2p2p2", // Code Review Assistant
		"p3p3p3p3-p3p3-p3p3-p3p3-p3p3p3p3p3p3", // Document Summarizer
		"p4p4p4p4-p4p4-p4p4-p4p4-p4p4p4p4p4p4", // Content Generator
		"p5p5p5p5-p5p5-p5p5-p5p5-p5p5p5p5p5p5", // RAG Knowledge Base
	}

	models = []ModelConfig{
		{"gpt-4", "openai", 0.03, 0.06, 800, 1500, 0.92},
		{"gpt-3.5-turbo", "openai", 0.0015, 0.002, 300, 800, 0.85},
		{"claude-3-opus", "anthropic", 0.015, 0.075, 1000, 2000, 0.94},
		{"claude-3-sonnet", "anthropic", 0.003, 0.015, 400, 900, 0.88},
		{"claude-3-haiku", "anthropic", 0.00025, 0.00125, 200, 500, 0.82},
	}

	eventTypes = []string{
		"llm_request", "evaluation", "optimization", "analysis",
	}

	operationNames = []string{
		"chat_completion", "text_generation", "code_review", "summarization",
		"evaluation_run", "prompt_optimization", "quality_check",
	}

	statuses = []StatusConfig{
		{"success", 0.92},
		{"error", 0.05},
		{"timeout", 0.02},
		{"cancelled", 0.01},
	}

	errorTypes = []string{
		"rate_limit_exceeded", "invalid_request", "model_unavailable",
		"timeout_error", "authentication_failed", "quota_exceeded",
	}

	regions = []string{
		"us-east-1", "us-west-2", "eu-west-1", "ap-southeast-1",
	}

	prompts = []string{
		"How do I reset my password?",
		"What are your business hours?",
		"I need help with billing",
		"Can you explain how this feature works?",
		"Review this code for potential bugs",
		"Summarize this document in 3 paragraphs",
		"Generate a product description for this item",
		"What's the weather like today?",
		"Help me debug this error message",
		"Translate this text to Spanish",
	}

	responses = []string{
		"I'd be happy to help you with that. Here's what you need to do...",
		"Based on your request, I recommend the following approach...",
		"Let me walk you through the process step by step...",
		"I've analyzed your request and here's my response...",
		"After reviewing the information, I can provide these insights...",
		"Here's a comprehensive answer to your question...",
		"I understand your concern. Let me help you resolve this...",
		"Thank you for your question. Here's what I found...",
		"I've processed your request and generated the following...",
		"Based on the context provided, my recommendation is...",
	}
)

type ModelConfig struct {
	Name             string
	Provider         string
	InputCostPer1K   float64
	OutputCostPer1K  float64
	MinLatencyMs     int
	MaxLatencyMs     int
	AvgQualityScore  float64
}

type StatusConfig struct {
	Status      string
	Probability float64
}

func main() {
	config := parseFlags()

	if config.Verbose {
		log.Printf("Starting data generation with config: %+v", config)
	}

	generator := NewDataGenerator(config)
	if err := generator.Generate(); err != nil {
		log.Fatalf("Failed to generate data: %v", err)
	}

	log.Println("Data generation completed successfully!")
}

func parseFlags() Config {
	config := Config{}

	flag.StringVar(&config.ClickHouseURL, "clickhouse-url", "clickhouse://evalforge:evalforge_dev@localhost:9000/evalforge", "ClickHouse connection URL")
	flag.IntVar(&config.ProjectCount, "projects", 3, "Number of projects to generate data for")
	flag.IntVar(&config.EventsPerDay, "events-per-day", 10000, "Number of events to generate per day")
	flag.IntVar(&config.DaysBack, "days-back", 7, "Number of days back to generate data for")
	flag.BoolVar(&config.Verbose, "verbose", false, "Enable verbose logging")

	flag.Parse()

	return config
}

type DataGenerator struct {
	config Config
	conn   clickhouse.Conn
}

func NewDataGenerator(config Config) *DataGenerator {
	return &DataGenerator{
		config: config,
	}
}

func (g *DataGenerator) Generate() error {
	if err := g.connect(); err != nil {
		return fmt.Errorf("failed to connect to ClickHouse: %w", err)
	}
	defer g.conn.Close()

	totalEvents := g.config.EventsPerDay * g.config.DaysBack
	batchSize := 1000

	log.Printf("Generating %d events across %d days (%d events/day)", 
		totalEvents, g.config.DaysBack, g.config.EventsPerDay)

	for day := 0; day < g.config.DaysBack; day++ {
		dayStart := time.Now().AddDate(0, 0, -day)
		eventsForDay := g.config.EventsPerDay

		if g.config.Verbose {
			log.Printf("Generating %d events for day %s", eventsForDay, dayStart.Format("2006-01-02"))
		}

		for batch := 0; batch < eventsForDay; batch += batchSize {
			batchEnd := batch + batchSize
			if batchEnd > eventsForDay {
				batchEnd = eventsForDay
			}

			events := g.generateEventBatch(dayStart, batchEnd-batch)
			if err := g.insertEvents(events); err != nil {
				return fmt.Errorf("failed to insert batch: %w", err)
			}

			if g.config.Verbose && batch%5000 == 0 {
				log.Printf("Inserted %d/%d events for day %s", batch, eventsForDay, dayStart.Format("2006-01-02"))
			}
		}
	}

	return nil
}

func (g *DataGenerator) connect() error {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{"localhost:9000"},
		Auth: clickhouse.Auth{
			Database: "evalforge",
			Username: "evalforge",
			Password: "evalforge_dev",
		},
	})
	if err != nil {
		return err
	}

	if err := conn.Ping(ctx); err != nil {
		return err
	}

	g.conn = conn
	return nil
}

func (g *DataGenerator) generateEventBatch(dayStart time.Time, count int) []Event {
	events := make([]Event, count)

	for i := 0; i < count; i++ {
		events[i] = g.generateEvent(dayStart)
	}

	return events
}

func (g *DataGenerator) generateEvent(dayStart time.Time) Event {
	// Generate realistic timestamp distribution throughout the day
	// More activity during business hours
	hour := g.generateBusinessHourBiased()
	minute := rand.Intn(60)
	second := rand.Intn(60)
	
	timestamp := time.Date(
		dayStart.Year(), dayStart.Month(), dayStart.Day(),
		hour, minute, second, 0, time.UTC,
	)

	// Select project and model
	projectID := projects[rand.Intn(min(g.config.ProjectCount, len(projects)))]
	model := models[rand.Intn(len(models))]
	
	// Generate trace and span IDs
	traceID := fmt.Sprintf("trace_%s_%d", uuid.New().String()[:8], timestamp.Unix())
	spanID := fmt.Sprintf("span_%s_%d", uuid.New().String()[:8], rand.Intn(1000))

	// Select status based on probability
	status := g.selectStatus()
	
	// Generate tokens and timing
	inputTokens := uint32(rand.Intn(800) + 50)  // 50-850 tokens
	outputTokens := uint32(rand.Intn(500) + 25) // 25-525 tokens
	tokensUsed := inputTokens + outputTokens
	
	durationMs := uint32(rand.Intn(model.MaxLatencyMs-model.MinLatencyMs) + model.MinLatencyMs)
	
	// Calculate costs
	inputCostCents := uint32(float64(inputTokens) * model.InputCostPer1K / 10.0)  // Convert to cents
	outputCostCents := uint32(float64(outputTokens) * model.OutputCostPer1K / 10.0)
	costCents := inputCostCents + outputCostCents

	// Generate quality scores
	relevanceScore := g.generateQualityScore(model.AvgQualityScore, 0.1)
	accuracyScore := g.generateQualityScore(model.AvgQualityScore, 0.08)
	safetyScore := g.generateQualityScore(0.95, 0.05) // Safety is generally high
	overallScore := (relevanceScore + accuracyScore + safetyScore) / 3.0

	// Select content
	prompt := prompts[rand.Intn(len(prompts))]
	response := responses[rand.Intn(len(responses))]

	// Error handling
	var errorType, errorMessage string
	if status == "error" {
		errorType = errorTypes[rand.Intn(len(errorTypes))]
		errorMessage = fmt.Sprintf("Mock error: %s occurred during processing", errorType)
		response = "" // No response on error
		relevanceScore = 0
		accuracyScore = 0
		overallScore = 0
	}

	// Generate metadata
	metadata := map[string]interface{}{
		"version":     "1.0",
		"sdk_version": "0.1.0",
		"experiment":  fmt.Sprintf("exp_%d", rand.Intn(5)),
	}
	metadataJSON, _ := json.Marshal(metadata)

	return Event{
		Timestamp:       timestamp,
		ProjectID:       projectID,
		TraceID:         traceID,
		SpanID:          spanID,
		ParentSpanID:    "",
		EventType:       eventTypes[rand.Intn(len(eventTypes))],
		OperationName:   operationNames[rand.Intn(len(operationNames))],
		Status:          status,
		Model:           model.Name,
		Provider:        model.Provider,
		DurationMs:      durationMs,
		TokensUsed:      tokensUsed,
		InputTokens:     inputTokens,
		OutputTokens:    outputTokens,
		CostCents:       costCents,
		InputCostCents:  inputCostCents,
		OutputCostCents: outputCostCents,
		RelevanceScore:  relevanceScore,
		AccuracyScore:   accuracyScore,
		SafetyScore:     safetyScore,
		OverallScore:    overallScore,
		Prompt:          prompt,
		Response:        response,
		ErrorType:       errorType,
		ErrorMessage:    errorMessage,
		Metadata:        string(metadataJSON),
		UserID:          fmt.Sprintf("user_%d", rand.Intn(100)),
		SessionID:       fmt.Sprintf("session_%d", rand.Intn(1000)),
		Region:          regions[rand.Intn(len(regions))],
		Environment:     "production",
	}
}

func (g *DataGenerator) generateBusinessHourBiased() int {
	// Weight business hours more heavily
	weights := []float64{
		0.5, 0.3, 0.2, 0.1, 0.1, 0.2, 0.5, 1.0, // 0-7 AM
		2.0, 3.0, 3.5, 3.0, 2.5, 3.0, 4.0, 3.5, // 8-15 (business hours)
		2.0, 1.5, 1.0, 0.8, 0.6, 0.4, 0.3, 0.2, // 16-23 PM
	}

	totalWeight := 0.0
	for _, w := range weights {
		totalWeight += w
	}

	r := rand.Float64() * totalWeight
	cumulative := 0.0

	for i, weight := range weights {
		cumulative += weight
		if r <= cumulative {
			return i
		}
	}

	return 12 // Default to noon
}

func (g *DataGenerator) selectStatus() string {
	r := rand.Float64()
	cumulative := 0.0

	for _, status := range statuses {
		cumulative += status.Probability
		if r <= cumulative {
			return status.Status
		}
	}

	return "success" // Default
}

func (g *DataGenerator) generateQualityScore(mean, stddev float64) float32 {
	// Generate a normally distributed score
	score := rand.NormFloat64()*stddev + mean
	
	// Clamp to [0, 1] range
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}

	return float32(score)
}

func (g *DataGenerator) insertEvents(events []Event) error {
	batch, err := g.conn.PrepareBatch(ctx, `
		INSERT INTO events (
			timestamp, project_id, trace_id, span_id, parent_span_id,
			event_type, operation_name, status, model, provider,
			duration_ms, tokens_used, input_tokens, output_tokens,
			cost_cents, input_cost_cents, output_cost_cents,
			relevance_score, accuracy_score, safety_score, overall_score,
			prompt, response, error_type, error_message, metadata,
			user_id, session_id, region, environment
		)
	`)
	if err != nil {
		return err
	}

	for _, event := range events {
		err := batch.Append(
			event.Timestamp, event.ProjectID, event.TraceID, event.SpanID, event.ParentSpanID,
			event.EventType, event.OperationName, event.Status, event.Model, event.Provider,
			event.DurationMs, event.TokensUsed, event.InputTokens, event.OutputTokens,
			event.CostCents, event.InputCostCents, event.OutputCostCents,
			event.RelevanceScore, event.AccuracyScore, event.SafetyScore, event.OverallScore,
			event.Prompt, event.Response, event.ErrorType, event.ErrorMessage, event.Metadata,
			event.UserID, event.SessionID, event.Region, event.Environment,
		)
		if err != nil {
			return err
		}
	}

	return batch.Send()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}