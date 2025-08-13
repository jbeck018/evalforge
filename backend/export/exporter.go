package export

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"time"
)

type ExportFormat string

const (
	FormatCSV  ExportFormat = "csv"
	FormatJSON ExportFormat = "json"
)

type Exporter struct {
	format ExportFormat
}

func NewExporter(format ExportFormat) *Exporter {
	return &Exporter{format: format}
}

// ExportAnalytics exports analytics data in the specified format
func (e *Exporter) ExportAnalytics(data []AnalyticsRow) ([]byte, error) {
	switch e.format {
	case FormatCSV:
		return e.exportCSV(data)
	case FormatJSON:
		return e.exportJSON(data)
	default:
		return nil, fmt.Errorf("unsupported format: %s", e.format)
	}
}

func (e *Exporter) exportCSV(data []AnalyticsRow) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	// Write header
	header := []string{
		"timestamp", "project_id", "event_type", "model", "provider",
		"latency_ms", "tokens_input", "tokens_output", "cost",
		"error_rate", "success_rate", "p50_latency", "p95_latency", "p99_latency",
	}
	if err := w.Write(header); err != nil {
		return nil, err
	}

	// Write data rows
	for _, row := range data {
		record := []string{
			row.Timestamp.Format(time.RFC3339),
			fmt.Sprintf("%d", row.ProjectID),
			row.EventType,
			row.Model,
			row.Provider,
			fmt.Sprintf("%.2f", row.LatencyMS),
			fmt.Sprintf("%d", row.TokensInput),
			fmt.Sprintf("%d", row.TokensOutput),
			fmt.Sprintf("%.4f", row.Cost),
			fmt.Sprintf("%.4f", row.ErrorRate),
			fmt.Sprintf("%.4f", row.SuccessRate),
			fmt.Sprintf("%.2f", row.P50Latency),
			fmt.Sprintf("%.2f", row.P95Latency),
			fmt.Sprintf("%.2f", row.P99Latency),
		}
		if err := w.Write(record); err != nil {
			return nil, err
		}
	}

	w.Flush()
	return buf.Bytes(), w.Error()
}

func (e *Exporter) exportJSON(data []AnalyticsRow) ([]byte, error) {
	return json.MarshalIndent(data, "", "  ")
}

// ExportTraces exports trace data
func (e *Exporter) ExportTraces(traces []TraceExport) ([]byte, error) {
	switch e.format {
	case FormatCSV:
		return e.exportTracesCSV(traces)
	case FormatJSON:
		return e.exportTracesJSON(traces)
	default:
		return nil, fmt.Errorf("unsupported format: %s", e.format)
	}
}

func (e *Exporter) exportTracesCSV(traces []TraceExport) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	// Write header
	header := []string{
		"trace_id", "span_id", "parent_span_id", "project_id", "timestamp",
		"duration_ms", "name", "type", "status", "model", "provider",
		"tokens_input", "tokens_output", "cost", "error_message",
	}
	if err := w.Write(header); err != nil {
		return nil, err
	}

	// Write data rows
	for _, trace := range traces {
		record := []string{
			trace.TraceID,
			trace.SpanID,
			trace.ParentSpanID,
			fmt.Sprintf("%d", trace.ProjectID),
			trace.Timestamp.Format(time.RFC3339),
			fmt.Sprintf("%.2f", trace.DurationMS),
			trace.Name,
			trace.Type,
			trace.Status,
			trace.Model,
			trace.Provider,
			fmt.Sprintf("%d", trace.TokensInput),
			fmt.Sprintf("%d", trace.TokensOutput),
			fmt.Sprintf("%.4f", trace.Cost),
			trace.ErrorMessage,
		}
		if err := w.Write(record); err != nil {
			return nil, err
		}
	}

	w.Flush()
	return buf.Bytes(), w.Error()
}

func (e *Exporter) exportTracesJSON(traces []TraceExport) ([]byte, error) {
	return json.MarshalIndent(traces, "", "  ")
}

// ExportEvaluations exports evaluation results
func (e *Exporter) ExportEvaluations(evals []EvaluationExport) ([]byte, error) {
	switch e.format {
	case FormatCSV:
		return e.exportEvaluationsCSV(evals)
	case FormatJSON:
		return e.exportEvaluationsJSON(evals)
	default:
		return nil, fmt.Errorf("unsupported format: %s", e.format)
	}
}

func (e *Exporter) exportEvaluationsCSV(evals []EvaluationExport) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	// Write header
	header := []string{
		"evaluation_id", "project_id", "name", "timestamp", "status",
		"total_samples", "passed_samples", "failed_samples", "success_rate",
		"avg_latency", "avg_cost", "model", "provider", "metrics",
	}
	if err := w.Write(header); err != nil {
		return nil, err
	}

	// Write data rows
	for _, eval := range evals {
		record := []string{
			fmt.Sprintf("%d", eval.EvaluationID),
			fmt.Sprintf("%d", eval.ProjectID),
			eval.Name,
			eval.Timestamp.Format(time.RFC3339),
			eval.Status,
			fmt.Sprintf("%d", eval.TotalSamples),
			fmt.Sprintf("%d", eval.PassedSamples),
			fmt.Sprintf("%d", eval.FailedSamples),
			fmt.Sprintf("%.4f", eval.SuccessRate),
			fmt.Sprintf("%.2f", eval.AvgLatency),
			fmt.Sprintf("%.4f", eval.AvgCost),
			eval.Model,
			eval.Provider,
			eval.Metrics,
		}
		if err := w.Write(record); err != nil {
			return nil, err
		}
	}

	w.Flush()
	return buf.Bytes(), w.Error()
}

func (e *Exporter) exportEvaluationsJSON(evals []EvaluationExport) ([]byte, error) {
	return json.MarshalIndent(evals, "", "  ")
}