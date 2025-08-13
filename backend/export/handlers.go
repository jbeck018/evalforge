package export

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ExportHandler struct {
	db *sql.DB
}

func NewExportHandler(db *sql.DB) *ExportHandler {
	return &ExportHandler{db: db}
}

// HandleExport handles export requests
func (h *ExportHandler) HandleExport(c *gin.Context) {
	var req ExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user context
	userID, _ := c.Get("user_id")
	
	// Validate project access
	if req.ProjectID != nil {
		var count int
		err := h.db.QueryRow(
			"SELECT COUNT(*) FROM projects WHERE id = $1 AND user_id = $2",
			*req.ProjectID, userID,
		).Scan(&count)
		
		if err != nil || count == 0 {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to project"})
			return
		}
	}

	// Export data based on type
	var data []byte
	var filename string
	var err error

	switch req.DataType {
	case "analytics":
		data, filename, err = h.exportAnalytics(req)
	case "traces":
		data, filename, err = h.exportTraces(req)
	case "evaluations":
		data, filename, err = h.exportEvaluations(req)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data type"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set content type based on format
	contentType := getContentType(req.Format)
	
	// Send file
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, contentType, data)
}

func (h *ExportHandler) exportAnalytics(req ExportRequest) ([]byte, string, error) {
	var rows []AnalyticsRow
	
	query := `
		SELECT 
			start_time as timestamp,
			project_id,
			operation_type as event_type,
			COALESCE(metadata->>'model', '') as model,
			COALESCE(metadata->>'provider', '') as provider,
			duration_ms::float as latency_ms,
			prompt_tokens as tokens_input,
			completion_tokens as tokens_output,
			cost::float as cost,
			0 as error_rate,
			1 as success_rate,
			0 as p50_latency,
			0 as p95_latency,
			0 as p99_latency
		FROM trace_events
		WHERE 1=1
	`
	
	args := []interface{}{}
	argCount := 0
	
	// Apply filters
	if req.ProjectID != nil {
		argCount++
		query += fmt.Sprintf(" AND project_id = $%d", argCount)
		args = append(args, *req.ProjectID)
	}
	if req.StartDate != nil {
		argCount++
		query += fmt.Sprintf(" AND start_time >= $%d", argCount)
		args = append(args, *req.StartDate)
	}
	if req.EndDate != nil {
		argCount++
		query += fmt.Sprintf(" AND start_time <= $%d", argCount)
		args = append(args, *req.EndDate)
	}
	
	result, err := h.db.Query(query, args...)
	if err != nil {
		return nil, "", err
	}
	defer result.Close()
	
	for result.Next() {
		var row AnalyticsRow
		err := result.Scan(
			&row.Timestamp, &row.ProjectID, &row.EventType,
			&row.Model, &row.Provider, &row.LatencyMS,
			&row.TokensInput, &row.TokensOutput, &row.Cost,
			&row.ErrorRate, &row.SuccessRate,
			&row.P50Latency, &row.P95Latency, &row.P99Latency,
		)
		if err != nil {
			return nil, "", err
		}
		rows = append(rows, row)
	}
	
	// Create exporter and export data
	exporter := NewExporter(req.Format)
	data, err := exporter.ExportAnalytics(rows)
	if err != nil {
		return nil, "", err
	}
	
	// Generate filename
	timestamp := time.Now().Format("20060102_150405")
	extension := getFileExtension(req.Format)
	filename := fmt.Sprintf("analytics_%s.%s", timestamp, extension)
	
	return data, filename, nil
}

func (h *ExportHandler) exportTraces(req ExportRequest) ([]byte, string, error) {
	var traces []TraceExport
	
	query := `
		SELECT 
			trace_id,
			span_id,
			parent_span_id,
			project_id,
			start_time as timestamp,
			duration_ms::float as duration_ms,
			operation_type as name,
			operation_type as type,
			status,
			COALESCE(metadata->>'model', '') as model,
			COALESCE(metadata->>'provider', '') as provider,
			prompt_tokens as tokens_input,
			completion_tokens as tokens_output,
			cost::float as cost,
			COALESCE(metadata->>'error', '') as error_message
		FROM trace_events
		WHERE 1=1
	`
	
	args := []interface{}{}
	argCount := 0
	
	// Apply filters
	if req.ProjectID != nil {
		argCount++
		query += fmt.Sprintf(" AND project_id = $%d", argCount)
		args = append(args, *req.ProjectID)
	}
	if req.StartDate != nil {
		argCount++
		query += fmt.Sprintf(" AND start_time >= $%d", argCount)
		args = append(args, *req.StartDate)
	}
	if req.EndDate != nil {
		argCount++
		query += fmt.Sprintf(" AND start_time <= $%d", argCount)
		args = append(args, *req.EndDate)
	}
	
	result, err := h.db.Query(query, args...)
	if err != nil {
		return nil, "", err
	}
	defer result.Close()
	
	for result.Next() {
		var trace TraceExport
		err := result.Scan(
			&trace.TraceID, &trace.SpanID, &trace.ParentSpanID,
			&trace.ProjectID, &trace.Timestamp, &trace.DurationMS,
			&trace.Name, &trace.Type, &trace.Status,
			&trace.Model, &trace.Provider,
			&trace.TokensInput, &trace.TokensOutput,
			&trace.Cost, &trace.ErrorMessage,
		)
		if err != nil {
			return nil, "", err
		}
		traces = append(traces, trace)
	}
	
	// Create exporter and export data
	exporter := NewExporter(req.Format)
	data, err := exporter.ExportTraces(traces)
	if err != nil {
		return nil, "", err
	}
	
	// Generate filename
	timestamp := time.Now().Format("20060102_150405")
	extension := getFileExtension(req.Format)
	filename := fmt.Sprintf("traces_%s.%s", timestamp, extension)
	
	return data, filename, nil
}

func (h *ExportHandler) exportEvaluations(req ExportRequest) ([]byte, string, error) {
	var evals []EvaluationExport
	
	query := `
		SELECT 
			id as evaluation_id,
			project_id,
			name,
			created_at as timestamp,
			status,
			0 as total_samples,
			0 as passed_samples,
			0 as failed_samples,
			0.0 as success_rate,
			0.0 as avg_latency,
			0.0 as avg_cost,
			'' as model,
			'' as provider,
			'{}' as metrics
		FROM evaluations
		WHERE 1=1
	`
	
	args := []interface{}{}
	argCount := 0
	
	// Apply filters
	if req.ProjectID != nil {
		argCount++
		query += fmt.Sprintf(" AND project_id = $%d", argCount)
		args = append(args, *req.ProjectID)
	}
	if req.StartDate != nil {
		argCount++
		query += fmt.Sprintf(" AND start_time >= $%d", argCount)
		args = append(args, *req.StartDate)
	}
	if req.EndDate != nil {
		argCount++
		query += fmt.Sprintf(" AND start_time <= $%d", argCount)
		args = append(args, *req.EndDate)
	}
	
	result, err := h.db.Query(query, args...)
	if err != nil {
		return nil, "", err
	}
	defer result.Close()
	
	for result.Next() {
		var eval EvaluationExport
		err := result.Scan(
			&eval.EvaluationID, &eval.ProjectID, &eval.Name,
			&eval.Timestamp, &eval.Status,
			&eval.TotalSamples, &eval.PassedSamples, &eval.FailedSamples,
			&eval.SuccessRate, &eval.AvgLatency, &eval.AvgCost,
			&eval.Model, &eval.Provider, &eval.Metrics,
		)
		if err != nil {
			return nil, "", err
		}
		evals = append(evals, eval)
	}
	
	// Create exporter and export data
	exporter := NewExporter(req.Format)
	data, err := exporter.ExportEvaluations(evals)
	if err != nil {
		return nil, "", err
	}
	
	// Generate filename
	timestamp := time.Now().Format("20060102_150405")
	extension := getFileExtension(req.Format)
	filename := fmt.Sprintf("evaluations_%s.%s", timestamp, extension)
	
	return data, filename, nil
}

// HandleExportStatus returns available export formats and data types
func (h *ExportHandler) HandleExportStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"formats": []string{"csv", "json"},
		"data_types": []string{"analytics", "traces", "evaluations"},
		"max_export_size": "100MB",
		"features": gin.H{
			"csv": gin.H{
				"description": "Comma-separated values format",
				"compression": false,
				"excel_compatible": true,
			},
			"json": gin.H{
				"description": "JavaScript Object Notation format",
				"compression": false,
				"structured": true,
			},
		},
	})
}

// HandleScheduledExport creates a scheduled export job
func (h *ExportHandler) HandleScheduledExport(c *gin.Context) {
	var req struct {
		ExportRequest
		Schedule string `json:"schedule" binding:"required"` // cron expression
		Email    string `json:"email" binding:"required,email"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	userID, _ := c.Get("user_id")
	
	// Create scheduled export record
	configJSON, _ := json.Marshal(req.ExportRequest)
	_, err := h.db.Exec(
		"INSERT INTO scheduled_exports (user_id, project_id, export_config, schedule, email, active) VALUES ($1, $2, $3, $4, $5, true)",
		userID, req.ProjectID, string(configJSON), req.Schedule, req.Email,
	)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create scheduled export"})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"message": "Scheduled export created successfully",
		"schedule": req.Schedule,
		"email": req.Email,
	})
}

func getContentType(format ExportFormat) string {
	switch format {
	case FormatCSV:
		return "text/csv"
	case FormatJSON:
		return "application/json"
	default:
		return "application/octet-stream"
	}
}

func getFileExtension(format ExportFormat) string {
	switch format {
	case FormatCSV:
		return "csv"
	case FormatJSON:
		return "json"
	default:
		return "dat"
	}
}