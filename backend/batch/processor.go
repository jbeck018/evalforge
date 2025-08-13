package batch

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
)

// Event represents a trace event
type Event struct {
	ID            string                 `json:"id"`
	ProjectID     int                    `json:"project_id"`
	TraceID       string                 `json:"trace_id"`
	SpanID        string                 `json:"span_id"`
	OperationType string                 `json:"operation_type"`
	StartTime     time.Time              `json:"start_time"`
	EndTime       time.Time              `json:"end_time"`
	DurationMS    int                    `json:"duration_ms"`
	Status        string                 `json:"status"`
	Input         map[string]interface{} `json:"input"`
	Output        map[string]interface{} `json:"output"`
	Metadata      map[string]interface{} `json:"metadata"`
	Tokens        map[string]interface{} `json:"tokens"`
	Cost          float64                `json:"cost"`
	Provider      string                 `json:"provider"`
	Model         string                 `json:"model"`
}

// BatchProcessor handles batch processing of events
type BatchProcessor struct {
	postgres      *sql.DB
	clickhouse    clickhouse.Conn
	batchSize     int
	flushInterval time.Duration
	eventChan     chan Event
	stopChan      chan struct{}
	wg            sync.WaitGroup
	buffer        []Event
	bufferMu      sync.Mutex
}

// NewBatchProcessor creates a new batch processor
func NewBatchProcessor(pg *sql.DB, ch clickhouse.Conn, batchSize int, flushInterval time.Duration) *BatchProcessor {
	return &BatchProcessor{
		postgres:      pg,
		clickhouse:    ch,
		batchSize:     batchSize,
		flushInterval: flushInterval,
		eventChan:     make(chan Event, batchSize*2),
		stopChan:      make(chan struct{}),
		buffer:        make([]Event, 0, batchSize),
	}
}

// Start begins the batch processing
func (bp *BatchProcessor) Start(ctx context.Context) {
	bp.wg.Add(2)
	
	// Start the processor goroutine
	go bp.processEvents(ctx)
	
	// Start the flush timer goroutine
	go bp.flushTimer(ctx)
}

// Stop gracefully stops the batch processor
func (bp *BatchProcessor) Stop() {
	close(bp.stopChan)
	bp.wg.Wait()
	close(bp.eventChan)
	
	// Flush any remaining events
	bp.flush(context.Background())
}

// AddEvent adds an event to the batch queue
func (bp *BatchProcessor) AddEvent(event Event) error {
	select {
	case bp.eventChan <- event:
		return nil
	default:
		return fmt.Errorf("event queue is full")
	}
}

// processEvents processes events from the channel
func (bp *BatchProcessor) processEvents(ctx context.Context) {
	defer bp.wg.Done()
	
	for {
		select {
		case <-bp.stopChan:
			return
		case event := <-bp.eventChan:
			bp.bufferMu.Lock()
			bp.buffer = append(bp.buffer, event)
			
			// Check if buffer is full
			if len(bp.buffer) >= bp.batchSize {
				// Make a copy of the buffer to process
				batch := make([]Event, len(bp.buffer))
				copy(batch, bp.buffer)
				bp.buffer = bp.buffer[:0]
				bp.bufferMu.Unlock()
				
				// Process the batch
				go bp.processBatch(ctx, batch)
			} else {
				bp.bufferMu.Unlock()
			}
		}
	}
}

// flushTimer periodically flushes the buffer
func (bp *BatchProcessor) flushTimer(ctx context.Context) {
	defer bp.wg.Done()
	
	ticker := time.NewTicker(bp.flushInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-bp.stopChan:
			return
		case <-ticker.C:
			bp.flush(ctx)
		}
	}
}

// flush processes any events in the buffer
func (bp *BatchProcessor) flush(ctx context.Context) {
	bp.bufferMu.Lock()
	if len(bp.buffer) == 0 {
		bp.bufferMu.Unlock()
		return
	}
	
	// Make a copy of the buffer to process
	batch := make([]Event, len(bp.buffer))
	copy(batch, bp.buffer)
	bp.buffer = bp.buffer[:0]
	bp.bufferMu.Unlock()
	
	// Process the batch
	bp.processBatch(ctx, batch)
}

// processBatch inserts a batch of events into the database
func (bp *BatchProcessor) processBatch(ctx context.Context, events []Event) {
	if len(events) == 0 {
		return
	}
	
	start := time.Now()
	
	// Insert into PostgreSQL
	if err := bp.insertPostgresBatch(ctx, events); err != nil {
		log.Printf("Error inserting batch to PostgreSQL: %v", err)
	}
	
	// Insert into ClickHouse if available
	if bp.clickhouse != nil {
		if err := bp.insertClickHouseBatch(ctx, events); err != nil {
			log.Printf("Error inserting batch to ClickHouse: %v", err)
		}
	}
	
	duration := time.Since(start)
	log.Printf("Processed batch of %d events in %v", len(events), duration)
}

// insertPostgresBatch inserts events into PostgreSQL
func (bp *BatchProcessor) insertPostgresBatch(ctx context.Context, events []Event) error {
	tx, err := bp.postgres.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO trace_events (
			id, project_id, trace_id, span_id, operation_type,
			start_time, end_time, duration_ms, status,
			input, output, metadata, tokens,
			cost, provider, model, created_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12, $13,
			$14, $15, $16, NOW()
		) ON CONFLICT (id) DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()
	
	for _, event := range events {
		inputJSON, _ := json.Marshal(event.Input)
		outputJSON, _ := json.Marshal(event.Output)
		metadataJSON, _ := json.Marshal(event.Metadata)
		tokensJSON, _ := json.Marshal(event.Tokens)
		
		_, err = stmt.ExecContext(ctx,
			event.ID, event.ProjectID, event.TraceID, event.SpanID, event.OperationType,
			event.StartTime, event.EndTime, event.DurationMS, event.Status,
			string(inputJSON), string(outputJSON), string(metadataJSON), string(tokensJSON),
			event.Cost, event.Provider, event.Model,
		)
		if err != nil {
			log.Printf("Error inserting event %s: %v", event.ID, err)
		}
	}
	
	return tx.Commit()
}

// insertClickHouseBatch inserts events into ClickHouse
func (bp *BatchProcessor) insertClickHouseBatch(ctx context.Context, events []Event) error {
	batch, err := bp.clickhouse.PrepareBatch(ctx, `
		INSERT INTO trace_events (
			id, project_id, trace_id, span_id, operation_type,
			start_time, end_time, duration_ms, status,
			input, output, metadata, tokens,
			cost, provider, model, created_at
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare batch: %w", err)
	}
	
	for _, event := range events {
		inputJSON, _ := json.Marshal(event.Input)
		outputJSON, _ := json.Marshal(event.Output)
		metadataJSON, _ := json.Marshal(event.Metadata)
		tokensJSON, _ := json.Marshal(event.Tokens)
		
		err = batch.Append(
			event.ID, event.ProjectID, event.TraceID, event.SpanID, event.OperationType,
			event.StartTime, event.EndTime, event.DurationMS, event.Status,
			string(inputJSON), string(outputJSON), string(metadataJSON), string(tokensJSON),
			event.Cost, event.Provider, event.Model, time.Now(),
		)
		if err != nil {
			log.Printf("Error appending event %s to batch: %v", event.ID, err)
		}
	}
	
	return batch.Send()
}

// GetStats returns statistics about the batch processor
func (bp *BatchProcessor) GetStats() map[string]interface{} {
	bp.bufferMu.Lock()
	bufferSize := len(bp.buffer)
	bp.bufferMu.Unlock()
	
	return map[string]interface{}{
		"buffer_size":    bufferSize,
		"queue_size":     len(bp.eventChan),
		"queue_capacity": cap(bp.eventChan),
		"batch_size":     bp.batchSize,
		"flush_interval": bp.flushInterval.String(),
	}
}