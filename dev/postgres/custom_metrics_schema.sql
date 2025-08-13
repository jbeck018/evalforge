-- Custom Metrics Schema for EvalForge

-- ============================================
-- CUSTOM METRICS TABLE
-- ============================================

CREATE TABLE IF NOT EXISTS custom_metrics (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL, -- numeric, boolean, string, percentage, score, custom
    aggregation VARCHAR(50) DEFAULT 'average', -- average, sum, min, max, median, p95, p99, count
    formula TEXT, -- Custom formula for calculated metrics
    thresholds JSONB NOT NULL DEFAULT '{}', -- Pass/fail thresholds
    weight DECIMAL(3,2) DEFAULT 1.0, -- Weight for composite scores
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(project_id, name)
);

-- ============================================
-- METRIC VALUES TABLE
-- ============================================

CREATE TABLE IF NOT EXISTS metric_values (
    id SERIAL PRIMARY KEY,
    metric_id INTEGER NOT NULL REFERENCES custom_metrics(id) ON DELETE CASCADE,
    evaluation_id INTEGER NOT NULL REFERENCES evaluations(id) ON DELETE CASCADE,
    sample_id VARCHAR(255),
    value JSONB NOT NULL,
    passed BOOLEAN DEFAULT FALSE,
    metadata JSONB DEFAULT '{}',
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_metric_values_metric ON metric_values(metric_id);
CREATE INDEX IF NOT EXISTS idx_metric_values_evaluation ON metric_values(evaluation_id);
CREATE INDEX IF NOT EXISTS idx_metric_values_sample ON metric_values(sample_id);

-- ============================================
-- METRIC RESULTS TABLE
-- ============================================

CREATE TABLE IF NOT EXISTS metric_results (
    id SERIAL PRIMARY KEY,
    evaluation_id INTEGER NOT NULL REFERENCES evaluations(id) ON DELETE CASCADE,
    metric_id INTEGER NOT NULL REFERENCES custom_metrics(id) ON DELETE CASCADE,
    aggregated_value DECIMAL(10,4),
    passed BOOLEAN DEFAULT FALSE,
    pass_rate DECIMAL(5,4),
    sample_count INTEGER DEFAULT 0,
    details JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(evaluation_id, metric_id)
);

-- ============================================
-- METRIC TEMPLATES TABLE
-- ============================================

CREATE TABLE IF NOT EXISTS metric_templates (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    category VARCHAR(100),
    description TEXT,
    type VARCHAR(50) NOT NULL,
    aggregation VARCHAR(50) DEFAULT 'average',
    formula TEXT,
    default_thresholds JSONB DEFAULT '{}',
    is_public BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================
-- COMPOSITE METRICS TABLE
-- ============================================

CREATE TABLE IF NOT EXISTS composite_metrics (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    metric_ids INTEGER[] NOT NULL, -- Array of component metric IDs
    weights DECIMAL(3,2)[] NOT NULL, -- Weights for each component
    aggregation_method VARCHAR(50) DEFAULT 'weighted_average',
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(project_id, name)
);

-- ============================================
-- INDEXES
-- ============================================

CREATE INDEX IF NOT EXISTS idx_custom_metrics_project 
ON custom_metrics(project_id);

CREATE INDEX IF NOT EXISTS idx_custom_metrics_enabled 
ON custom_metrics(enabled) WHERE enabled = TRUE;

CREATE INDEX IF NOT EXISTS idx_metric_results_evaluation 
ON metric_results(evaluation_id);

CREATE INDEX IF NOT EXISTS idx_composite_metrics_project 
ON composite_metrics(project_id);

-- ============================================
-- VIEWS
-- ============================================

-- View for metric performance over time
CREATE OR REPLACE VIEW v_metric_performance AS
SELECT 
    cm.id as metric_id,
    cm.project_id,
    cm.name as metric_name,
    cm.type as metric_type,
    e.name as evaluation_name,
    mr.aggregated_value,
    mr.passed,
    mr.pass_rate,
    mr.sample_count,
    mr.created_at
FROM metric_results mr
JOIN custom_metrics cm ON mr.metric_id = cm.id
JOIN evaluations e ON mr.evaluation_id = e.id
ORDER BY mr.created_at DESC;

-- View for metric statistics
CREATE OR REPLACE VIEW v_metric_statistics AS
SELECT 
    cm.id as metric_id,
    cm.project_id,
    cm.name as metric_name,
    COUNT(mr.id) as evaluation_count,
    AVG(mr.aggregated_value) as avg_value,
    MIN(mr.aggregated_value) as min_value,
    MAX(mr.aggregated_value) as max_value,
    AVG(mr.pass_rate) as avg_pass_rate,
    SUM(mr.sample_count) as total_samples
FROM custom_metrics cm
LEFT JOIN metric_results mr ON cm.id = mr.metric_id
WHERE cm.enabled = TRUE
GROUP BY cm.id, cm.project_id, cm.name;

-- ============================================
-- FUNCTIONS
-- ============================================

-- Function to calculate composite metric score
CREATE OR REPLACE FUNCTION calculate_composite_score(
    p_composite_id INTEGER,
    p_evaluation_id INTEGER
) RETURNS DECIMAL AS $$
DECLARE
    v_metric_ids INTEGER[];
    v_weights DECIMAL[];
    v_total_score DECIMAL := 0;
    v_total_weight DECIMAL := 0;
    v_metric_value DECIMAL;
    i INTEGER;
BEGIN
    -- Get composite metric configuration
    SELECT metric_ids, weights INTO v_metric_ids, v_weights
    FROM composite_metrics
    WHERE id = p_composite_id;
    
    -- Calculate weighted score
    FOR i IN 1..array_length(v_metric_ids, 1) LOOP
        SELECT aggregated_value INTO v_metric_value
        FROM metric_results
        WHERE metric_id = v_metric_ids[i] AND evaluation_id = p_evaluation_id;
        
        IF v_metric_value IS NOT NULL THEN
            v_total_score := v_total_score + (v_metric_value * v_weights[i]);
            v_total_weight := v_total_weight + v_weights[i];
        END IF;
    END LOOP;
    
    IF v_total_weight > 0 THEN
        RETURN v_total_score / v_total_weight;
    ELSE
        RETURN NULL;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Function to update metric result statistics
CREATE OR REPLACE FUNCTION update_metric_statistics()
RETURNS TRIGGER AS $$
BEGIN
    -- Update evaluation statistics when metric results change
    UPDATE evaluations
    SET 
        metrics = metrics || jsonb_build_object(
            'custom_metrics_evaluated', (
                SELECT COUNT(*) FROM metric_results 
                WHERE evaluation_id = NEW.evaluation_id
            ),
            'custom_metrics_passed', (
                SELECT COUNT(*) FROM metric_results 
                WHERE evaluation_id = NEW.evaluation_id AND passed = TRUE
            )
        ),
        updated_at = NOW()
    WHERE id = NEW.evaluation_id;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- TRIGGERS
-- ============================================

CREATE TRIGGER trigger_update_metric_statistics
    AFTER INSERT OR UPDATE ON metric_results
    FOR EACH ROW
    EXECUTE FUNCTION update_metric_statistics();

CREATE TRIGGER trigger_update_custom_metrics_timestamp
    BEFORE UPDATE ON custom_metrics
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trigger_update_composite_metrics_timestamp
    BEFORE UPDATE ON composite_metrics
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- SAMPLE METRIC TEMPLATES
-- ============================================

INSERT INTO metric_templates (name, category, description, type, aggregation, default_thresholds) VALUES
(
    'Response Relevance',
    'Quality',
    'Measures how relevant the response is to the input query',
    'score',
    'average',
    '{"pass_value": 0.7, "warning_value": 0.5, "fail_value": 0.3, "operator": ">="}'::jsonb
),
(
    'Token Efficiency',
    'Performance',
    'Ratio of meaningful tokens to total tokens',
    'percentage',
    'average',
    '{"pass_value": 0.8, "warning_value": 0.6, "fail_value": 0.4, "operator": ">="}'::jsonb
),
(
    'Hallucination Detection',
    'Safety',
    'Detects factual inconsistencies in responses',
    'boolean',
    'average',
    '{"pass_value": 0.95, "operator": ">="}'::jsonb
),
(
    'Response Time SLA',
    'Performance',
    'Checks if response time meets SLA requirements',
    'numeric',
    'p95',
    '{"pass_value": 1000, "operator": "<="}'::jsonb
),
(
    'Sentiment Consistency',
    'Quality',
    'Ensures consistent sentiment in responses',
    'score',
    'average',
    '{"pass_value": 0.85, "operator": ">="}'::jsonb
),
(
    'PII Detection',
    'Safety',
    'Detects personally identifiable information in outputs',
    'boolean',
    'sum',
    '{"pass_value": 0, "operator": "=="}'::jsonb
),
(
    'Code Syntax Validity',
    'Quality',
    'Validates generated code syntax',
    'boolean',
    'average',
    '{"pass_value": 1.0, "operator": "=="}'::jsonb
),
(
    'Citation Accuracy',
    'Quality',
    'Verifies accuracy of citations and references',
    'percentage',
    'average',
    '{"pass_value": 0.95, "operator": ">="}'::jsonb
)
ON CONFLICT (name) DO NOTHING;