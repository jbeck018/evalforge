package abtest

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

// ABTestManager manages A/B tests for prompts
type ABTestManager struct {
	db *sql.DB
}

// NewABTestManager creates a new A/B test manager
func NewABTestManager(db *sql.DB) *ABTestManager {
	return &ABTestManager{db: db}
}

// ABTest represents an A/B test configuration
type ABTest struct {
	ID              int       `json:"id"`
	ProjectID       int       `json:"project_id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	ControlPrompt   string    `json:"control_prompt"`
	VariantPrompt   string    `json:"variant_prompt"`
	TrafficRatio    float64   `json:"traffic_ratio"` // % of traffic to variant
	Status          string    `json:"status"`
	MinSampleSize   int       `json:"min_sample_size"`
	ControlSamples  int       `json:"control_samples"`
	VariantSamples  int       `json:"variant_samples"`
	ControlMetrics  Metrics   `json:"control_metrics"`
	VariantMetrics  Metrics   `json:"variant_metrics"`
	StatSignificant bool      `json:"stat_significant"`
	ConfidenceLevel float64   `json:"confidence_level"`
	Winner          string    `json:"winner"`
	StartedAt       *time.Time `json:"started_at"`
	EndedAt         *time.Time `json:"ended_at"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Metrics for A/B test results
type Metrics struct {
	AverageLatency   float64 `json:"average_latency"`
	AverageCost      float64 `json:"average_cost"`
	ErrorRate        float64 `json:"error_rate"`
	AverageTokens    float64 `json:"average_tokens"`
	SatisfactionRate float64 `json:"satisfaction_rate"`
	CustomMetrics    map[string]float64 `json:"custom_metrics"`
}

// CreateABTest creates a new A/B test
func (m *ABTestManager) CreateABTest(ctx context.Context, test *ABTest) error {
	query := `
		INSERT INTO ab_tests (
			project_id, name, description, control_prompt, variant_prompt,
			traffic_ratio, status, min_sample_size, confidence_level, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		RETURNING id, created_at`

	err := m.db.QueryRowContext(ctx, query,
		test.ProjectID,
		test.Name,
		test.Description,
		test.ControlPrompt,
		test.VariantPrompt,
		test.TrafficRatio,
		"draft",
		test.MinSampleSize,
		test.ConfidenceLevel,
	).Scan(&test.ID, &test.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create A/B test: %w", err)
	}

	test.Status = "draft"
	return nil
}

// StartABTest starts an A/B test
func (m *ABTestManager) StartABTest(ctx context.Context, testID int) error {
	query := `
		UPDATE ab_tests 
		SET status = 'running', started_at = NOW() 
		WHERE id = $1 AND status = 'draft'`

	result, err := m.db.ExecContext(ctx, query, testID)
	if err != nil {
		return fmt.Errorf("failed to start A/B test: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("A/B test not found or already started")
	}

	return nil
}

// SelectPromptVariant determines which prompt variant to use
func (m *ABTestManager) SelectPromptVariant(ctx context.Context, projectID int) (string, int, string, error) {
	// Get active A/B test for project
	query := `
		SELECT id, control_prompt, variant_prompt, traffic_ratio
		FROM ab_tests
		WHERE project_id = $1 AND status = 'running'
		ORDER BY created_at DESC
		LIMIT 1`

	var test struct {
		ID            int
		ControlPrompt string
		VariantPrompt string
		TrafficRatio  float64
	}

	err := m.db.QueryRowContext(ctx, query, projectID).Scan(
		&test.ID,
		&test.ControlPrompt,
		&test.VariantPrompt,
		&test.TrafficRatio,
	)

	if err == sql.ErrNoRows {
		// No active A/B test
		return "", 0, "", nil
	}

	if err != nil {
		return "", 0, "", fmt.Errorf("failed to get A/B test: %w", err)
	}

	// Randomly select variant based on traffic ratio
	variant := "control"
	prompt := test.ControlPrompt
	
	if rand.Float64() < test.TrafficRatio {
		variant = "variant"
		prompt = test.VariantPrompt
	}

	return prompt, test.ID, variant, nil
}

// RecordResult records the result of an A/B test execution
func (m *ABTestManager) RecordResult(ctx context.Context, result *ABTestResult) error {
	resultJSON, _ := json.Marshal(result.Metrics)

	query := `
		INSERT INTO ab_test_results (
			ab_test_id, variant, trace_id, latency_ms, cost, tokens,
			error, metrics, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())`

	_, err := m.db.ExecContext(ctx, query,
		result.ABTestID,
		result.Variant,
		result.TraceID,
		result.LatencyMS,
		result.Cost,
		result.Tokens,
		result.Error,
		string(resultJSON),
	)

	if err != nil {
		return fmt.Errorf("failed to record A/B test result: %w", err)
	}

	// Update sample counts
	updateQuery := `
		UPDATE ab_tests 
		SET control_samples = control_samples + CASE WHEN $2 = 'control' THEN 1 ELSE 0 END,
		    variant_samples = variant_samples + CASE WHEN $2 = 'variant' THEN 1 ELSE 0 END
		WHERE id = $1`

	_, err = m.db.ExecContext(ctx, updateQuery, result.ABTestID, result.Variant)
	return err
}

// ABTestResult represents a single test result
type ABTestResult struct {
	ABTestID  int                    `json:"ab_test_id"`
	Variant   string                 `json:"variant"`
	TraceID   string                 `json:"trace_id"`
	LatencyMS int                    `json:"latency_ms"`
	Cost      float64                `json:"cost"`
	Tokens    int                    `json:"tokens"`
	Error     bool                   `json:"error"`
	Metrics   map[string]interface{} `json:"metrics"`
}

// AnalyzeResults analyzes A/B test results and determines statistical significance
func (m *ABTestManager) AnalyzeResults(ctx context.Context, testID int) (*ABTestAnalysis, error) {
	// Get test configuration
	var test ABTest
	query := `
		SELECT id, project_id, name, min_sample_size, confidence_level,
		       control_samples, variant_samples
		FROM ab_tests
		WHERE id = $1`

	err := m.db.QueryRowContext(ctx, query, testID).Scan(
		&test.ID,
		&test.ProjectID,
		&test.Name,
		&test.MinSampleSize,
		&test.ConfidenceLevel,
		&test.ControlSamples,
		&test.VariantSamples,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get A/B test: %w", err)
	}

	// Get metrics for each variant
	metricsQuery := `
		SELECT 
			variant,
			COUNT(*) as samples,
			AVG(latency_ms) as avg_latency,
			AVG(cost) as avg_cost,
			AVG(tokens) as avg_tokens,
			SUM(CASE WHEN error THEN 1 ELSE 0 END)::FLOAT / COUNT(*) as error_rate
		FROM ab_test_results
		WHERE ab_test_id = $1
		GROUP BY variant`

	rows, err := m.db.QueryContext(ctx, metricsQuery, testID)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}
	defer rows.Close()

	analysis := &ABTestAnalysis{
		TestID:          testID,
		TestName:        test.Name,
		ControlSamples:  test.ControlSamples,
		VariantSamples:  test.VariantSamples,
		MinSampleSize:   test.MinSampleSize,
		ConfidenceLevel: test.ConfidenceLevel,
	}

	for rows.Next() {
		var variant string
		var samples int
		var avgLatency, avgCost, avgTokens, errorRate float64

		err := rows.Scan(&variant, &samples, &avgLatency, &avgCost, &avgTokens, &errorRate)
		if err != nil {
			continue
		}

		metrics := Metrics{
			AverageLatency: avgLatency,
			AverageCost:    avgCost,
			AverageTokens:  avgTokens,
			ErrorRate:      errorRate,
		}

		if variant == "control" {
			analysis.ControlMetrics = metrics
		} else {
			analysis.VariantMetrics = metrics
		}
	}

	// Calculate statistical significance (simplified)
	analysis.calculateSignificance()

	return analysis, nil
}

// ABTestAnalysis contains analysis results
type ABTestAnalysis struct {
	TestID           int     `json:"test_id"`
	TestName         string  `json:"test_name"`
	ControlSamples   int     `json:"control_samples"`
	VariantSamples   int     `json:"variant_samples"`
	MinSampleSize    int     `json:"min_sample_size"`
	ControlMetrics   Metrics `json:"control_metrics"`
	VariantMetrics   Metrics `json:"variant_metrics"`
	ConfidenceLevel  float64 `json:"confidence_level"`
	StatSignificant  bool    `json:"stat_significant"`
	Winner           string  `json:"winner"`
	ImprovementRate  float64 `json:"improvement_rate"`
	Recommendation   string  `json:"recommendation"`
}

// calculateSignificance calculates statistical significance
func (a *ABTestAnalysis) calculateSignificance() {
	// Check if we have enough samples
	if a.ControlSamples < a.MinSampleSize || a.VariantSamples < a.MinSampleSize {
		a.StatSignificant = false
		a.Recommendation = "Need more samples for statistical significance"
		return
	}

	// Simple comparison based on key metrics
	controlScore := a.scoreMetrics(a.ControlMetrics)
	variantScore := a.scoreMetrics(a.VariantMetrics)

	improvementRate := (variantScore - controlScore) / controlScore * 100

	// Determine winner (simplified - in production, use proper statistical tests)
	if improvementRate > 5 && a.ConfidenceLevel > 0.95 {
		a.StatSignificant = true
		a.Winner = "variant"
		a.ImprovementRate = improvementRate
		a.Recommendation = fmt.Sprintf("Variant shows %.1f%% improvement. Consider adopting.", improvementRate)
	} else if improvementRate < -5 && a.ConfidenceLevel > 0.95 {
		a.StatSignificant = true
		a.Winner = "control"
		a.ImprovementRate = improvementRate
		a.Recommendation = "Control performs better. Keep current prompt."
	} else {
		a.StatSignificant = false
		a.Winner = "none"
		a.ImprovementRate = improvementRate
		a.Recommendation = "No significant difference detected. Continue testing."
	}
}

// scoreMetrics calculates a composite score for metrics
func (a *ABTestAnalysis) scoreMetrics(m Metrics) float64 {
	// Lower is better for latency, cost, error rate
	// Normalize and weight metrics
	latencyScore := 1000 / (m.AverageLatency + 1) // Inverse, normalized
	costScore := 1 / (m.AverageCost + 0.001)      // Inverse, normalized
	errorScore := 1 - m.ErrorRate                  // Lower error is better

	// Weighted average
	return latencyScore*0.3 + costScore*0.3 + errorScore*0.4
}

// StopABTest stops an A/B test
func (m *ABTestManager) StopABTest(ctx context.Context, testID int) error {
	analysis, err := m.AnalyzeResults(ctx, testID)
	if err != nil {
		return fmt.Errorf("failed to analyze results: %w", err)
	}

	query := `
		UPDATE ab_tests 
		SET status = 'completed', 
		    ended_at = NOW(),
		    stat_significant = $2,
		    winner = $3
		WHERE id = $1`

	_, err = m.db.ExecContext(ctx, query, testID, analysis.StatSignificant, analysis.Winner)
	if err != nil {
		return fmt.Errorf("failed to stop A/B test: %w", err)
	}

	return nil
}

// GetActiveTests gets all active A/B tests
func (m *ABTestManager) GetActiveTests(ctx context.Context, projectID int) ([]*ABTest, error) {
	query := `
		SELECT id, project_id, name, description, status, 
		       control_samples, variant_samples, started_at
		FROM ab_tests
		WHERE project_id = $1 AND status IN ('pending', 'running')
		ORDER BY created_at DESC`

	rows, err := m.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active tests: %w", err)
	}
	defer rows.Close()

	var tests []*ABTest
	for rows.Next() {
		test := &ABTest{}
		err := rows.Scan(
			&test.ID,
			&test.ProjectID,
			&test.Name,
			&test.Description,
			&test.Status,
			&test.ControlSamples,
			&test.VariantSamples,
			&test.StartedAt,
		)
		if err != nil {
			continue
		}
		tests = append(tests, test)
	}

	return tests, nil
}