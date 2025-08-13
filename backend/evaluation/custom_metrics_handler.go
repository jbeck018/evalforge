package evaluation

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CustomMetricsHandler handles custom metrics API endpoints
type CustomMetricsHandler struct {
	evaluator *CustomMetricsEvaluator
	db        *sql.DB
}

// NewCustomMetricsHandler creates a new custom metrics handler
func NewCustomMetricsHandler(db *sql.DB) *CustomMetricsHandler {
	return &CustomMetricsHandler{
		evaluator: NewCustomMetricsEvaluator(db),
		db:        db,
	}
}

// GetCustomMetrics returns custom metrics for a project
func (h *CustomMetricsHandler) GetCustomMetrics(c *gin.Context) {
	projectID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Verify user has access
	userID := c.GetInt("user_id")
	var hasAccess bool
	err = h.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM projects 
			WHERE id = $1 AND user_id = $2
		)
	`, projectID, userID).Scan(&hasAccess)

	if err != nil || !hasAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Load metrics
	if err := h.evaluator.LoadMetrics(projectID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load metrics"})
		return
	}

	// Convert map to slice for response
	metrics := make([]*CustomMetric, 0, len(h.evaluator.metrics))
	for _, metric := range h.evaluator.metrics {
		metrics = append(metrics, metric)
	}

	c.JSON(http.StatusOK, gin.H{"metrics": metrics})
}

// CreateCustomMetric creates a new custom metric
func (h *CustomMetricsHandler) CreateCustomMetric(c *gin.Context) {
	projectID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Verify user has access
	userID := c.GetInt("user_id")
	var hasAccess bool
	err = h.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM projects 
			WHERE id = $1 AND user_id = $2
		)
	`, projectID, userID).Scan(&hasAccess)

	if err != nil || !hasAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var metric CustomMetric
	if err := c.ShouldBindJSON(&metric); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	metric.ProjectID = projectID
	
	// Set defaults
	if metric.Weight == 0 {
		metric.Weight = 1.0
	}
	if metric.Aggregation == "" {
		metric.Aggregation = AggregationAverage
	}
	metric.Enabled = true

	if err := h.evaluator.SaveMetric(&metric); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save metric"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      metric.ID,
		"message": "Custom metric created successfully",
		"metric":  metric,
	})
}

// UpdateCustomMetric updates an existing custom metric
func (h *CustomMetricsHandler) UpdateCustomMetric(c *gin.Context) {
	metricID, err := strconv.Atoi(c.Param("metricId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid metric ID"})
		return
	}

	// Verify user has access
	userID := c.GetInt("user_id")
	var hasAccess bool
	var projectID int
	err = h.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM custom_metrics cm
			JOIN projects p ON cm.project_id = p.id
			WHERE cm.id = $1 AND p.user_id = $2
		), cm.project_id
		FROM custom_metrics cm
		WHERE cm.id = $1
	`, metricID, userID).Scan(&hasAccess, &projectID)

	if err != nil || !hasAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Load existing metric
	var metric CustomMetric
	err = h.db.QueryRow(`
		SELECT id, project_id, name, description, type, aggregation, 
		       formula, thresholds, weight, enabled
		FROM custom_metrics
		WHERE id = $1
	`, metricID).Scan(
		&metric.ID, &metric.ProjectID, &metric.Name, &metric.Description,
		&metric.Type, &metric.Aggregation, &metric.Formula,
		&metric.Thresholds, &metric.Weight, &metric.Enabled,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Metric not found"})
		return
	}

	// Apply updates
	if name, ok := updates["name"].(string); ok {
		metric.Name = name
	}
	if desc, ok := updates["description"].(string); ok {
		metric.Description = desc
	}
	if metricType, ok := updates["type"].(string); ok {
		metric.Type = MetricType(metricType)
	}
	if agg, ok := updates["aggregation"].(string); ok {
		metric.Aggregation = AggregationType(agg)
	}
	if formula, ok := updates["formula"].(string); ok {
		metric.Formula = formula
	}
	if weight, ok := updates["weight"].(float64); ok {
		metric.Weight = weight
	}
	if enabled, ok := updates["enabled"].(bool); ok {
		metric.Enabled = enabled
	}
	if thresholds, ok := updates["thresholds"].(map[string]interface{}); ok {
		thresholdsJSON, _ := json.Marshal(thresholds)
		json.Unmarshal(thresholdsJSON, &metric.Thresholds)
	}

	// Save updated metric
	if err := h.evaluator.SaveMetric(&metric); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update metric"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Metric updated successfully",
		"metric":  metric,
	})
}

// DeleteCustomMetric deletes a custom metric
func (h *CustomMetricsHandler) DeleteCustomMetric(c *gin.Context) {
	metricID, err := strconv.Atoi(c.Param("metricId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid metric ID"})
		return
	}

	// Verify user has access
	userID := c.GetInt("user_id")
	var hasAccess bool
	err = h.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM custom_metrics cm
			JOIN projects p ON cm.project_id = p.id
			WHERE cm.id = $1 AND p.user_id = $2
		)
	`, metricID, userID).Scan(&hasAccess)

	if err != nil || !hasAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	if err := h.evaluator.DeleteMetric(metricID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete metric"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Metric deleted successfully"})
}

// GetMetricTemplates returns available metric templates
func (h *CustomMetricsHandler) GetMetricTemplates(c *gin.Context) {
	rows, err := h.db.Query(`
		SELECT id, name, category, description, type, aggregation, 
		       formula, default_thresholds
		FROM metric_templates
		WHERE is_public = TRUE
		ORDER BY category, name
	`)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch templates"})
		return
	}
	defer rows.Close()

	templates := []gin.H{}
	for rows.Next() {
		var id int
		var name, category, description, metricType, aggregation string
		var formula sql.NullString
		var thresholdsJSON string

		err := rows.Scan(&id, &name, &category, &description, &metricType,
			&aggregation, &formula, &thresholdsJSON)
		if err != nil {
			continue
		}

		template := gin.H{
			"id":          id,
			"name":        name,
			"category":    category,
			"description": description,
			"type":        metricType,
			"aggregation": aggregation,
			"thresholds":  thresholdsJSON,
		}

		if formula.Valid {
			template["formula"] = formula.String
		}

		templates = append(templates, template)
	}

	c.JSON(http.StatusOK, gin.H{"templates": templates})
}

// CreateMetricFromTemplate creates a metric from a template
func (h *CustomMetricsHandler) CreateMetricFromTemplate(c *gin.Context) {
	projectID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Verify user has access
	userID := c.GetInt("user_id")
	var hasAccess bool
	err = h.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM projects 
			WHERE id = $1 AND user_id = $2
		)
	`, projectID, userID).Scan(&hasAccess)

	if err != nil || !hasAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var req struct {
		TemplateID int    `json:"template_id" binding:"required"`
		Name       string `json:"name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Load template
	var metric CustomMetric
	var thresholdsJSON string
	err = h.db.QueryRow(`
		SELECT name, description, type, aggregation, formula, default_thresholds
		FROM metric_templates
		WHERE id = $1
	`, req.TemplateID).Scan(
		&metric.Name, &metric.Description, &metric.Type,
		&metric.Aggregation, &metric.Formula, &thresholdsJSON,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}

	// Override name if provided
	if req.Name != "" {
		metric.Name = req.Name
	}

	// Parse thresholds
	json.Unmarshal([]byte(thresholdsJSON), &metric.Thresholds)

	// Set project and defaults
	metric.ProjectID = projectID
	metric.Weight = 1.0
	metric.Enabled = true

	// Save metric
	if err := h.evaluator.SaveMetric(&metric); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create metric"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      metric.ID,
		"message": "Metric created from template",
		"metric":  metric,
	})
}

// GetMetricResults returns results for a specific evaluation
func (h *CustomMetricsHandler) GetMetricResults(c *gin.Context) {
	evaluationID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid evaluation ID"})
		return
	}

	// Verify user has access
	userID := c.GetInt("user_id")
	var hasAccess bool
	err = h.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM evaluations e
			JOIN projects p ON e.project_id = p.id
			WHERE e.id = $1 AND p.user_id = $2
		)
	`, evaluationID, userID).Scan(&hasAccess)

	if err != nil || !hasAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Get metric results
	rows, err := h.db.Query(`
		SELECT mr.id, mr.metric_id, cm.name, cm.type, mr.aggregated_value,
		       mr.passed, mr.pass_rate, mr.sample_count, mr.details
		FROM metric_results mr
		JOIN custom_metrics cm ON mr.metric_id = cm.id
		WHERE mr.evaluation_id = $1
		ORDER BY cm.name
	`, evaluationID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch results"})
		return
	}
	defer rows.Close()

	results := []gin.H{}
	for rows.Next() {
		var id, metricID, sampleCount int
		var name, metricType string
		var value, passRate float64
		var passed bool
		var detailsJSON sql.NullString

		err := rows.Scan(&id, &metricID, &name, &metricType, &value,
			&passed, &passRate, &sampleCount, &detailsJSON)
		if err != nil {
			continue
		}

		result := gin.H{
			"id":           id,
			"metric_id":    metricID,
			"metric_name":  name,
			"metric_type":  metricType,
			"value":        value,
			"passed":       passed,
			"pass_rate":    passRate,
			"sample_count": sampleCount,
		}

		if detailsJSON.Valid {
			result["details"] = detailsJSON.String
		}

		results = append(results, result)
	}

	c.JSON(http.StatusOK, gin.H{"results": results})
}