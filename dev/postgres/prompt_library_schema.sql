-- Prompt Library Schema for EvalForge

-- ============================================
-- PROMPT TEMPLATES TABLE
-- ============================================

CREATE TABLE IF NOT EXISTS prompt_templates (
    id SERIAL PRIMARY KEY,
    project_id INTEGER REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100),
    template_text TEXT NOT NULL,
    variables JSONB DEFAULT '[]', -- Array of variable definitions
    tags TEXT[] DEFAULT '{}',
    version INTEGER DEFAULT 1,
    is_public BOOLEAN DEFAULT FALSE,
    usage_count INTEGER DEFAULT 0,
    average_rating DECIMAL(3,2) DEFAULT 0.0,
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================
-- PROMPT TEMPLATE VERSIONS TABLE
-- ============================================

CREATE TABLE IF NOT EXISTS prompt_template_versions (
    id SERIAL PRIMARY KEY,
    template_id INTEGER NOT NULL REFERENCES prompt_templates(id) ON DELETE CASCADE,
    version INTEGER NOT NULL,
    template_text TEXT NOT NULL,
    variables JSONB DEFAULT '[]',
    change_notes TEXT,
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(template_id, version)
);

-- ============================================
-- PROMPT TEMPLATE RATINGS TABLE
-- ============================================

CREATE TABLE IF NOT EXISTS prompt_template_ratings (
    id SERIAL PRIMARY KEY,
    template_id INTEGER NOT NULL REFERENCES prompt_templates(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id),
    rating INTEGER CHECK (rating >= 1 AND rating <= 5),
    comment TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(template_id, user_id)
);

-- ============================================
-- PROMPT TEMPLATE USAGE TABLE
-- ============================================

CREATE TABLE IF NOT EXISTS prompt_template_usage (
    id SERIAL PRIMARY KEY,
    template_id INTEGER NOT NULL REFERENCES prompt_templates(id) ON DELETE CASCADE,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id),
    variables_used JSONB,
    evaluation_id INTEGER REFERENCES evaluations(id),
    performance_metrics JSONB, -- Stores latency, cost, tokens, etc.
    used_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================
-- COMMON PROMPT CATEGORIES
-- ============================================

CREATE TABLE IF NOT EXISTS prompt_categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    icon VARCHAR(50),
    display_order INTEGER DEFAULT 0
);

-- Insert default categories
INSERT INTO prompt_categories (name, description, icon, display_order) VALUES
    ('Chat', 'Conversational AI and chatbot prompts', 'chat', 1),
    ('Content Generation', 'Blog posts, articles, and creative writing', 'edit', 2),
    ('Code Generation', 'Programming and code-related prompts', 'code', 3),
    ('Data Analysis', 'Data extraction and analysis prompts', 'chart', 4),
    ('Translation', 'Language translation prompts', 'translate', 5),
    ('Summarization', 'Text summarization prompts', 'compress', 6),
    ('Question Answering', 'Q&A and information retrieval', 'help', 7),
    ('Classification', 'Text classification and categorization', 'tag', 8),
    ('Custom', 'User-defined prompt categories', 'folder', 99)
ON CONFLICT (name) DO NOTHING;

-- ============================================
-- INDEXES
-- ============================================

CREATE INDEX IF NOT EXISTS idx_prompt_templates_project 
ON prompt_templates(project_id);

CREATE INDEX IF NOT EXISTS idx_prompt_templates_category 
ON prompt_templates(category);

CREATE INDEX IF NOT EXISTS idx_prompt_templates_public 
ON prompt_templates(is_public) WHERE is_public = TRUE;

CREATE INDEX IF NOT EXISTS idx_prompt_templates_tags 
ON prompt_templates USING GIN(tags);

CREATE INDEX IF NOT EXISTS idx_prompt_template_usage_template 
ON prompt_template_usage(template_id);

CREATE INDEX IF NOT EXISTS idx_prompt_template_usage_project 
ON prompt_template_usage(project_id);

-- ============================================
-- VIEWS
-- ============================================

-- View for popular templates
CREATE OR REPLACE VIEW v_popular_prompt_templates AS
SELECT 
    t.id,
    t.name,
    t.description,
    t.category,
    t.tags,
    t.usage_count,
    t.average_rating,
    t.is_public,
    u.email as created_by_email,
    t.created_at
FROM prompt_templates t
LEFT JOIN users u ON t.created_by = u.id
WHERE t.is_public = TRUE
ORDER BY t.usage_count DESC, t.average_rating DESC
LIMIT 50;

-- View for template performance metrics
CREATE OR REPLACE VIEW v_template_performance AS
SELECT 
    t.id as template_id,
    t.name as template_name,
    COUNT(u.id) as total_uses,
    AVG((u.performance_metrics->>'latency_ms')::FLOAT) as avg_latency,
    AVG((u.performance_metrics->>'cost')::FLOAT) as avg_cost,
    AVG((u.performance_metrics->>'tokens')::INT) as avg_tokens,
    AVG((u.performance_metrics->>'error_rate')::FLOAT) as avg_error_rate
FROM prompt_templates t
LEFT JOIN prompt_template_usage u ON t.id = u.template_id
GROUP BY t.id, t.name;

-- ============================================
-- FUNCTIONS
-- ============================================

-- Function to update average rating
CREATE OR REPLACE FUNCTION update_template_average_rating()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE prompt_templates
    SET average_rating = (
        SELECT AVG(rating)::DECIMAL(3,2)
        FROM prompt_template_ratings
        WHERE template_id = NEW.template_id
    )
    WHERE id = NEW.template_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Function to increment usage count
CREATE OR REPLACE FUNCTION increment_template_usage_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE prompt_templates
    SET usage_count = usage_count + 1
    WHERE id = NEW.template_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Function to create version on update
CREATE OR REPLACE FUNCTION create_template_version()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.template_text != NEW.template_text OR 
       OLD.variables::text != NEW.variables::text THEN
        INSERT INTO prompt_template_versions (
            template_id, version, template_text, variables, created_by
        ) VALUES (
            NEW.id, NEW.version, OLD.template_text, OLD.variables, NEW.created_by
        );
        NEW.version = NEW.version + 1;
    END IF;
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- TRIGGERS
-- ============================================

CREATE TRIGGER trigger_update_template_rating
    AFTER INSERT OR UPDATE ON prompt_template_ratings
    FOR EACH ROW
    EXECUTE FUNCTION update_template_average_rating();

CREATE TRIGGER trigger_increment_usage_count
    AFTER INSERT ON prompt_template_usage
    FOR EACH ROW
    EXECUTE FUNCTION increment_template_usage_count();

CREATE TRIGGER trigger_create_template_version
    BEFORE UPDATE ON prompt_templates
    FOR EACH ROW
    EXECUTE FUNCTION create_template_version();

-- ============================================
-- SAMPLE TEMPLATES
-- ============================================

-- Insert some sample templates
INSERT INTO prompt_templates (name, description, category, template_text, variables, tags, is_public) VALUES
(
    'Customer Support Agent',
    'Professional customer service chatbot prompt',
    'Chat',
    'You are a helpful customer support agent for {{company_name}}. Your role is to assist customers with their inquiries about {{product_or_service}}. 

Key guidelines:
- Be professional, friendly, and empathetic
- Provide accurate information
- If you don''t know something, offer to escalate to a human agent
- Always thank the customer for their patience

Company policies:
{{company_policies}}

How can I help you today?',
    '[{"name": "company_name", "type": "string", "required": true}, {"name": "product_or_service", "type": "string", "required": true}, {"name": "company_policies", "type": "text", "required": false}]'::jsonb,
    '{customer-service, chatbot, support}',
    TRUE
),
(
    'Code Review Assistant',
    'AI code reviewer for pull requests',
    'Code Generation',
    'You are an expert code reviewer. Review the following {{language}} code for:

1. Code quality and best practices
2. Potential bugs or errors
3. Performance issues
4. Security vulnerabilities
5. Readability and maintainability

Code to review:
```{{language}}
{{code}}
```

Provide constructive feedback with specific suggestions for improvement.',
    '[{"name": "language", "type": "string", "required": true}, {"name": "code", "type": "text", "required": true}]'::jsonb,
    '{code-review, programming, development}',
    TRUE
),
(
    'Blog Post Generator',
    'SEO-optimized blog post creator',
    'Content Generation',
    'Write a {{word_count}}-word blog post about "{{topic}}" for {{target_audience}}.

Requirements:
- SEO-optimized with relevant keywords
- Engaging introduction and conclusion
- Use subheadings for better readability
- Include {{num_examples}} practical examples
- Tone: {{tone}}

Keywords to include: {{keywords}}',
    '[{"name": "word_count", "type": "number", "default": 1000}, {"name": "topic", "type": "string", "required": true}, {"name": "target_audience", "type": "string", "required": true}, {"name": "num_examples", "type": "number", "default": 3}, {"name": "tone", "type": "string", "default": "professional"}, {"name": "keywords", "type": "array"}]'::jsonb,
    '{content, blog, seo, writing}',
    TRUE
)
ON CONFLICT DO NOTHING;