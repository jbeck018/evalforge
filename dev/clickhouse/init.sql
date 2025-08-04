-- ClickHouse initialization script for EvalForge
-- This script sets up the OLAP database for event storage and analytics

-- Create database if not exists
CREATE DATABASE IF NOT EXISTS evalforge;
USE evalforge;

-- Main events table for all trace data
CREATE TABLE IF NOT EXISTS events (
    timestamp DateTime64(3) CODEC(Delta, ZSTD(3)),
    project_id String CODEC(ZSTD(3)),
    trace_id String CODEC(ZSTD(3)),
    span_id String CODEC(ZSTD(3)),
    parent_span_id String CODEC(ZSTD(3)),
    
    -- Event metadata
    event_type LowCardinality(String),
    operation_name LowCardinality(String),
    status LowCardinality(String), -- success, error, timeout
    
    -- LLM specific fields
    model LowCardinality(String),
    provider LowCardinality(String), -- openai, anthropic, cohere, etc.
    
    -- Performance metrics
    duration_ms UInt32 CODEC(ZSTD(3)),
    tokens_used UInt32 CODEC(ZSTD(3)),
    input_tokens UInt32 CODEC(ZSTD(3)),
    output_tokens UInt32 CODEC(ZSTD(3)),
    
    -- Cost tracking
    cost_cents UInt32 CODEC(ZSTD(3)),
    input_cost_cents UInt32 CODEC(ZSTD(3)),
    output_cost_cents UInt32 CODEC(ZSTD(3)),
    
    -- Quality metrics (from evaluations)
    relevance_score Float32 CODEC(ZSTD(3)),
    accuracy_score Float32 CODEC(ZSTD(3)),
    safety_score Float32 CODEC(ZSTD(3)),
    overall_score Float32 CODEC(ZSTD(3)),
    
    -- Request/Response data (compressed)
    prompt String CODEC(ZSTD(9)),
    response String CODEC(ZSTD(9)),
    
    -- Error information
    error_type LowCardinality(String),
    error_message String CODEC(ZSTD(3)),
    
    -- Additional metadata as JSON
    metadata String CODEC(ZSTD(3)),
    
    -- User context
    user_id String CODEC(ZSTD(3)),
    session_id String CODEC(ZSTD(3)),
    
    -- Geographical and environment info
    region LowCardinality(String),
    environment LowCardinality(String) -- production, staging, development
) ENGINE = MergeTree()
PARTITION BY (toYYYYMM(timestamp), cityHash64(project_id) % 16)
ORDER BY (project_id, timestamp, trace_id, span_id)
TTL timestamp + INTERVAL 90 DAY
SETTINGS index_granularity = 8192;

-- Materialized view for 1-minute aggregates
CREATE TABLE IF NOT EXISTS events_1min (
    minute DateTime CODEC(Delta, ZSTD(1)),
    project_id String CODEC(ZSTD(1)),
    model LowCardinality(String),
    provider LowCardinality(String),
    status LowCardinality(String),
    
    -- Counts
    event_count UInt64 CODEC(ZSTD(1)),
    error_count UInt64 CODEC(ZSTD(1)),
    
    -- Token statistics
    total_tokens UInt64 CODEC(ZSTD(1)),
    total_input_tokens UInt64 CODEC(ZSTD(1)),
    total_output_tokens UInt64 CODEC(ZSTD(1)),
    
    -- Cost statistics
    total_cost_cents UInt64 CODEC(ZSTD(1)),
    
    -- Duration percentiles (quantile states for merging)
    duration_quantiles AggregateFunction(quantiles(0.5, 0.9, 0.95, 0.99), UInt32) CODEC(ZSTD(3)),
    
    -- Quality score averages
    avg_relevance_score Float64 CODEC(ZSTD(1)),
    avg_accuracy_score Float64 CODEC(ZSTD(1)),
    avg_safety_score Float64 CODEC(ZSTD(1)),
    avg_overall_score Float64 CODEC(ZSTD(1))
) ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(minute)
ORDER BY (project_id, minute, model, provider, status)
TTL minute + INTERVAL 365 DAY
SETTINGS index_granularity = 8192;

-- Materialized view to populate 1-minute aggregates
CREATE MATERIALIZED VIEW IF NOT EXISTS events_1min_mv TO events_1min AS
SELECT
    toStartOfMinute(timestamp) as minute,
    project_id,
    model,
    provider,
    status,
    
    count() as event_count,
    countIf(status = 'error') as error_count,
    
    sum(tokens_used) as total_tokens,
    sum(input_tokens) as total_input_tokens,
    sum(output_tokens) as total_output_tokens,
    
    sum(cost_cents) as total_cost_cents,
    
    quantilesState(0.5, 0.9, 0.95, 0.99)(duration_ms) as duration_quantiles,
    
    avg(relevance_score) as avg_relevance_score,
    avg(accuracy_score) as avg_accuracy_score,
    avg(safety_score) as avg_safety_score,
    avg(overall_score) as avg_overall_score
FROM events
WHERE timestamp >= toStartOfMinute(now() - INTERVAL 1 HOUR)
GROUP BY minute, project_id, model, provider, status;

-- Hourly aggregates table
CREATE TABLE IF NOT EXISTS events_1hour (
    hour DateTime CODEC(Delta, ZSTD(1)),
    project_id String CODEC(ZSTD(1)),
    model LowCardinality(String),
    provider LowCardinality(String),
    
    -- Aggregated metrics
    event_count UInt64 CODEC(ZSTD(1)),
    error_count UInt64 CODEC(ZSTD(1)),
    error_rate Float64 CODEC(ZSTD(1)),
    
    total_tokens UInt64 CODEC(ZSTD(1)),
    total_cost_cents UInt64 CODEC(ZSTD(1)),
    
    -- Performance metrics
    p50_duration_ms Float64 CODEC(ZSTD(1)),
    p90_duration_ms Float64 CODEC(ZSTD(1)),
    p95_duration_ms Float64 CODEC(ZSTD(1)),
    p99_duration_ms Float64 CODEC(ZSTD(1)),
    
    -- Quality metrics
    avg_relevance_score Float64 CODEC(ZSTD(1)),
    avg_accuracy_score Float64 CODEC(ZSTD(1)),
    avg_safety_score Float64 CODEC(ZSTD(1)),
    avg_overall_score Float64 CODEC(ZSTD(1))
) ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(hour)
ORDER BY (project_id, hour, model, provider)
TTL hour + INTERVAL 730 DAY -- 2 years
SETTINGS index_granularity = 8192;

-- Materialized view for hourly aggregates from minute data
CREATE MATERIALIZED VIEW IF NOT EXISTS events_1hour_mv TO events_1hour AS
SELECT
    toStartOfHour(minute) as hour,
    project_id,
    model,
    provider,
    
    sum(event_count) as event_count,
    sum(error_count) as error_count,
    if(sum(event_count) > 0, sum(error_count) / sum(event_count), 0) as error_rate,
    
    sum(total_tokens) as total_tokens,
    sum(total_cost_cents) as total_cost_cents,
    
    -- Extract percentiles from quantile states
    quantilesMerge(0.5)(duration_quantiles)[1] as p50_duration_ms,
    quantilesMerge(0.9)(duration_quantiles)[1] as p90_duration_ms,
    quantilesMerge(0.95)(duration_quantiles)[1] as p95_duration_ms,
    quantilesMerge(0.99)(duration_quantiles)[1] as p99_duration_ms,
    
    avg(avg_relevance_score) as avg_relevance_score,
    avg(avg_accuracy_score) as avg_accuracy_score,
    avg(avg_safety_score) as avg_safety_score,
    avg(avg_overall_score) as avg_overall_score
FROM events_1min
GROUP BY hour, project_id, model, provider;

-- Table for storing evaluation results
CREATE TABLE IF NOT EXISTS evaluation_results (
    timestamp DateTime64(3) CODEC(Delta, ZSTD(3)),
    evaluation_run_id String CODEC(ZSTD(3)),
    project_id String CODEC(ZSTD(3)),
    prompt_version_id String CODEC(ZSTD(3)),
    
    -- Test case information
    test_case_id String CODEC(ZSTD(3)),
    input_text String CODEC(ZSTD(9)),
    expected_output String CODEC(ZSTD(9)),
    actual_output String CODEC(ZSTD(9)),
    
    -- Scores
    relevance_score Float32,
    accuracy_score Float32,
    safety_score Float32,
    fluency_score Float32,
    coherence_score Float32,
    overall_score Float32,
    
    -- Performance metrics
    duration_ms UInt32,
    tokens_used UInt32,
    cost_cents UInt32,
    
    -- Metadata
    model LowCardinality(String),
    evaluator_model LowCardinality(String),
    evaluation_criteria String CODEC(ZSTD(3))
) ENGINE = MergeTree()
PARTITION BY (toYYYYMM(timestamp), cityHash64(project_id) % 8)
ORDER BY (project_id, timestamp, evaluation_run_id, test_case_id)
TTL timestamp + INTERVAL 365 DAY
SETTINGS index_granularity = 8192;

-- User sessions table for analytics
CREATE TABLE IF NOT EXISTS user_sessions (
    session_start DateTime64(3) CODEC(Delta, ZSTD(3)),
    session_id String CODEC(ZSTD(3)),
    user_id String CODEC(ZSTD(3)),
    project_id String CODEC(ZSTD(3)),
    
    -- Session metrics
    session_duration_ms UInt32,
    total_requests UInt32,
    total_tokens UInt32,
    total_cost_cents UInt32,
    
    -- User agent and environment
    user_agent String CODEC(ZSTD(3)),
    ip_address String CODEC(ZSTD(3)),
    region LowCardinality(String),
    
    -- Session quality
    avg_response_time_ms Float32,
    error_rate Float32,
    avg_quality_score Float32
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(session_start)
ORDER BY (project_id, session_start, user_id)
TTL session_start + INTERVAL 90 DAY
SETTINGS index_granularity = 8192;

-- Create useful views for common queries
CREATE VIEW IF NOT EXISTS recent_events AS
SELECT *
FROM events
WHERE timestamp >= now() - INTERVAL 24 HOUR
ORDER BY timestamp DESC;

CREATE VIEW IF NOT EXISTS project_daily_stats AS
SELECT
    toDate(hour) as date,
    project_id,
    sum(event_count) as daily_events,
    sum(error_count) as daily_errors,
    avg(error_rate) as avg_error_rate,
    sum(total_tokens) as daily_tokens,
    sum(total_cost_cents) as daily_cost_cents,
    avg(p95_duration_ms) as avg_p95_duration_ms,
    avg(avg_overall_score) as avg_quality_score
FROM events_1hour
WHERE hour >= today() - INTERVAL 30 DAY
GROUP BY date, project_id
ORDER BY date DESC, project_id;

CREATE VIEW IF NOT EXISTS model_performance_comparison AS
SELECT
    model,
    provider,
    count() as total_requests,
    avg(duration_ms) as avg_duration_ms,
    quantile(0.95)(duration_ms) as p95_duration_ms,
    avg(cost_cents) as avg_cost_cents,
    avg(overall_score) as avg_quality_score,
    countIf(status = 'error') / count() as error_rate
FROM events
WHERE timestamp >= now() - INTERVAL 7 DAY
GROUP BY model, provider
ORDER BY avg_quality_score DESC, p95_duration_ms ASC;

-- Create indexes for common query patterns
-- Note: ClickHouse doesn't use traditional indexes, but we can create skip indexes
ALTER TABLE events ADD INDEX idx_user_id user_id TYPE bloom_filter(0.01) GRANULARITY 1;
ALTER TABLE events ADD INDEX idx_session_id session_id TYPE bloom_filter(0.01) GRANULARITY 1;
ALTER TABLE events ADD INDEX idx_error_type error_type TYPE set(100) GRANULARITY 1;

-- Insert some sample data for development
INSERT INTO events VALUES
    (now() - INTERVAL 1 HOUR, 'p1p1p1p1-p1p1-p1p1-p1p1-p1p1p1p1p1p1', 'trace_001', 'span_001', '', 'llm_request', 'chat_completion', 'success', 'gpt-4', 'openai', 523, 150, 100, 50, 45, 30, 15, 0.85, 0.90, 0.95, 0.88, 'How do I reset my password?', 'To reset your password, please visit our password reset page and follow the instructions.', '', '', '{}', 'user_123', 'session_456', 'us-east-1', 'production'),
    (now() - INTERVAL 2 HOUR, 'p1p1p1p1-p1p1-p1p1-p1p1-p1p1p1p1p1p1', 'trace_002', 'span_002', '', 'llm_request', 'chat_completion', 'success', 'gpt-3.5-turbo', 'openai', 298, 120, 80, 40, 24, 16, 8, 0.82, 0.88, 0.91, 0.85, 'What are your business hours?', 'Our business hours are Monday through Friday, 9 AM to 5 PM EST.', '', '', '{}', 'user_124', 'session_457', 'us-west-2', 'production'),
    (now() - INTERVAL 3 HOUR, 'p2p2p2p2-p2p2-p2p2-p2p2-p2p2p2p2p2p2', 'trace_003', 'span_003', '', 'llm_request', 'code_review', 'success', 'claude-3-opus', 'anthropic', 1205, 280, 200, 80, 89, 67, 22, 0.92, 0.89, 0.96, 0.91, 'function validateEmail(email) { return /^[^@]+@[^@]+\.[^@]+$/.test(email); }', 'The email validation regex is too simplistic. Consider using a more robust validation library or a more comprehensive regex pattern.', '', '', '{}', 'user_125', 'session_458', 'us-east-1', 'production');

-- Create a function to generate realistic test data
-- Note: ClickHouse doesn't support stored procedures like PostgreSQL, but we can create this as a template

-- Optimization settings for better performance
OPTIMIZE TABLE events FINAL;
OPTIMIZE TABLE events_1min FINAL;
OPTIMIZE TABLE events_1hour FINAL;

-- Show some useful information
SELECT 'ClickHouse Event Storage Initialized' as status;
SELECT 'Tables created:' as info, count() as table_count FROM system.tables WHERE database = 'evalforge';
SELECT 'Sample events inserted:' as info, count() as event_count FROM events;