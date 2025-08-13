# EvalForge Local Development Guide

## Quick Start (Under 5 Minutes)

```bash
# Clone and setup
git clone https://github.com/evalforge/evalforge.git
cd evalforge

# One command to rule them all
make dev

# You're now running:
# - API Server: http://localhost:8000
# - Frontend: http://localhost:3000
# - Mock LLMs: http://localhost:8080 (OpenAI), http://localhost:8081 (Anthropic)
# - ClickHouse: http://localhost:8123
# - PostgreSQL: localhost:5432
# - Redis: localhost:6379
```

## Development Environment Architecture

### Directory Structure
```
evalforge/
├── backend/                 # Go backend services
│   ├── cmd/
│   │   ├── api/           # Main API server
│   │   ├── worker/        # Background job processor
│   │   └── seed/          # Data seeding tool
│   ├── internal/
│   │   ├── ingestion/     # Event ingestion pipeline
│   │   ├── evaluation/    # Auto-evaluation engine
│   │   ├── storage/       # Database interfaces
│   │   └── mock/          # Mock implementations
│   └── pkg/               # Public Go packages
├── frontend/               # React frontend
│   ├── src/
│   │   ├── components/
│   │   ├── pages/
│   │   ├── hooks/
│   │   └── api/
│   └── public/
├── dev/                    # Development tools
│   ├── mock-llm/          # Mock LLM service
│   ├── data-generator/    # Synthetic data generation
│   └── scripts/           # Utility scripts
├── docker/                 # Docker configurations
│   ├── local/
│   ├── staging/
│   └── production/
└── docs/                   # Documentation
```

### Mock LLM Service

The mock LLM service simulates OpenAI and Anthropic APIs with realistic behavior:

```go
// dev/mock-llm/responses.go
package main

import (
    "crypto/md5"
    "encoding/hex"
    "fmt"
    "strings"
)

// Deterministic response generation based on prompt
func generateResponse(prompt string) string {
    // Hash prompt for consistent responses
    hash := md5.Sum([]byte(prompt))
    hashStr := hex.EncodeToString(hash[:])
    
    // Response templates based on hash
    templates := []string{
        "Based on the analysis, I recommend: %s",
        "The key insight here is: %s",
        "After careful consideration: %s",
        "The optimal approach would be: %s",
    }
    
    // Select template based on hash
    templateIdx := int(hash[0]) % len(templates)
    
    // Generate response content
    wordCount := 50 + int(hash[1])%100
    response := generateWords(hashStr, wordCount)
    
    return fmt.Sprintf(templates[templateIdx], response)
}

// Configurable response behaviors
type MockConfig struct {
    // Latency simulation
    MinLatencyMs int `json:"min_latency_ms"`
    MaxLatencyMs int `json:"max_latency_ms"`
    
    // Error simulation
    ErrorRate     float32 `json:"error_rate"`
    TimeoutRate   float32 `json:"timeout_rate"`
    
    // Cost simulation
    InputTokenCost  float32 `json:"input_token_cost"`
    OutputTokenCost float32 `json:"output_token_cost"`
    
    // Response variations
    ResponseVariations []ResponseVariation `json:"response_variations"`
}
```

### Data Generation Tools

```python
# dev/data-generator/scenarios.py
import random
from datetime import datetime, timedelta
from typing import List, Dict

class ScenarioGenerator:
    """Generate realistic usage patterns for testing"""
    
    def __init__(self):
        self.scenarios = {
            'chatbot': self.generate_chatbot_scenario,
            'code_assistant': self.generate_code_scenario,
            'content_generation': self.generate_content_scenario,
            'rag_pipeline': self.generate_rag_scenario,
        }
    
    def generate_chatbot_scenario(self, duration_hours: int = 24) -> List[Dict]:
        """Simulate realistic chatbot usage patterns"""
        events = []
        now = datetime.now()
        
        # Simulate daily pattern: low at night, peaks at 10am and 3pm
        for hour in range(duration_hours):
            timestamp = now - timedelta(hours=duration_hours-hour)
            
            # Calculate load based on hour of day
            hour_of_day = timestamp.hour
            if 0 <= hour_of_day < 6:
                base_load = 10  # Night time
            elif 9 <= hour_of_day <= 11:
                base_load = 100  # Morning peak
            elif 14 <= hour_of_day <= 16:
                base_load = 120  # Afternoon peak
            else:
                base_load = 50  # Normal hours
            
            # Add some randomness
            num_events = int(base_load * (0.8 + random.random() * 0.4))
            
            for _ in range(num_events):
                events.append(self._create_chat_event(timestamp))
        
        return events
```

### Database Seeding

```sql
-- dev/postgres/seed.sql
-- Create test organizations and projects
INSERT INTO organizations (id, name, created_at) VALUES
  ('org_test_1', 'Acme Corp', NOW()),
  ('org_test_2', 'TechStartup Inc', NOW()),
  ('org_test_3', 'Enterprise Co', NOW());

INSERT INTO projects (id, organization_id, name, created_at) VALUES
  ('proj_acme_chatbot', 'org_test_1', 'Customer Support Bot', NOW()),
  ('proj_acme_analytics', 'org_test_1', 'Analytics Assistant', NOW()),
  ('proj_tech_coder', 'org_test_2', 'Code Generator', NOW()),
  ('proj_enterprise_rag', 'org_test_3', 'Knowledge Base QA', NOW());

-- Create test users with different permission levels
INSERT INTO users (id, email, role, organization_id) VALUES
  ('user_admin', 'admin@example.com', 'admin', 'org_test_1'),
  ('user_dev', 'dev@example.com', 'developer', 'org_test_1'),
  ('user_viewer', 'viewer@example.com', 'viewer', 'org_test_1');
```

### Hot Reloading Configuration

```yaml
# backend/.air.toml
root = "."
tmp_dir = "tmp"

[build]
cmd = "go build -o ./tmp/main ./cmd/api"
bin = "tmp/main"
full_bin = "./tmp/main"
include_ext = ["go", "yaml", "toml"]
exclude_dir = ["tmp", "vendor", "frontend"]
include_dir = []
exclude_file = []
delay = 1000
stop_on_error = true
log = "air_errors.log"

[log]
time = true

[color]
main = "magenta"
watcher = "cyan"
build = "yellow"
runner = "green"
```

### Frontend Development Setup

```typescript
// frontend/vite.config.ts
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8000',
        changeOrigin: true,
      },
      '/ws': {
        target: 'ws://localhost:8000',
        ws: true,
      }
    }
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
      '@components': path.resolve(__dirname, './src/components'),
      '@hooks': path.resolve(__dirname, './src/hooks'),
      '@api': path.resolve(__dirname, './src/api'),
    }
  },
  build: {
    sourcemap: true,
    rollupOptions: {
      output: {
        manualChunks: {
          'react-vendor': ['react', 'react-dom', 'react-router-dom'],
          'ui-vendor': ['@radix-ui/react-dialog', '@radix-ui/react-dropdown-menu'],
          'chart-vendor': ['recharts', 'd3-scale', 'd3-shape'],
        }
      }
    }
  }
})
```

### Local Testing Utilities

```bash
#!/bin/bash
# dev/scripts/test-ingestion.sh
# Test event ingestion at various scales

echo "Testing EvalForge ingestion pipeline..."

# Test single event
echo "1. Testing single event ingestion..."
curl -X POST http://localhost:8000/api/v1/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test_token" \
  -d '{
    "trace_id": "test_trace_1",
    "span_id": "test_span_1",
    "project_id": "proj_acme_chatbot",
    "model": "gpt-4",
    "tokens": 150,
    "latency_ms": 523,
    "cost_cents": 45
  }'

# Test batch ingestion
echo "2. Testing batch ingestion (100 events)..."
python3 - <<EOF
import requests
import json
import time

events = []
for i in range(100):
    events.append({
        "trace_id": f"test_trace_{i}",
        "span_id": f"test_span_{i}",
        "project_id": "proj_acme_chatbot",
        "model": "gpt-4",
        "tokens": 100 + i,
        "latency_ms": 400 + i * 10,
        "cost_cents": 30 + i
    })

start = time.time()
response = requests.post(
    "http://localhost:8000/api/v1/events/batch",
    json={"events": events},
    headers={"Authorization": "Bearer test_token"}
)
elapsed = time.time() - start

print(f"Status: {response.status_code}")
print(f"Time: {elapsed:.2f}s")
print(f"Rate: {100/elapsed:.0f} events/sec")
EOF

# Test load
echo "3. Testing sustained load (1000 events/sec for 10 seconds)..."
go run dev/scripts/loadtest.go \
  --rate=1000 \
  --duration=10s \
  --endpoint=http://localhost:8000/api/v1/events
```

### Debugging Tools

```go
// dev/tools/trace-viewer/main.go
package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "github.com/gorilla/mux"
)

// Simple trace viewer for local development
func main() {
    r := mux.NewRouter()
    
    // Serve trace viewer UI
    r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
    
    // API endpoints
    r.HandleFunc("/api/traces/{traceId}", getTrace).Methods("GET")
    r.HandleFunc("/api/traces/{traceId}/spans", getSpans).Methods("GET")
    r.HandleFunc("/api/traces/{traceId}/waterfall", getWaterfall).Methods("GET")
    
    fmt.Println("Trace viewer running at http://localhost:8090")
    http.ListenAndServe(":8090", r)
}
```

### Performance Profiling

```makefile
# Makefile targets for profiling
.PHONY: profile-cpu profile-mem profile-trace

profile-cpu:
	@echo "Starting CPU profiling..."
	@go test -cpuprofile=cpu.prof -bench=. ./internal/ingestion
	@go tool pprof -http=:8091 cpu.prof

profile-mem:
	@echo "Starting memory profiling..."
	@go test -memprofile=mem.prof -bench=. ./internal/ingestion
	@go tool pprof -http=:8092 mem.prof

profile-trace:
	@echo "Starting execution trace..."
	@go test -trace=trace.out -bench=. ./internal/ingestion
	@go tool trace trace.out
```

### Docker Compose Override for Development

```yaml
# docker-compose.override.yml
# Automatically loaded in development
version: '3.8'

services:
  postgres:
    ports:
      - "5432:5432"
    volumes:
      - ./dev/postgres/init.sql:/docker-entrypoint-initdb.d/01-init.sql
      - ./dev/postgres/seed.sql:/docker-entrypoint-initdb.d/02-seed.sql
  
  clickhouse:
    ports:
      - "8123:8123"
      - "9000:9000"
    volumes:
      - ./dev/clickhouse/users.xml:/etc/clickhouse-server/users.xml
      - ./dev/clickhouse/config.xml:/etc/clickhouse-server/config.xml
  
  # Development-only services
  mailhog:
    image: mailhog/mailhog
    ports:
      - "1025:1025"  # SMTP
      - "8025:8025"  # Web UI
  
  swagger-ui:
    image: swaggerapi/swagger-ui
    ports:
      - "8089:8080"
    environment:
      SWAGGER_JSON_URL: "http://localhost:8000/api/v1/swagger.json"
```

### VS Code Configuration

```json
// .vscode/launch.json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug API Server",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/backend/cmd/api",
      "env": {
        "EVALFORGE_ENV": "development",
        "EVALFORGE_MOCK_LLMS": "true",
        "EVALFORGE_LOG_LEVEL": "debug"
      }
    },
    {
      "name": "Debug Worker",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/backend/cmd/worker"
    },
    {
      "name": "Debug Frontend",
      "type": "chrome",
      "request": "launch",
      "url": "http://localhost:3000",
      "webRoot": "${workspaceFolder}/frontend/src",
      "sourceMaps": true
    }
  ]
}
```

### Git Hooks for Quality

```bash
#!/bin/bash
# .githooks/pre-commit
# Ensure code quality before commits

echo "Running pre-commit checks..."

# Go formatting
echo "Checking Go formatting..."
if ! gofmt -l backend/ | grep -q .; then
    echo "✓ Go formatting OK"
else
    echo "✗ Go formatting issues found. Run: make fmt"
    exit 1
fi

# Go linting
echo "Running Go linter..."
if golangci-lint run backend/...; then
    echo "✓ Go linting OK"
else
    echo "✗ Go linting issues found"
    exit 1
fi

# Frontend linting
echo "Running Frontend linter..."
cd frontend && npm run lint
if [ $? -eq 0 ]; then
    echo "✓ Frontend linting OK"
else
    echo "✗ Frontend linting issues found"
    exit 1
fi

# Run fast tests
echo "Running quick tests..."
go test -short ./backend/...

echo "✓ All pre-commit checks passed!"
```

## Troubleshooting

### Common Issues

```yaml
issue: "ClickHouse connection refused"
solution: |
  1. Check if ClickHouse is running: docker ps | grep clickhouse
  2. Wait for initialization: docker logs evalforge_clickhouse_1
  3. Test connection: curl http://localhost:8123/ping

issue: "Frontend hot reload not working"
solution: |
  1. Check Vite is running: ps aux | grep vite
  2. Clear Vite cache: rm -rf frontend/node_modules/.vite
  3. Restart dev server: cd frontend && npm run dev

issue: "Mock LLM returning same responses"
solution: |
  Mock LLM is deterministic by design. To add variety:
  1. Set MOCK_LLM_VARIETY=true in docker-compose.yml
  2. Or use the /api/mock/config endpoint to adjust behavior

issue: "Database migrations failing"
solution: |
  1. Reset database: make dev-reset
  2. Check migration files: ls backend/migrations/
  3. Run migrations manually: make migrate-up
```

## Performance Testing Locally

```bash
# Quick performance test
make perf-test

# Detailed performance analysis
./dev/scripts/perf-test.sh --detailed

# Generates report showing:
# - Ingestion throughput (events/sec)
# - Query latency (p50, p95, p99)
# - Resource usage (CPU, memory, disk I/O)
# - Bottleneck analysis
```

This setup ensures developers can work completely offline while maintaining realistic behavior for testing and development.