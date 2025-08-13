-- Evaluation system schema for EvalForge
-- This file adds tables for the auto-evaluation engine

-- Enable JSON extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Evaluations table
CREATE TABLE IF NOT EXISTS evaluations (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    prompt_text TEXT NOT NULL,
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'running', 'completed', 'failed', 'cancelled')),
    progress DECIMAL(5,2) DEFAULT 0.0 CHECK (progress >= 0 AND progress <= 100),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Prompt analyses table
CREATE TABLE IF NOT EXISTS prompt_analyses (
    id SERIAL PRIMARY KEY,
    evaluation_id INTEGER NOT NULL REFERENCES evaluations(id) ON DELETE CASCADE,
    prompt_id INTEGER, -- Reference to original prompt if applicable
    prompt_text TEXT NOT NULL,
    task_type VARCHAR(50) NOT NULL CHECK (task_type IN ('classification', 'generation', 'extraction', 'summarization', 'question_answering', 'transformation', 'completion')),
    input_schema JSONB DEFAULT '{}',
    output_schema JSONB DEFAULT '{}',
    constraints JSONB DEFAULT '[]',
    examples JSONB DEFAULT '[]',
    confidence DECIMAL(4,3) DEFAULT 0.0 CHECK (confidence >= 0 AND confidence <= 1),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Test cases table
CREATE TABLE IF NOT EXISTS test_cases (
    id SERIAL PRIMARY KEY,
    evaluation_id INTEGER NOT NULL REFERENCES evaluations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    input JSONB NOT NULL DEFAULT '{}',
    expected_output JSONB NOT NULL DEFAULT '{}',
    actual_output JSONB DEFAULT '{}',
    category VARCHAR(50) NOT NULL DEFAULT 'normal' CHECK (category IN ('normal', 'edge_case', 'adversarial')),
    weight DECIMAL(4,2) DEFAULT 1.0 CHECK (weight >= 0),
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'running', 'passed', 'failed', 'error', 'skipped')),
    score DECIMAL(4,3) DEFAULT 0.0 CHECK (score >= 0 AND score <= 1),
    execution_time_ms INTEGER DEFAULT 0,
    error_message TEXT,
    executed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Evaluation metrics table
CREATE TABLE IF NOT EXISTS evaluation_metrics (
    id SERIAL PRIMARY KEY,
    evaluation_id INTEGER NOT NULL REFERENCES evaluations(id) ON DELETE CASCADE,
    overall_score DECIMAL(4,3) DEFAULT 0.0 CHECK (overall_score >= 0 AND overall_score <= 1),
    pass_rate DECIMAL(4,3) DEFAULT 0.0 CHECK (pass_rate >= 0 AND pass_rate <= 1),
    test_cases_passed INTEGER DEFAULT 0,
    test_cases_total INTEGER DEFAULT 0,
    classification_metrics JSONB DEFAULT NULL,
    generation_metrics JSONB DEFAULT NULL,
    custom_metrics JSONB DEFAULT '{}',
    calculated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Error analyses table
CREATE TABLE IF NOT EXISTS error_analyses (
    id SERIAL PRIMARY KEY,
    evaluation_id INTEGER NOT NULL REFERENCES evaluations(id) ON DELETE CASCADE,
    common_errors JSONB DEFAULT '[]',
    error_patterns JSONB DEFAULT '{}',
    ambiguous_cases DECIMAL(4,3) DEFAULT 0.0 CHECK (ambiguous_cases >= 0 AND ambiguous_cases <= 1),
    format_errors DECIMAL(4,3) DEFAULT 0.0 CHECK (format_errors >= 0 AND format_errors <= 1),
    logic_errors DECIMAL(4,3) DEFAULT 0.0 CHECK (logic_errors >= 0 AND logic_errors <= 1),
    inconsistent_cases DECIMAL(4,3) DEFAULT 0.0 CHECK (inconsistent_cases >= 0 AND inconsistent_cases <= 1),
    error_categories JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Optimization suggestions table
CREATE TABLE IF NOT EXISTS optimization_suggestions (
    id SERIAL PRIMARY KEY,
    evaluation_id INTEGER NOT NULL REFERENCES evaluations(id) ON DELETE CASCADE,
    type VARCHAR(100) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    old_prompt TEXT NOT NULL,
    new_prompt TEXT NOT NULL,
    expected_impact DECIMAL(4,3) DEFAULT 0.0 CHECK (expected_impact >= 0 AND expected_impact <= 1),
    confidence DECIMAL(4,3) DEFAULT 0.0 CHECK (confidence >= 0 AND confidence <= 1),
    priority VARCHAR(20) DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high')),
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'rejected', 'applied')),
    reasoning TEXT,
    examples JSONB DEFAULT '[]',
    metadata JSONB DEFAULT '{}',
    applied_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- A/B tests table
CREATE TABLE IF NOT EXISTS ab_tests (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    control_prompt TEXT NOT NULL,
    variant_prompt TEXT NOT NULL,
    traffic_ratio DECIMAL(4,3) DEFAULT 0.5 CHECK (traffic_ratio >= 0 AND traffic_ratio <= 1),
    status VARCHAR(20) DEFAULT 'draft' CHECK (status IN ('draft', 'running', 'completed', 'cancelled')),
    min_sample_size INTEGER DEFAULT 100 CHECK (min_sample_size > 0),
    control_samples INTEGER DEFAULT 0,
    variant_samples INTEGER DEFAULT 0,
    stat_significant BOOLEAN DEFAULT FALSE,
    confidence_level DECIMAL(4,3) DEFAULT 0.95 CHECK (confidence_level >= 0 AND confidence_level <= 1),
    winner VARCHAR(20) CHECK (winner IN ('control', 'variant', 'inconclusive')),
    effect_size DECIMAL(6,4) DEFAULT 0.0,
    p_value DECIMAL(10,8) DEFAULT 1.0,
    started_at TIMESTAMP WITH TIME ZONE,
    ended_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- A/B test results table
CREATE TABLE IF NOT EXISTS ab_test_results (
    id SERIAL PRIMARY KEY,
    ab_test_id UUID NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE,
    variant VARCHAR(20) NOT NULL CHECK (variant IN ('control', 'variant')),
    evaluation_id UUID REFERENCES evaluations(id) ON DELETE SET NULL,
    metrics JSONB DEFAULT '{}',
    sample_size INTEGER DEFAULT 0,
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Test execution logs table
CREATE TABLE IF NOT EXISTS test_execution_logs (
    id SERIAL PRIMARY KEY,
    test_case_id UUID NOT NULL REFERENCES test_cases(id) ON DELETE CASCADE,
    execution_start TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    execution_end TIMESTAMP WITH TIME ZONE,
    duration_ms INTEGER,
    status VARCHAR(50) NOT NULL,
    input_used JSONB DEFAULT '{}',
    output_received JSONB DEFAULT '{}',
    error_details TEXT,
    llm_provider VARCHAR(100),
    llm_model VARCHAR(100),
    tokens_used JSONB DEFAULT '{}',
    cost DECIMAL(10,6) DEFAULT 0.0,
    metadata JSONB DEFAULT '{}'
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_evaluations_project_id ON evaluations(project_id);
CREATE INDEX IF NOT EXISTS idx_evaluations_status ON evaluations(status);
CREATE INDEX IF NOT EXISTS idx_evaluations_created_at ON evaluations(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_prompt_analyses_evaluation_id ON prompt_analyses(evaluation_id);
CREATE INDEX IF NOT EXISTS idx_prompt_analyses_task_type ON prompt_analyses(task_type);

CREATE INDEX IF NOT EXISTS idx_test_cases_evaluation_id ON test_cases(evaluation_id);
CREATE INDEX IF NOT EXISTS idx_test_cases_category ON test_cases(category);
CREATE INDEX IF NOT EXISTS idx_test_cases_status ON test_cases(status);

CREATE INDEX IF NOT EXISTS idx_evaluation_metrics_evaluation_id ON evaluation_metrics(evaluation_id);

CREATE INDEX IF NOT EXISTS idx_error_analyses_evaluation_id ON error_analyses(evaluation_id);

CREATE INDEX IF NOT EXISTS idx_optimization_suggestions_evaluation_id ON optimization_suggestions(evaluation_id);
CREATE INDEX IF NOT EXISTS idx_optimization_suggestions_type ON optimization_suggestions(type);
CREATE INDEX IF NOT EXISTS idx_optimization_suggestions_status ON optimization_suggestions(status);
CREATE INDEX IF NOT EXISTS idx_optimization_suggestions_priority ON optimization_suggestions(priority);

CREATE INDEX IF NOT EXISTS idx_ab_tests_project_id ON ab_tests(project_id);
CREATE INDEX IF NOT EXISTS idx_ab_tests_status ON ab_tests(status);

CREATE INDEX IF NOT EXISTS idx_ab_test_results_ab_test_id ON ab_test_results(ab_test_id);
CREATE INDEX IF NOT EXISTS idx_ab_test_results_variant ON ab_test_results(variant);

CREATE INDEX IF NOT EXISTS idx_test_execution_logs_test_case_id ON test_execution_logs(test_case_id);
CREATE INDEX IF NOT EXISTS idx_test_execution_logs_status ON test_execution_logs(status);

-- Create composite indexes for common queries
CREATE INDEX IF NOT EXISTS idx_evaluations_project_status ON evaluations(project_id, status);
CREATE INDEX IF NOT EXISTS idx_test_cases_eval_category ON test_cases(evaluation_id, category);
CREATE INDEX IF NOT EXISTS idx_suggestions_eval_priority ON optimization_suggestions(evaluation_id, priority);

-- Add trigger to update the updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply the trigger to tables with updated_at columns
CREATE TRIGGER update_evaluations_updated_at BEFORE UPDATE ON evaluations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_prompt_analyses_updated_at BEFORE UPDATE ON prompt_analyses
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_optimization_suggestions_updated_at BEFORE UPDATE ON optimization_suggestions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_ab_tests_updated_at BEFORE UPDATE ON ab_tests
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Add some helpful views for common queries
CREATE OR REPLACE VIEW evaluation_summary AS
SELECT 
    e.id,
    e.name,
    e.description,
    e.status,
    e.progress,
    e.created_at,
    e.completed_at,
    pa.task_type,
    pa.confidence as analysis_confidence,
    em.overall_score,
    em.pass_rate,
    em.test_cases_total,
    em.test_cases_passed,
    COUNT(DISTINCT os.id) as suggestion_count,
    COUNT(DISTINCT CASE WHEN os.priority = 'high' THEN os.id END) as high_priority_suggestions
FROM evaluations e
LEFT JOIN prompt_analyses pa ON e.id = pa.evaluation_id
LEFT JOIN evaluation_metrics em ON e.id = em.evaluation_id
LEFT JOIN optimization_suggestions os ON e.id = os.evaluation_id
GROUP BY e.id, e.name, e.description, e.status, e.progress, e.created_at, e.completed_at,
         pa.task_type, pa.confidence, em.overall_score, em.pass_rate, em.test_cases_total, em.test_cases_passed;

-- View for test case performance by category
CREATE OR REPLACE VIEW test_case_performance AS
SELECT 
    tc.evaluation_id,
    tc.category,
    COUNT(*) as total_cases,
    COUNT(CASE WHEN tc.status = 'passed' THEN 1 END) as passed_cases,
    COUNT(CASE WHEN tc.status = 'failed' THEN 1 END) as failed_cases,
    COUNT(CASE WHEN tc.status = 'error' THEN 1 END) as error_cases,
    AVG(tc.score) as avg_score,
    AVG(tc.weight) as avg_weight,
    SUM(tc.score * tc.weight) / NULLIF(SUM(tc.weight), 0) as weighted_score
FROM test_cases tc
GROUP BY tc.evaluation_id, tc.category;

-- View for suggestion statistics
CREATE OR REPLACE VIEW suggestion_stats AS
SELECT 
    os.evaluation_id,
    os.type,
    os.priority,
    COUNT(*) as count,
    AVG(os.expected_impact) as avg_expected_impact,
    AVG(os.confidence) as avg_confidence,
    COUNT(CASE WHEN os.status = 'applied' THEN 1 END) as applied_count,
    COUNT(CASE WHEN os.status = 'accepted' THEN 1 END) as accepted_count,
    COUNT(CASE WHEN os.status = 'rejected' THEN 1 END) as rejected_count
FROM optimization_suggestions os
GROUP BY os.evaluation_id, os.type, os.priority;

-- Add some sample data for development (optional)
-- This will be useful for testing the evaluation system

-- Note: In production, this sample data should be removed or placed in a separate seed file