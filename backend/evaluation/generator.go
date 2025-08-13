package evaluation

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

)

// DefaultTestGenerator implements the TestGenerator interface
type DefaultTestGenerator struct {
	llmClient LLMClient
	random    *rand.Rand
}

// NewTestGenerator creates a new test case generator
func NewTestGenerator(llmClient LLMClient) *DefaultTestGenerator {
	return &DefaultTestGenerator{
		llmClient: llmClient,
		random:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GenerateTestCases generates test cases based on prompt analysis
func (tg *DefaultTestGenerator) GenerateTestCases(ctx context.Context, analysis *PromptAnalysis, options TestGeneratorOptions) ([]TestCase, error) {
	if analysis == nil {
		return nil, fmt.Errorf("prompt analysis cannot be nil")
	}

	// Set default options if not provided
	if options.NormalCases == 0 {
		options.NormalCases = 15
	}
	if options.EdgeCases == 0 {
		options.EdgeCases = 8
	}
	if options.AdversarialCases == 0 {
		options.AdversarialCases = 5
	}

	var allTestCases []TestCase
	evaluationID := 0

	// Generate normal test cases
	normalCases, err := tg.GenerateCategory(ctx, analysis, "normal", options.NormalCases)
	if err != nil {
		return nil, fmt.Errorf("failed to generate normal test cases: %w", err)
	}
	allTestCases = append(allTestCases, normalCases...)

	// Generate edge cases
	edgeCases, err := tg.GenerateCategory(ctx, analysis, "edge_case", options.EdgeCases)
	if err != nil {
		return nil, fmt.Errorf("failed to generate edge test cases: %w", err)
	}
	allTestCases = append(allTestCases, edgeCases...)

	// Generate adversarial cases
	adversarialCases, err := tg.GenerateCategory(ctx, analysis, "adversarial", options.AdversarialCases)
	if err != nil {
		return nil, fmt.Errorf("failed to generate adversarial test cases: %w", err)
	}
	allTestCases = append(allTestCases, adversarialCases...)

	// Set evaluation ID for all test cases
	for i := range allTestCases {
		allTestCases[i].EvaluationID = evaluationID
		allTestCases[i].Status = "pending"
		allTestCases[i].CreatedAt = time.Now()
	}

	return allTestCases, nil
}

// GenerateCategory generates test cases for a specific category
func (tg *DefaultTestGenerator) GenerateCategory(ctx context.Context, analysis *PromptAnalysis, category string, count int) ([]TestCase, error) {
	if count <= 0 {
		return []TestCase{}, nil
	}

	prompt := tg.buildGenerationPrompt(analysis, category, count)
	
	response, err := tg.llmClient.CompleteWithOptions(ctx, prompt, LLMOptions{
		Temperature: 0.7, // Higher temperature for diverse test cases
		MaxTokens:   3000,
		Model:       "gpt-4",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate test cases: %w", err)
	}

	testCases, err := tg.parseTestCasesResponse(response, category)
	if err != nil {
		return nil, fmt.Errorf("failed to parse test cases response: %w", err)
	}

	// Validate and enhance test cases
	validTestCases := make([]TestCase, 0, len(testCases))
	for _, testCase := range testCases {
		if err := tg.ValidateTestCase(&testCase, analysis); err != nil {
			// Log warning but continue with other test cases
			fmt.Printf("Warning: Invalid test case %s: %v\n", testCase.Name, err)
			continue
		}
		
		// Set additional fields
		testCase.ID = 0
		testCase.Category = category
		testCase.Weight = tg.calculateWeight(category, analysis.TaskType)
		
		validTestCases = append(validTestCases, testCase)
	}

	// If we didn't get enough valid test cases, generate fallback cases
	if len(validTestCases) < count/2 {
		fallbackCases := tg.generateFallbackCases(analysis, category, count-len(validTestCases))
		validTestCases = append(validTestCases, fallbackCases...)
	}

	return validTestCases, nil
}

// buildGenerationPrompt creates the prompt for generating test cases
func (tg *DefaultTestGenerator) buildGenerationPrompt(analysis *PromptAnalysis, category string, count int) string {
	basePrompt := fmt.Sprintf(`Generate %d test cases for evaluating a %s prompt. 

Prompt Analysis:
- Task Type: %s
- Input Schema: %s
- Output Schema: %s
- Constraints: %s

Category: %s

`, count, analysis.TaskType, analysis.TaskType, 
		tg.schemaToString(analysis.InputSchema), 
		tg.schemaToString(analysis.OutputSchema),
		tg.constraintsToString(analysis.Constraints),
		category)

	// Add category-specific instructions
	switch category {
	case "normal":
		basePrompt += `Generate typical, straightforward test cases that represent common usage patterns. These should be clear examples that the prompt should handle well.

Guidelines:
- Use realistic, varied inputs
- Ensure expected outputs are correct and consistent
- Cover different aspects of the task
- Make cases representative of normal usage`

	case "edge_case":
		basePrompt += `Generate edge cases and boundary conditions that test the limits of the prompt. These should be valid inputs that might be challenging.

Guidelines:
- Empty or minimal inputs
- Maximum length inputs
- Ambiguous or borderline cases
- Unusual but valid formatting
- Corner cases specific to the task type`

	case "adversarial":
		basePrompt += `Generate adversarial test cases designed to potentially trick or confuse the prompt. These should be tricky but fair tests.

Guidelines:
- Misleading or deceptive inputs
- Cases that might trigger common errors
- Inputs with conflicting signals
- Subtle variations that change meaning
- Cases that test robustness`
	}

	// Add task-specific guidelines
	basePrompt += tg.getTaskSpecificGuidelines(analysis.TaskType, category)

	basePrompt += `

Return as a JSON array with this exact structure:
[
  {
    "name": "descriptive name",
    "description": "what this test case validates",
    "input": {"field": "value"},
    "expected_output": {"field": "value"},
    "category": "` + category + `",
    "weight": 1.0
  }
]

Ensure all test cases are valid JSON and follow the input/output schemas specified above.`

	return basePrompt
}

// getTaskSpecificGuidelines returns guidelines specific to each task type
func (tg *DefaultTestGenerator) getTaskSpecificGuidelines(taskType TaskType, category string) string {
	switch taskType {
	case TaskClassification:
		return `
Task-specific guidelines for classification:
- Ensure expected outputs use valid class labels
- Include cases for each class if possible
- For edge cases: borderline examples, ambiguous cases
- For adversarial: misleading context, conflicting signals`

	case TaskGeneration:
		return `
Task-specific guidelines for generation:
- Expected outputs should be realistic examples
- Vary input complexity and length
- For edge cases: very short/long prompts, unusual requests
- For adversarial: contradictory instructions, impossible requests`

	case TaskExtraction:
		return `
Task-specific guidelines for extraction:
- Include text with clear extractable information
- Expected outputs should list found entities/data
- For edge cases: no extractable data, minimal text
- For adversarial: misleading information, false entities`

	case TaskSummarization:
		return `
Task-specific guidelines for summarization:
- Provide text of varying lengths and complexity
- Expected outputs should be concise summaries
- For edge cases: very short/long text, repetitive content
- For adversarial: contradictory information, irrelevant details`

	case TaskQA:
		return `
Task-specific guidelines for Q&A:
- Include clear questions with definitive answers
- Provide necessary context in input
- For edge cases: ambiguous questions, insufficient context
- For adversarial: trick questions, misleading context`

	default:
		return `
General task guidelines:
- Ensure inputs match the expected schema
- Make expected outputs realistic and correct
- Consider task complexity and requirements`
	}
}

// parseTestCasesResponse parses the LLM response into TestCase structs
func (tg *DefaultTestGenerator) parseTestCasesResponse(response, category string) ([]TestCase, error) {
	// Clean the response
	response = strings.TrimSpace(response)
	if strings.HasPrefix(response, "```json") {
		response = strings.TrimPrefix(response, "```json")
		response = strings.TrimSuffix(response, "```")
	}
	response = strings.TrimSpace(response)

	var testCasesData []struct {
		Name           string                 `json:"name"`
		Description    string                 `json:"description"`
		Input          map[string]interface{} `json:"input"`
		ExpectedOutput map[string]interface{} `json:"expected_output"`
		Category       string                 `json:"category"`
		Weight         float64                `json:"weight"`
	}

	if err := json.Unmarshal([]byte(response), &testCasesData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal test cases response: %w", err)
	}

	testCases := make([]TestCase, len(testCasesData))
	for i, data := range testCasesData {
		testCases[i] = TestCase{
			Name:           data.Name,
			Description:    data.Description,
			Input:          data.Input,
			ExpectedOutput: data.ExpectedOutput,
			Category:       category, // Use the requested category
			Weight:         data.Weight,
		}
		
		// Ensure weight is reasonable
		if testCases[i].Weight <= 0 {
			testCases[i].Weight = 1.0
		}
	}

	return testCases, nil
}

// ValidateTestCase validates a single test case
func (tg *DefaultTestGenerator) ValidateTestCase(testCase *TestCase, analysis *PromptAnalysis) error {
	if testCase == nil {
		return fmt.Errorf("test case cannot be nil")
	}

	if testCase.Name == "" {
		return fmt.Errorf("test case name cannot be empty")
	}

	if testCase.Input == nil {
		return fmt.Errorf("test case input cannot be nil")
	}

	if testCase.ExpectedOutput == nil {
		return fmt.Errorf("test case expected output cannot be nil")
	}

	if testCase.Weight < 0 {
		return fmt.Errorf("test case weight cannot be negative")
	}

	// Validate category
	validCategories := map[string]bool{
		"normal":      true,
		"edge_case":   true,
		"adversarial": true,
	}
	if !validCategories[testCase.Category] {
		return fmt.Errorf("invalid test case category: %s", testCase.Category)
	}

	// Task-specific validation
	if analysis.TaskType == TaskClassification {
		if err := tg.validateClassificationTestCase(testCase, analysis); err != nil {
			return fmt.Errorf("classification validation failed: %w", err)
		}
	}

	return nil
}

// validateClassificationTestCase validates classification-specific test cases
func (tg *DefaultTestGenerator) validateClassificationTestCase(testCase *TestCase, analysis *PromptAnalysis) error {
	// Check if expected output contains valid class
	if len(analysis.OutputSchema.Classes) > 0 {
		expectedClass := ""
		if classVal, ok := testCase.ExpectedOutput["class"]; ok {
			expectedClass = fmt.Sprintf("%v", classVal)
		} else if sentimentVal, ok := testCase.ExpectedOutput["sentiment"]; ok {
			expectedClass = fmt.Sprintf("%v", sentimentVal)
		} else if labelVal, ok := testCase.ExpectedOutput["label"]; ok {
			expectedClass = fmt.Sprintf("%v", labelVal)
		}

		if expectedClass != "" {
			validClass := false
			for _, validClassLabel := range analysis.OutputSchema.Classes {
				if expectedClass == validClassLabel {
					validClass = true
					break
				}
			}
			if !validClass {
				return fmt.Errorf("expected output class '%s' not in valid classes: %v", expectedClass, analysis.OutputSchema.Classes)
			}
		}
	}

	return nil
}

// calculateWeight calculates the weight for a test case based on category and task type
func (tg *DefaultTestGenerator) calculateWeight(category string, taskType TaskType) float64 {
	baseWeight := 1.0

	switch category {
	case "normal":
		baseWeight = 1.0
	case "edge_case":
		baseWeight = 1.3
	case "adversarial":
		baseWeight = 1.5
	}

	// Adjust based on task type complexity
	switch taskType {
	case TaskClassification:
		return baseWeight
	case TaskGeneration:
		return baseWeight * 1.1
	case TaskExtraction:
		return baseWeight * 1.2
	case TaskSummarization:
		return baseWeight * 1.2
	case TaskQA:
		return baseWeight * 1.1
	default:
		return baseWeight
	}
}

// generateFallbackCases generates simple fallback test cases when LLM generation fails
func (tg *DefaultTestGenerator) generateFallbackCases(analysis *PromptAnalysis, category string, count int) []TestCase {
	testCases := make([]TestCase, 0, count)

	for i := 0; i < count && i < 3; i++ { // Limit fallback cases
		testCase := TestCase{
			ID:          0,
			Name:        fmt.Sprintf("Fallback %s case %d", category, i+1),
			Description: fmt.Sprintf("Generated fallback test case for %s category", category),
			Category:    category,
			Weight:      tg.calculateWeight(category, analysis.TaskType),
			Status:      "pending",
			CreatedAt:   time.Now(),
		}

		// Generate simple input/output based on task type
		switch analysis.TaskType {
		case TaskClassification:
			testCase.Input = map[string]interface{}{
				"text": fmt.Sprintf("Sample text for classification %d", i+1),
			}
			if len(analysis.OutputSchema.Classes) > 0 {
				testCase.ExpectedOutput = map[string]interface{}{
					"class": analysis.OutputSchema.Classes[i%len(analysis.OutputSchema.Classes)],
				}
			} else {
				testCase.ExpectedOutput = map[string]interface{}{
					"class": "positive",
				}
			}
		default:
			testCase.Input = map[string]interface{}{
				"text": fmt.Sprintf("Sample input %d", i+1),
			}
			testCase.ExpectedOutput = map[string]interface{}{
				"result": fmt.Sprintf("Sample output %d", i+1),
			}
		}

		testCases = append(testCases, testCase)
	}

	return testCases
}

// EnhanceTestCases improves existing test cases based on evaluation results
func (tg *DefaultTestGenerator) EnhanceTestCases(ctx context.Context, testCases []TestCase, metrics *EvaluationMetrics) ([]TestCase, error) {
	if metrics == nil || len(testCases) == 0 {
		return testCases, nil
	}

	// If overall performance is good, return as-is
	if metrics.OverallScore > 0.85 {
		return testCases, nil
	}

	// Identify weak areas and generate additional test cases
	// This is a simplified implementation - could be much more sophisticated
	enhancedCases := make([]TestCase, len(testCases))
	copy(enhancedCases, testCases)

	// If pass rate is low, add more edge cases
	if metrics.PassRate < 0.7 {
		// Generate a few more edge cases
		// This would typically involve analyzing which types of cases are failing
		// and generating more of those types
	}

	return enhancedCases, nil
}

// Helper functions
func (tg *DefaultTestGenerator) schemaToString(schema interface{}) string {
	data, _ := json.Marshal(schema)
	return string(data)
}

func (tg *DefaultTestGenerator) constraintsToString(constraints []Constraint) string {
	if len(constraints) == 0 {
		return "None"
	}
	
	var constraintStrings []string
	for _, c := range constraints {
		constraintStrings = append(constraintStrings, fmt.Sprintf("%s: %s", c.Type, c.Description))
	}
	return strings.Join(constraintStrings, "; ")
}