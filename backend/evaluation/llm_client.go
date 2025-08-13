package evaluation

import (
	"context"
	"encoding/json"
	"strings"
)

// MockLLMClient implements LLMClient for testing and development
type MockLLMClient struct{}

// NewMockLLMClient creates a new mock LLM client
func NewMockLLMClient() *MockLLMClient {
	return &MockLLMClient{}
}

// Complete implements LLMClient.Complete
func (m *MockLLMClient) Complete(ctx context.Context, prompt string) (string, error) {
	return m.CompleteWithOptions(ctx, prompt, LLMOptions{})
}

// CompleteWithOptions implements LLMClient.CompleteWithOptions
func (m *MockLLMClient) CompleteWithOptions(ctx context.Context, prompt string, options LLMOptions) (string, error) {
	// Simple pattern matching to return appropriate responses
	promptLower := strings.ToLower(prompt)
	
	// If this is a prompt analysis request
	if strings.Contains(promptLower, "analyze the following prompt") {
		return m.generateAnalysisResponse(prompt), nil
	}
	
	// If this is a test case generation request
	if strings.Contains(promptLower, "generate") && strings.Contains(promptLower, "test cases") {
		return m.generateTestCasesResponse(prompt), nil
	}
	
	// If this is an optimization request
	if strings.Contains(promptLower, "suggest improvements") || strings.Contains(promptLower, "optimize") {
		return m.generateOptimizationResponse(prompt), nil
	}
	
	// Default response
	return "Mock LLM response for: " + prompt[:min(100, len(prompt))], nil
}

// ValidateConnection implements LLMClient.ValidateConnection
func (m *MockLLMClient) ValidateConnection(ctx context.Context) error {
	return nil // Mock always validates successfully
}

// generateAnalysisResponse generates a mock analysis response
func (m *MockLLMClient) generateAnalysisResponse(prompt string) string {
	// Extract the actual prompt being analyzed
	lines := strings.Split(prompt, "\n")
	var targetPrompt string
	inPromptSection := false
	
	for _, line := range lines {
		if strings.Contains(line, `"""`) {
			inPromptSection = !inPromptSection
			continue
		}
		if inPromptSection {
			targetPrompt += line + " "
		}
	}
	
	targetPrompt = strings.TrimSpace(targetPrompt)
	taskType, outputSchema := m.detectTaskType(targetPrompt)
	
	analysis := map[string]interface{}{
		"task_type":  taskType,
		"confidence": 0.85,
		"input_schema": map[string]interface{}{
			"type":        "text",
			"fields":      map[string]interface{}{},
			"required":    []string{},
			"constraints": map[string]interface{}{},
		},
		"output_schema": outputSchema,
		"constraints": []map[string]interface{}{
			{
				"type":        "format",
				"description": "Output must follow the specified format",
				"rule":        "structured_output",
				"severity":    "error",
			},
		},
		"examples": []map[string]interface{}{},
	}
	
	response, _ := json.MarshalIndent(analysis, "", "  ")
	return string(response)
}

// detectTaskType detects the task type from a prompt
func (m *MockLLMClient) detectTaskType(prompt string) (string, map[string]interface{}) {
	promptLower := strings.ToLower(prompt)
	
	// Classification indicators
	if strings.Contains(promptLower, "classify") || strings.Contains(promptLower, "categorize") || 
	   strings.Contains(promptLower, "sentiment") || strings.Contains(promptLower, "positive") && strings.Contains(promptLower, "negative") {
		return "classification", map[string]interface{}{
			"type":    "classification",
			"format":  "single_class",
			"classes": []string{"positive", "negative", "neutral"},
			"fields":  map[string]interface{}{},
			"constraints": map[string]interface{}{
				"valid_classes": []string{"positive", "negative", "neutral"},
			},
		}
	}
	
	// Generation indicators
	if strings.Contains(promptLower, "generate") || strings.Contains(promptLower, "create") || 
	   strings.Contains(promptLower, "write") || strings.Contains(promptLower, "compose") {
		return "generation", map[string]interface{}{
			"type":   "text",
			"format": "free_text",
			"classes": []string{},
			"fields": map[string]interface{}{},
			"constraints": map[string]interface{}{
				"min_length": 10,
				"max_length": 1000,
			},
		}
	}
	
	// Extraction indicators
	if strings.Contains(promptLower, "extract") || strings.Contains(promptLower, "find") || strings.Contains(promptLower, "identify") {
		return "extraction", map[string]interface{}{
			"type":   "structured",
			"format": "json",
			"classes": []string{},
			"fields": map[string]interface{}{
				"extracted_entities": "array",
				"confidence": "number",
			},
			"constraints": map[string]interface{}{},
		}
	}
	
	// Summarization indicators
	if strings.Contains(promptLower, "summarize") || strings.Contains(promptLower, "summary") {
		return "summarization", map[string]interface{}{
			"type":   "text",
			"format": "paragraph",
			"classes": []string{},
			"fields": map[string]interface{}{},
			"constraints": map[string]interface{}{
				"max_length": 200,
			},
		}
	}
	
	// Q&A indicators
	if strings.Contains(promptLower, "answer") || strings.Contains(promptLower, "question") {
		return "question_answering", map[string]interface{}{
			"type":   "text",
			"format": "answer",
			"classes": []string{},
			"fields": map[string]interface{}{},
			"constraints": map[string]interface{}{},
		}
	}
	
	// Default to generation
	return "generation", map[string]interface{}{
		"type":   "text",
		"format": "free_text",
		"classes": []string{},
		"fields": map[string]interface{}{},
		"constraints": map[string]interface{}{},
	}
}

// generateTestCasesResponse generates mock test cases
func (m *MockLLMClient) generateTestCasesResponse(prompt string) string {
	// Parse the prompt to understand what type of test cases to generate
	promptLower := strings.ToLower(prompt)
	
	var testCases []map[string]interface{}
	
	if strings.Contains(promptLower, "classification") {
		testCases = []map[string]interface{}{
			{
				"name":        "Positive sentiment example",
				"description": "Clear positive sentiment text",
				"input":       map[string]interface{}{"text": "I love this product! It's amazing and works perfectly."},
				"expected_output": map[string]interface{}{"sentiment": "positive"},
				"category":    "normal",
				"weight":      1.0,
			},
			{
				"name":        "Negative sentiment example", 
				"description": "Clear negative sentiment text",
				"input":       map[string]interface{}{"text": "This is terrible and doesn't work at all."},
				"expected_output": map[string]interface{}{"sentiment": "negative"},
				"category":    "normal",
				"weight":      1.0,
			},
			{
				"name":        "Neutral sentiment example",
				"description": "Neutral sentiment text",
				"input":       map[string]interface{}{"text": "The product has standard features."},
				"expected_output": map[string]interface{}{"sentiment": "neutral"},
				"category":    "normal",
				"weight":      1.0,
			},
			{
				"name":        "Ambiguous case",
				"description": "Sarcastic text that could be misclassified",
				"input":       map[string]interface{}{"text": "Oh great, another broken feature."},
				"expected_output": map[string]interface{}{"sentiment": "negative"},
				"category":    "edge_case",
				"weight":      1.5,
			},
		}
	} else {
		// Generic test cases
		testCases = []map[string]interface{}{
			{
				"name":        "Basic example",
				"description": "Standard input case",
				"input":       map[string]interface{}{"text": "Sample input text"},
				"expected_output": map[string]interface{}{"result": "Sample output"},
				"category":    "normal",
				"weight":      1.0,
			},
			{
				"name":        "Edge case",
				"description": "Boundary condition",
				"input":       map[string]interface{}{"text": ""},
				"expected_output": map[string]interface{}{"result": "Empty input handled"},
				"category":    "edge_case",
				"weight":      1.2,
			},
		}
	}
	
	response, _ := json.MarshalIndent(testCases, "", "  ")
	return string(response)
}

// generateOptimizationResponse generates mock optimization suggestions
func (m *MockLLMClient) generateOptimizationResponse(prompt string) string {
	suggestions := []map[string]interface{}{
		{
			"type":            "clarity",
			"title":           "Improve prompt clarity",
			"description":     "Add more specific instructions to reduce ambiguity",
			"old_prompt":      "Classify the sentiment",
			"new_prompt":      "Classify the sentiment of the following text as positive, negative, or neutral. Consider the overall emotional tone.",
			"expected_impact": 0.15,
			"confidence":      0.8,
			"priority":        "high",
			"reasoning":       "The original prompt lacks specific output format and classification criteria",
		},
		{
			"type":            "examples",
			"title":           "Add few-shot examples",
			"description":     "Include examples to improve consistency",
			"old_prompt":      "Current prompt without examples",
			"new_prompt":      "Enhanced prompt with examples:\nExample 1: Input: 'Great product!' Output: positive\nExample 2: Input: 'Poor quality' Output: negative",
			"expected_impact": 0.12,
			"confidence":      0.85,
			"priority":        "medium",
			"reasoning":       "Few-shot examples help establish consistent output patterns",
		},
	}
	
	response, _ := json.MarshalIndent(suggestions, "", "  ")
	return string(response)
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}