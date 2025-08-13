-- Notifications Schema for EvalForge

-- ============================================
-- NOTIFICATION CONFIGS TABLE
-- ============================================

CREATE TABLE IF NOT EXISTS notification_configs (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    channel VARCHAR(50) NOT NULL, -- slack, email, webhook
    type VARCHAR(50) NOT NULL, -- evaluation_complete, error_alert, cost_alert, etc
    config JSONB NOT NULL, -- Channel-specific configuration
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================
-- NOTIFICATION HISTORY TABLE
-- ============================================

CREATE TABLE IF NOT EXISTS notification_history (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    config_id INTEGER REFERENCES notification_configs(id) ON DELETE SET NULL,
    channel VARCHAR(50) NOT NULL,
    type VARCHAR(50) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending', -- pending, sent, failed
    payload JSONB,
    error_message TEXT,
    sent_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================
-- ALERT THRESHOLDS TABLE
-- ============================================

CREATE TABLE IF NOT EXISTS alert_thresholds (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    metric_type VARCHAR(50) NOT NULL, -- error_rate, cost, latency_p95, etc
    threshold_value DECIMAL(10,4) NOT NULL,
    comparison_operator VARCHAR(10) DEFAULT '>', -- >, <, >=, <=, =
    time_window_minutes INTEGER DEFAULT 5,
    cooldown_minutes INTEGER DEFAULT 30, -- Prevent alert spam
    enabled BOOLEAN DEFAULT TRUE,
    last_triggered TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(project_id, metric_type)
);

-- ============================================
-- NOTIFICATION TEMPLATES TABLE
-- ============================================

CREATE TABLE IF NOT EXISTS notification_templates (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    channel VARCHAR(50) NOT NULL,
    type VARCHAR(50) NOT NULL,
    template_config JSONB NOT NULL,
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================
-- INDEXES
-- ============================================

CREATE INDEX IF NOT EXISTS idx_notification_configs_project 
ON notification_configs(project_id);

CREATE INDEX IF NOT EXISTS idx_notification_configs_enabled 
ON notification_configs(enabled) WHERE enabled = TRUE;

CREATE INDEX IF NOT EXISTS idx_notification_history_project 
ON notification_history(project_id);

CREATE INDEX IF NOT EXISTS idx_notification_history_status 
ON notification_history(status);

CREATE INDEX IF NOT EXISTS idx_notification_history_created 
ON notification_history(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_alert_thresholds_project 
ON alert_thresholds(project_id);

CREATE INDEX IF NOT EXISTS idx_alert_thresholds_enabled 
ON alert_thresholds(enabled) WHERE enabled = TRUE;

-- ============================================
-- VIEWS
-- ============================================

-- View for active notification configurations
CREATE OR REPLACE VIEW v_active_notifications AS
SELECT 
    nc.id,
    nc.project_id,
    p.name as project_name,
    nc.channel,
    nc.type,
    nc.config,
    nc.enabled,
    nc.created_at
FROM notification_configs nc
JOIN projects p ON nc.project_id = p.id
WHERE nc.enabled = TRUE;

-- View for notification statistics
CREATE OR REPLACE VIEW v_notification_stats AS
SELECT 
    nh.project_id,
    p.name as project_name,
    nh.channel,
    nh.type,
    COUNT(*) as total_sent,
    COUNT(CASE WHEN nh.status = 'sent' THEN 1 END) as successful,
    COUNT(CASE WHEN nh.status = 'failed' THEN 1 END) as failed,
    MAX(nh.sent_at) as last_sent
FROM notification_history nh
JOIN projects p ON nh.project_id = p.id
WHERE nh.created_at > NOW() - INTERVAL '30 days'
GROUP BY nh.project_id, p.name, nh.channel, nh.type;

-- ============================================
-- FUNCTIONS
-- ============================================

-- Function to check if alert should be triggered
CREATE OR REPLACE FUNCTION should_trigger_alert(
    p_project_id INTEGER,
    p_metric_type VARCHAR,
    p_current_value DECIMAL
) RETURNS BOOLEAN AS $$
DECLARE
    v_threshold alert_thresholds%ROWTYPE;
    v_should_trigger BOOLEAN := FALSE;
BEGIN
    SELECT * INTO v_threshold
    FROM alert_thresholds
    WHERE project_id = p_project_id
      AND metric_type = p_metric_type
      AND enabled = TRUE;
    
    IF NOT FOUND THEN
        RETURN FALSE;
    END IF;
    
    -- Check if in cooldown period
    IF v_threshold.last_triggered IS NOT NULL AND
       v_threshold.last_triggered + (v_threshold.cooldown_minutes || ' minutes')::INTERVAL > NOW() THEN
        RETURN FALSE;
    END IF;
    
    -- Check threshold
    CASE v_threshold.comparison_operator
        WHEN '>' THEN
            v_should_trigger := p_current_value > v_threshold.threshold_value;
        WHEN '>=' THEN
            v_should_trigger := p_current_value >= v_threshold.threshold_value;
        WHEN '<' THEN
            v_should_trigger := p_current_value < v_threshold.threshold_value;
        WHEN '<=' THEN
            v_should_trigger := p_current_value <= v_threshold.threshold_value;
        WHEN '=' THEN
            v_should_trigger := p_current_value = v_threshold.threshold_value;
        ELSE
            v_should_trigger := FALSE;
    END CASE;
    
    -- Update last triggered if alert should fire
    IF v_should_trigger THEN
        UPDATE alert_thresholds
        SET last_triggered = NOW()
        WHERE id = v_threshold.id;
    END IF;
    
    RETURN v_should_trigger;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- TRIGGERS
-- ============================================

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_notification_configs_updated_at
    BEFORE UPDATE ON notification_configs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_alert_thresholds_updated_at
    BEFORE UPDATE ON alert_thresholds
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- SAMPLE DATA
-- ============================================

-- Insert default notification templates
INSERT INTO notification_templates (name, channel, type, template_config, is_default) VALUES
(
    'Default Slack Evaluation',
    'slack',
    'evaluation_complete',
    '{"message_format": "standard", "include_metrics": true}'::jsonb,
    TRUE
),
(
    'Default Slack Error Alert',
    'slack',
    'error_alert',
    '{"message_format": "alert", "priority": "high"}'::jsonb,
    TRUE
),
(
    'Default Webhook',
    'webhook',
    '*',
    '{"headers": {"Content-Type": "application/json"}, "retry_count": 3}'::jsonb,
    TRUE
)
ON CONFLICT DO NOTHING;

-- Insert sample alert thresholds (for demo project if exists)
INSERT INTO alert_thresholds (project_id, metric_type, threshold_value, comparison_operator, time_window_minutes) 
SELECT 
    id,
    'error_rate',
    0.05, -- 5% error rate
    '>',
    5
FROM projects
WHERE name = 'Demo Project'
ON CONFLICT DO NOTHING;

INSERT INTO alert_thresholds (project_id, metric_type, threshold_value, comparison_operator, time_window_minutes) 
SELECT 
    id,
    'cost',
    100.00, -- $100 cost threshold
    '>',
    60
FROM projects
WHERE name = 'Demo Project'
ON CONFLICT DO NOTHING;

INSERT INTO alert_thresholds (project_id, metric_type, threshold_value, comparison_operator, time_window_minutes) 
SELECT 
    id,
    'latency_p95',
    1000.00, -- 1000ms latency
    '>',
    5
FROM projects
WHERE name = 'Demo Project'
ON CONFLICT DO NOTHING;