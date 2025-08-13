-- Performance Optimization Indexes for EvalForge
-- Run this after initial schema setup

-- ============================================
-- TRACE EVENTS INDEXES
-- ============================================

-- Composite index for project filtering with time range
CREATE INDEX IF NOT EXISTS idx_trace_events_project_time 
ON trace_events(project_id, created_at DESC);

-- Index for trace_id lookups
CREATE INDEX IF NOT EXISTS idx_trace_events_trace_id 
ON trace_events(trace_id);

-- Index for operation type filtering
CREATE INDEX IF NOT EXISTS idx_trace_events_operation_type 
ON trace_events(operation_type, created_at DESC);

-- Index for status filtering
CREATE INDEX IF NOT EXISTS idx_trace_events_status 
ON trace_events(status) 
WHERE status != 'success'; -- Partial index for errors

-- Index for model/provider analytics
CREATE INDEX IF NOT EXISTS idx_trace_events_provider_model 
ON trace_events(provider, model);

-- Index for cost analytics
CREATE INDEX IF NOT EXISTS idx_trace_events_cost 
ON trace_events(project_id, cost) 
WHERE cost > 0;

-- ============================================
-- EVALUATIONS INDEXES
-- ============================================

-- Index for project evaluations listing
CREATE INDEX IF NOT EXISTS idx_evaluations_project_status 
ON evaluations(project_id, status, created_at DESC);

-- Index for running evaluations
CREATE INDEX IF NOT EXISTS idx_evaluations_running 
ON evaluations(status) 
WHERE status IN ('pending', 'running');

-- ============================================
-- TEST CASES INDEXES
-- ============================================

-- Index for evaluation test cases
CREATE INDEX IF NOT EXISTS idx_test_cases_evaluation 
ON test_cases(evaluation_id, status);

-- Index for failed test cases
CREATE INDEX IF NOT EXISTS idx_test_cases_failed 
ON test_cases(evaluation_id) 
WHERE status = 'failed';

-- ============================================
-- PROJECTS INDEXES
-- ============================================

-- Index for user projects
CREATE INDEX IF NOT EXISTS idx_projects_user 
ON projects(user_id, created_at DESC);

-- Index for active projects
CREATE INDEX IF NOT EXISTS idx_projects_active 
ON projects(created_at DESC);

-- ============================================
-- OPTIMIZATION SUGGESTIONS INDEXES
-- ============================================

-- Index for pending suggestions
CREATE INDEX IF NOT EXISTS idx_suggestions_pending 
ON optimization_suggestions(evaluation_id, status) 
WHERE status = 'pending';

-- Index for high priority suggestions
CREATE INDEX IF NOT EXISTS idx_suggestions_priority 
ON optimization_suggestions(evaluation_id, priority, expected_impact DESC);

-- ============================================
-- PROMPT ANALYSES INDEXES
-- ============================================

-- Index for task type analytics
CREATE INDEX IF NOT EXISTS idx_prompt_analyses_task_type 
ON prompt_analyses(task_type, confidence DESC);

-- ============================================
-- STATISTICS AND MAINTENANCE
-- ============================================

-- Update table statistics for query planner
ANALYZE trace_events;
ANALYZE evaluations;
ANALYZE test_cases;
ANALYZE projects;
ANALYZE optimization_suggestions;
ANALYZE prompt_analyses;

-- Create extension for query monitoring if not exists
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

-- ============================================
-- QUERY PERFORMANCE VIEWS
-- ============================================

-- View for slow queries monitoring
CREATE OR REPLACE VIEW v_slow_queries AS
SELECT 
    query,
    calls,
    total_exec_time AS total_time,
    mean_exec_time AS mean,
    max_exec_time AS max_time,
    rows
FROM pg_stat_statements
WHERE mean_exec_time > 100 -- queries taking more than 100ms on average
ORDER BY mean_exec_time DESC
LIMIT 20;

-- View for project analytics summary
CREATE OR REPLACE VIEW v_project_analytics AS
SELECT 
    p.id AS project_id,
    p.name AS project_name,
    COUNT(DISTINCT te.id) AS total_events,
    COUNT(DISTINCT te.trace_id) AS total_traces,
    SUM(COALESCE((te.tokens->>'total')::int, 0)) AS total_tokens,
    SUM(te.cost) AS total_cost,
    AVG(te.duration_ms) AS avg_latency_ms,
    SUM(CASE WHEN te.status = 'error' THEN 1 ELSE 0 END)::FLOAT / 
        NULLIF(COUNT(te.id), 0) AS error_rate,
    MAX(te.created_at) AS last_event_at
FROM projects p
LEFT JOIN trace_events te ON p.id = te.project_id
GROUP BY p.id, p.name;

-- View for evaluation performance metrics
CREATE OR REPLACE VIEW v_evaluation_metrics AS
SELECT 
    e.id AS evaluation_id,
    e.project_id,
    e.name AS evaluation_name,
    e.status,
    e.progress,
    COUNT(DISTINCT tc.id) AS total_test_cases,
    SUM(CASE WHEN tc.status = 'passed' THEN 1 ELSE 0 END) AS passed_tests,
    AVG(tc.score) AS avg_score,
    e.created_at,
    e.completed_at,
    EXTRACT(EPOCH FROM (e.completed_at - e.started_at)) AS duration_seconds
FROM evaluations e
LEFT JOIN test_cases tc ON e.id = tc.evaluation_id
GROUP BY e.id;

-- ============================================
-- PERFORMANCE CONFIGURATION
-- ============================================

-- Set appropriate work_mem for sorting operations
-- ALTER SYSTEM SET work_mem = '16MB';

-- Set effective_cache_size based on available RAM
-- ALTER SYSTEM SET effective_cache_size = '2GB';

-- Enable parallel query execution
-- ALTER SYSTEM SET max_parallel_workers_per_gather = 2;

-- Optimize for SSD storage
-- ALTER SYSTEM SET random_page_cost = 1.1;

-- RELOAD CONFIGURATION
-- SELECT pg_reload_conf();