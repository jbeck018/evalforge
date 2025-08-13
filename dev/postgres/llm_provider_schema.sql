-- LLM Provider Management Schema

-- Table for storing LLM provider configurations
CREATE TABLE IF NOT EXISTS llm_providers (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    provider VARCHAR(50) NOT NULL, -- openai, anthropic, google, azure, huggingface, custom
    api_key TEXT NOT NULL, -- encrypted in production
    api_base TEXT,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Table for storing model configurations
CREATE TABLE IF NOT EXISTS llm_models (
    id SERIAL PRIMARY KEY,
    provider_id INTEGER NOT NULL REFERENCES llm_providers(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    max_tokens INTEGER DEFAULT 4096,
    cost_per_1k_input DECIMAL(10, 6) DEFAULT 0,
    cost_per_1k_output DECIMAL(10, 6) DEFAULT 0,
    supports_functions BOOLEAN DEFAULT false,
    supports_vision BOOLEAN DEFAULT false,
    enabled BOOLEAN DEFAULT true,
    rate_limit INTEGER, -- requests per minute
    timeout INTEGER, -- timeout in seconds
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_llm_providers_user_id ON llm_providers(user_id);
CREATE INDEX idx_llm_providers_enabled ON llm_providers(enabled);
CREATE INDEX idx_llm_models_provider_id ON llm_models(provider_id);
CREATE INDEX idx_llm_models_enabled ON llm_models(enabled);

-- Comments
COMMENT ON TABLE llm_providers IS 'Stores LLM provider configurations for each user';
COMMENT ON TABLE llm_models IS 'Stores model configurations for each provider';
COMMENT ON COLUMN llm_providers.api_key IS 'API key for the provider (should be encrypted in production)';
COMMENT ON COLUMN llm_models.cost_per_1k_input IS 'Cost per 1000 input tokens in USD';
COMMENT ON COLUMN llm_models.cost_per_1k_output IS 'Cost per 1000 output tokens in USD';