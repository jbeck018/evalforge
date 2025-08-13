-- Export and Scheduled Jobs Schema for EvalForge

-- ============================================
-- SCHEDULED EXPORTS TABLE
-- ============================================

CREATE TABLE IF NOT EXISTS scheduled_exports (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    project_id INTEGER REFERENCES projects(id),
    export_config JSONB NOT NULL, -- Stores ExportRequest configuration
    schedule VARCHAR(100) NOT NULL, -- Cron expression
    email VARCHAR(255) NOT NULL,
    active BOOLEAN DEFAULT TRUE,
    last_run TIMESTAMP WITH TIME ZONE,
    next_run TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================
-- EXPORT HISTORY TABLE
-- ============================================

CREATE TABLE IF NOT EXISTS export_history (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    project_id INTEGER REFERENCES projects(id),
    export_type VARCHAR(50) NOT NULL,
    format VARCHAR(20) NOT NULL,
    file_size BIGINT,
    row_count INTEGER,
    status VARCHAR(20) DEFAULT 'pending', -- pending, processing, completed, failed
    error_message TEXT,
    file_url TEXT,
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}'
);

-- ============================================
-- EXPORT TEMPLATES TABLE
-- ============================================

CREATE TABLE IF NOT EXISTS export_templates (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    export_config JSONB NOT NULL,
    is_public BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================
-- INDEXES
-- ============================================

CREATE INDEX IF NOT EXISTS idx_scheduled_exports_user 
ON scheduled_exports(user_id);

CREATE INDEX IF NOT EXISTS idx_scheduled_exports_active 
ON scheduled_exports(active) WHERE active = TRUE;

CREATE INDEX IF NOT EXISTS idx_scheduled_exports_next_run 
ON scheduled_exports(next_run) WHERE active = TRUE;

CREATE INDEX IF NOT EXISTS idx_export_history_user 
ON export_history(user_id);

CREATE INDEX IF NOT EXISTS idx_export_history_project 
ON export_history(project_id);

CREATE INDEX IF NOT EXISTS idx_export_history_status 
ON export_history(status);

CREATE INDEX IF NOT EXISTS idx_export_templates_user 
ON export_templates(user_id);

CREATE INDEX IF NOT EXISTS idx_export_templates_public 
ON export_templates(is_public) WHERE is_public = TRUE;

-- ============================================
-- VIEWS
-- ============================================

-- View for export statistics
CREATE OR REPLACE VIEW v_export_statistics AS
SELECT 
    u.email as user_email,
    p.name as project_name,
    COUNT(eh.id) as total_exports,
    SUM(eh.file_size) as total_size_bytes,
    SUM(eh.row_count) as total_rows_exported,
    AVG(EXTRACT(EPOCH FROM (eh.completed_at - eh.started_at))) as avg_export_time_seconds,
    COUNT(CASE WHEN eh.status = 'completed' THEN 1 END) as successful_exports,
    COUNT(CASE WHEN eh.status = 'failed' THEN 1 END) as failed_exports,
    MAX(eh.started_at) as last_export_at
FROM export_history eh
LEFT JOIN users u ON eh.user_id = u.id
LEFT JOIN projects p ON eh.project_id = p.id
GROUP BY u.email, p.name;

-- View for popular export templates
CREATE OR REPLACE VIEW v_popular_export_templates AS
SELECT 
    et.id,
    et.name,
    et.description,
    u.email as created_by,
    et.is_public,
    COUNT(eh.id) as usage_count,
    et.created_at
FROM export_templates et
LEFT JOIN users u ON et.user_id = u.id
LEFT JOIN export_history eh ON eh.metadata->>'template_id' = et.id::text
WHERE et.is_public = TRUE
GROUP BY et.id, et.name, et.description, u.email, et.is_public, et.created_at
ORDER BY usage_count DESC
LIMIT 20;

-- ============================================
-- FUNCTIONS
-- ============================================

-- Function to calculate next run time for scheduled exports
CREATE OR REPLACE FUNCTION calculate_next_run(cron_expr VARCHAR, last_run TIMESTAMP WITH TIME ZONE)
RETURNS TIMESTAMP WITH TIME ZONE AS $$
DECLARE
    next_run TIMESTAMP WITH TIME ZONE;
BEGIN
    -- Simple implementation for common patterns
    -- In production, use pg_cron or similar for proper cron parsing
    IF cron_expr = '0 0 * * *' THEN
        -- Daily at midnight
        next_run := date_trunc('day', COALESCE(last_run, NOW())) + INTERVAL '1 day';
    ELSIF cron_expr = '0 0 * * 0' THEN
        -- Weekly on Sunday
        next_run := date_trunc('week', COALESCE(last_run, NOW())) + INTERVAL '1 week';
    ELSIF cron_expr = '0 0 1 * *' THEN
        -- Monthly on the 1st
        next_run := date_trunc('month', COALESCE(last_run, NOW())) + INTERVAL '1 month';
    ELSE
        -- Default to daily
        next_run := COALESCE(last_run, NOW()) + INTERVAL '1 day';
    END IF;
    
    RETURN next_run;
END;
$$ LANGUAGE plpgsql;

-- Function to update next run time after export
CREATE OR REPLACE FUNCTION update_scheduled_export_next_run()
RETURNS TRIGGER AS $$
BEGIN
    NEW.next_run := calculate_next_run(NEW.schedule, NEW.last_run);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- TRIGGERS
-- ============================================

CREATE TRIGGER trigger_update_next_run
    BEFORE INSERT OR UPDATE OF last_run ON scheduled_exports
    FOR EACH ROW
    EXECUTE FUNCTION update_scheduled_export_next_run();

-- ============================================
-- SAMPLE DATA
-- ============================================

-- Insert sample export templates
INSERT INTO export_templates (user_id, name, description, export_config, is_public) VALUES
(
    1,
    'Weekly Analytics Report',
    'Export weekly analytics data in CSV format',
    '{"format": "csv", "data_type": "analytics", "filters": {"period": "7d"}}'::jsonb,
    TRUE
),
(
    1,
    'Monthly Trace Export',
    'Export all traces for the past month in Parquet format',
    '{"format": "parquet", "data_type": "traces", "filters": {"period": "30d"}}'::jsonb,
    TRUE
),
(
    1,
    'Evaluation Results Export',
    'Export evaluation results with detailed metrics',
    '{"format": "csv", "data_type": "evaluations", "filters": {"include_metrics": true}}'::jsonb,
    TRUE
)
ON CONFLICT DO NOTHING;