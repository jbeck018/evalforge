package evaluation

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
)

// PostgreSQLRepository implements EvaluationRepository using PostgreSQL
type PostgreSQLRepository struct {
	db *sql.DB
}

// NewPostgreSQLRepository creates a new PostgreSQL repository
func NewPostgreSQLRepository(db *sql.DB) *PostgreSQLRepository {
	return &PostgreSQLRepository{db: db}
}

// Evaluation CRUD operations
func (r *PostgreSQLRepository) CreateEvaluation(ctx context.Context, evaluation *Evaluation) error {
	query := `
		INSERT INTO evaluations (project_id, name, description, prompt_text, status, progress)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`
	
	// Extract prompt text safely
	var promptText string
	if evaluation.PromptAnalysis != nil {
		promptText = evaluation.PromptAnalysis.PromptText
	}
	
	err := r.db.QueryRowContext(ctx, query,
		evaluation.ProjectID,
		evaluation.Name,
		evaluation.Description,
		promptText, // Store prompt text directly
		evaluation.Status,
		evaluation.Progress,
	).Scan(&evaluation.ID)
	
	if err != nil {
		return fmt.Errorf("failed to create evaluation: %w", err)
	}

	// Create prompt analysis if provided
	if evaluation.PromptAnalysis != nil {
		evaluation.PromptAnalysis.EvaluationID = evaluation.ID
		// Set default task type if not provided
		if evaluation.PromptAnalysis.TaskType == "" {
			evaluation.PromptAnalysis.TaskType = "generation" // Default to generation task
		}
		if err := r.createPromptAnalysis(ctx, evaluation.PromptAnalysis); err != nil {
			return fmt.Errorf("failed to create prompt analysis: %w", err)
		}
	}

	return nil
}

func (r *PostgreSQLRepository) GetEvaluation(ctx context.Context, id string) (*Evaluation, error) {
	evaluation := &Evaluation{}
	
	query := `
		SELECT id, project_id, name, description, status, progress, started_at, completed_at, created_at, updated_at
		FROM evaluations 
		WHERE id = $1`
	
	row := r.db.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&evaluation.ID,
		&evaluation.ProjectID,
		&evaluation.Name,
		&evaluation.Description,
		&evaluation.Status,
		&evaluation.Progress,
		&evaluation.StartedAt,
		&evaluation.CompletedAt,
		&evaluation.CreatedAt,
		&evaluation.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("evaluation not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get evaluation: %w", err)
	}

	// Load related data
	if err := r.loadEvaluationRelations(ctx, evaluation); err != nil {
		return nil, fmt.Errorf("failed to load evaluation relations: %w", err)
	}

	return evaluation, nil
}

func (r *PostgreSQLRepository) UpdateEvaluation(ctx context.Context, evaluation *Evaluation) error {
	query := `
		UPDATE evaluations 
		SET name = $2, description = $3, status = $4, progress = $5, started_at = $6, completed_at = $7
		WHERE id = $1`
	
	_, err := r.db.ExecContext(ctx, query,
		evaluation.ID,
		evaluation.Name,
		evaluation.Description,
		evaluation.Status,
		evaluation.Progress,
		evaluation.StartedAt,
		evaluation.CompletedAt,
	)
	
	return err
}

func (r *PostgreSQLRepository) DeleteEvaluation(ctx context.Context, id string) error {
	query := `DELETE FROM evaluations WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *PostgreSQLRepository) ListEvaluations(ctx context.Context, projectID int, options ListOptions) ([]Evaluation, error) {
	query := `
		SELECT id, project_id, name, description, status, progress, started_at, completed_at, created_at, updated_at
		FROM evaluations 
		WHERE project_id = $1`
	
	args := []interface{}{projectID}
	argIndex := 2

	// Add status filter if specified
	if options.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, options.Status)
		argIndex++
	}

	// Add ordering
	if options.OrderBy != "" {
		direction := "ASC"
		if options.SortDesc {
			direction = "DESC"
		}
		query += fmt.Sprintf(" ORDER BY %s %s", options.OrderBy, direction)
	} else {
		query += " ORDER BY created_at DESC"
	}

	// Add pagination
	if options.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, options.Limit)
		argIndex++
		
		if options.Offset > 0 {
			query += fmt.Sprintf(" OFFSET $%d", argIndex)
			args = append(args, options.Offset)
		}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list evaluations: %w", err)
	}
	defer rows.Close()

	var evaluations []Evaluation
	for rows.Next() {
		evaluation := Evaluation{}
		err := rows.Scan(
			&evaluation.ID,
			&evaluation.ProjectID,
			&evaluation.Name,
			&evaluation.Description,
			&evaluation.Status,
			&evaluation.Progress,
			&evaluation.StartedAt,
			&evaluation.CompletedAt,
			&evaluation.CreatedAt,
			&evaluation.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan evaluation: %w", err)
		}
		evaluations = append(evaluations, evaluation)
	}

	return evaluations, nil
}

// Test cases operations
func (r *PostgreSQLRepository) SaveTestCases(ctx context.Context, testCases []TestCase) error {
	if len(testCases) == 0 {
		return nil
	}

	query := `
		INSERT INTO test_cases (id, evaluation_id, name, description, input, expected_output, actual_output, category, weight, status, score, executed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, testCase := range testCases {
		inputJSON, _ := json.Marshal(testCase.Input)
		expectedOutputJSON, _ := json.Marshal(testCase.ExpectedOutput)
		actualOutputJSON, _ := json.Marshal(testCase.ActualOutput)

		_, err = tx.ExecContext(ctx, query,
			testCase.ID,
			testCase.EvaluationID,
			testCase.Name,
			testCase.Description,
			string(inputJSON),
			string(expectedOutputJSON),
			string(actualOutputJSON),
			testCase.Category,
			testCase.Weight,
			testCase.Status,
			testCase.Score,
			testCase.ExecutedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert test case %s: %w", testCase.ID, err)
		}
	}

	return tx.Commit()
}

func (r *PostgreSQLRepository) GetTestCases(ctx context.Context, evaluationID string) ([]TestCase, error) {
	query := `
		SELECT id, evaluation_id, name, description, input, expected_output, actual_output, category, weight, status, score, executed_at, created_at
		FROM test_cases 
		WHERE evaluation_id = $1
		ORDER BY created_at ASC`

	// Convert string ID to int
	evalID, err := strconv.Atoi(evaluationID)
	if err != nil {
		return nil, fmt.Errorf("invalid evaluation ID: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, query, evalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get test cases: %w", err)
	}
	defer rows.Close()

	var testCases []TestCase
	for rows.Next() {
		testCase := TestCase{}
		var inputJSON, expectedOutputJSON, actualOutputJSON string

		err := rows.Scan(
			&testCase.ID,
			&testCase.EvaluationID,
			&testCase.Name,
			&testCase.Description,
			&inputJSON,
			&expectedOutputJSON,
			&actualOutputJSON,
			&testCase.Category,
			&testCase.Weight,
			&testCase.Status,
			&testCase.Score,
			&testCase.ExecutedAt,
			&testCase.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan test case: %w", err)
		}

		// Parse JSON fields
		if err := json.Unmarshal([]byte(inputJSON), &testCase.Input); err != nil {
			testCase.Input = make(map[string]interface{})
		}
		if err := json.Unmarshal([]byte(expectedOutputJSON), &testCase.ExpectedOutput); err != nil {
			testCase.ExpectedOutput = make(map[string]interface{})
		}
		if err := json.Unmarshal([]byte(actualOutputJSON), &testCase.ActualOutput); err != nil {
			testCase.ActualOutput = make(map[string]interface{})
		}

		testCases = append(testCases, testCase)
	}

	return testCases, nil
}

func (r *PostgreSQLRepository) UpdateTestCase(ctx context.Context, testCase *TestCase) error {
	actualOutputJSON, _ := json.Marshal(testCase.ActualOutput)
	
	query := `
		UPDATE test_cases 
		SET actual_output = $2, status = $3, score = $4, executed_at = $5
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query,
		testCase.ID,
		string(actualOutputJSON),
		testCase.Status,
		testCase.Score,
		testCase.ExecutedAt,
	)

	return err
}

// Metrics operations
func (r *PostgreSQLRepository) SaveMetrics(ctx context.Context, metrics *EvaluationMetrics) error {
	classificationMetricsJSON, _ := json.Marshal(metrics.ClassificationMetrics)
	generationMetricsJSON, _ := json.Marshal(metrics.GenerationMetrics)
	customMetricsJSON, _ := json.Marshal(metrics.CustomMetrics)

	query := `
		INSERT INTO evaluation_metrics (id, evaluation_id, overall_score, pass_rate, test_cases_passed, test_cases_total, classification_metrics, generation_metrics, custom_metrics)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (evaluation_id) DO UPDATE SET
			overall_score = EXCLUDED.overall_score,
			pass_rate = EXCLUDED.pass_rate,
			test_cases_passed = EXCLUDED.test_cases_passed,
			test_cases_total = EXCLUDED.test_cases_total,
			classification_metrics = EXCLUDED.classification_metrics,
			generation_metrics = EXCLUDED.generation_metrics,
			custom_metrics = EXCLUDED.custom_metrics,
			calculated_at = NOW()`

	_, err := r.db.ExecContext(ctx, query,
		metrics.ID,
		metrics.EvaluationID,
		metrics.OverallScore,
		metrics.PassRate,
		metrics.TestCasesPassed,
		metrics.TestCasesTotal,
		string(classificationMetricsJSON),
		string(generationMetricsJSON),
		string(customMetricsJSON),
	)

	return err
}

func (r *PostgreSQLRepository) GetMetrics(ctx context.Context, evaluationID string) (*EvaluationMetrics, error) {
	metrics := &EvaluationMetrics{}
	var classificationMetricsJSON, generationMetricsJSON, customMetricsJSON sql.NullString

	// Convert string ID to int
	evalID, err := strconv.Atoi(evaluationID)
	if err != nil {
		return nil, fmt.Errorf("invalid evaluation ID: %w", err)
	}

	query := `
		SELECT id, evaluation_id, overall_score, pass_rate, test_cases_passed, test_cases_total, 
		       COALESCE(classification_metrics, '{}'), COALESCE(generation_metrics, '{}'), COALESCE(custom_metrics, '{}'), calculated_at
		FROM evaluation_metrics 
		WHERE evaluation_id = $1`

	row := r.db.QueryRowContext(ctx, query, evalID)
	err = row.Scan(
		&metrics.ID,
		&metrics.EvaluationID,
		&metrics.OverallScore,
		&metrics.PassRate,
		&metrics.TestCasesPassed,
		&metrics.TestCasesTotal,
		&classificationMetricsJSON,
		&generationMetricsJSON,
		&customMetricsJSON,
		&metrics.CalculatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("metrics not found for evaluation: %s", evaluationID)
		}
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}

	// Parse JSON fields
	if classificationMetricsJSON.Valid && classificationMetricsJSON.String != "" && classificationMetricsJSON.String != "null" {
		json.Unmarshal([]byte(classificationMetricsJSON.String), &metrics.ClassificationMetrics)
	}
	if generationMetricsJSON.Valid && generationMetricsJSON.String != "" && generationMetricsJSON.String != "null" {
		json.Unmarshal([]byte(generationMetricsJSON.String), &metrics.GenerationMetrics)
	}
	if customMetricsJSON.Valid && customMetricsJSON.String != "" && customMetricsJSON.String != "null" {
		json.Unmarshal([]byte(customMetricsJSON.String), &metrics.CustomMetrics)
	}

	return metrics, nil
}

// Suggestions operations
func (r *PostgreSQLRepository) SaveSuggestions(ctx context.Context, suggestions []OptimizationSuggestion) error {
	if len(suggestions) == 0 {
		return nil
	}

	query := `
		INSERT INTO optimization_suggestions (id, evaluation_id, type, title, description, old_prompt, new_prompt, expected_impact, confidence, priority, status, reasoning, examples, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, suggestion := range suggestions {
		examplesJSON, _ := json.Marshal(suggestion.Examples)
		metadataJSON, _ := json.Marshal(suggestion.Metadata)

		_, err = tx.ExecContext(ctx, query,
			suggestion.ID,
			suggestion.EvaluationID,
			suggestion.Type,
			suggestion.Title,
			suggestion.Description,
			suggestion.OldPrompt,
			suggestion.NewPrompt,
			suggestion.ExpectedImpact,
			suggestion.Confidence,
			suggestion.Priority,
			suggestion.Status,
			suggestion.Reasoning,
			string(examplesJSON),
			string(metadataJSON),
		)
		if err != nil {
			return fmt.Errorf("failed to insert suggestion %s: %w", suggestion.ID, err)
		}
	}

	return tx.Commit()
}

func (r *PostgreSQLRepository) GetSuggestions(ctx context.Context, evaluationID string) ([]OptimizationSuggestion, error) {
	query := `
		SELECT id, evaluation_id, type, title, description, old_prompt, new_prompt, expected_impact, confidence, priority, status, 
		       COALESCE(reasoning, ''), COALESCE(examples::text, '[]'), COALESCE(metadata::text, '{}'), applied_at, created_at, updated_at
		FROM optimization_suggestions 
		WHERE evaluation_id = $1
		ORDER BY priority DESC, expected_impact DESC`

	// Convert string ID to int
	evalID, err := strconv.Atoi(evaluationID)
	if err != nil {
		return nil, fmt.Errorf("invalid evaluation ID: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, query, evalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get suggestions: %w", err)
	}
	defer rows.Close()

	var suggestions []OptimizationSuggestion
	for rows.Next() {
		suggestion := OptimizationSuggestion{}
		var examplesJSON, metadataJSON string

		var appliedAt sql.NullTime
		err := rows.Scan(
			&suggestion.ID,
			&suggestion.EvaluationID,
			&suggestion.Type,
			&suggestion.Title,
			&suggestion.Description,
			&suggestion.OldPrompt,
			&suggestion.NewPrompt,
			&suggestion.ExpectedImpact,
			&suggestion.Confidence,
			&suggestion.Priority,
			&suggestion.Status,
			&suggestion.Reasoning,
			&examplesJSON,
			&metadataJSON,
			&appliedAt,
			&suggestion.CreatedAt,
			&suggestion.UpdatedAt,
		)
		if appliedAt.Valid {
			suggestion.AppliedAt = &appliedAt.Time
		}
		if err != nil {
			return nil, fmt.Errorf("failed to scan suggestion: %w", err)
		}

		// Parse JSON fields
		if err := json.Unmarshal([]byte(examplesJSON), &suggestion.Examples); err != nil {
			suggestion.Examples = []Example{}
		}
		if err := json.Unmarshal([]byte(metadataJSON), &suggestion.Metadata); err != nil {
			suggestion.Metadata = make(map[string]interface{})
		}

		suggestions = append(suggestions, suggestion)
	}

	return suggestions, nil
}

func (r *PostgreSQLRepository) UpdateSuggestion(ctx context.Context, suggestion *OptimizationSuggestion) error {
	query := `
		UPDATE optimization_suggestions 
		SET status = $2, applied_at = CASE WHEN $2 = 'applied' THEN NOW() ELSE applied_at END
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, suggestion.ID, suggestion.Status)
	return err
}

// A/B test operations
func (r *PostgreSQLRepository) CreateABTest(ctx context.Context, test *ABTest) error {
	query := `
		INSERT INTO ab_tests (id, project_id, name, description, control_prompt, variant_prompt, traffic_ratio, status, min_sample_size, confidence_level)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.db.ExecContext(ctx, query,
		test.ID,
		test.ProjectID,
		test.Name,
		test.Description,
		test.ControlPrompt,
		test.VariantPrompt,
		test.TrafficRatio,
		test.Status,
		test.MinSampleSize,
		test.ConfidenceLevel,
	)

	return err
}

func (r *PostgreSQLRepository) GetABTest(ctx context.Context, id string) (*ABTest, error) {
	test := &ABTest{}

	query := `
		SELECT id, project_id, name, description, control_prompt, variant_prompt, traffic_ratio, status, 
		       min_sample_size, control_samples, variant_samples, stat_significant, confidence_level, 
		       winner, started_at, ended_at, created_at, updated_at
		FROM ab_tests 
		WHERE id = $1`

	row := r.db.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&test.ID,
		&test.ProjectID,
		&test.Name,
		&test.Description,
		&test.ControlPrompt,
		&test.VariantPrompt,
		&test.TrafficRatio,
		&test.Status,
		&test.MinSampleSize,
		&test.ControlSamples,
		&test.VariantSamples,
		&test.StatSignificant,
		&test.ConfidenceLevel,
		&test.Winner,
		&test.StartedAt,
		&test.EndedAt,
		&test.CreatedAt,
		&test.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("A/B test not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get A/B test: %w", err)
	}

	return test, nil
}

func (r *PostgreSQLRepository) UpdateABTest(ctx context.Context, test *ABTest) error {
	query := `
		UPDATE ab_tests 
		SET name = $2, description = $3, status = $4, control_samples = $5, variant_samples = $6, 
		    stat_significant = $7, winner = $8, started_at = $9, ended_at = $10
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query,
		test.ID,
		test.Name,
		test.Description,
		test.Status,
		test.ControlSamples,
		test.VariantSamples,
		test.StatSignificant,
		test.Winner,
		test.StartedAt,
		test.EndedAt,
	)

	return err
}

func (r *PostgreSQLRepository) ListABTests(ctx context.Context, projectID int, options ListOptions) ([]ABTest, error) {
	query := `
		SELECT id, project_id, name, description, status, traffic_ratio, min_sample_size, 
		       control_samples, variant_samples, stat_significant, winner, created_at
		FROM ab_tests 
		WHERE project_id = $1`

	args := []interface{}{projectID}
	argIndex := 2

	if options.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, options.Status)
		argIndex++
	}

	query += " ORDER BY created_at DESC"

	if options.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, options.Limit)
		argIndex++

		if options.Offset > 0 {
			query += fmt.Sprintf(" OFFSET $%d", argIndex)
			args = append(args, options.Offset)
		}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list A/B tests: %w", err)
	}
	defer rows.Close()

	var tests []ABTest
	for rows.Next() {
		test := ABTest{}
		err := rows.Scan(
			&test.ID,
			&test.ProjectID,
			&test.Name,
			&test.Description,
			&test.Status,
			&test.TrafficRatio,
			&test.MinSampleSize,
			&test.ControlSamples,
			&test.VariantSamples,
			&test.StatSignificant,
			&test.Winner,
			&test.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan A/B test: %w", err)
		}
		tests = append(tests, test)
	}

	return tests, nil
}

// Helper methods
func (r *PostgreSQLRepository) createPromptAnalysis(ctx context.Context, analysis *PromptAnalysis) error {
	inputSchemaJSON, _ := json.Marshal(analysis.InputSchema)
	outputSchemaJSON, _ := json.Marshal(analysis.OutputSchema)
	constraintsJSON, _ := json.Marshal(analysis.Constraints)
	examplesJSON, _ := json.Marshal(analysis.Examples)

	query := `
		INSERT INTO prompt_analyses (evaluation_id, prompt_id, prompt_text, task_type, input_schema, output_schema, constraints, examples, confidence)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		analysis.EvaluationID,
		analysis.PromptID,
		analysis.PromptText,
		analysis.TaskType,
		string(inputSchemaJSON),
		string(outputSchemaJSON),
		string(constraintsJSON),
		string(examplesJSON),
		analysis.Confidence,
	).Scan(&analysis.ID)

	return err
}

func (r *PostgreSQLRepository) loadEvaluationRelations(ctx context.Context, evaluation *Evaluation) error {
	// Load prompt analysis
	analysisQuery := `
		SELECT id, prompt_id, prompt_text, task_type, input_schema, output_schema, constraints, examples, confidence, created_at, updated_at
		FROM prompt_analyses 
		WHERE evaluation_id = $1`

	row := r.db.QueryRowContext(ctx, analysisQuery, evaluation.ID)
	
	analysis := &PromptAnalysis{}
	var inputSchemaJSON, outputSchemaJSON, constraintsJSON, examplesJSON string

	err := row.Scan(
		&analysis.ID,
		&analysis.PromptID,
		&analysis.PromptText,
		&analysis.TaskType,
		&inputSchemaJSON,
		&outputSchemaJSON,
		&constraintsJSON,
		&examplesJSON,
		&analysis.Confidence,
		&analysis.CreatedAt,
		&analysis.UpdatedAt,
	)

	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to load prompt analysis: %w", err)
	}

	if err == nil {
		// Parse JSON fields
		json.Unmarshal([]byte(inputSchemaJSON), &analysis.InputSchema)
		json.Unmarshal([]byte(outputSchemaJSON), &analysis.OutputSchema)
		json.Unmarshal([]byte(constraintsJSON), &analysis.Constraints)
		json.Unmarshal([]byte(examplesJSON), &analysis.Examples)
		
		analysis.EvaluationID = evaluation.ID
		evaluation.PromptAnalysis = analysis
	}

	// Load test cases
	testCases, err := r.GetTestCases(ctx, strconv.Itoa(evaluation.ID))
	if err == nil {
		evaluation.TestCases = testCases
	}

	// Load metrics
	metrics, err := r.GetMetrics(ctx, strconv.Itoa(evaluation.ID))
	if err == nil {
		evaluation.Metrics = metrics
	} else {
		// Log but don't fail - metrics are optional
		log.Printf("Failed to load metrics for evaluation %d: %v", evaluation.ID, err)
	}

	// Load suggestions
	suggestions, err := r.GetSuggestions(ctx, strconv.Itoa(evaluation.ID))
	if err == nil {
		evaluation.Suggestions = suggestions
	} else {
		// Log but don't fail - suggestions are optional
		log.Printf("Failed to load suggestions for evaluation %d: %v", evaluation.ID, err)
	}

	return nil
}