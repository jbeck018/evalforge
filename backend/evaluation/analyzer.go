package evaluation

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

)

// DefaultPromptAnalyzer implements the PromptAnalyzer interface
type DefaultPromptAnalyzer struct {
	llmClient LLMClient
}

// NewPromptAnalyzer creates a new prompt analyzer
func NewPromptAnalyzer(llmClient LLMClient) *DefaultPromptAnalyzer {
	return &DefaultPromptAnalyzer{
		llmClient: llmClient,
	}
}

// AnalyzePrompt analyzes a prompt and returns analysis results
func (pa *DefaultPromptAnalyzer) AnalyzePrompt(ctx context.Context, prompt string, examples []Example) (*PromptAnalysis, error) {
	analysisPrompt := pa.buildAnalysisPrompt(prompt, examples)
	
	response, err := pa.llmClient.CompleteWithOptions(ctx, analysisPrompt, LLMOptions{
		Temperature: 0.1, // Low temperature for consistent analysis
		MaxTokens:   2000,
		Model:       "gpt-4",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to analyze prompt: %w", err)
	}

	analysis, err := pa.parseAnalysisResponse(response, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse analysis response: %w", err)
	}

	// Validate and enhance the analysis
	if err := pa.ValidateAnalysis(analysis); err != nil {
		return nil, fmt.Errorf("invalid analysis: %w", err)
	}

	// Set metadata
	analysis.ID = 0
	analysis.PromptID = 0 // This should be set by the caller
	analysis.PromptText = prompt
	analysis.CreatedAt = time.Now()
	analysis.UpdatedAt = time.Now()

	return analysis, nil
}

// buildAnalysisPrompt creates the analysis prompt for the LLM
func (pa *DefaultPromptAnalyzer) buildAnalysisPrompt(prompt string, examples []Example) string {
	examplesText := ""
	if len(examples) > 0 {
		examplesJSON, _ := json.MarshalIndent(examples, "  ", "  ")
		examplesText = fmt.Sprintf("\n\nExamples provided in prompt:\n%s", string(examplesJSON))
	}

	return fmt.Sprintf(`You are an expert prompt analyzer. Analyze the following prompt and determine its characteristics.

Prompt to analyze:
"""
%s
"""
%s

Please provide a detailed analysis in JSON format with the following structure:

{
  "task_type": "classification|generation|extraction|summarization|question_answering|transformation|completion",
  "confidence": 0.95,
  "input_schema": {
    "type": "text|json|structured",
    "fields": {},
    "required": [],
    "constraints": {}
  },
  "output_schema": {
    "type": "text|json|classification|structured", 
    "format": "specific format requirements",
    "classes": ["class1", "class2"] // only for classification
    "fields": {},
    "constraints": {}
  },
  "constraints": [
    {
      "type": "format|length|value|pattern",
      "description": "human readable description",
      "rule": "the actual constraint",
      "severity": "error|warning|info"
    }
  ],
  "examples": [
    {
      "input": {},
      "output": {},
      "explanation": "optional explanation"
    }
  ]
}

Analysis Guidelines:
1. **Task Type**: Identify the primary task:
   - classification: Categorizing input into predefined classes
   - generation: Creating new content (text, code, etc.)
   - extraction: Pulling specific information from input
   - summarization: Condensing information
   - question_answering: Answering questions based on context
   - transformation: Converting input from one format to another
   - completion: Finishing incomplete input

2. **Input Schema**: Describe expected input format and constraints
3. **Output Schema**: Describe expected output format, including classes for classification tasks
4. **Constraints**: Identify any rules, formatting requirements, or limitations
5. **Examples**: Extract any few-shot examples from the prompt

Provide only the JSON response, no additional text.`, prompt, examplesText)
}

// parseAnalysisResponse parses the LLM response into a PromptAnalysis struct
func (pa *DefaultPromptAnalyzer) parseAnalysisResponse(response, originalPrompt string) (*PromptAnalysis, error) {
	// Clean the response to extract JSON
	response = strings.TrimSpace(response)
	if strings.HasPrefix(response, "```json") {
		response = strings.TrimPrefix(response, "```json")
		response = strings.TrimSuffix(response, "```")
	}
	response = strings.TrimSpace(response)

	var analysisData struct {
		TaskType     string        `json:"task_type"`
		Confidence   float64       `json:"confidence"`
		InputSchema  InputSchema   `json:"input_schema"`
		OutputSchema OutputSchema  `json:"output_schema"`
		Constraints  []Constraint  `json:"constraints"`
		Examples     []Example     `json:"examples"`
	}

	if err := json.Unmarshal([]byte(response), &analysisData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal analysis response: %w", err)
	}

	// Convert to our internal types
	analysis := &PromptAnalysis{
		TaskType:     TaskType(analysisData.TaskType),
		InputSchema:  analysisData.InputSchema,
		OutputSchema: analysisData.OutputSchema,
		Constraints:  analysisData.Constraints,
		Examples:     analysisData.Examples,
		Confidence:   analysisData.Confidence,
	}

	// If confidence wasn't provided, estimate it
	if analysis.Confidence == 0 {
		analysis.Confidence = pa.estimateConfidence(originalPrompt, analysis)
	}

	return analysis, nil
}

// estimateConfidence estimates confidence based on prompt characteristics
func (pa *DefaultPromptAnalyzer) estimateConfidence(prompt string, analysis *PromptAnalysis) float64 {
	confidence := 0.5 // Base confidence

	// Higher confidence for clear task indicators
	taskIndicators := map[TaskType][]string{
		TaskClassification:  {"classify", "categorize", "determine", "identify the type", "label"},
		TaskGeneration:      {"generate", "create", "write", "produce", "compose"},
		TaskExtraction:      {"extract", "find", "identify", "pull out", "get"},
		TaskSummarization:   {"summarize", "summary", "condense", "brief"},
		TaskQA:              {"answer", "respond", "what", "how", "why", "explain"},
		TaskTransformation:  {"convert", "transform", "translate", "change"},
		TaskCompletion:      {"complete", "finish", "continue"},
	}

	promptLower := strings.ToLower(prompt)
	if indicators, exists := taskIndicators[analysis.TaskType]; exists {
		for _, indicator := range indicators {
			if strings.Contains(promptLower, indicator) {
				confidence += 0.1
				break
			}
		}
	}

	// Higher confidence for structured prompts
	if strings.Contains(prompt, "Format:") || strings.Contains(prompt, "Output:") {
		confidence += 0.15
	}

	// Higher confidence for examples
	if len(analysis.Examples) > 0 {
		confidence += 0.2
	}

	// Higher confidence for specific constraints
	if len(analysis.Constraints) > 0 {
		confidence += 0.1
	}

	// Cap confidence at 0.95
	if confidence > 0.95 {
		confidence = 0.95
	}

	return confidence
}

// ValidateAnalysis validates the analysis results
func (pa *DefaultPromptAnalyzer) ValidateAnalysis(analysis *PromptAnalysis) error {
	if analysis == nil {
		return fmt.Errorf("analysis cannot be nil")
	}

	// Validate task type
	validTaskTypes := map[TaskType]bool{
		TaskClassification:  true,
		TaskGeneration:      true,
		TaskExtraction:      true,
		TaskSummarization:   true,
		TaskQA:              true,
		TaskTransformation:  true,
		TaskCompletion:      true,
	}

	if !validTaskTypes[analysis.TaskType] {
		return fmt.Errorf("invalid task type: %s", analysis.TaskType)
	}

	// Validate confidence
	if analysis.Confidence < 0 || analysis.Confidence > 1 {
		return fmt.Errorf("confidence must be between 0 and 1, got: %f", analysis.Confidence)
	}

	// Validate classification-specific fields
	if analysis.TaskType == TaskClassification {
		if len(analysis.OutputSchema.Classes) == 0 {
			return fmt.Errorf("classification tasks must have output classes defined")
		}
	}

	// Validate constraints
	for i, constraint := range analysis.Constraints {
		if constraint.Type == "" {
			return fmt.Errorf("constraint %d missing type", i)
		}
		if constraint.Description == "" {
			return fmt.Errorf("constraint %d missing description", i)
		}
		validSeverities := map[string]bool{"error": true, "warning": true, "info": true}
		if !validSeverities[constraint.Severity] {
			return fmt.Errorf("constraint %d has invalid severity: %s", i, constraint.Severity)
		}
	}

	return nil
}

// UpdateAnalysis updates an existing analysis with a new prompt
func (pa *DefaultPromptAnalyzer) UpdateAnalysis(ctx context.Context, analysisID string, prompt string) (*PromptAnalysis, error) {
	// Parse the analysisID - now we expect an integer
	// For now, we'll just use 0 since we're migrating to integer IDs
	id := 0

	// Perform new analysis
	analysis, err := pa.AnalyzePrompt(ctx, prompt, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze updated prompt: %w", err)
	}

	// Update the ID and timestamp
	analysis.ID = id
	analysis.UpdatedAt = time.Now()

	return analysis, nil
}

// GetSupportedTaskTypes returns all supported task types
func (pa *DefaultPromptAnalyzer) GetSupportedTaskTypes() []TaskType {
	return []TaskType{
		TaskClassification,
		TaskGeneration,
		TaskExtraction,
		TaskSummarization,
		TaskQA,
		TaskTransformation,
		TaskCompletion,
	}
}

// AnalyzePromptBatch analyzes multiple prompts in batch
func (pa *DefaultPromptAnalyzer) AnalyzePromptBatch(ctx context.Context, prompts []string) ([]*PromptAnalysis, error) {
	results := make([]*PromptAnalysis, len(prompts))
	
	// TODO: Implement concurrent processing with rate limiting
	for i, prompt := range prompts {
		analysis, err := pa.AnalyzePrompt(ctx, prompt, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze prompt %d: %w", i, err)
		}
		results[i] = analysis
	}
	
	return results, nil
}

// ExtractPromptFeatures extracts key features from a prompt for analysis
func (pa *DefaultPromptAnalyzer) ExtractPromptFeatures(prompt string) map[string]interface{} {
	features := make(map[string]interface{})
	
	promptLower := strings.ToLower(prompt)
	
	// Basic statistics
	features["length"] = len(prompt)
	features["word_count"] = len(strings.Fields(prompt))
	features["line_count"] = len(strings.Split(prompt, "\n"))
	
	// Structural features
	features["has_examples"] = strings.Contains(prompt, "Example:") || strings.Contains(prompt, "Examples:")
	features["has_format_spec"] = strings.Contains(prompt, "Format:") || strings.Contains(prompt, "Output:")
	features["has_constraints"] = strings.Contains(promptLower, "must") || strings.Contains(promptLower, "should") || strings.Contains(promptLower, "cannot")
	
	// Task indicators
	taskKeywords := map[string][]string{
		"classification": {"classify", "categorize", "determine", "identify the type", "label"},
		"generation":     {"generate", "create", "write", "produce", "compose"},
		"extraction":     {"extract", "find", "identify", "pull out", "get"},
		"summarization":  {"summarize", "summary", "condense", "brief"},
		"qa":             {"answer", "respond", "what", "how", "why", "explain"},
		"transformation": {"convert", "transform", "translate", "change"},
		"completion":     {"complete", "finish", "continue"},
	}
	
	taskScores := make(map[string]int)
	for taskType, keywords := range taskKeywords {
		score := 0
		for _, keyword := range keywords {
			if strings.Contains(promptLower, keyword) {
				score++
			}
		}
		taskScores[taskType] = score
	}
	features["task_scores"] = taskScores
	
	return features
}