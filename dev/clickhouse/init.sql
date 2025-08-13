-- Create database if not exists
CREATE DATABASE IF NOT EXISTS evalforge;

-- Create events table
CREATE TABLE IF NOT EXISTS evalforge.events
(
    id UUID DEFAULT generateUUIDv4(),
    project_id UUID,
    trace_id UUID,
    span_id UUID,
    parent_span_id UUID,
    type String,
    name String,
    timestamp DateTime64(3),
    duration_ms UInt32,
    input String,
    output String,
    metadata String,
    metrics String,
    error String,
    created_at DateTime DEFAULT now()
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (project_id, timestamp, trace_id)
SETTINGS index_granularity = 8192;

-- Create metrics table for aggregated data
CREATE TABLE IF NOT EXISTS evalforge.metrics
(
    project_id UUID,
    metric_name String,
    metric_value Float64,
    timestamp DateTime,
    dimensions String
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (project_id, metric_name, timestamp)
SETTINGS index_granularity = 8192;