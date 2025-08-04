# EvalForge JavaScript/TypeScript SDK

Zero-overhead LLM observability and evaluation platform for Node.js and browsers.

## Installation

```bash
npm install @evalforge/sdk
```

## Quick Start

### Basic Setup

```typescript
import { configure } from '@evalforge/sdk';

// Configure the SDK
configure({
  apiKey: 'your-evalforge-api-key',
  projectId: 1
});
```

### Automatic OpenAI Tracing

```typescript
import OpenAI from 'openai';
import { wrapOpenAI } from '@evalforge/sdk';

// Wrap your OpenAI client
const WrappedOpenAI = wrapOpenAI(OpenAI);
const openai = new WrappedOpenAI({
  apiKey: 'your-openai-api-key',
});

// Use normally - traces are automatically captured
const completion = await openai.chat.completions.create({
  model: 'gpt-3.5-turbo',
  messages: [{ role: 'user', content: 'Hello, world!' }],
});
```

### Manual Tracing

```typescript
import { EvalForge } from '@evalforge/sdk';

const client = new EvalForge({
  apiKey: 'your-api-key',
  projectId: 1
});

client.trace({
  operationType: 'llm_call',
  input: { prompt: 'Hello, world!' },
  output: { response: 'Hello! How can I help you today?' },
  tokens: { prompt: 10, completion: 15, total: 25 },
  cost: 0.000045,
  provider: 'openai',
  model: 'gpt-3.5-turbo'
});
```

## Features

### 🚀 Zero-Overhead Design
- Asynchronous event processing with automatic batching
- Minimal performance impact on your application
- Automatic retries with exponential backoff
- Graceful degradation when EvalForge is unavailable

### 🔌 Provider Integration
- **OpenAI**: Automatic tracing for chat completions and embeddings
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

```typescript
import { trace } from '@evalforge/sdk';

class LLMService {
  @trace({ operationType: 'data_processing' })
  async processUserQuery(query: string): Promise<string> {
    // Your processing logic
    return processedQuery;
  }

  @trace() // Uses method name as operation type
  async generateResponse(processedQuery: string): Promise<string> {
    // Your generation logic
    return response;
  }
}
```

### Context Manager for Complex Workflows

```typescript
import { TraceContext } from '@evalforge/sdk';

async function handleUserInteraction(userInput: string) {
  const mainContext = new TraceContext('user_interaction', { input: userInput });
  
  try {
    // Step 1: Process query
    const queryContext = mainContext.child('query_processing');
    const processedQuery = await queryContext.run(async () => {
      return processQuery(userInput);
    });
    
    // Step 2: Call LLM
    const llmContext = mainContext.child('llm_call');
    const response = await llmContext.run(async () => {
      return callLLM(processedQuery);
    });
    
    mainContext.setOutput({ response });
    mainContext.finish('success');
    
    return response;
  } catch (error) {
    mainContext.setOutput({ error: error.message });
    mainContext.finish('error');
    throw error;
  }
}
```

### Manual Client Management

```typescript
import { EvalForge } from '@evalforge/sdk';

const client = new EvalForge({
  apiKey: 'your-api-key',
  projectId: 1,
  batchSize: 50,        // Send events in batches of 50
  flushInterval: 2000,  // Flush every 2 seconds
  debug: true           // Enable debug logging
});

// Use the client
client.trace({
  operationType: 'custom_operation',
  input: { data: 'input' },
  output: { result: 'output' },
  metadata: { version: '1.0' }
});

// Clean up when done
await client.close();
```

## Configuration

### Environment Variables

```bash
export EVALFORGE_API_KEY="your-api-key"
export EVALFORGE_PROJECT_ID="1"
export EVALFORGE_BASE_URL="https://api.evalforge.dev"
export EVALFORGE_BATCH_SIZE="100"
export EVALFORGE_FLUSH_INTERVAL="5"  # seconds
export EVALFORGE_DEBUG="true"
```

### Configuration Options

```typescript
import { configure } from '@evalforge/sdk';

configure({
  apiKey: 'your-api-key',          // Required: Your EvalForge API key
  projectId: 1,                    // Required: Project ID
  baseUrl: 'https://api.evalforge.dev',  // API base URL
  batchSize: 100,                  // Events per batch
  flushInterval: 5000,             // Milliseconds between flushes
  maxRetries: 3,                   // Retry attempts
  timeout: 30000,                  // Request timeout in milliseconds
  debug: false                     // Debug logging
});
```

## Error Handling

The SDK is designed to fail gracefully:

```typescript
import { wrapOpenAI } from '@evalforge/sdk';
import OpenAI from 'openai';

// Even if EvalForge is misconfigured, your OpenAI calls will work
const WrappedOpenAI = wrapOpenAI(OpenAI);
const openai = new WrappedOpenAI({ apiKey: 'your-openai-key' });

try {
  const result = await openai.chat.completions.create({
    model: 'gpt-3.5-turbo',
    messages: [{ role: 'user', content: 'Hello!' }],
  });
  // Your result will be returned normally
} catch (error) {
  // Handle OpenAI errors as usual
  // EvalForge errors won't interfere
}
```

## Browser Usage

The SDK works in modern browsers with some limitations:

```html
<!DOCTYPE html>
<html>
<head>
  <script type="module">
    import { configure, EvalForge } from 'https://unpkg.com/@evalforge/sdk/dist/index.esm.js';
    
    configure({
      apiKey: 'your-api-key',
      projectId: 1
    });
    
    const client = new EvalForge({
      apiKey: 'your-api-key',
      projectId: 1
    });
    
    client.trace({
      operationType: 'user_action',
      input: { action: 'button_click' },
      output: { result: 'success' }
    });
  </script>
</head>
</html>
```

**Note**: Direct LLM API calls from browsers are not recommended due to API key exposure. Use server-side proxies instead.

## Performance Considerations

### Background Processing
- Events are queued immediately (< 1ms overhead)
- Background processing handles HTTP requests
- Automatic batching reduces network overhead

### Memory Usage
- Efficient event queuing with automatic flushing
- Configurable batch sizes to control memory usage
- Built-in backpressure handling

### Network Optimization
- HTTP/2 support with connection pooling
- Automatic request compression
- Configurable retry logic with exponential backoff

## TypeScript Support

Full TypeScript support with comprehensive type definitions:

```typescript
import { EvalForge, TraceEvent, TokenUsage } from '@evalforge/sdk';

const client = new EvalForge({
  apiKey: 'your-key',
  projectId: 1
});

const tokens: TokenUsage = {
  prompt: 10,
  completion: 15,
  total: 25
};

const traceId = client.trace({
  operationType: 'llm_call',
  tokens,
  cost: 0.000045
});
```

## Development

### Building

```bash
npm install
npm run build
```

### Testing

```bash
npm test
```

### Linting

```bash
npm run lint
npm run lint:fix
```

## License

MIT License - see LICENSE file for details.

## Support

- Documentation: https://docs.evalforge.dev
- Issues: https://github.com/evalforge/evalforge/issues
- Email: support@evalforge.dev