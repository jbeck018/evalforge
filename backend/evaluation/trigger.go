package evaluation

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

// AutoEvaluationTrigger handles automatically triggering evaluations
type AutoEvaluationTrigger struct {
	orchestrator EvaluationOrchestrator
	analyzer     PromptAnalyzer
	repository   EvaluationRepository
	config       TriggerConfig
}

// TriggerConfig contains configuration for auto-evaluation triggers
type TriggerConfig struct {
	EnableAutoEvaluation bool
	MinPromptLength      int
	MaxPromptLength      int
	TriggerThreshold     int    // Number of agent calls before triggering evaluation
	ExcludePatterns      []string // Patterns to exclude from auto-evaluation
	IncludeTaskTypes     []TaskType
	DelayBetweenRuns     time.Duration
}

// AgentTrackingEvent represents an agent execution event
type AgentTrackingEvent struct {
	ID           string                 `json:"id"`
	ProjectID    int                    `json:"project_id"`
	TraceID      string                 `json:"trace_id"`
	SpanID       string                 `json:"span_id"`
	OperationType string                `json:"operation_type"`
	StartTime    time.Time              `json:"start_time"`
	EndTime      time.Time              `json:"end_time"`
	Input        map[string]interface{} `json:"input"`
	Output       map[string]interface{} `json:"output"`
	Metadata     map[string]interface{} `json:"metadata"`
	Provider     string                 `json:"provider"`
	Model        string                 `json:"model"`
}

// PromptCandidate represents a prompt that might need evaluation
type PromptCandidate struct {
	ProjectID      int
	Prompt         string
	FirstSeen      time.Time
	LastSeen       time.Time
	ExecutionCount int
	TaskType       TaskType
	Examples       []Example
	SampleInputs   []map[string]interface{}
	SampleOutputs  []map[string]interface{}
}

// NewAutoEvaluationTrigger creates a new auto-evaluation trigger
func NewAutoEvaluationTrigger(
	orchestrator EvaluationOrchestrator,
	analyzer PromptAnalyzer,
	repository EvaluationRepository,
	config TriggerConfig,
) *AutoEvaluationTrigger {
	// Set default config values
	if config.MinPromptLength == 0 {
		config.MinPromptLength = 10
	}
	if config.MaxPromptLength == 0 {
		config.MaxPromptLength = 10000
	}
	if config.TriggerThreshold == 0 {
		config.TriggerThreshold = 10
	}
	if config.DelayBetweenRuns == 0 {
		config.DelayBetweenRuns = 1 * time.Hour
	}
	if len(config.IncludeTaskTypes) == 0 {
		config.IncludeTaskTypes = []TaskType{
			TaskClassification,
			TaskGeneration,
			TaskExtraction,
			TaskSummarization,
			TaskQA,
		}
	}

	return &AutoEvaluationTrigger{
		orchestrator: orchestrator,
		analyzer:     analyzer,
		repository:   repository,
		config:       config,
	}
}

// ProcessTrackingEvent processes an agent tracking event and determines if evaluation should be triggered  
func (aet *AutoEvaluationTrigger) ProcessTrackingEvent(ctx context.Context, event AgentTrackingEvent) error {
	if !aet.config.EnableAutoEvaluation {
		return nil
	}

	// Extract prompt from the event
	prompt := aet.extractPromptFromEvent(event)
	if prompt == "" {
		return nil // No prompt found, skip
	}

	// Validate prompt
	if !aet.isValidPrompt(prompt) {
		log.Printf("Skipping invalid prompt for auto-evaluation: length=%d", len(prompt))
		return nil
	}

	// Check if prompt should be excluded
	if aet.shouldExcludePrompt(prompt) {
		log.Printf("Skipping excluded prompt for auto-evaluation")
		return nil
	}

	// Analyze prompt to determine task type
	analysis, err := aet.analyzer.AnalyzePrompt(ctx, prompt, nil)
	if err != nil {
		log.Printf("Failed to analyze prompt for auto-evaluation: %v", err)
		return nil // Don't fail the whole process
	}

	// Check if task type is included
	if !aet.isIncludedTaskType(analysis.TaskType) {
		log.Printf("Skipping prompt with excluded task type: %s", analysis.TaskType)
		return nil
	}

	// Get or create prompt candidate
	candidate, err := aet.getOrCreatePromptCandidate(event.ProjectID, prompt, analysis.TaskType, event)
	if err != nil {
		return fmt.Errorf("failed to get prompt candidate: %w", err)
	}

	// Update candidate with new execution
	candidate.LastSeen = event.EndTime
	candidate.ExecutionCount++
	
	// Add sample data
	if len(candidate.SampleInputs) < 20 {
		candidate.SampleInputs = append(candidate.SampleInputs, event.Input)
		candidate.SampleOutputs = append(candidate.SampleOutputs, event.Output)
	}

	// Check if we should trigger evaluation
	if aet.shouldTriggerEvaluation(candidate) {
		log.Printf("Triggering auto-evaluation for prompt in project %d", event.ProjectID)
		
		// Check if evaluation already exists for this prompt
		if exists, err := aet.evaluationExistsForPrompt(ctx, event.ProjectID, prompt); err == nil && !exists {
			if err := aet.triggerEvaluation(ctx, candidate); err != nil {
				log.Printf("Failed to trigger auto-evaluation: %v", err)
				return err
			}
		} else if err != nil {
			log.Printf("Error checking existing evaluations: %v", err)
		}
	}

	return nil
}

// extractPromptFromEvent extracts the prompt text from various possible fields in the event
func (aet *AutoEvaluationTrigger) extractPromptFromEvent(event AgentTrackingEvent) string {
	// Check common field names for prompts
	promptFields := []string{"prompt", "system_prompt", "user_prompt", "instruction", "query", "messages"}
	
	// First check input
	for _, field := range promptFields {
		if value, exists := event.Input[field]; exists {
			if prompt := aet.convertToString(value); prompt != "" {
				return prompt
			}
		}
	}

	// Check metadata
	for _, field := range promptFields {
		if value, exists := event.Metadata[field]; exists {
			if prompt := aet.convertToString(value); prompt != "" {
				return prompt
			}
		}
	}

	// Check for messages array (chat format)
	if messages, exists := event.Input["messages"]; exists {
		if prompt := aet.extractFromMessages(messages); prompt != "" {
			return prompt
		}
	}

	return ""
}

// convertToString safely converts various types to string
func (aet *AutoEvaluationTrigger) convertToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		// Try JSON marshaling for complex types
		if bytes, err := json.Marshal(v); err == nil {
			return string(bytes)
		}
		return fmt.Sprintf("%v", v)
	}
}

// extractFromMessages extracts prompt from chat messages format
func (aet *AutoEvaluationTrigger) extractFromMessages(messages interface{}) string {
	messagesJSON, err := json.Marshal(messages)
	if err != nil {
		return ""
	}

	var messageArray []map[string]interface{}
	if err := json.Unmarshal(messagesJSON, &messageArray); err != nil {
		return ""
	}

	var prompts []string
	for _, message := range messageArray {
		if content, exists := message["content"]; exists {
			if contentStr := aet.convertToString(content); contentStr != "" {
				prompts = append(prompts, contentStr)
			}
		}
	}

	return strings.Join(prompts, "\n")
}

// isValidPrompt checks if a prompt is valid for evaluation
func (aet *AutoEvaluationTrigger) isValidPrompt(prompt string) bool {
	length := len(prompt)
	return length >= aet.config.MinPromptLength && length <= aet.config.MaxPromptLength
}

// shouldExcludePrompt checks if a prompt should be excluded based on patterns
func (aet *AutoEvaluationTrigger) shouldExcludePrompt(prompt string) bool {
	promptLower := strings.ToLower(prompt)
	
	for _, pattern := range aet.config.ExcludePatterns {
		if strings.Contains(promptLower, strings.ToLower(pattern)) {
			return true
		}
	}
	
	return false
}

// isIncludedTaskType checks if a task type should trigger evaluation
func (aet *AutoEvaluationTrigger) isIncludedTaskType(taskType TaskType) bool {
	for _, included := range aet.config.IncludeTaskTypes {
		if taskType == included {
			return true
		}
	}
	return false
}

// getOrCreatePromptCandidate gets or creates a prompt candidate
func (aet *AutoEvaluationTrigger) getOrCreatePromptCandidate(projectID int, prompt string, taskType TaskType, event AgentTrackingEvent) (*PromptCandidate, error) {
	// In a real implementation, this would be stored in a database or cache
	// For now, we'll create a new candidate each time
	return &PromptCandidate{
		ProjectID:      projectID,
		Prompt:         prompt,
		FirstSeen:      event.StartTime,
		LastSeen:       event.EndTime,
		ExecutionCount: 1,
		TaskType:       taskType,
		Examples:       []Example{},
		SampleInputs:   []map[string]interface{}{event.Input},
		SampleOutputs:  []map[string]interface{}{event.Output},
	}, nil
}

// shouldTriggerEvaluation determines if an evaluation should be triggered
func (aet *AutoEvaluationTrigger) shouldTriggerEvaluation(candidate *PromptCandidate) bool {
	// Check if we have enough executions
	if candidate.ExecutionCount < aet.config.TriggerThreshold {
		return false
	}

	// Check if enough time has passed since first seen
	timeSinceFirst := time.Since(candidate.FirstSeen)
	if timeSinceFirst < aet.config.DelayBetweenRuns {
		return false
	}

	// Additional heuristics could be added here
	// - Check if prompt has changed significantly
	// - Check if performance seems to be declining
	// - Check if this is a new type of prompt we haven't seen

	return true
}

// evaluationExistsForPrompt checks if an evaluation already exists for this prompt
func (aet *AutoEvaluationTrigger) evaluationExistsForPrompt(ctx context.Context, projectID int, prompt string) (bool, error) {
	// List recent evaluations for the project
	evaluations, err := aet.orchestrator.ListEvaluations(ctx, projectID, ListOptions{
		Limit:   50,
		OrderBy: "created_at",
		SortDesc: true,
	})
	if err != nil {
		return false, err
	}

	// Check if any evaluation has a similar prompt
	for _, eval := range evaluations {
		if eval.PromptAnalysis != nil && aet.isSimilarPrompt(prompt, eval.PromptAnalysis.PromptText) {
			// If evaluation is recent (within last 24 hours), don't create another
			if time.Since(eval.CreatedAt) < 24*time.Hour {
				return true, nil
			}
		}
	}

	return false, nil
}

// isSimilarPrompt checks if two prompts are similar (simple implementation)
func (aet *AutoEvaluationTrigger) isSimilarPrompt(prompt1, prompt2 string) bool {
	// Simple similarity check - in production, you might use more sophisticated methods
	if prompt1 == prompt2 {
		return true
	}

	// Check if one is a substring of the other (for minor variations)
	p1Lower := strings.ToLower(strings.TrimSpace(prompt1))
	p2Lower := strings.ToLower(strings.TrimSpace(prompt2))
	
	// Calculate simple similarity ratio
	similarity := aet.calculateSimilarity(p1Lower, p2Lower)
	return similarity > 0.8 // 80% similarity threshold
}

// calculateSimilarity calculates a simple similarity ratio between two strings
func (aet *AutoEvaluationTrigger) calculateSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}
	
	if len(s1) == 0 || len(s2) == 0 {
		return 0.0
	}

	// Simple character-based similarity
	longer := s1
	shorter := s2
	if len(s2) > len(s1) {
		longer = s2
		shorter = s1
	}

	matches := 0
	for i, char := range shorter {
		if i < len(longer) && longer[i] == byte(char) {
			matches++
		}
	}

	return float64(matches) / float64(len(longer))
}

// triggerEvaluation creates and runs an evaluation for the prompt candidate
func (aet *AutoEvaluationTrigger) triggerEvaluation(ctx context.Context, candidate *PromptCandidate) error {
	// Generate evaluation name
	evaluationName := fmt.Sprintf("Auto-evaluation %s", time.Now().Format("2006-01-02 15:04"))
	
	// Create evaluation options
	options := EvaluationOptions{
		Name:        evaluationName,
		Description: fmt.Sprintf("Automatically generated evaluation for %s prompt (triggered after %d executions)", candidate.TaskType, candidate.ExecutionCount),
		TestGeneratorOptions: TestGeneratorOptions{
			NormalCases:      15,
			EdgeCases:        8,
			AdversarialCases: 5,
		},
		AutoSuggest: true,
	}

	// Create the evaluation
	evaluation, err := aet.orchestrator.CreateEvaluation(ctx, candidate.ProjectID, candidate.Prompt, options)
	if err != nil {
		return fmt.Errorf("failed to create auto-evaluation: %w", err)
	}

	// Run the evaluation asynchronously
	go func() {
		evalCtx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		
		_, err := aet.orchestrator.RunEvaluation(evalCtx, strconv.Itoa(evaluation.ID))
		if err != nil {
			log.Printf("Auto-evaluation %s failed: %v", evaluation.ID, err)
		} else {
			log.Printf("Auto-evaluation %s completed successfully", evaluation.ID)
		}
	}()

	log.Printf("Created auto-evaluation %s for project %d", evaluation.ID, candidate.ProjectID)
	return nil
}

// GetTriggerStats returns statistics about the auto-evaluation trigger
func (aet *AutoEvaluationTrigger) GetTriggerStats() map[string]interface{} {
	return map[string]interface{}{
		"enabled":               aet.config.EnableAutoEvaluation,
		"trigger_threshold":     aet.config.TriggerThreshold,
		"min_prompt_length":     aet.config.MinPromptLength,
		"max_prompt_length":     aet.config.MaxPromptLength,
		"delay_between_runs":    aet.config.DelayBetweenRuns.String(),
		"included_task_types":   aet.config.IncludeTaskTypes,
		"exclude_patterns":      aet.config.ExcludePatterns,
	}
}

// UpdateConfig updates the trigger configuration
func (aet *AutoEvaluationTrigger) UpdateConfig(config TriggerConfig) {
	aet.config = config
}