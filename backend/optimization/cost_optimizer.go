package optimization

import (
	"database/sql"
	"fmt"
	"sort"
	"time"
)

// CostOptimizer provides cost optimization recommendations
type CostOptimizer struct {
	db *sql.DB
}

// NewCostOptimizer creates a new cost optimizer
func NewCostOptimizer(db *sql.DB) *CostOptimizer {
	return &CostOptimizer{db: db}
}

// CostRecommendation represents a cost optimization recommendation
type CostRecommendation struct {
	Type             string  `json:"type"`
	Priority         string  `json:"priority"`
	Title            string  `json:"title"`
	Description      string  `json:"description"`
	EstimatedSavings string  `json:"estimated_savings"`
	Impact           string  `json:"impact"`
	Actions          []string `json:"actions"`
	ModelDetails     *ModelCostDetails `json:"model_details,omitempty"`
}

// ModelCostDetails provides detailed cost analysis for a model
type ModelCostDetails struct {
	Model              string  `json:"model"`
	Provider           string  `json:"provider"`
	CurrentCost        float64 `json:"current_cost"`
	CostPerToken       float64 `json:"cost_per_token"`
	TokensPerDay       int     `json:"tokens_per_day"`
	AverageLatency     float64 `json:"average_latency"`
	ErrorRate          float64 `json:"error_rate"`
	UsagePattern       string  `json:"usage_pattern"`
	RecommendedActions []string `json:"recommended_actions"`
}

// CostOptimizationReport contains comprehensive cost analysis
type CostOptimizationReport struct {
	ProjectID           int                    `json:"project_id"`
	AnalysisPeriod      string                 `json:"analysis_period"`
	TotalCost           float64                `json:"total_cost"`
	PotentialSavings    float64                `json:"potential_savings"`
	SavingsPercentage   float64                `json:"savings_percentage"`
	Recommendations     []CostRecommendation   `json:"recommendations"`
	ModelBreakdown      []ModelCostDetails     `json:"model_breakdown"`
	CostTrends          []CostTrendPoint       `json:"cost_trends"`
	OptimizationScore   float64                `json:"optimization_score"`
	Summary             CostOptimizationSummary `json:"summary"`
}

// CostOptimizationSummary provides high-level insights
type CostOptimizationSummary struct {
	HighestCostModel     string  `json:"highest_cost_model"`
	LeastEfficientModel  string  `json:"least_efficient_model"`
	RecommendedModel     string  `json:"recommended_model"`
	QuickWins            []string `json:"quick_wins"`
	LongTermStrategies   []string `json:"long_term_strategies"`
}

// CostTrendPoint represents cost data over time
type CostTrendPoint struct {
	Date      string  `json:"date"`
	TotalCost float64 `json:"total_cost"`
	EventCount int    `json:"event_count"`
	AvgCostPerEvent float64 `json:"avg_cost_per_event"`
}

// AnalyzeCosts generates comprehensive cost optimization recommendations
func (co *CostOptimizer) AnalyzeCosts(projectID int, days int) (*CostOptimizationReport, error) {
	// Get overall cost metrics
	totalCost, err := co.getTotalCost(projectID, days)
	if err != nil {
		return nil, fmt.Errorf("failed to get total cost: %w", err)
	}

	// Get model breakdown
	modelBreakdown, err := co.getModelCostBreakdown(projectID, days)
	if err != nil {
		return nil, fmt.Errorf("failed to get model breakdown: %w", err)
	}

	// Get cost trends
	costTrends, err := co.getCostTrends(projectID, days)
	if err != nil {
		return nil, fmt.Errorf("failed to get cost trends: %w", err)
	}

	// Generate recommendations
	recommendations := co.generateRecommendations(modelBreakdown, totalCost)
	
	// Calculate potential savings
	potentialSavings := co.calculatePotentialSavings(recommendations)
	savingsPercentage := 0.0
	if totalCost > 0 {
		savingsPercentage = (potentialSavings / totalCost) * 100
	}

	// Calculate optimization score
	optimizationScore := co.calculateOptimizationScore(modelBreakdown, recommendations)

	// Generate summary
	summary := co.generateSummary(modelBreakdown, recommendations)

	return &CostOptimizationReport{
		ProjectID:           projectID,
		AnalysisPeriod:      fmt.Sprintf("%d days", days),
		TotalCost:           totalCost,
		PotentialSavings:    potentialSavings,
		SavingsPercentage:   savingsPercentage,
		Recommendations:     recommendations,
		ModelBreakdown:      modelBreakdown,
		CostTrends:          costTrends,
		OptimizationScore:   optimizationScore,
		Summary:             summary,
	}, nil
}

// getTotalCost calculates total cost for the period
func (co *CostOptimizer) getTotalCost(projectID int, days int) (float64, error) {
	query := `
		SELECT COALESCE(SUM(cost), 0) as total_cost
		FROM trace_events
		WHERE project_id = $1 
		AND created_at >= NOW() - INTERVAL '1 day' * $2`

	var totalCost float64
	err := co.db.QueryRow(query, projectID, days).Scan(&totalCost)
	return totalCost, err
}

// getModelCostBreakdown provides detailed cost analysis per model
func (co *CostOptimizer) getModelCostBreakdown(projectID int, days int) ([]ModelCostDetails, error) {
	query := `
		SELECT 
			COALESCE(model, 'unknown') as model,
			COALESCE(provider, 'unknown') as provider,
			SUM(cost) as total_cost,
			AVG(cost) as avg_cost,
			CASE 
				WHEN SUM(total_tokens) > 0 THEN SUM(cost) / SUM(total_tokens)
				ELSE 0 
			END as cost_per_token,
			COUNT(*) as event_count,
			SUM(total_tokens) as total_tokens,
			AVG(duration_ms) as avg_latency,
			AVG(CASE WHEN status != 'success' THEN 1.0 ELSE 0.0 END) as error_rate
		FROM trace_events
		WHERE project_id = $1 
		AND created_at >= NOW() - INTERVAL '1 day' * $2
		AND cost > 0
		GROUP BY model, provider
		ORDER BY total_cost DESC`

	rows, err := co.db.Query(query, projectID, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var breakdown []ModelCostDetails
	for rows.Next() {
		var detail ModelCostDetails
		var eventCount int
		var totalTokens int

		var avgCost float64
		err := rows.Scan(
			&detail.Model, &detail.Provider, &detail.CurrentCost,
			&avgCost, &detail.CostPerToken, &eventCount,
			&totalTokens, &detail.AverageLatency, &detail.ErrorRate,
		)
		if err != nil {
			continue
		}

		// Calculate tokens per day
		detail.TokensPerDay = totalTokens / days
		if detail.TokensPerDay < 1 {
			detail.TokensPerDay = totalTokens // For short periods
		}

		// Determine usage pattern
		detail.UsagePattern = co.determineUsagePattern(eventCount, days)

		// Generate model-specific recommendations
		detail.RecommendedActions = co.generateModelRecommendations(detail)

		breakdown = append(breakdown, detail)
	}

	return breakdown, nil
}

// getCostTrends returns cost trends over time
func (co *CostOptimizer) getCostTrends(projectID int, days int) ([]CostTrendPoint, error) {
	interval := "1 day"
	if days <= 7 {
		interval = "1 hour"
	}

	query := fmt.Sprintf(`
		SELECT 
			DATE_TRUNC('%s', created_at) as time_bucket,
			SUM(cost) as total_cost,
			COUNT(*) as event_count,
			AVG(cost) as avg_cost_per_event
		FROM trace_events
		WHERE project_id = $1 
		AND created_at >= NOW() - INTERVAL '1 day' * $2
		AND cost > 0
		GROUP BY time_bucket
		ORDER BY time_bucket`, interval)

	rows, err := co.db.Query(query, projectID, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trends []CostTrendPoint
	for rows.Next() {
		var trend CostTrendPoint
		var timestamp time.Time

		err := rows.Scan(&timestamp, &trend.TotalCost, &trend.EventCount, &trend.AvgCostPerEvent)
		if err != nil {
			continue
		}

		trend.Date = timestamp.Format(time.RFC3339)
		trends = append(trends, trend)
	}

	return trends, nil
}

// generateRecommendations creates cost optimization recommendations
func (co *CostOptimizer) generateRecommendations(models []ModelCostDetails, totalCost float64) []CostRecommendation {
	var recommendations []CostRecommendation

	if len(models) == 0 {
		return []CostRecommendation{{
			Type:        "general",
			Priority:    "low",
			Title:       "No Cost Data Available",
			Description: "No cost data found for analysis. Ensure cost tracking is enabled.",
			Actions:     []string{"Enable cost tracking in your model configurations"},
		}}
	}

	// Sort models by cost to identify highest cost models
	sort.Slice(models, func(i, j int) bool {
		return models[i].CurrentCost > models[j].CurrentCost
	})

	// Recommendation 1: High-cost model optimization
	if len(models) > 0 && models[0].CurrentCost > totalCost*0.3 {
		recommendations = append(recommendations, CostRecommendation{
			Type:             "model_optimization",
			Priority:         "high",
			Title:            "Optimize High-Cost Model Usage",
			Description:      fmt.Sprintf("%s (%s) accounts for %.1f%% of total costs", models[0].Model, models[0].Provider, (models[0].CurrentCost/totalCost)*100),
			EstimatedSavings: fmt.Sprintf("$%.2f", models[0].CurrentCost*0.3),
			Impact:           "High",
			Actions: []string{
				"Consider using a more cost-effective model for non-critical tasks",
				"Implement request caching to reduce redundant API calls",
				"Optimize prompt length to reduce token usage",
			},
			ModelDetails: &models[0],
		})
	}

	// Recommendation 2: Error rate optimization
	for _, model := range models {
		if model.ErrorRate > 0.05 { // More than 5% error rate
			recommendations = append(recommendations, CostRecommendation{
				Type:             "error_reduction",
				Priority:         "medium",
				Title:            "Reduce Error Rate Costs",
				Description:      fmt.Sprintf("%s (%s) has %.1f%% error rate, wasting costs on failed requests", model.Model, model.Provider, model.ErrorRate*100),
				EstimatedSavings: fmt.Sprintf("$%.2f", model.CurrentCost*model.ErrorRate),
				Impact:           "Medium",
				Actions: []string{
					"Review and fix common error patterns",
					"Implement better input validation",
					"Add retry logic with exponential backoff",
				},
				ModelDetails: &model,
			})
			break // Only show one error rate recommendation
		}
	}

	// Recommendation 3: Model consolidation
	if len(models) > 3 {
		recommendations = append(recommendations, CostRecommendation{
			Type:        "consolidation",
			Priority:    "medium",
			Title:       "Consolidate Model Usage",
			Description: fmt.Sprintf("Using %d different models increases complexity and costs", len(models)),
			EstimatedSavings: fmt.Sprintf("$%.2f", totalCost*0.15),
			Impact:      "Medium",
			Actions: []string{
				"Evaluate if multiple models are necessary",
				"Standardize on 1-2 primary models",
				"Negotiate volume discounts with preferred providers",
			},
		})
	}

	// Recommendation 4: Token optimization
	highTokenModels := make([]ModelCostDetails, 0)
	for _, model := range models {
		if model.TokensPerDay > 10000 && model.CostPerToken > 0.00003 { // High usage + expensive
			highTokenModels = append(highTokenModels, model)
		}
	}

	if len(highTokenModels) > 0 {
		recommendations = append(recommendations, CostRecommendation{
			Type:        "token_optimization",
			Priority:    "high",
			Title:       "Optimize Token Usage",
			Description: "High token usage detected on expensive models",
			EstimatedSavings: fmt.Sprintf("$%.2f", totalCost*0.25),
			Impact:      "High",
			Actions: []string{
				"Implement prompt compression techniques",
				"Use shorter system prompts",
				"Cache common responses",
				"Consider cheaper models for simple tasks",
			},
		})
	}

	// Recommendation 5: Cost monitoring
	if totalCost > 100 {
		recommendations = append(recommendations, CostRecommendation{
			Type:        "monitoring",
			Priority:    "medium",
			Title:       "Implement Cost Alerts",
			Description: "Set up automated cost monitoring to prevent unexpected charges",
			EstimatedSavings: "Prevent overruns",
			Impact:      "Preventive",
			Actions: []string{
				"Set daily/monthly cost thresholds",
				"Configure Slack/email alerts",
				"Implement automatic request throttling",
				"Regular cost reviews",
			},
		})
	}

	return recommendations
}

// calculatePotentialSavings estimates total potential savings
func (co *CostOptimizer) calculatePotentialSavings(recommendations []CostRecommendation) float64 {
	totalSavings := 0.0
	for _, rec := range recommendations {
		if rec.EstimatedSavings != "" && rec.EstimatedSavings[0] == '$' {
			var savings float64
			fmt.Sscanf(rec.EstimatedSavings[1:], "%f", &savings)
			totalSavings += savings
		}
	}
	return totalSavings
}

// calculateOptimizationScore provides a score from 0-100
func (co *CostOptimizer) calculateOptimizationScore(models []ModelCostDetails, recommendations []CostRecommendation) float64 {
	if len(models) == 0 {
		return 0
	}

	score := 100.0

	// Deduct points for high-priority issues
	for _, rec := range recommendations {
		switch rec.Priority {
		case "high":
			score -= 20
		case "medium":
			score -= 10
		case "low":
			score -= 5
		}
	}

	// Deduct points for high error rates
	for _, model := range models {
		if model.ErrorRate > 0.1 {
			score -= 15
		} else if model.ErrorRate > 0.05 {
			score -= 10
		}
	}

	// Deduct points for too many models
	if len(models) > 5 {
		score -= 10
	}

	if score < 0 {
		score = 0
	}

	return score
}

// generateSummary creates high-level summary
func (co *CostOptimizer) generateSummary(models []ModelCostDetails, recommendations []CostRecommendation) CostOptimizationSummary {
	summary := CostOptimizationSummary{
		QuickWins:          []string{},
		LongTermStrategies: []string{},
	}

	if len(models) > 0 {
		// Sort by cost
		sort.Slice(models, func(i, j int) bool {
			return models[i].CurrentCost > models[j].CurrentCost
		})
		summary.HighestCostModel = fmt.Sprintf("%s (%s)", models[0].Model, models[0].Provider)

		// Find least efficient (highest cost per token with high error rate)
		leastEfficient := models[0]
		for _, model := range models {
			efficiency := model.CostPerToken * (1 + model.ErrorRate)
			currentWorst := leastEfficient.CostPerToken * (1 + leastEfficient.ErrorRate)
			if efficiency > currentWorst {
				leastEfficient = model
			}
		}
		summary.LeastEfficientModel = fmt.Sprintf("%s (%s)", leastEfficient.Model, leastEfficient.Provider)

		// Find most efficient model for recommendation
		mostEfficient := models[0]
		for _, model := range models {
			efficiency := model.CostPerToken * (1 + model.ErrorRate)
			currentBest := mostEfficient.CostPerToken * (1 + mostEfficient.ErrorRate)
			if efficiency < currentBest && model.ErrorRate < 0.05 {
				mostEfficient = model
			}
		}
		summary.RecommendedModel = fmt.Sprintf("%s (%s)", mostEfficient.Model, mostEfficient.Provider)
	}

	// Extract quick wins and long-term strategies
	for _, rec := range recommendations {
		if rec.Priority == "high" {
			summary.QuickWins = append(summary.QuickWins, rec.Title)
		} else if rec.Priority == "medium" {
			summary.LongTermStrategies = append(summary.LongTermStrategies, rec.Title)
		}
	}

	return summary
}

// determineUsagePattern analyzes usage patterns
func (co *CostOptimizer) determineUsagePattern(eventCount, days int) string {
	avgPerDay := float64(eventCount) / float64(days)
	
	if avgPerDay < 10 {
		return "low"
	} else if avgPerDay < 100 {
		return "moderate"
	} else if avgPerDay < 1000 {
		return "high"
	} else {
		return "very_high"
	}
}

// generateModelRecommendations creates model-specific recommendations
func (co *CostOptimizer) generateModelRecommendations(model ModelCostDetails) []string {
	recommendations := []string{}

	if model.ErrorRate > 0.1 {
		recommendations = append(recommendations, "High error rate - review API usage patterns")
	}

	if model.CostPerToken > 0.00005 {
		recommendations = append(recommendations, "Expensive model - consider alternatives for simple tasks")
	}

	if model.AverageLatency > 2000 {
		recommendations = append(recommendations, "High latency - consider faster alternatives")
	}

	if model.TokensPerDay > 50000 {
		recommendations = append(recommendations, "High usage - implement caching and optimization")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Model performing well - continue monitoring")
	}

	return recommendations
}