package comparison

import (
	"database/sql"
	"fmt"
	"time"
)

// ModelComparator handles model performance comparisons
type ModelComparator struct {
	db *sql.DB
}

// NewModelComparator creates a new model comparator
func NewModelComparator(db *sql.DB) *ModelComparator {
	return &ModelComparator{db: db}
}

// ModelMetrics represents performance metrics for a model
type ModelMetrics struct {
	Model             string  `json:"model"`
	Provider          string  `json:"provider"`
	TotalEvents       int     `json:"total_events"`
	SuccessRate       float64 `json:"success_rate"`
	AvgLatency        float64 `json:"avg_latency_ms"`
	P95Latency        float64 `json:"p95_latency_ms"`
	P99Latency        float64 `json:"p99_latency_ms"`
	AvgCost           float64 `json:"avg_cost"`
	TotalCost         float64 `json:"total_cost"`
	AvgTokensInput    float64 `json:"avg_tokens_input"`
	AvgTokensOutput   float64 `json:"avg_tokens_output"`
	AvgTokensTotal    float64 `json:"avg_tokens_total"`
	CostPerToken      float64 `json:"cost_per_token"`
	TokensPerSecond   float64 `json:"tokens_per_second"`
	ErrorRate         float64 `json:"error_rate"`
	LastUsed          string  `json:"last_used"`
}

// ComparisonResult represents a comparison between models
type ComparisonResult struct {
	TimeRange      string         `json:"time_range"`
	Models         []ModelMetrics `json:"models"`
	Summary        ComparisonSummary `json:"summary"`
	Recommendations []string      `json:"recommendations"`
}

// ComparisonSummary provides high-level comparison insights
type ComparisonSummary struct {
	FastestModel      string  `json:"fastest_model"`
	CheapestModel     string  `json:"cheapest_model"`
	MostReliableModel string  `json:"most_reliable_model"`
	BestValueModel    string  `json:"best_value_model"`
	PerformanceSpread float64 `json:"performance_spread"`
	CostSpread        float64 `json:"cost_spread"`
}

// CompareModels compares performance across different models for a project
func (mc *ModelComparator) CompareModels(projectID int, timeRange string) (*ComparisonResult, error) {
	// Determine time filter
	var timeFilter string
	switch timeRange {
	case "1h":
		timeFilter = "created_at >= NOW() - INTERVAL '1 hour'"
	case "24h":
		timeFilter = "created_at >= NOW() - INTERVAL '24 hours'"
	case "7d":
		timeFilter = "created_at >= NOW() - INTERVAL '7 days'"
	case "30d":
		timeFilter = "created_at >= NOW() - INTERVAL '30 days'"
	default:
		timeFilter = "created_at >= NOW() - INTERVAL '24 hours'"
		timeRange = "24h"
	}

	query := fmt.Sprintf(`
		SELECT 
			COALESCE(model, 'unknown') as model,
			COALESCE(provider, 'unknown') as provider,
			COUNT(*) as total_events,
			AVG(CASE WHEN status = 'success' THEN 1.0 ELSE 0.0 END) as success_rate,
			AVG(duration_ms) as avg_latency,
			PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY duration_ms) as p95_latency,
			PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY duration_ms) as p99_latency,
			AVG(cost) as avg_cost,
			SUM(cost) as total_cost,
			AVG(prompt_tokens) as avg_tokens_input,
			AVG(completion_tokens) as avg_tokens_output,
			AVG(total_tokens) as avg_tokens_total,
			CASE 
				WHEN SUM(total_tokens) > 0 THEN SUM(cost) / SUM(total_tokens)
				ELSE 0 
			END as cost_per_token,
			CASE 
				WHEN AVG(duration_ms) > 0 THEN (AVG(total_tokens) * 1000.0) / AVG(duration_ms)
				ELSE 0 
			END as tokens_per_second,
			AVG(CASE WHEN status != 'success' THEN 1.0 ELSE 0.0 END) as error_rate,
			MAX(created_at) as last_used
		FROM trace_events 
		WHERE project_id = $1 AND %s
		GROUP BY model, provider
		HAVING COUNT(*) >= 5
		ORDER BY total_events DESC, avg_latency ASC
	`, timeFilter)

	rows, err := mc.db.Query(query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to query model metrics: %w", err)
	}
	defer rows.Close()

	var models []ModelMetrics
	for rows.Next() {
		var m ModelMetrics
		var lastUsed time.Time

		err := rows.Scan(
			&m.Model, &m.Provider, &m.TotalEvents, &m.SuccessRate,
			&m.AvgLatency, &m.P95Latency, &m.P99Latency,
			&m.AvgCost, &m.TotalCost,
			&m.AvgTokensInput, &m.AvgTokensOutput, &m.AvgTokensTotal,
			&m.CostPerToken, &m.TokensPerSecond, &m.ErrorRate, &lastUsed,
		)
		if err != nil {
			continue
		}

		m.LastUsed = lastUsed.Format(time.RFC3339)
		models = append(models, m)
	}

	if len(models) == 0 {
		return &ComparisonResult{
			TimeRange: timeRange,
			Models:    []ModelMetrics{},
			Summary:   ComparisonSummary{},
			Recommendations: []string{"No sufficient data for comparison. Need at least 5 events per model."},
		}, nil
	}

	// Generate summary and recommendations
	summary := mc.generateSummary(models)
	recommendations := mc.generateRecommendations(models)

	return &ComparisonResult{
		TimeRange:       timeRange,
		Models:          models,
		Summary:         summary,
		Recommendations: recommendations,
	}, nil
}

// generateSummary creates a high-level comparison summary
func (mc *ModelComparator) generateSummary(models []ModelMetrics) ComparisonSummary {
	if len(models) == 0 {
		return ComparisonSummary{}
	}

	var fastestModel, cheapestModel, mostReliableModel, bestValueModel string
	var minLatency, minCost, maxReliability float64 = 999999, 999999, 0
	var bestValue float64

	latencies := make([]float64, len(models))
	costs := make([]float64, len(models))

	for i, model := range models {
		// Track latencies and costs for spread calculation
		latencies[i] = model.AvgLatency
		costs[i] = model.AvgCost

		// Find fastest model (lowest latency)
		if model.AvgLatency < minLatency {
			minLatency = model.AvgLatency
			fastestModel = fmt.Sprintf("%s (%s)", model.Model, model.Provider)
		}

		// Find cheapest model (lowest cost per token)
		if model.CostPerToken < minCost && model.CostPerToken > 0 {
			minCost = model.CostPerToken
			cheapestModel = fmt.Sprintf("%s (%s)", model.Model, model.Provider)
		}

		// Find most reliable model (highest success rate)
		if model.SuccessRate > maxReliability {
			maxReliability = model.SuccessRate
			mostReliableModel = fmt.Sprintf("%s (%s)", model.Model, model.Provider)
		}

		// Calculate value score (success rate / cost per token)
		valueScore := 0.0
		if model.CostPerToken > 0 {
			valueScore = model.SuccessRate / model.CostPerToken
		}
		if valueScore > bestValue {
			bestValue = valueScore
			bestValueModel = fmt.Sprintf("%s (%s)", model.Model, model.Provider)
		}
	}

	// Calculate spreads
	performanceSpread := mc.calculateSpread(latencies)
	costSpread := mc.calculateSpread(costs)

	return ComparisonSummary{
		FastestModel:      fastestModel,
		CheapestModel:     cheapestModel,
		MostReliableModel: mostReliableModel,
		BestValueModel:    bestValueModel,
		PerformanceSpread: performanceSpread,
		CostSpread:        costSpread,
	}
}

// generateRecommendations provides actionable insights
func (mc *ModelComparator) generateRecommendations(models []ModelMetrics) []string {
	recommendations := []string{}

	if len(models) < 2 {
		return []string{"More models needed for meaningful comparison"}
	}

	// Analyze patterns and provide recommendations
	var highCostModels, slowModels, unreliableModels []string
	var avgCost, avgLatency, avgReliability float64

	for _, model := range models {
		avgCost += model.CostPerToken
		avgLatency += model.AvgLatency
		avgReliability += model.SuccessRate
	}
	avgCost /= float64(len(models))
	avgLatency /= float64(len(models))
	avgReliability /= float64(len(models))

	for _, model := range models {
		modelName := fmt.Sprintf("%s (%s)", model.Model, model.Provider)
		
		if model.CostPerToken > avgCost*1.5 {
			highCostModels = append(highCostModels, modelName)
		}
		if model.AvgLatency > avgLatency*1.5 {
			slowModels = append(slowModels, modelName)
		}
		if model.SuccessRate < avgReliability*0.9 {
			unreliableModels = append(unreliableModels, modelName)
		}
	}

	// Generate specific recommendations
	if len(highCostModels) > 0 {
		recommendations = append(recommendations, 
			fmt.Sprintf("Consider alternatives to high-cost models: %v", highCostModels))
	}
	
	if len(slowModels) > 0 {
		recommendations = append(recommendations, 
			fmt.Sprintf("Optimize latency by replacing slow models: %v", slowModels))
	}
	
	if len(unreliableModels) > 0 {
		recommendations = append(recommendations, 
			fmt.Sprintf("Review error handling for unreliable models: %v", unreliableModels))
	}

	// Performance insights
	if avgLatency > 2000 {
		recommendations = append(recommendations, 
			"Overall response times are high. Consider using faster models for real-time applications.")
	}

	// Cost optimization
	totalCost := 0.0
	for _, model := range models {
		totalCost += model.TotalCost
	}
	if totalCost > 100 {
		recommendations = append(recommendations, 
			"Significant model costs detected. Consider cost-optimized models for non-critical tasks.")
	}

	return recommendations
}

// calculateSpread calculates the coefficient of variation for a dataset
func (mc *ModelComparator) calculateSpread(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	// Calculate mean
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))

	if mean == 0 {
		return 0
	}

	// Calculate standard deviation
	varianceSum := 0.0
	for _, v := range values {
		diff := v - mean
		varianceSum += diff * diff
	}
	variance := varianceSum / float64(len(values))
	stdDev := variance // Simplified approximation

	// Coefficient of variation (relative variability)
	return stdDev / mean
}

// GetModelTrends returns performance trends for models over time
func (mc *ModelComparator) GetModelTrends(projectID int, model, provider string, days int) ([]TrendPoint, error) {
	query := `
		SELECT 
			DATE_TRUNC('hour', created_at) as time_bucket,
			AVG(duration_ms) as avg_latency,
			AVG(cost) as avg_cost,
			AVG(CASE WHEN status = 'success' THEN 1.0 ELSE 0.0 END) as success_rate,
			COUNT(*) as event_count
		FROM trace_events 
		WHERE project_id = $1 
			AND model = $2 
			AND provider = $3
			AND created_at >= NOW() - INTERVAL '%d days'
		GROUP BY time_bucket
		ORDER BY time_bucket
	`

	formattedQuery := fmt.Sprintf(query, days)
	rows, err := mc.db.Query(formattedQuery, projectID, model, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to query model trends: %w", err)
	}
	defer rows.Close()

	var trends []TrendPoint
	for rows.Next() {
		var trend TrendPoint
		var timestamp time.Time

		err := rows.Scan(
			&timestamp, &trend.AvgLatency, &trend.AvgCost, 
			&trend.SuccessRate, &trend.EventCount,
		)
		if err != nil {
			continue
		}

		trend.Timestamp = timestamp.Format(time.RFC3339)
		trends = append(trends, trend)
	}

	return trends, nil
}

// TrendPoint represents a data point in a model's performance trend
type TrendPoint struct {
	Timestamp   string  `json:"timestamp"`
	AvgLatency  float64 `json:"avg_latency"`
	AvgCost     float64 `json:"avg_cost"`
	SuccessRate float64 `json:"success_rate"`
	EventCount  int     `json:"event_count"`
}