package evaluation

import (
	"context"
)

// LLMClient interface for communicating with language models
type LLMClient interface {
	// Complete sends a prompt to the LLM and returns the response
	Complete(ctx context.Context, prompt string) (string, error)
	
	// CompleteWithOptions sends a prompt with additional options
	CompleteWithOptions(ctx context.Context, prompt string, options LLMOptions) (string, error)
	
	// ValidateConnection checks if the LLM client is properly configured
	ValidateConnection(ctx context.Context) error
}

// LLMOptions contains options for LLM completion
type LLMOptions struct {
	Temperature   float64            `json:"temperature,omitempty"`
	MaxTokens     int                `json:"max_tokens,omitempty"`
	Model         string             `json:"model,omitempty"`
	SystemPrompt  string             `json:"system_prompt,omitempty"`
	StopSequences []string           `json:"stop_sequences,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// PromptAnalyzer interface for analyzing prompts
type PromptAnalyzer interface {
	// AnalyzePrompt analyzes a prompt and returns analysis results
	AnalyzePrompt(ctx context.Context, prompt string, examples []Example) (*PromptAnalysis, error)
	
	// ValidateAnalysis validates the analysis results
	ValidateAnalysis(analysis *PromptAnalysis) error
	
	// UpdateAnalysis updates an existing analysis
	UpdateAnalysis(ctx context.Context, analysisID string, prompt string) (*PromptAnalysis, error)
}

// TestGenerator interface for generating test cases
type TestGenerator interface {
	// GenerateTestCases generates test cases based on prompt analysis
	GenerateTestCases(ctx context.Context, analysis *PromptAnalysis, options TestGeneratorOptions) ([]TestCase, error)
	
	// GenerateCategory generates test cases for a specific category
	GenerateCategory(ctx context.Context, analysis *PromptAnalysis, category string, count int) ([]TestCase, error)
	
	// ValidateTestCase validates a single test case
	ValidateTestCase(testCase *TestCase, analysis *PromptAnalysis) error
	
	// EnhanceTestCases improves existing test cases based on results
	EnhanceTestCases(ctx context.Context, testCases []TestCase, metrics *EvaluationMetrics) ([]TestCase, error)
}

// TestGeneratorOptions contains options for test case generation
type TestGeneratorOptions struct {
	NormalCases      int                    `json:"normal_cases"`
	EdgeCases        int                    `json:"edge_cases"`
	AdversarialCases int                    `json:"adversarial_cases"`
	Categories       []string               `json:"categories"`
	CustomRules      map[string]interface{} `json:"custom_rules"`
	Seed             int64                  `json:"seed"`
}

// TestExecutor interface for executing test cases
type TestExecutor interface {
	// ExecuteTestCase executes a single test case
	ExecuteTestCase(ctx context.Context, testCase *TestCase, prompt string) (*TestCase, error)
	
	// ExecuteTestCases executes multiple test cases in parallel
	ExecuteTestCases(ctx context.Context, testCases []TestCase, prompt string, options ExecutorOptions) ([]TestCase, error)
	
	// ValidateExecution validates the execution results
	ValidateExecution(testCase *TestCase, analysis *PromptAnalysis) error
}

// ExecutorOptions contains options for test execution
type ExecutorOptions struct {
	MaxConcurrency int                    `json:"max_concurrency"`
	Timeout        int                    `json:"timeout_seconds"`
	RetryCount     int                    `json:"retry_count"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// MetricsCalculator interface for calculating evaluation metrics
type MetricsCalculator interface {
	// CalculateMetrics calculates metrics based on test results
	CalculateMetrics(ctx context.Context, testCases []TestCase, analysis *PromptAnalysis) (*EvaluationMetrics, error)
	
	// CalculateClassificationMetrics calculates classification-specific metrics
	CalculateClassificationMetrics(predictions, groundTruth []string, classes []string) (*ClassificationMetrics, error)
	
	// CalculateGenerationMetrics calculates generation-specific metrics
	CalculateGenerationMetrics(predictions, references []string) (*GenerationMetrics, error)
	
	// CalculateCustomMetrics calculates custom metrics based on rules
	CalculateCustomMetrics(testCases []TestCase, rules map[string]interface{}) (map[string]float64, error)
}

// ErrorAnalyzer interface for analyzing errors in test results
type ErrorAnalyzer interface {
	// AnalyzeErrors analyzes failed test cases to identify patterns
	AnalyzeErrors(ctx context.Context, testCases []TestCase, analysis *PromptAnalysis) (*ErrorAnalysis, error)
	
	// CategorizeErrors categorizes errors by type
	CategorizeErrors(testCases []TestCase) (map[string][]TestCase, error)
	
	// FindPatterns finds common patterns in errors
	FindPatterns(errors []TestCase) (map[string]int, error)
}

// PromptOptimizer interface for suggesting prompt improvements
type PromptOptimizer interface {
	// SuggestImprovements suggests improvements based on evaluation results
	SuggestImprovements(ctx context.Context, prompt string, metrics *EvaluationMetrics, errorAnalysis *ErrorAnalysis) ([]OptimizationSuggestion, error)
	
	// OptimizeForClarity suggests improvements for prompt clarity
	OptimizeForClarity(ctx context.Context, prompt string, errorAnalysis *ErrorAnalysis) (*OptimizationSuggestion, error)
	
	// OptimizeForAccuracy suggests improvements for better accuracy
	OptimizeForAccuracy(ctx context.Context, prompt string, metrics *EvaluationMetrics) (*OptimizationSuggestion, error)
	
	// OptimizeForConsistency suggests improvements for consistency
	OptimizeForConsistency(ctx context.Context, prompt string, testCases []TestCase) (*OptimizationSuggestion, error)
	
	// ValidateSuggestion validates an optimization suggestion
	ValidateSuggestion(suggestion *OptimizationSuggestion) error
}

// EvaluationOrchestrator interface for coordinating the evaluation pipeline
type EvaluationOrchestrator interface {
	// CreateEvaluation creates a new evaluation
	CreateEvaluation(ctx context.Context, projectID int, prompt string, options EvaluationOptions) (*Evaluation, error)
	
	// RunEvaluation runs a complete evaluation pipeline
	RunEvaluation(ctx context.Context, evaluationID string) (*Evaluation, error)
	
	// GetEvaluation retrieves an evaluation by ID
	GetEvaluation(ctx context.Context, evaluationID string) (*Evaluation, error)
	
	// ListEvaluations lists evaluations for a project
	ListEvaluations(ctx context.Context, projectID int, options ListOptions) ([]Evaluation, error)
	
	// DeleteEvaluation deletes an evaluation
	DeleteEvaluation(ctx context.Context, evaluationID string) error
}

// EvaluationOptions contains options for creating evaluations
type EvaluationOptions struct {
	Name                string                 `json:"name"`
	Description         string                 `json:"description"`
	TestGeneratorOptions TestGeneratorOptions   `json:"test_generator_options"`
	ExecutorOptions     ExecutorOptions        `json:"executor_options"`
	AutoSuggest         bool                   `json:"auto_suggest"`
	Metadata            map[string]interface{} `json:"metadata"`
}

// ListOptions contains options for listing evaluations
type ListOptions struct {
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
	OrderBy  string `json:"order_by"`
	Status   string `json:"status"`
	SortDesc bool   `json:"sort_desc"`
}

// ABTestManager interface for managing A/B tests
type ABTestManager interface {
	// CreateABTest creates a new A/B test
	CreateABTest(ctx context.Context, projectID int, options ABTestOptions) (*ABTest, error)
	
	// StartABTest starts an A/B test
	StartABTest(ctx context.Context, testID string) error
	
	// StopABTest stops an A/B test
	StopABTest(ctx context.Context, testID string) error
	
	// GetABTest retrieves an A/B test by ID
	GetABTest(ctx context.Context, testID string) (*ABTest, error)
	
	// ListABTests lists A/B tests for a project
	ListABTests(ctx context.Context, projectID int, options ListOptions) ([]ABTest, error)
	
	// AnalyzeABTest analyzes A/B test results
	AnalyzeABTest(ctx context.Context, testID string) (*ABTestResults, error)
}

// ABTestOptions contains options for creating A/B tests
type ABTestOptions struct {
	Name            string  `json:"name"`
	Description     string  `json:"description"`
	ControlPrompt   string  `json:"control_prompt"`
	VariantPrompt   string  `json:"variant_prompt"`
	TrafficRatio    float64 `json:"traffic_ratio"`
	MinSampleSize   int     `json:"min_sample_size"`
	ConfidenceLevel float64 `json:"confidence_level"`
}

// ABTestResults contains A/B test analysis results
type ABTestResults struct {
	TestID           string                 `json:"test_id"`
	ControlMetrics   *EvaluationMetrics     `json:"control_metrics"`
	VariantMetrics   *EvaluationMetrics     `json:"variant_metrics"`
	StatSignificant  bool                   `json:"stat_significant"`
	ConfidenceLevel  float64                `json:"confidence_level"`
	Winner           string                 `json:"winner"`
	EffectSize       float64                `json:"effect_size"`
	PValue           float64                `json:"p_value"`
	Recommendation   string                 `json:"recommendation"`
	Details          map[string]interface{} `json:"details"`
}

// EvaluationRepository interface for data persistence
type EvaluationRepository interface {
	// Evaluation CRUD operations
	CreateEvaluation(ctx context.Context, evaluation *Evaluation) error
	GetEvaluation(ctx context.Context, id string) (*Evaluation, error)
	UpdateEvaluation(ctx context.Context, evaluation *Evaluation) error
	DeleteEvaluation(ctx context.Context, id string) error
	ListEvaluations(ctx context.Context, projectID int, options ListOptions) ([]Evaluation, error)
	
	// Test cases operations
	SaveTestCases(ctx context.Context, testCases []TestCase) error
	GetTestCases(ctx context.Context, evaluationID string) ([]TestCase, error)
	UpdateTestCase(ctx context.Context, testCase *TestCase) error
	
	// Metrics operations
	SaveMetrics(ctx context.Context, metrics *EvaluationMetrics) error
	GetMetrics(ctx context.Context, evaluationID string) (*EvaluationMetrics, error)
	
	// Suggestions operations
	SaveSuggestions(ctx context.Context, suggestions []OptimizationSuggestion) error
	GetSuggestions(ctx context.Context, evaluationID string) ([]OptimizationSuggestion, error)
	UpdateSuggestion(ctx context.Context, suggestion *OptimizationSuggestion) error
	
	// A/B test operations
	CreateABTest(ctx context.Context, test *ABTest) error
	GetABTest(ctx context.Context, id string) (*ABTest, error)
	UpdateABTest(ctx context.Context, test *ABTest) error
	ListABTests(ctx context.Context, projectID int, options ListOptions) ([]ABTest, error)
}