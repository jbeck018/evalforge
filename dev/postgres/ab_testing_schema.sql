-- A/B Testing Schema for EvalForge

-- ============================================
-- A/B TESTS TABLE
-- ============================================

CREATE TABLE IF NOT EXISTS ab_tests (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    control_prompt TEXT NOT NULL,
    variant_prompt TEXT NOT NULL,
    traffic_ratio DECIMAL(3,2) DEFAULT 0.50 CHECK (traffic_ratio >= 0 AND traffic_ratio <= 1),
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'running', 'completed', 'cancelled')),
    min_sample_size INTEGER DEFAULT 100,
    control_samples INTEGER DEFAULT 0,
    variant_samples INTEGER DEFAULT 0,
    stat_significant BOOLEAN DEFAULT FALSE,
    confidence_level DECIMAL(3,2) DEFAULT 0.95,
    winner VARCHAR(20) CHECK (winner IN ('control', 'variant', 'none')),
    started_at TIMESTAMP WITH TIME ZONE,
    ended_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================
-- A/B TEST RESULTS TABLE
-- ============================================

CREATE TABLE IF NOT EXISTS ab_test_results (
    id SERIAL PRIMARY KEY,
    ab_test_id INTEGER NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE,
    variant VARCHAR(20) NOT NULL CHECK (variant IN ('control', 'variant')),
    trace_id VARCHAR(255),
    latency_ms INTEGER,
    cost DECIMAL(10,4),
    tokens INTEGER,
    error BOOLEAN DEFAULT FALSE,
    metrics JSONB DEFAULT '{}',
    user_feedback INTEGER CHECK (user_feedback >= 1 AND user_feedback <= 5),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================
-- INDEXES
-- ============================================

CREATE INDEX IF NOT EXISTS idx_ab_tests_project_status 
ON ab_tests(project_id, status);

CREATE INDEX IF NOT EXISTS idx_ab_tests_running 
ON ab_tests(status) WHERE status = 'running';

CREATE INDEX IF NOT EXISTS idx_ab_test_results_test_variant 
ON ab_test_results(ab_test_id, variant);

CREATE INDEX IF NOT EXISTS idx_ab_test_results_created 
ON ab_test_results(created_at DESC);

-- ============================================
-- VIEWS
-- ============================================

-- View for A/B test summary
CREATE OR REPLACE VIEW v_ab_test_summary AS
SELECT 
    t.id,
    t.project_id,
    t.name,
    t.status,
    t.control_samples,
    t.variant_samples,
    t.traffic_ratio,
    t.started_at,
    t.ended_at,
    COALESCE(
        (SELECT AVG(latency_ms) FROM ab_test_results WHERE ab_test_id = t.id AND variant = 'control'),
        0
    ) as control_avg_latency,
    COALESCE(
        (SELECT AVG(latency_ms) FROM ab_test_results WHERE ab_test_id = t.id AND variant = 'variant'),
        0
    ) as variant_avg_latency,
    COALESCE(
        (SELECT AVG(cost) FROM ab_test_results WHERE ab_test_id = t.id AND variant = 'control'),
        0
    ) as control_avg_cost,
    COALESCE(
        (SELECT AVG(cost) FROM ab_test_results WHERE ab_test_id = t.id AND variant = 'variant'),
        0
    ) as variant_avg_cost,
    COALESCE(
        (SELECT SUM(CASE WHEN error THEN 1 ELSE 0 END)::FLOAT / NULLIF(COUNT(*), 0) 
         FROM ab_test_results WHERE ab_test_id = t.id AND variant = 'control'),
        0
    ) as control_error_rate,
    COALESCE(
        (SELECT SUM(CASE WHEN error THEN 1 ELSE 0 END)::FLOAT / NULLIF(COUNT(*), 0)
         FROM ab_test_results WHERE ab_test_id = t.id AND variant = 'variant'),
        0
    ) as variant_error_rate,
    t.stat_significant,
    t.winner
FROM ab_tests t;

-- ============================================
-- FUNCTIONS
-- ============================================

-- Function to calculate statistical significance (simplified Z-test)
CREATE OR REPLACE FUNCTION calculate_ab_test_significance(
    control_successes INT,
    control_trials INT,
    variant_successes INT,
    variant_trials INT,
    confidence_level DECIMAL DEFAULT 0.95
) RETURNS TABLE(
    is_significant BOOLEAN,
    p_value DECIMAL,
    z_score DECIMAL
) AS $$
DECLARE
    p1 DECIMAL;
    p2 DECIMAL;
    p_pooled DECIMAL;
    se DECIMAL;
    z DECIMAL;
    p_val DECIMAL;
BEGIN
    -- Calculate conversion rates
    p1 := control_successes::DECIMAL / NULLIF(control_trials, 0);
    p2 := variant_successes::DECIMAL / NULLIF(variant_trials, 0);
    
    -- Calculate pooled probability
    p_pooled := (control_successes + variant_successes)::DECIMAL / 
                NULLIF(control_trials + variant_trials, 0);
    
    -- Calculate standard error
    se := SQRT(p_pooled * (1 - p_pooled) * 
               (1.0/NULLIF(control_trials, 0) + 1.0/NULLIF(variant_trials, 0)));
    
    -- Calculate z-score
    z := (p2 - p1) / NULLIF(se, 0);
    
    -- Simplified p-value calculation (would use normal CDF in production)
    p_val := CASE 
        WHEN ABS(z) > 2.58 THEN 0.01  -- 99% confidence
        WHEN ABS(z) > 1.96 THEN 0.05  -- 95% confidence
        WHEN ABS(z) > 1.64 THEN 0.10  -- 90% confidence
        ELSE 0.50
    END;
    
    RETURN QUERY SELECT 
        p_val < (1 - confidence_level),
        p_val,
        z;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- TRIGGERS
-- ============================================

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_ab_test_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_ab_tests_updated_at
    BEFORE UPDATE ON ab_tests
    FOR EACH ROW
    EXECUTE FUNCTION update_ab_test_updated_at();