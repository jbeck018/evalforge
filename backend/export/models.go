package export

import "time"

// AnalyticsRow represents a row of analytics data for export
type AnalyticsRow struct {
	Timestamp     time.Time `json:"timestamp"`
	ProjectID     int       `json:"project_id"`
	EventType     string    `json:"event_type"`
	Model         string    `json:"model"`
	Provider      string    `json:"provider"`
	LatencyMS     float64   `json:"latency_ms"`
	TokensInput   int       `json:"tokens_input"`
	TokensOutput  int       `json:"tokens_output"`
	Cost          float64   `json:"cost"`
	ErrorRate     float64   `json:"error_rate"`
	SuccessRate   float64   `json:"success_rate"`
	P50Latency    float64   `json:"p50_latency"`
	P95Latency    float64   `json:"p95_latency"`
	P99Latency    float64   `json:"p99_latency"`
}

// TraceExport represents trace data for export
type TraceExport struct {
	TraceID       string    `json:"trace_id"`
	SpanID        string    `json:"span_id"`
	ParentSpanID  string    `json:"parent_span_id"`
	ProjectID     int       `json:"project_id"`
	Timestamp     time.Time `json:"timestamp"`
	DurationMS    float64   `json:"duration_ms"`
	Name          string    `json:"name"`
	Type          string    `json:"type"`
	Status        string    `json:"status"`
	Model         string    `json:"model"`
	Provider      string    `json:"provider"`
	TokensInput   int       `json:"tokens_input"`
	TokensOutput  int       `json:"tokens_output"`
	Cost          float64   `json:"cost"`
	ErrorMessage  string    `json:"error_message"`
}

// EvaluationExport represents evaluation data for export
type EvaluationExport struct {
	EvaluationID  int       `json:"evaluation_id"`
	ProjectID     int       `json:"project_id"`
	Name          string    `json:"name"`
	Timestamp     time.Time `json:"timestamp"`
	Status        string    `json:"status"`
	TotalSamples  int       `json:"total_samples"`
	PassedSamples int       `json:"passed_samples"`
	FailedSamples int       `json:"failed_samples"`
	SuccessRate   float64   `json:"success_rate"`
	AvgLatency    float64   `json:"avg_latency"`
	AvgCost       float64   `json:"avg_cost"`
	Model         string    `json:"model"`
	Provider      string    `json:"provider"`
	Metrics       string    `json:"metrics"` // JSON string
}

// ExportRequest represents an export request
type ExportRequest struct {
	Format    ExportFormat `json:"format" binding:"required,oneof=csv json"`
	DataType  string       `json:"data_type" binding:"required,oneof=analytics traces evaluations"`
	StartDate *time.Time   `json:"start_date"`
	EndDate   *time.Time   `json:"end_date"`
	ProjectID *int         `json:"project_id"`
	Filters   map[string]interface{} `json:"filters"`
}