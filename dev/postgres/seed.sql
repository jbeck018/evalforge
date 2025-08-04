-- Development seed data for EvalForge
-- This script creates realistic test data for local development

-- Create test organizations
INSERT INTO organizations (id, name, slug, settings) VALUES
  ('11111111-1111-1111-1111-111111111111', 'Acme Corporation', 'acme-corp', '{"plan": "enterprise", "max_projects": 100}'),
  ('22222222-2222-2222-2222-222222222222', 'TechStartup Inc', 'techstartup', '{"plan": "pro", "max_projects": 10}'),
  ('33333333-3333-3333-3333-333333333333', 'Enterprise Solutions', 'enterprise-sol', '{"plan": "enterprise", "max_projects": 50}');

-- Create test users
INSERT INTO users (id, email, name, role, organization_id, settings) VALUES
  ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'admin@acme.com', 'Alice Admin', 'admin', '11111111-1111-1111-1111-111111111111', '{"notifications": true, "theme": "dark"}'),
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'dev@acme.com', 'Bob Developer', 'developer', '11111111-1111-1111-1111-111111111111', '{"notifications": true}'),
  ('cccccccc-cccc-cccc-cccc-cccccccccccc', 'viewer@acme.com', 'Carol Viewer', 'viewer', '11111111-1111-1111-1111-111111111111', '{"notifications": false}'),
  ('dddddddd-dddd-dddd-dddd-dddddddddddd', 'founder@techstartup.com', 'David Founder', 'admin', '22222222-2222-2222-2222-222222222222', '{"notifications": true}'),
  ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'engineer@techstartup.com', 'Eva Engineer', 'developer', '22222222-2222-2222-2222-222222222222', '{}');

-- Create test API keys (using bcrypt hash for password "test-key-123")
INSERT INTO api_keys (id, name, key_hash, key_prefix, user_id, organization_id, permissions) VALUES
  ('f1f1f1f1-f1f1-f1f1-f1f1-f1f1f1f1f1f1', 'Development Key', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'ef_dev', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '11111111-1111-1111-1111-111111111111', '["read", "write", "admin"]'),
  ('f2f2f2f2-f2f2-f2f2-f2f2-f2f2f2f2f2f2', 'Testing Key', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'ef_test', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '11111111-1111-1111-1111-111111111111', '["read", "write"]');

-- Create test projects
INSERT INTO projects (id, name, description, organization_id, status, settings, created_by) VALUES
  ('p1p1p1p1-p1p1-p1p1-p1p1-p1p1p1p1p1p1', 'Customer Support Chatbot', 'AI-powered customer support assistant for handling common inquiries', '11111111-1111-1111-1111-111111111111', 'active', '{"model": "gpt-4", "max_tokens": 500}', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'),
  ('p2p2p2p2-p2p2-p2p2-p2p2-p2p2p2p2p2p2', 'Code Review Assistant', 'Automated code review and suggestions system', '11111111-1111-1111-1111-111111111111', 'active', '{"model": "gpt-4", "temperature": 0.2}', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'),
  ('p3p3p3p3-p3p3-p3p3-p3p3-p3p3p3p3p3p3', 'Document Summarizer', 'Intelligent document summarization for legal documents', '11111111-1111-1111-1111-111111111111', 'active', '{"model": "claude-3-opus"}', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'),
  ('p4p4p4p4-p4p4-p4p4-p4p4-p4p4p4p4p4p4', 'Content Generator', 'Marketing content generation and optimization', '22222222-2222-2222-2222-222222222222', 'active', '{"model": "gpt-3.5-turbo"}', 'dddddddd-dddd-dddd-dddd-dddddddddddd'),
  ('p5p5p5p5-p5p5-p5p5-p5p5-p5p5p5p5p5p5', 'RAG Knowledge Base', 'Retrieval-augmented generation for company knowledge base', '33333333-3333-3333-3333-333333333333', 'active', '{"model": "gpt-4", "embedding_model": "text-embedding-ada-002"}', 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee');

-- Create prompt versions
INSERT INTO prompt_versions (id, project_id, version_number, name, template, variables, metadata, created_by) VALUES
  ('pv1pv1pv-1pv1-pv1p-v1pv-1pv1pv1pv1pv', 'p1p1p1p1-p1p1-p1p1-p1p1-p1p1p1p1p1p1', 1, 'Initial Support Prompt', 
   'You are a helpful customer support assistant for {{company_name}}. Please help the customer with their inquiry: {{customer_question}}. Be polite, helpful, and concise.',
   '{"company_name": "Acme Corporation", "customer_question": ""}',
   '{"created_reason": "Initial version", "performance_notes": "Baseline performance"}',
   'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'),
  
  ('pv2pv2pv-2pv2-pv2p-v2pv-2pv2pv2pv2pv', 'p1p1p1p1-p1p1-p1p1-p1p1-p1p1p1p1p1p1', 2, 'Enhanced Support Prompt',
   'You are a professional customer support representative for {{company_name}}. Your goal is to provide accurate, helpful, and empathetic responses. Customer inquiry: {{customer_question}}. If you cannot help, explain what the customer should do next.',
   '{"company_name": "Acme Corporation", "customer_question": ""}',
   '{"created_reason": "Improved empathy and accuracy", "performance_notes": "20% better customer satisfaction"}',
   'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'),
   
  ('pv3pv3pv-3pv3-pv3p-v3pv-3pv3pv3pv3pv', 'p2p2p2p2-p2p2-p2p2-p2p2-p2p2p2p2p2p2', 1, 'Code Review Prompt',
   'Review the following {{language}} code for best practices, potential bugs, and improvements:\n\n{{code}}\n\nProvide specific, actionable feedback with examples where appropriate.',
   '{"language": "JavaScript", "code": ""}',
   '{"created_reason": "Initial code review prompt"}',
   'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb');

-- Create LLM provider configurations (mock data)
INSERT INTO llm_providers (id, organization_id, name, provider_type, config, is_active) VALUES
  ('llm1llm1-llm1-llm1-llm1-llm1llm1llm1', '11111111-1111-1111-1111-111111111111', 'OpenAI Production', 'openai', 
   '{"api_key": "encrypted_key_here", "base_url": "https://api.openai.com/v1", "models": ["gpt-4", "gpt-3.5-turbo"]}', true),
  ('llm2llm2-llm2-llm2-llm2-llm2llm2llm2', '11111111-1111-1111-1111-111111111111', 'Anthropic Claude', 'anthropic',
   '{"api_key": "encrypted_key_here", "base_url": "https://api.anthropic.com", "models": ["claude-3-opus", "claude-3-sonnet"]}', true),
  ('llm3llm3-llm3-llm3-llm3-llm3llm3llm3', '22222222-2222-2222-2222-222222222222', 'OpenAI Startup', 'openai',
   '{"api_key": "encrypted_key_here", "base_url": "https://api.openai.com/v1", "models": ["gpt-3.5-turbo"]}', true);

-- Create evaluation definitions
INSERT INTO evaluation_definitions (id, project_id, name, description, criteria, test_cases, is_active, created_by) VALUES
  ('ed1ed1ed-1ed1-ed1e-d1ed-1ed1ed1ed1ed', 'p1p1p1p1-p1p1-p1p1-p1p1-p1p1p1p1p1p1', 'Customer Satisfaction Eval',
   'Evaluates customer support responses for helpfulness, politeness, and accuracy',
   '{"relevance": {"weight": 0.4, "description": "How relevant is the response to the question"}, "politeness": {"weight": 0.3, "description": "How polite and professional is the tone"}, "accuracy": {"weight": 0.3, "description": "Is the information provided accurate"}}',
   '[{"input": "I cant login to my account", "expected_criteria": {"relevance": 0.9, "politeness": 0.8, "accuracy": 0.9}}, {"input": "How do I cancel my subscription?", "expected_criteria": {"relevance": 0.9, "politeness": 0.8, "accuracy": 0.9}}]',
   true, 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'),
   
  ('ed2ed2ed-2ed2-ed2e-d2ed-2ed2ed2ed2ed', 'p2p2p2p2-p2p2-p2p2-p2p2-p2p2p2p2p2p2', 'Code Quality Eval',
   'Evaluates code review responses for technical accuracy and usefulness',
   '{"technical_accuracy": {"weight": 0.5, "description": "Are the technical suggestions correct"}, "completeness": {"weight": 0.3, "description": "Does it cover all important issues"}, "clarity": {"weight": 0.2, "description": "Are suggestions clear and actionable"}}',
   '[{"input": "function getData() { return fetch(\"/api/data\").then(r => r.json()) }", "expected_criteria": {"technical_accuracy": 0.8, "completeness": 0.7, "clarity": 0.8}}]',
   true, 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb');

-- Create some evaluation runs with sample results
INSERT INTO evaluation_runs (id, evaluation_definition_id, prompt_version_id, status, started_at, completed_at, results, metrics) VALUES
  ('er1er1er-1er1-er1e-r1er-1er1er1er1er', 'ed1ed1ed-1ed1-ed1e-d1ed-1ed1ed1ed1ed', 'pv1pv1pv-1pv1-pv1p-v1pv-1pv1pv1pv1pv', 'completed',
   NOW() - INTERVAL '2 hours', NOW() - INTERVAL '1 hour 45 minutes',
   '{"test_results": [{"input": "I cant login to my account", "output": "I understand you are having trouble logging in. Please try resetting your password using the forgot password link.", "scores": {"relevance": 0.85, "politeness": 0.9, "accuracy": 0.8}}]}',
   '{"overall_score": 0.85, "total_tests": 5, "passed_tests": 4, "cost_cents": 245}'),
   
  ('er2er2er-2er2-er2e-r2er-2er2er2er2er', 'ed1ed1ed-1ed1-ed1e-d1ed-1ed1ed1ed1ed', 'pv2pv2pv-2pv2-pv2p-v2pv-2pv2pv2pv2pv', 'completed',
   NOW() - INTERVAL '1 hour', NOW() - INTERVAL '30 minutes',
   '{"test_results": [{"input": "I cant login to my account", "output": "I am sorry to hear you are experiencing difficulties logging into your account. Let me help you resolve this issue. Please try the following steps: 1) Clear your browser cache, 2) Try resetting your password using our secure reset link.", "scores": {"relevance": 0.95, "politeness": 0.95, "accuracy": 0.9}}]}',
   '{"overall_score": 0.93, "total_tests": 5, "passed_tests": 5, "cost_cents": 287}'),
   
  ('er3er3er-3er3-er3e-r3er-3er3er3er3er', 'ed2ed2ed-2ed2-ed2e-d2ed-2ed2ed2ed2ed', 'pv3pv3pv-3pv3-pv3p-v3pv-3pv3pv3pv3pv', 'running',
   NOW() - INTERVAL '15 minutes', NULL,
   '{}',
   '{}');

-- Create webhooks for testing
INSERT INTO webhooks (id, project_id, url, events, secret, is_active) VALUES
  ('wh1wh1wh-1wh1-wh1w-h1wh-1wh1wh1wh1wh', 'p1p1p1p1-p1p1-p1p1-p1p1-p1p1p1p1p1p1', 'https://acme.com/webhooks/evalforge', 
   ARRAY['evaluation.completed', 'evaluation.failed'], 'webhook_secret_123', true),
  ('wh2wh2wh-2wh2-wh2w-h2wh-2wh2wh2wh2wh', 'p2p2p2p2-p2p2-p2p2-p2p2-p2p2p2p2p2p2', 'https://slack.com/api/webhooks/12345',
   ARRAY['evaluation.completed'], NULL, true);

-- Update some timestamps to make data look more realistic
UPDATE users SET last_login_at = NOW() - INTERVAL '2 hours' WHERE id = 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa';
UPDATE users SET last_login_at = NOW() - INTERVAL '1 day' WHERE id = 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb';
UPDATE users SET last_login_at = NOW() - INTERVAL '3 days' WHERE id = 'cccccccc-cccc-cccc-cccc-cccccccccccc';

UPDATE api_keys SET last_used_at = NOW() - INTERVAL '30 minutes' WHERE id = 'f1f1f1f1-f1f1-f1f1-f1f1-f1f1f1f1f1f1';
UPDATE api_keys SET last_used_at = NOW() - INTERVAL '2 hours' WHERE id = 'f2f2f2f2-f2f2-f2f2-f2f2-f2f2f2f2f2f2';

-- Create a function to generate more test data dynamically
CREATE OR REPLACE FUNCTION generate_test_traces(project_uuid UUID, num_traces INTEGER)
RETURNS VOID AS $$
DECLARE
    i INTEGER;
    trace_id TEXT;
    span_id TEXT;
BEGIN
    FOR i IN 1..num_traces LOOP
        trace_id := 'trace_' || i || '_' || extract(epoch from now())::text;
        span_id := 'span_' || i || '_' || extract(epoch from now())::text;
        
        -- This would typically insert into ClickHouse, but for this seed script
        -- we're just creating the metadata in PostgreSQL
        -- The actual trace data will be generated by the data generator script
        
        RAISE NOTICE 'Generated trace % for project %', trace_id, project_uuid;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Grant permissions for the test data
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO evalforge;

-- Create some useful queries for development
COMMENT ON TABLE organizations IS 'Test organizations: Acme Corp (enterprise), TechStartup (pro), Enterprise Solutions (enterprise)';
COMMENT ON TABLE users IS 'Test users: admin@acme.com (admin), dev@acme.com (developer), viewer@acme.com (viewer)';
COMMENT ON TABLE api_keys IS 'Test API keys available with prefix "ef_dev" and "ef_test", both use password "test-key-123"';
COMMENT ON TABLE projects IS 'Test projects covering different AI use cases: chatbot, code review, document summarization, content generation, RAG';

-- Print helpful information
DO $$
BEGIN
    RAISE NOTICE '=== EvalForge Development Database Seeded ===';
    RAISE NOTICE '';
    RAISE NOTICE 'Test Organizations:';
    RAISE NOTICE '  - Acme Corporation (enterprise)';
    RAISE NOTICE '  - TechStartup Inc (pro)';
    RAISE NOTICE '  - Enterprise Solutions (enterprise)';
    RAISE NOTICE '';
    RAISE NOTICE 'Test Users:';
    RAISE NOTICE '  - admin@acme.com (Admin)';
    RAISE NOTICE '  - dev@acme.com (Developer)';  
    RAISE NOTICE '  - viewer@acme.com (Viewer)';
    RAISE NOTICE '';
    RAISE NOTICE 'Test API Keys:';
    RAISE NOTICE '  - ef_dev_... (admin permissions)';
    RAISE NOTICE '  - ef_test_... (read/write permissions)';
    RAISE NOTICE '';
    RAISE NOTICE 'Test Projects:';
    RAISE NOTICE '  - Customer Support Chatbot';
    RAISE NOTICE '  - Code Review Assistant';
    RAISE NOTICE '  - Document Summarizer';
    RAISE NOTICE '  - Content Generator';
    RAISE NOTICE '  - RAG Knowledge Base';
    RAISE NOTICE '';
    RAISE NOTICE 'Database is ready for development!';
END $$;