package evaluation

import (
	"time"
)

// TaskType represents the type of task the prompt is designed for
type TaskType string

const (
	TaskClassification  TaskType = "classification"
	TaskGeneration      TaskType = "generation"
	TaskExtraction      TaskType = "extraction"
	TaskSummarization   TaskType = "summarization"
	TaskQA              TaskType = "question_answering"
	TaskTransformation  TaskType = "transformation"
	TaskCompletion      TaskType = "completion"
)

// InputSchema defines the expected input format for a prompt
type InputSchema struct {
	Type        string                 `json:"type"`        // text, json, structured
	Fields      map[string]interface{} `json:"fields"`      // field definitions for structured input
	Required    []string               `json:"required"`    // required fields
	Constraints map[string]interface{} `json:"constraints"` // validation constraints
}

// OutputSchema defines the expected output format for a prompt
type OutputSchema struct {
	Type        string                 `json:"type"`        // text, json, classification, structured
	Format      string                 `json:"format"`      // specific format requirements
	Classes     []string               `json:"classes"`     // for classification tasks
	Fields      map[string]interface{} `json:"fields"`      // for structured output
	Constraints map[string]interface{} `json:"constraints"` // validation constraints
}

// Constraint represents a rule or constraint that the output should follow
type Constraint struct {
	Type        string      `json:"type"`        // format, length, value, pattern
	Description string      `json:"description"` // human-readable description
	Rule        interface{} `json:"rule"`        // the actual constraint rule
	Severity    string      `json:"severity"`    // error, warning, info
}

// Example represents a few-shot example in the prompt
type Example struct {
	Input       map[string]interface{} `json:"input"`       // example input
	Output      map[string]interface{} `json:"output"`      // expected output
	Explanation string                 `json:"explanation"` // optional explanation
}

// PromptAnalysis contains the analysis results of a prompt
type PromptAnalysis struct {
	ID           int           `json:"id" db:"id"`
	EvaluationID int           `json:"evaluation_id" db:"evaluation_id"`
	PromptID     int           `json:"prompt_id" db:"prompt_id"`
	PromptText   string        `json:"prompt_text" db:"prompt_text"`
	TaskType     TaskType      `json:"task_type" db:"task_type"`
	InputSchema  InputSchema   `json:"input_schema" db:"input_schema"`
	OutputSchema OutputSchema  `json:"output_schema" db:"output_schema"`
	Constraints  []Constraint  `json:"constraints" db:"constraints"`
	Examples     []Example     `json:"examples" db:"examples"`
	Confidence   float64       `json:"confidence" db:"confidence"`
	CreatedAt    time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at" db:"updated_at"`
}

// TestCase represents a generated test case for evaluation
type TestCase struct {
	ID             int                    `json:"id" db:"id"`
	EvaluationID   int                    `json:"evaluation_id" db:"evaluation_id"`
	Name           string                 `json:"name" db:"name"`
	Description    string                 `json:"description" db:"description"`
	Input          map[string]interface{} `json:"input" db:"input"`
	ExpectedOutput map[string]interface{} `json:"expected_output" db:"expected_output"`
	ActualOutput   map[string]interface{} `json:"actual_output" db:"actual_output"`
	Category       string                 `json:"category" db:"category"` // normal, edge_case, adversarial
	Weight         float64                `json:"weight" db:"weight"`
	Status         string                 `json:"status" db:"status"` // pending, passed, failed, error
	Score          float64                `json:"score" db:"score"`
	ExecutedAt     *time.Time             `json:"executed_at" db:"executed_at"`
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
}

// ClassificationMetrics contains metrics for classification tasks
type ClassificationMetrics struct {
	Accuracy        float64                       `json:"accuracy"`
	Precision       map[string]float64            `json:"precision"`       // per-class precision
	Recall          map[string]float64            `json:"recall"`          // per-class recall
	F1Score         map[string]float64            `json:"f1_score"`        // per-class F1
	MacroF1         float64                       `json:"macro_f1"`        // macro-averaged F1
	WeightedF1      float64                       `json:"weighted_f1"`     // weighted F1
	ConfusionMatrix map[string]map[string]int     `json:"confusion_matrix"`
	Support         map[string]int                `json:"support"`         // per-class support
}

// GenerationMetrics contains metrics for generation tasks
type GenerationMetrics struct {
	BLEU        float64 `json:"bleu"`        // BLEU score
	ROUGE1      float64 `json:"rouge_1"`     // ROUGE-1 score
	ROUGE2      float64 `json:"rouge_2"`     // ROUGE-2 score
	ROUGEL      float64 `json:"rouge_l"`     // ROUGE-L score
	BERTScore   float64 `json:"bert_score"`  // BERTScore
	Perplexity  float64 `json:"perplexity"`  // perplexity score
	Diversity   float64 `json:"diversity"`   // lexical diversity
	Coherence   float64 `json:"coherence"`   // coherence score
	Relevance   float64 `json:"relevance"`   // relevance score
}

// EvaluationMetrics contains all metrics for an evaluation
type EvaluationMetrics struct {
	ID                    int                    `json:"id" db:"id"`
	EvaluationID          int                    `json:"evaluation_id" db:"evaluation_id"`
	OverallScore          float64                `json:"overall_score" db:"overall_score"`
	PassRate              float64                `json:"pass_rate" db:"pass_rate"`
	TestCasesPassed       int                    `json:"test_cases_passed" db:"test_cases_passed"`
	TestCasesTotal        int                    `json:"test_cases_total" db:"test_cases_total"`
	ClassificationMetrics *ClassificationMetrics `json:"classification_metrics,omitempty" db:"classification_metrics"`
	GenerationMetrics     *GenerationMetrics     `json:"generation_metrics,omitempty" db:"generation_metrics"`
	CustomMetrics         map[string]float64     `json:"custom_metrics" db:"custom_metrics"`
	CalculatedAt          time.Time              `json:"calculated_at" db:"calculated_at"`
}

// ErrorAnalysis contains analysis of common errors
type ErrorAnalysis struct {
	ID               int               `json:"id" db:"id"`
	EvaluationID     int               `json:"evaluation_id" db:"evaluation_id"`
	CommonErrors     []string          `json:"common_errors" db:"common_errors"`
	ErrorPatterns    map[string]int    `json:"error_patterns" db:"error_patterns"`
	AmbiguousCases   float64           `json:"ambiguous_cases" db:"ambiguous_cases"`
	FormatErrors     float64           `json:"format_errors" db:"format_errors"`
	LogicErrors      float64           `json:"logic_errors" db:"logic_errors"`
	InconsistentCases float64          `json:"inconsistent_cases" db:"inconsistent_cases"`
	ErrorCategories  map[string]float64 `json:"error_categories" db:"error_categories"`
	CreatedAt        time.Time         `json:"created_at" db:"created_at"`
}

// OptimizationSuggestion represents a suggestion for improving a prompt
type OptimizationSuggestion struct {
	ID                int                    `json:"id" db:"id"`
	EvaluationID      int                    `json:"evaluation_id" db:"evaluation_id"`
	Type              string                 `json:"type" db:"type"` // clarity, specificity, examples, format, constraints
	Title             string                 `json:"title" db:"title"`
	Description       string                 `json:"description" db:"description"`
	OldPrompt         string                 `json:"old_prompt" db:"old_prompt"`
	NewPrompt         string                 `json:"new_prompt" db:"new_prompt"`
	ExpectedImpact    float64                `json:"expected_impact" db:"expected_impact"`
	Confidence        float64                `json:"confidence" db:"confidence"`
	Priority          string                 `json:"priority" db:"priority"` // high, medium, low
	Status            string                 `json:"status" db:"status"`     // pending, accepted, rejected, applied
	Reasoning         string                 `json:"reasoning" db:"reasoning"`
	Examples          []Example              `json:"examples" db:"examples"`
	Metadata          map[string]interface{} `json:"metadata" db:"metadata"`
	AppliedAt         *time.Time             `json:"applied_at,omitempty" db:"applied_at"`
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" db:"updated_at"`
}

// Evaluation represents a complete evaluation run
type Evaluation struct {
	ID              int                     `json:"id" db:"id"`
	ProjectID       int                     `json:"project_id" db:"project_id"`
	Name            string                  `json:"name" db:"name"`
	Description     string                  `json:"description" db:"description"`
	PromptAnalysis  *PromptAnalysis         `json:"prompt_analysis,omitempty"`
	TestCases       []TestCase              `json:"test_cases,omitempty"`
	Metrics         *EvaluationMetrics      `json:"metrics,omitempty"`
	ErrorAnalysis   *ErrorAnalysis          `json:"error_analysis,omitempty"`
	Suggestions     []OptimizationSuggestion `json:"suggestions,omitempty"`
	Status          string                  `json:"status" db:"status"` // pending, running, completed, failed
	Progress        float64                 `json:"progress" db:"progress"`
	StartedAt       *time.Time              `json:"started_at" db:"started_at"`
	CompletedAt     *time.Time              `json:"completed_at" db:"completed_at"`
	CreatedAt       time.Time               `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time               `json:"updated_at" db:"updated_at"`
}

// ABTest represents an A/B test between prompt variations
type ABTest struct {
	ID              int         `json:"id" db:"id"`
	ProjectID       int         `json:"project_id" db:"project_id"`
	Name            string      `json:"name" db:"name"`
	Description     string      `json:"description" db:"description"`
	ControlPrompt   string      `json:"control_prompt" db:"control_prompt"`
	VariantPrompt   string      `json:"variant_prompt" db:"variant_prompt"`
	TrafficRatio    float64     `json:"traffic_ratio" db:"traffic_ratio"` // 0.0 to 1.0
	Status          string      `json:"status" db:"status"`               // draft, running, completed, cancelled
	MinSampleSize   int         `json:"min_sample_size" db:"min_sample_size"`
	ControlSamples  int         `json:"control_samples" db:"control_samples"`
	VariantSamples  int         `json:"variant_samples" db:"variant_samples"`
	StatSignificant bool        `json:"stat_significant" db:"stat_significant"`
	ConfidenceLevel float64     `json:"confidence_level" db:"confidence_level"`
	Winner          string      `json:"winner" db:"winner"` // control, variant, inconclusive
	StartedAt       *time.Time  `json:"started_at" db:"started_at"`
	EndedAt         *time.Time  `json:"ended_at" db:"ended_at"`
	CreatedAt       time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at" db:"updated_at"`
}