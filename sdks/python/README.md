# EvalForge Python SDK

Zero-overhead LLM observability and evaluation platform.

## Quick Start

### Installation

```bash
pip install evalforge
```

### Basic Usage

```python
import evalforge

# Configure the client
evalforge.configure(
    api_key="your-evalforge-api-key",
    project_id=1
)

# Option 1: Automatic tracing with LLM provider wrappers
import openai

# Wrap your OpenAI client
OpenAI = evalforge.wrap_openai(openai.OpenAI)
client = OpenAI(api_key="your-openai-key")

# Use normally - traces are automatically captured
response = client.chat.completions.create(
    model="gpt-3.5-turbo",
    messages=[{"role": "user", "content": "Hello, world!"}]
)

# Option 2: Manual tracing
evalforge.trace(
    operation_type="custom_llm_call",
    input_data={"prompt": "Hello, world!"},
    output_data={"response": "Hello! How can I help you today?"},
    tokens=evalforge.TokenUsage(prompt=10, completion=15),
    cost=0.000045,
    provider="openai",
    model="gpt-3.5-turbo"
)
```

## Features

### 🚀 Zero-Overhead Design
- Asynchronous event processing with background threads
- Local buffering and batching for minimal latency impact
- Automatic retries with exponential backoff
- Graceful degradation when EvalForge is unavailable

### 🔌 LLM Provider Integration
- **OpenAI**: Automatic tracing for chat completions and embeddings
- **Anthropic**: Automatic tracing for Claude messages
- **Custom Providers**: Easy integration with any LLM provider

### 📊 Comprehensive Observability
- Token usage and cost tracking
- Latency and performance metrics
- Error tracking and debugging
- Request/response data capture

### 🎯 Flexible Tracing Options
- Automatic tracing with provider wrappers
- Manual tracing for custom integrations
- Decorator-based tracing for functions
- Context manager for complex workflows

## Advanced Usage

### Decorator-Based Tracing

```python
import evalforge

@evalforge.trace(operation_type="data_processing")
def process_user_query(query: str) -> str:
    # Your processing logic
    return processed_query

@evalforge.trace()  # Uses function name as operation_type
def generate_response(processed_query: str) -> str:
    # Your generation logic
    return response
```

### Context Manager for Complex Workflows

```python
import evalforge

with evalforge.TraceContext("user_interaction") as ctx:
    # Main workflow
    
    with ctx.child("query_processing") as query_ctx:
        processed_query = process_query(user_input)
        query_ctx.set_output({"processed_query": processed_query})
    
    with ctx.child("llm_call") as llm_ctx:
        response = call_llm(processed_query)
        llm_ctx.set_output({"response": response})
```

### Anthropic Integration

```python
import anthropic
import evalforge

# Configure EvalForge
evalforge.configure(api_key="your-key", project_id=1)

# Wrap Anthropic client
Anthropic = evalforge.wrap_anthropic(anthropic.Anthropic)
client = Anthropic(api_key="your-anthropic-key")

# Use normally - traces are captured automatically
response = client.messages.create(
    model="claude-3-sonnet-20240229",
    max_tokens=1024,
    messages=[{"role": "user", "content": "Explain quantum computing"}]
)
```

### Manual Client Usage

```python
import evalforge

# Create client instance
client = evalforge.EvalForge(
    api_key="your-api-key",
    project_id=1,
    batch_size=50,  # Send events in batches of 50
    flush_interval=2.0,  # Flush every 2 seconds
    debug=True  # Enable debug logging
)

# Use with context manager for automatic cleanup
with client:
    client.trace(
        operation_type="custom_operation",
        input_data={"input": "data"},
        output_data={"output": "result"},
        metadata={"version": "1.0"}
    )
    
    # Client will automatically flush when exiting context
```

## Configuration

### Environment Variables

```bash
export EVALFORGE_API_KEY="your-api-key"
export EVALFORGE_PROJECT_ID="1"
export EVALFORGE_BASE_URL="https://api.evalforge.dev"
export EVALFORGE_BATCH_SIZE="100"
export EVALFORGE_FLUSH_INTERVAL="5.0"
export EVALFORGE_DEBUG="true"
```

### Configuration Options

```python
evalforge.configure(
    api_key="your-api-key",          # Required: Your EvalForge API key
    project_id=1,                    # Required: Project ID
    base_url="https://api.evalforge.dev",  # API base URL
    batch_size=100,                  # Events per batch
    flush_interval=5.0,              # Seconds between flushes
    max_retries=3,                   # Retry attempts
    timeout=30.0,                    # Request timeout
    debug=False                      # Debug logging
)
```

## Error Handling

The SDK is designed to fail gracefully:

```python
import evalforge

# If EvalForge is not configured, operations continue normally
try:
    result = some_llm_call()
except Exception as e:
    # Your error handling
    pass

# SDK will not interfere with your application's error handling
```

## Performance Considerations

### Background Processing
All events are processed in a background thread to minimize impact on your application:
- Events are queued immediately (< 1ms overhead)
- Background thread handles batching and HTTP requests
- Automatic backpressure handling prevents memory issues

### Batching and Compression
- Events are batched to reduce HTTP overhead
- Automatic compression for large payloads
- Configurable batch sizes and flush intervals

### Resource Usage
- Minimal memory footprint (< 10MB typical)
- Single background thread for all processing
- Automatic connection pooling and keep-alive

## Development

### Running Tests

```bash
# Install development dependencies
pip install -e ".[dev]"

# Run tests
pytest

# Run tests with coverage
pytest --cov=evalforge --cov-report=html
```

### Code Quality

```bash
# Format code
black evalforge/
isort evalforge/

# Lint code
flake8 evalforge/
mypy evalforge/
```

## License

MIT License - see LICENSE file for details.

## Support

- Documentation: https://docs.evalforge.dev
- Issues: https://github.com/evalforge/evalforge/issues
- Email: support@evalforge.dev