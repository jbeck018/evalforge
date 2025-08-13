package evaluation

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

)

// DefaultPromptOptimizer implements the PromptOptimizer interface
type DefaultPromptOptimizer struct {
	llmClient LLMClient
}

// NewPromptOptimizer creates a new prompt optimizer
func NewPromptOptimizer(llmClient LLMClient) *DefaultPromptOptimizer {
	return &DefaultPromptOptimizer{
		llmClient: llmClient,
	}
}

// SuggestImprovements suggests improvements based on evaluation results
func (po *DefaultPromptOptimizer) SuggestImprovements(ctx context.Context, prompt string, metrics *EvaluationMetrics, errorAnalysis *ErrorAnalysis) ([]OptimizationSuggestion, error) {
	var suggestions []OptimizationSuggestion

	// Analyze different aspects and generate suggestions
	if metrics.OverallScore < 0.8 {
		// Low overall performance - suggest clarity improvements
		if suggestion, err := po.OptimizeForClarity(ctx, prompt, errorAnalysis); err == nil && suggestion != nil {
			suggestions = append(suggestions, *suggestion)
		}
	}

	if metrics.PassRate < 0.7 {
		// Low pass rate - suggest accuracy improvements
		if suggestion, err := po.OptimizeForAccuracy(ctx, prompt, metrics); err == nil && suggestion != nil {
			suggestions = append(suggestions, *suggestion)
		}
	}

	// Check for consistency issues
	if errorAnalysis != nil && errorAnalysis.InconsistentCases > 0.15 {
		if suggestion, err := po.OptimizeForConsistency(ctx, prompt, nil); err == nil && suggestion != nil {
			suggestions = append(suggestions, *suggestion)
		}
	}

	// Check for format issues
	if errorAnalysis != nil && errorAnalysis.FormatErrors > 0.1 {
		if suggestion, err := po.optimizeForFormat(ctx, prompt, errorAnalysis); err == nil && suggestion != nil {
			suggestions = append(suggestions, *suggestion)
		}
	}

	// Suggest examples if prompt lacks them and performance is poor
	if !po.hasExamples(prompt) && metrics.OverallScore < 0.85 {
		if suggestion, err := po.addExamples(ctx, prompt, metrics, errorAnalysis); err == nil && suggestion != nil {
			suggestions = append(suggestions, *suggestion)
		}
	}

	// Classification-specific suggestions
	if metrics.ClassificationMetrics != nil {
		classificationSuggestions := po.generateClassificationSuggestions(ctx, prompt, metrics.ClassificationMetrics)
		suggestions = append(suggestions, classificationSuggestions...)
	}

	// Generation-specific suggestions
	if metrics.GenerationMetrics != nil {
		generationSuggestions := po.generateGenerationSuggestions(ctx, prompt, metrics.GenerationMetrics)
		suggestions = append(suggestions, generationSuggestions...)
	}

	// Set timestamps and IDs
	for i := range suggestions {
		suggestions[i].ID = 0
		suggestions[i].CreatedAt = time.Now()
		suggestions[i].UpdatedAt = time.Now()
		suggestions[i].Status = "pending"
	}

	return suggestions, nil
}

// OptimizeForClarity suggests improvements for prompt clarity
func (po *DefaultPromptOptimizer) OptimizeForClarity(ctx context.Context, prompt string, errorAnalysis *ErrorAnalysis) (*OptimizationSuggestion, error) {
	clarityPrompt := po.buildClarityOptimizationPrompt(prompt, errorAnalysis)

	response, err := po.llmClient.CompleteWithOptions(ctx, clarityPrompt, LLMOptions{
		Temperature: 0.3,
		MaxTokens:   1500,
		Model:       "gpt-4",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate clarity optimization: %w", err)
	}

	return po.parseSuggestionResponse(response, "clarity", prompt)
}

// OptimizeForAccuracy suggests improvements for better accuracy
func (po *DefaultPromptOptimizer) OptimizeForAccuracy(ctx context.Context, prompt string, metrics *EvaluationMetrics) (*OptimizationSuggestion, error) {
	accuracyPrompt := po.buildAccuracyOptimizationPrompt(prompt, metrics)

	response, err := po.llmClient.CompleteWithOptions(ctx, accuracyPrompt, LLMOptions{
		Temperature: 0.3,
		MaxTokens:   1500,
		Model:       "gpt-4",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate accuracy optimization: %w", err)
	}

	return po.parseSuggestionResponse(response, "accuracy", prompt)
}

// OptimizeForConsistency suggests improvements for consistency
func (po *DefaultPromptOptimizer) OptimizeForConsistency(ctx context.Context, prompt string, testCases []TestCase) (*OptimizationSuggestion, error) {
	consistencyPrompt := po.buildConsistencyOptimizationPrompt(prompt, testCases)

	response, err := po.llmClient.CompleteWithOptions(ctx, consistencyPrompt, LLMOptions{
		Temperature: 0.2, // Lower temperature for consistency
		MaxTokens:   1500,
		Model:       "gpt-4",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate consistency optimization: %w", err)
	}

	return po.parseSuggestionResponse(response, "consistency", prompt)
}

// optimizeForFormat suggests improvements for output format
func (po *DefaultPromptOptimizer) optimizeForFormat(ctx context.Context, prompt string, errorAnalysis *ErrorAnalysis) (*OptimizationSuggestion, error) {
	formatPrompt := po.buildFormatOptimizationPrompt(prompt, errorAnalysis)

	response, err := po.llmClient.CompleteWithOptions(ctx, formatPrompt, LLMOptions{
		Temperature: 0.2,
		MaxTokens:   1500,
		Model:       "gpt-4",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate format optimization: %w", err)
	}

	return po.parseSuggestionResponse(response, "format", prompt)
}

// addExamples suggests adding few-shot examples
func (po *DefaultPromptOptimizer) addExamples(ctx context.Context, prompt string, metrics *EvaluationMetrics, errorAnalysis *ErrorAnalysis) (*OptimizationSuggestion, error) {
	examplesPrompt := po.buildExamplesOptimizationPrompt(prompt, metrics, errorAnalysis)

	response, err := po.llmClient.CompleteWithOptions(ctx, examplesPrompt, LLMOptions{
		Temperature: 0.4,
		MaxTokens:   2000,
		Model:       "gpt-4",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate examples optimization: %w", err)
	}

	return po.parseSuggestionResponse(response, "examples", prompt)
}

// Build optimization prompts
func (po *DefaultPromptOptimizer) buildClarityOptimizationPrompt(prompt string, errorAnalysis *ErrorAnalysis) string {
	basePrompt := fmt.Sprintf(`You are an expert prompt engineer. Analyze the following prompt and suggest improvements to make it clearer and less ambiguous.

Original Prompt:
"""
%s
"""

`, prompt)

	if errorAnalysis != nil {
		basePrompt += fmt.Sprintf(`Error Analysis:
- Common Errors: %v
- Ambiguous Cases: %.2f%%
- Logic Errors: %.2f%%

`, errorAnalysis.CommonErrors, errorAnalysis.AmbiguousCases*100, errorAnalysis.LogicErrors*100)
	}

	basePrompt += `Please provide a JSON response with the following structure:
{
  "title": "Improve prompt clarity",
  "description": "Brief description of the improvement",
  "new_prompt": "The improved prompt text",
  "expected_impact": 0.15,
  "confidence": 0.8,
  "priority": "high|medium|low",
  "reasoning": "Detailed explanation of why this improvement will help",
  "examples": [
    {
      "input": {"example": "input"},
      "output": {"example": "output"},
      "explanation": "why this example helps"
    }
  ]
}

Focus on:
1. Removing ambiguous language
2. Adding specific instructions
3. Clarifying expected output format
4. Defining edge case handling
5. Providing clear constraints

Return only the JSON response.`

	return basePrompt
}

func (po *DefaultPromptOptimizer) buildAccuracyOptimizationPrompt(prompt string, metrics *EvaluationMetrics) string {
	basePrompt := fmt.Sprintf(`You are an expert prompt engineer. The following prompt has accuracy issues. Suggest improvements to increase performance.

Original Prompt:
"""
%s
"""

Performance Metrics:
- Overall Score: %.2f
- Pass Rate: %.2f%%
`, prompt, metrics.OverallScore, metrics.PassRate*100)

	if metrics.ClassificationMetrics != nil {
		basePrompt += fmt.Sprintf(`- Accuracy: %.2f%%
- Macro F1: %.2f
- Weighted F1: %.2f
`, metrics.ClassificationMetrics.Accuracy*100, metrics.ClassificationMetrics.MacroF1, metrics.ClassificationMetrics.WeightedF1)
	}

	basePrompt += `
Please provide a JSON response with improvements focused on accuracy:
{
  "title": "Improve prompt accuracy",
  "description": "Brief description of the improvement",
  "new_prompt": "The improved prompt text",
  "expected_impact": 0.20,
  "confidence": 0.85,
  "priority": "high",
  "reasoning": "Detailed explanation of accuracy improvements",
  "examples": []
}

Focus on:
1. More precise task definitions
2. Better constraint specifications
3. Clearer decision criteria
4. Improved output specifications
5. Better handling of edge cases

Return only the JSON response.`

	return basePrompt
}

func (po *DefaultPromptOptimizer) buildConsistencyOptimizationPrompt(prompt string, testCases []TestCase) string {
	basePrompt := fmt.Sprintf(`You are an expert prompt engineer. The following prompt produces inconsistent results. Suggest improvements for better consistency.

Original Prompt:
"""
%s
"""

The prompt needs to produce more consistent outputs across similar inputs.

`, prompt)

	if testCases != nil && len(testCases) > 0 {
		basePrompt += "Sample test cases show inconsistency patterns.\n\n"
	}

	basePrompt += `Please provide a JSON response focused on consistency:
{
  "title": "Improve prompt consistency",
  "description": "Brief description of the improvement",
  "new_prompt": "The improved prompt text",
  "expected_impact": 0.18,
  "confidence": 0.75,
  "priority": "medium",
  "reasoning": "Detailed explanation of consistency improvements",
  "examples": []
}

Focus on:
1. Standardizing output format
2. Adding explicit rules and constraints
3. Providing clear decision frameworks
4. Reducing subjective language
5. Adding consistency checks

Return only the JSON response.`

	return basePrompt
}

func (po *DefaultPromptOptimizer) buildFormatOptimizationPrompt(prompt string, errorAnalysis *ErrorAnalysis) string {
	basePrompt := fmt.Sprintf(`You are an expert prompt engineer. The following prompt has output format issues that need to be addressed.

Original Prompt:
"""
%s
"""

Format Issues:
- Format Errors: %.2f%%

`, prompt, errorAnalysis.FormatErrors*100)

	basePrompt += `Please provide a JSON response focused on format improvements:
{
  "title": "Improve output format specification",
  "description": "Brief description of the improvement", 
  "new_prompt": "The improved prompt text",
  "expected_impact": 0.12,
  "confidence": 0.9,
  "priority": "medium",
  "reasoning": "Detailed explanation of format improvements",
  "examples": []
}

Focus on:
1. Explicit output format specification
2. Clear structure requirements
3. Field definitions and types
4. Format validation rules
5. Example output formats

Return only the JSON response.`

	return basePrompt
}

func (po *DefaultPromptOptimizer) buildExamplesOptimizationPrompt(prompt string, metrics *EvaluationMetrics, errorAnalysis *ErrorAnalysis) string {
	basePrompt := fmt.Sprintf(`You are an expert prompt engineer. The following prompt would benefit from few-shot examples to improve performance.

Original Prompt:
"""
%s
"""

Current Performance:
- Overall Score: %.2f
- Pass Rate: %.2f%%

`, prompt, metrics.OverallScore, metrics.PassRate*100)

	if errorAnalysis != nil {
		basePrompt += fmt.Sprintf(`Common Issues:
- Ambiguous Cases: %.2f%%
- Format Errors: %.2f%%

`, errorAnalysis.AmbiguousCases*100, errorAnalysis.FormatErrors*100)
	}

	basePrompt += `Please provide a JSON response with few-shot examples:
{
  "title": "Add few-shot examples",
  "description": "Brief description of the improvement",
  "new_prompt": "The improved prompt text with examples",
  "expected_impact": 0.25,
  "confidence": 0.85,
  "priority": "high",
  "reasoning": "Detailed explanation of why examples will help",
  "examples": [
    {
      "input": {"example": "input"},
      "output": {"example": "output"},
      "explanation": "why this example is helpful"
    }
  ]
}

Focus on:
1. 2-4 high-quality examples
2. Cover different scenarios (normal, edge cases)
3. Show correct output format
4. Demonstrate decision criteria
5. Clear input-output relationships

Return only the JSON response.`

	return basePrompt
}

// Generate task-specific suggestions
func (po *DefaultPromptOptimizer) generateClassificationSuggestions(ctx context.Context, prompt string, metrics *ClassificationMetrics) []OptimizationSuggestion {
	var suggestions []OptimizationSuggestion

	// Low accuracy suggestion
	if metrics.Accuracy < 0.8 {
		suggestion := OptimizationSuggestion{
			Type:            "classification_accuracy",
			Title:           "Improve classification accuracy",
			Description:     "Add more specific classification criteria and examples",
			Priority:        "high",
			ExpectedImpact:  0.15,
			Confidence:      0.8,
			Reasoning:       fmt.Sprintf("Current accuracy is %.2f%%, below optimal threshold", metrics.Accuracy*100),
			OldPrompt:       prompt,
			NewPrompt:       po.generateClassificationAccuracyPrompt(prompt, metrics),
		}
		suggestions = append(suggestions, suggestion)
	}

	// Class imbalance suggestion
	if po.hasClassImbalance(metrics) {
		suggestion := OptimizationSuggestion{
			Type:            "class_balance",
			Title:           "Address class imbalance",
			Description:     "Improve handling of underperforming classes",
			Priority:        "medium",
			ExpectedImpact:  0.12,
			Confidence:      0.75,
			Reasoning:       "Some classes have significantly lower precision/recall",
			OldPrompt:       prompt,
			NewPrompt:       po.generateClassBalancePrompt(prompt, metrics),
		}
		suggestions = append(suggestions, suggestion)
	}

	return suggestions
}

func (po *DefaultPromptOptimizer) generateGenerationSuggestions(ctx context.Context, prompt string, metrics *GenerationMetrics) []OptimizationSuggestion {
	var suggestions []OptimizationSuggestion

	// Low BLEU score suggestion
	if metrics.BLEU < 0.6 {
		suggestion := OptimizationSuggestion{
			Type:            "generation_quality",
			Title:           "Improve generation quality",
			Description:     "Enhance prompt for better text generation",
			Priority:        "high",
			ExpectedImpact:  0.18,
			Confidence:      0.8,
			Reasoning:       fmt.Sprintf("BLEU score is %.2f, indicating poor text quality", metrics.BLEU),
			OldPrompt:       prompt,
			NewPrompt:       po.generateQualityPrompt(prompt, metrics),
		}
		suggestions = append(suggestions, suggestion)
	}

	// Low diversity suggestion
	if metrics.Diversity < 0.5 {
		suggestion := OptimizationSuggestion{
			Type:            "generation_diversity",
			Title:           "Improve lexical diversity",
			Description:     "Encourage more varied vocabulary and structures",
			Priority:        "medium",
			ExpectedImpact:  0.1,
			Confidence:      0.7,
			Reasoning:       fmt.Sprintf("Lexical diversity is %.2f, indicating repetitive output", metrics.Diversity),
			OldPrompt:       prompt,
			NewPrompt:       po.generateDiversityPrompt(prompt, metrics),
		}
		suggestions = append(suggestions, suggestion)
	}

	return suggestions
}

// Helper methods
func (po *DefaultPromptOptimizer) hasExamples(prompt string) bool {
	promptLower := strings.ToLower(prompt)
	examples := []string{"example", "for instance", "such as", "e.g.", "input:", "output:"}
	
	for _, example := range examples {
		if strings.Contains(promptLower, example) {
			return true
		}
	}
	return false
}

func (po *DefaultPromptOptimizer) hasClassImbalance(metrics *ClassificationMetrics) bool {
	if len(metrics.F1Score) < 2 {
		return false
	}

	var f1Scores []float64
	for _, f1 := range metrics.F1Score {
		f1Scores = append(f1Scores, f1)
	}

	// Simple check: if max F1 - min F1 > 0.3, consider it imbalanced
	min, max := f1Scores[0], f1Scores[0]
	for _, f1 := range f1Scores[1:] {
		if f1 < min {
			min = f1
		}
		if f1 > max {
			max = f1
		}
	}

	return max-min > 0.3
}

func (po *DefaultPromptOptimizer) parseSuggestionResponse(response, suggestionType, originalPrompt string) (*OptimizationSuggestion, error) {
	// Clean the response
	response = strings.TrimSpace(response)
	if strings.HasPrefix(response, "```json") {
		response = strings.TrimPrefix(response, "```json")
		response = strings.TrimSuffix(response, "```")
	}
	response = strings.TrimSpace(response)

	var suggestionData struct {
		Title           string    `json:"title"`
		Description     string    `json:"description"`
		NewPrompt       string    `json:"new_prompt"`
		ExpectedImpact  float64   `json:"expected_impact"`
		Confidence      float64   `json:"confidence"`
		Priority        string    `json:"priority"`
		Reasoning       string    `json:"reasoning"`
		Examples        []Example `json:"examples"`
	}

	if err := json.Unmarshal([]byte(response), &suggestionData); err != nil {
		return nil, fmt.Errorf("failed to parse suggestion response: %w", err)
	}

	suggestion := &OptimizationSuggestion{
		Type:            suggestionType,
		Title:           suggestionData.Title,
		Description:     suggestionData.Description,
		OldPrompt:       originalPrompt,
		NewPrompt:       suggestionData.NewPrompt,
		ExpectedImpact:  suggestionData.ExpectedImpact,
		Confidence:      suggestionData.Confidence,
		Priority:        suggestionData.Priority,
		Reasoning:       suggestionData.Reasoning,
		Examples:        suggestionData.Examples,
	}

	return suggestion, nil
}

// ValidateSuggestion validates an optimization suggestion
func (po *DefaultPromptOptimizer) ValidateSuggestion(suggestion *OptimizationSuggestion) error {
	if suggestion == nil {
		return fmt.Errorf("suggestion cannot be nil")
	}

	if suggestion.Title == "" {
		return fmt.Errorf("suggestion title cannot be empty")
	}

	if suggestion.Description == "" {
		return fmt.Errorf("suggestion description cannot be empty")
	}

	if suggestion.NewPrompt == "" {
		return fmt.Errorf("suggestion new_prompt cannot be empty")
	}

	if suggestion.ExpectedImpact < 0 || suggestion.ExpectedImpact > 1 {
		return fmt.Errorf("expected_impact must be between 0 and 1")
	}

	if suggestion.Confidence < 0 || suggestion.Confidence > 1 {
		return fmt.Errorf("confidence must be between 0 and 1")
	}

	validPriorities := map[string]bool{"high": true, "medium": true, "low": true}
	if !validPriorities[suggestion.Priority] {
		return fmt.Errorf("invalid priority: %s", suggestion.Priority)
	}

	validStatuses := map[string]bool{"pending": true, "accepted": true, "rejected": true, "applied": true}
	if suggestion.Status != "" && !validStatuses[suggestion.Status] {
		return fmt.Errorf("invalid status: %s", suggestion.Status)
	}

	return nil
}

// Simple prompt generation helpers for specific optimization types
func (po *DefaultPromptOptimizer) generateClassificationAccuracyPrompt(originalPrompt string, metrics *ClassificationMetrics) string {
	return originalPrompt + "\n\nPlease be more specific in your classification criteria and consider the following classes equally: " + 
		strings.Join(po.getClassesFromMetrics(metrics), ", ")
}

func (po *DefaultPromptOptimizer) generateClassBalancePrompt(originalPrompt string, metrics *ClassificationMetrics) string {
	weakClasses := po.getWeakClasses(metrics)
	return originalPrompt + fmt.Sprintf("\n\nPay special attention to these classes which may be underperformed: %s", 
		strings.Join(weakClasses, ", "))
}

func (po *DefaultPromptOptimizer) generateQualityPrompt(originalPrompt string, metrics *GenerationMetrics) string {
	return originalPrompt + "\n\nFocus on generating high-quality, coherent text that closely matches the expected output style and content."
}

func (po *DefaultPromptOptimizer) generateDiversityPrompt(originalPrompt string, metrics *GenerationMetrics) string {
	return originalPrompt + "\n\nUse varied vocabulary and sentence structures. Avoid repetitive phrasing and aim for lexical diversity."
}

func (po *DefaultPromptOptimizer) getClassesFromMetrics(metrics *ClassificationMetrics) []string {
	var classes []string
	for class := range metrics.F1Score {
		classes = append(classes, class)
	}
	return classes
}

func (po *DefaultPromptOptimizer) getWeakClasses(metrics *ClassificationMetrics) []string {
	var weakClasses []string
	for class, f1 := range metrics.F1Score {
		if f1 < 0.6 { // Consider classes with F1 < 0.6 as weak
			weakClasses = append(weakClasses, class)
		}
	}
	return weakClasses
}