package websocket

import (
	"context"
	"database/sql"
	"log"
	"time"
)

type MetricsAggregator struct {
	db     *sql.DB
	hub    *Hub
	ticker *time.Ticker
}

type RealtimeMetrics struct {
	TotalEvents      int64              `json:"totalEvents"`
	EventsPerMinute  float64            `json:"eventsPerMinute"`
	ErrorRate        float64            `json:"errorRate"`
	AvgLatency       float64            `json:"avgLatency"`
	ActiveProjects   int                `json:"activeProjects"`
	RecentEvaluations int               `json:"recentEvaluations"`
	OperationCounts  map[string]int64   `json:"operationCounts"`
	StatusCounts     map[string]int64   `json:"statusCounts"`
	Timestamp        time.Time          `json:"timestamp"`
}

func NewMetricsAggregator(db *sql.DB, hub *Hub) *MetricsAggregator {
	return &MetricsAggregator{
		db:     db,
		hub:    hub,
		ticker: time.NewTicker(5 * time.Second), // Update every 5 seconds
	}
}

func (ma *MetricsAggregator) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				ma.ticker.Stop()
				return
			case <-ma.ticker.C:
				metrics := ma.collectMetrics()
				if metrics != nil {
					ma.hub.BroadcastMetrics(metrics)
				}
			}
		}
	}()
}

func (ma *MetricsAggregator) collectMetrics() *RealtimeMetrics {
	metrics := &RealtimeMetrics{
		OperationCounts: make(map[string]int64),
		StatusCounts:    make(map[string]int64),
		Timestamp:       time.Now(),
	}

	// Get total events count
	var totalEvents sql.NullInt64
	err := ma.db.QueryRow(`
		SELECT COUNT(*) FROM trace_events
	`).Scan(&totalEvents)
	if err != nil {
		log.Printf("Error getting total events: %v", err)
		return nil
	}
	metrics.TotalEvents = totalEvents.Int64

	// Get events per minute (last 5 minutes)
	var eventsLastMinute sql.NullFloat64
	err = ma.db.QueryRow(`
		SELECT COUNT(*)::float / 5.0
		FROM trace_events
		WHERE created_at >= NOW() - INTERVAL '5 minutes'
	`).Scan(&eventsLastMinute)
	if err != nil {
		log.Printf("Error getting events per minute: %v", err)
	} else {
		metrics.EventsPerMinute = eventsLastMinute.Float64
	}

	// Get error rate
	var errorRate sql.NullFloat64
	err = ma.db.QueryRow(`
		SELECT 
			CASE 
				WHEN COUNT(*) > 0 THEN 
					(COUNT(CASE WHEN status = 'error' THEN 1 END)::float / COUNT(*)::float) * 100
				ELSE 0
			END
		FROM trace_events
		WHERE created_at >= NOW() - INTERVAL '1 hour'
	`).Scan(&errorRate)
	if err != nil {
		log.Printf("Error getting error rate: %v", err)
	} else {
		metrics.ErrorRate = errorRate.Float64
	}

	// Get average latency
	var avgLatency sql.NullFloat64
	err = ma.db.QueryRow(`
		SELECT AVG(duration_ms)
		FROM trace_events
		WHERE created_at >= NOW() - INTERVAL '1 hour'
		AND duration_ms IS NOT NULL
	`).Scan(&avgLatency)
	if err != nil {
		log.Printf("Error getting average latency: %v", err)
	} else {
		metrics.AvgLatency = avgLatency.Float64
	}

	// Get active projects count
	var activeProjects sql.NullInt64
	err = ma.db.QueryRow(`
		SELECT COUNT(DISTINCT project_id)
		FROM trace_events
		WHERE created_at >= NOW() - INTERVAL '24 hours'
	`).Scan(&activeProjects)
	if err != nil {
		log.Printf("Error getting active projects: %v", err)
	} else {
		metrics.ActiveProjects = int(activeProjects.Int64)
	}

	// Get recent evaluations count
	var recentEvaluations sql.NullInt64
	err = ma.db.QueryRow(`
		SELECT COUNT(*)
		FROM evaluations
		WHERE created_at >= NOW() - INTERVAL '1 hour'
	`).Scan(&recentEvaluations)
	if err != nil {
		log.Printf("Error getting recent evaluations: %v", err)
	} else {
		metrics.RecentEvaluations = int(recentEvaluations.Int64)
	}

	// Get operation counts
	rows, err := ma.db.Query(`
		SELECT operation, COUNT(*) as count
		FROM trace_events
		WHERE created_at >= NOW() - INTERVAL '1 hour'
		GROUP BY operation
		ORDER BY count DESC
		LIMIT 10
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var operation string
			var count int64
			if err := rows.Scan(&operation, &count); err == nil {
				metrics.OperationCounts[operation] = count
			}
		}
	}

	// Get status counts
	rows, err = ma.db.Query(`
		SELECT status, COUNT(*) as count
		FROM trace_events
		WHERE created_at >= NOW() - INTERVAL '1 hour'
		GROUP BY status
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var status string
			var count int64
			if err := rows.Scan(&status, &count); err == nil {
				metrics.StatusCounts[status] = count
			}
		}
	}

	return metrics
}