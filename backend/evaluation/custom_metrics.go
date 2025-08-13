package evaluation

import (
	"database/sql"
	"encoding/json"
	"math"
	"regexp"
	"strings"
	"time"
)

// MetricType represents the type of custom metric
type MetricType string

const (
	MetricTypeNumeric     MetricType = "numeric"
	MetricTypeBoolean     MetricType = "boolean"
	MetricTypeString      MetricType = "string"
	MetricTypePercentage  MetricType = "percentage"
	MetricTypeScore       MetricType = "score"
	MetricTypeCustom      MetricType = "custom"
)

// AggregationType represents how to aggregate metric values
type AggregationType string

const (
	AggregationAverage AggregationType = "average"
	AggregationSum     AggregationType = "sum"
	AggregationMin     AggregationType = "min"
	AggregationMax     AggregationType = "max"
	AggregationMedian  AggregationType = "median"
	AggregationP95     AggregationType = "p95"
	AggregationP99     AggregationType = "p99"
	AggregationCount   AggregationType = "count"
)

// CustomMetric represents a user-defined evaluation metric
type CustomMetric struct {
	ID          int             `json:"id"`
	ProjectID   int             `json:"project_id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Type        MetricType      `json:"type"`
	Aggregation AggregationType `json:"aggregation"`
	Formula     string          `json:"formula,omitempty"`
	Thresholds  MetricThresholds `json:"thresholds"`
	Weight      float64         `json:"weight"`
	Enabled     bool            `json:"enabled"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// MetricThresholds defines pass/fail thresholds for a metric
type MetricThresholds struct {
	PassValue    float64 `json:"pass_value"`
	WarningValue float64 `json:"warning_value,omitempty"`
	FailValue    float64 `json:"fail_value,omitempty"`
	Operator     string  `json:"operator"` // >, <, >=, <=, ==, !=
}

// MetricValue represents a single metric measurement
type MetricValue struct {
	MetricID     int         `json:"metric_id"`
	EvaluationID int         `json:"evaluation_id"`
	SampleID     string      `json:"sample_id"`
	Value        interface{} `json:"value"`
	Passed       bool        `json:"passed"`
	Timestamp    time.Time   `json:"timestamp"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// MetricResult represents aggregated results for a metric
type MetricResult struct {
	MetricID    int         `json:"metric_id"`
	MetricName  string      `json:"metric_name"`
	Value       float64     `json:"value"`
	Passed      bool        `json:"passed"`
	PassRate    float64     `json:"pass_rate"`
	SampleCount int         `json:"sample_count"`
	Details     interface{} `json:"details,omitempty"`
}

// CustomMetricsEvaluator evaluates custom metrics
type CustomMetricsEvaluator struct {
	db      *sql.DB
	metrics map[int]*CustomMetric
}

// NewCustomMetricsEvaluator creates a new custom metrics evaluator
func NewCustomMetricsEvaluator(db *sql.DB) *CustomMetricsEvaluator {
	return &CustomMetricsEvaluator{
		db:      db,
		metrics: make(map[int]*CustomMetric),
	}
}

// LoadMetrics loads custom metrics for a project
func (e *CustomMetricsEvaluator) LoadMetrics(projectID int) error {
	rows, err := e.db.Query(`
		SELECT id, name, description, type, aggregation, formula, 
		       thresholds, weight, enabled, created_at, updated_at
		FROM custom_metrics
		WHERE project_id = $1 AND enabled = true
	`, projectID)
	if err != nil {
		return err
	}
	defer rows.Close()

	e.metrics = make(map[int]*CustomMetric)
	
	for rows.Next() {
		var metric CustomMetric
		var thresholdsJSON string
		
		err := rows.Scan(
			&metric.ID,
			&metric.Name,
			&metric.Description,
			&metric.Type,
			&metric.Aggregation,
			&metric.Formula,
			&thresholdsJSON,
			&metric.Weight,
			&metric.Enabled,
			&metric.CreatedAt,
			&metric.UpdatedAt,
		)
		if err != nil {
			continue
		}
		
		metric.ProjectID = projectID
		
		// Parse thresholds JSON
		if err := json.Unmarshal([]byte(thresholdsJSON), &metric.Thresholds); err == nil {
			e.metrics[metric.ID] = &metric
		}
	}
	
	return nil
}

// EvaluateMetric evaluates a single metric for a sample
func (e *CustomMetricsEvaluator) EvaluateMetric(metric *CustomMetric, sample interface{}) (*MetricValue, error) {
	value := &MetricValue{
		MetricID:  metric.ID,
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}
	
	// Extract value based on metric type
	var rawValue interface{}
	var numericValue float64
	
	switch metric.Type {
	case MetricTypeNumeric, MetricTypePercentage, MetricTypeScore:
		// Extract numeric value from sample
		rawValue = e.extractNumericValue(sample, metric)
		if v, ok := rawValue.(float64); ok {
			numericValue = v
		} else if v, ok := rawValue.(int); ok {
			numericValue = float64(v)
		}
		
	case MetricTypeBoolean:
		// Extract boolean value
		rawValue = e.extractBooleanValue(sample, metric)
		if v, ok := rawValue.(bool); ok {
			if v {
				numericValue = 1.0
			} else {
				numericValue = 0.0
			}
		}
		
	case MetricTypeString:
		// Extract string value and check against patterns
		rawValue = e.extractStringValue(sample, metric)
		// For string metrics, we need custom evaluation logic
		numericValue = e.evaluateStringMetric(rawValue, metric)
		
	case MetricTypeCustom:
		// Evaluate custom formula
		numericValue = e.evaluateFormula(metric.Formula, sample)
		rawValue = numericValue
	}
	
	value.Value = rawValue
	
	// Check against thresholds
	value.Passed = e.checkThreshold(numericValue, metric.Thresholds)
	
	return value, nil
}

// checkThreshold checks if a value passes the threshold
func (e *CustomMetricsEvaluator) checkThreshold(value float64, thresholds MetricThresholds) bool {
	switch thresholds.Operator {
	case ">":
		return value > thresholds.PassValue
	case ">=":
		return value >= thresholds.PassValue
	case "<":
		return value < thresholds.PassValue
	case "<=":
		return value <= thresholds.PassValue
	case "==":
		return math.Abs(value-thresholds.PassValue) < 0.0001
	case "!=":
		return math.Abs(value-thresholds.PassValue) >= 0.0001
	default:
		return value >= thresholds.PassValue
	}
}

// AggregateResults aggregates metric values for an evaluation
func (e *CustomMetricsEvaluator) AggregateResults(metricID int, values []MetricValue) *MetricResult {
	metric, exists := e.metrics[metricID]
	if !exists {
		return nil
	}
	
	result := &MetricResult{
		MetricID:    metricID,
		MetricName:  metric.Name,
		SampleCount: len(values),
	}
	
	if len(values) == 0 {
		return result
	}
	
	// Extract numeric values for aggregation
	numericValues := make([]float64, 0, len(values))
	passCount := 0
	
	for _, v := range values {
		if v.Passed {
			passCount++
		}
		
		// Convert value to float64 for aggregation
		switch val := v.Value.(type) {
		case float64:
			numericValues = append(numericValues, val)
		case int:
			numericValues = append(numericValues, float64(val))
		case bool:
			if val {
				numericValues = append(numericValues, 1.0)
			} else {
				numericValues = append(numericValues, 0.0)
			}
		}
	}
	
	// Calculate aggregated value
	result.Value = e.aggregate(numericValues, metric.Aggregation)
	result.PassRate = float64(passCount) / float64(len(values))
	result.Passed = e.checkThreshold(result.Value, metric.Thresholds)
	
	// Add details for percentile aggregations
	if metric.Aggregation == AggregationP95 || metric.Aggregation == AggregationP99 {
		result.Details = map[string]interface{}{
			"min":    e.aggregate(numericValues, AggregationMin),
			"max":    e.aggregate(numericValues, AggregationMax),
			"median": e.aggregate(numericValues, AggregationMedian),
			"avg":    e.aggregate(numericValues, AggregationAverage),
		}
	}
	
	return result
}

// aggregate performs aggregation on numeric values
func (e *CustomMetricsEvaluator) aggregate(values []float64, aggregation AggregationType) float64 {
	if len(values) == 0 {
		return 0
	}
	
	switch aggregation {
	case AggregationAverage:
		sum := 0.0
		for _, v := range values {
			sum += v
		}
		return sum / float64(len(values))
		
	case AggregationSum:
		sum := 0.0
		for _, v := range values {
			sum += v
		}
		return sum
		
	case AggregationMin:
		min := values[0]
		for _, v := range values[1:] {
			if v < min {
				min = v
			}
		}
		return min
		
	case AggregationMax:
		max := values[0]
		for _, v := range values[1:] {
			if v > max {
				max = v
			}
		}
		return max
		
	case AggregationMedian:
		return percentile(values, 50)
		
	case AggregationP95:
		return percentile(values, 95)
		
	case AggregationP99:
		return percentile(values, 99)
		
	case AggregationCount:
		return float64(len(values))
		
	default:
		return e.aggregate(values, AggregationAverage)
	}
}

// percentile calculates the percentile value
func percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}
	
	// Sort values
	sorted := make([]float64, len(values))
	copy(sorted, values)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	
	// Calculate percentile index
	index := (p / 100.0) * float64(len(sorted)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))
	
	if lower == upper {
		return sorted[lower]
	}
	
	// Interpolate between two values
	weight := index - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

// Helper methods for extracting values from samples

func (e *CustomMetricsEvaluator) extractNumericValue(sample interface{}, metric *CustomMetric) interface{} {
	// Extract numeric value based on metric configuration
	// This would be customized based on your sample structure
	if m, ok := sample.(map[string]interface{}); ok {
		if v, exists := m[metric.Name]; exists {
			return v
		}
	}
	return 0.0
}

func (e *CustomMetricsEvaluator) extractBooleanValue(sample interface{}, metric *CustomMetric) interface{} {
	if m, ok := sample.(map[string]interface{}); ok {
		if v, exists := m[metric.Name]; exists {
			return v
		}
	}
	return false
}

func (e *CustomMetricsEvaluator) extractStringValue(sample interface{}, metric *CustomMetric) interface{} {
	if m, ok := sample.(map[string]interface{}); ok {
		if v, exists := m[metric.Name]; exists {
			return v
		}
	}
	return ""
}

func (e *CustomMetricsEvaluator) evaluateStringMetric(value interface{}, metric *CustomMetric) float64 {
	str, ok := value.(string)
	if !ok {
		return 0.0
	}
	
	// Example: Check if string matches a pattern
	if metric.Formula != "" {
		if matched, _ := regexp.MatchString(metric.Formula, str); matched {
			return 1.0
		}
	}
	
	// Example: Check string length
	if strings.Contains(metric.Name, "length") {
		return float64(len(str))
	}
	
	return 0.0
}

func (e *CustomMetricsEvaluator) evaluateFormula(formula string, sample interface{}) float64 {
	// Simple formula evaluation (would need a proper expression evaluator in production)
	// For now, just return a placeholder value
	return 0.0
}

// SaveMetric saves a custom metric definition
func (e *CustomMetricsEvaluator) SaveMetric(metric *CustomMetric) error {
	thresholdsJSON, err := json.Marshal(metric.Thresholds)
	if err != nil {
		return err
	}
	
	if metric.ID == 0 {
		// Insert new metric
		err = e.db.QueryRow(`
			INSERT INTO custom_metrics 
			(project_id, name, description, type, aggregation, formula, thresholds, weight, enabled)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id, created_at, updated_at
		`, metric.ProjectID, metric.Name, metric.Description, metric.Type,
			metric.Aggregation, metric.Formula, string(thresholdsJSON),
			metric.Weight, metric.Enabled).
			Scan(&metric.ID, &metric.CreatedAt, &metric.UpdatedAt)
	} else {
		// Update existing metric
		_, err = e.db.Exec(`
			UPDATE custom_metrics
			SET name = $2, description = $3, type = $4, aggregation = $5,
			    formula = $6, thresholds = $7, weight = $8, enabled = $9, updated_at = NOW()
			WHERE id = $1
		`, metric.ID, metric.Name, metric.Description, metric.Type,
			metric.Aggregation, metric.Formula, string(thresholdsJSON),
			metric.Weight, metric.Enabled)
		metric.UpdatedAt = time.Now()
	}
	
	if err != nil {
		return err
	}
	
	// Update in-memory cache
	e.metrics[metric.ID] = metric
	
	return nil
}

// DeleteMetric deletes a custom metric
func (e *CustomMetricsEvaluator) DeleteMetric(metricID int) error {
	_, err := e.db.Exec("DELETE FROM custom_metrics WHERE id = $1", metricID)
	if err != nil {
		return err
	}
	
	delete(e.metrics, metricID)
	return nil
}