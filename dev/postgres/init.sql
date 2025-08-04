-- PostgreSQL initialization script for EvalForge
-- This script sets up the core OLTP database schema

-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "pg_stat_statements";

-- Create custom types
CREATE TYPE user_role AS ENUM ('admin', 'developer', 'viewer');
CREATE TYPE project_status AS ENUM ('active', 'archived', 'paused');
CREATE TYPE evaluation_status AS ENUM ('pending', 'running', 'completed', 'failed');

-- Organizations table
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    role user_role DEFAULT 'viewer',
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    auth_provider VARCHAR(50) DEFAULT 'local',
    auth_provider_id VARCHAR(255),
    settings JSONB DEFAULT '{}',
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- API keys for authentication
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(255) UNIQUE NOT NULL, -- bcrypt hash of the actual key
    key_prefix VARCHAR(10) NOT NULL, -- First few chars for identification
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    permissions JSONB DEFAULT '[]',
    last_used_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Projects table
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    status project_status DEFAULT 'active',
    settings JSONB DEFAULT '{}',
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Prompt versions for tracking changes
CREATE TABLE prompt_versions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    version_number INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    template TEXT NOT NULL,
    variables JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(project_id, version_number)
);

-- Evaluation definitions
CREATE TABLE evaluation_definitions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    criteria JSONB NOT NULL, -- evaluation criteria and rules
    test_cases JSONB DEFAULT '[]',
    schedule_cron VARCHAR(100), -- for automated evaluations
    is_active BOOLEAN DEFAULT true,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Evaluation runs
CREATE TABLE evaluation_runs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    evaluation_definition_id UUID REFERENCES evaluation_definitions(id) ON DELETE CASCADE,
    prompt_version_id UUID REFERENCES prompt_versions(id),
    status evaluation_status DEFAULT 'pending',
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    results JSONB DEFAULT '{}',
    metrics JSONB DEFAULT '{}',
    error_message TEXT,
    triggered_by VARCHAR(50) DEFAULT 'manual', -- manual, scheduled, webhook
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- LLM provider configurations
CREATE TABLE llm_providers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    provider_type VARCHAR(50) NOT NULL, -- openai, anthropic, cohere, etc.
    config JSONB NOT NULL, -- encrypted API keys and settings
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(organization_id, name)
);

-- Webhooks for external integrations
CREATE TABLE webhooks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    url VARCHAR(1000) NOT NULL,
    events TEXT[] NOT NULL, -- array of event types
    secret VARCHAR(255), -- for signature verification
    is_active BOOLEAN DEFAULT true,
    last_triggered_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX idx_users_organization_id ON users(organization_id);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX idx_api_keys_organization_id ON api_keys(organization_id);
CREATE INDEX idx_projects_organization_id ON projects(organization_id);
CREATE INDEX idx_projects_status ON projects(status);
CREATE INDEX idx_prompt_versions_project_id ON prompt_versions(project_id);
CREATE INDEX idx_evaluation_definitions_project_id ON evaluation_definitions(project_id);
CREATE INDEX idx_evaluation_runs_definition_id ON evaluation_runs(evaluation_definition_id);
CREATE INDEX idx_evaluation_runs_status ON evaluation_runs(status);
CREATE INDEX idx_evaluation_runs_created_at ON evaluation_runs(created_at);
CREATE INDEX idx_llm_providers_organization_id ON llm_providers(organization_id);
CREATE INDEX idx_webhooks_project_id ON webhooks(project_id);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
CREATE TRIGGER update_organizations_updated_at BEFORE UPDATE ON organizations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_projects_updated_at BEFORE UPDATE ON projects
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_evaluation_definitions_updated_at BEFORE UPDATE ON evaluation_definitions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_llm_providers_updated_at BEFORE UPDATE ON llm_providers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create a function to generate API keys
CREATE OR REPLACE FUNCTION generate_api_key()
RETURNS TEXT AS $$
DECLARE
    key_length CONSTANT INTEGER := 32;
    chars CONSTANT TEXT := 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    result TEXT := 'ef_';
    i INTEGER;
BEGIN
    FOR i IN 1..key_length LOOP
        result := result || substr(chars, (random() * length(chars))::integer + 1, 1);
    END LOOP;
    RETURN result;
END;
$$ LANGUAGE plpgsql;

-- Create views for common queries
CREATE VIEW user_projects AS
SELECT 
    p.id,
    p.name,
    p.description,
    p.status,
    p.created_at,
    p.updated_at,
    u.name as created_by_name,
    o.name as organization_name
FROM projects p
JOIN users u ON p.created_by = u.id
JOIN organizations o ON p.organization_id = o.id;

CREATE VIEW evaluation_summary AS
SELECT 
    ed.id,
    ed.name,
    ed.description,
    p.name as project_name,
    COUNT(er.id) as total_runs,
    COUNT(CASE WHEN er.status = 'completed' THEN 1 END) as completed_runs,
    COUNT(CASE WHEN er.status = 'failed' THEN 1 END) as failed_runs,
    MAX(er.completed_at) as last_run_at
FROM evaluation_definitions ed
JOIN projects p ON ed.project_id = p.id
LEFT JOIN evaluation_runs er ON ed.id = er.evaluation_definition_id
WHERE ed.is_active = true
GROUP BY ed.id, ed.name, ed.description, p.name;

-- Grant permissions
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO evalforge;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO evalforge;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO evalforge;