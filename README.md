# EvalForge - LLM Observability & Evaluation Platform

EvalForge is a comprehensive observability and evaluation platform for Large Language Model (LLM) applications. It provides real-time monitoring, automatic evaluation, and optimization suggestions for your AI-powered applications.

## Features

### Core Capabilities
- **Real-time Observability**: Monitor LLM API calls, latency, costs, and errors in real-time
- **Automatic Evaluation**: Automatically evaluate prompt quality and generate test cases
- **Performance Analytics**: Track token usage, costs, latency distributions, and error rates
- **Prompt Optimization**: Get AI-powered suggestions for improving prompt performance
- **WebSocket Support**: Real-time metrics streaming for live dashboards
- **Multi-Provider Support**: Works with OpenAI, Anthropic, and other LLM providers

### Key Components
- **Event Ingestion**: High-throughput event collection with batch processing
- **Analytics Engine**: Real-time and historical analytics with cost tracking
- **Evaluation System**: Automated prompt analysis and test generation
- **SDKs**: Python SDK for easy integration (Node.js coming soon)
- **Web Dashboard**: React-based UI with real-time updates

## Quick Start

### Prerequisites
- Docker and Docker Compose
- Python 3.8+ (for SDK)
- Node.js 20+ (for local frontend development)

### Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/evalforge.git
cd evalforge
```

2. Set up environment variables:
```bash
cp .env.example .env
# Edit .env with your configuration
```

3. Start the platform:
```bash
docker-compose up -d
```

4. Access the dashboard:
- Web UI: http://localhost:3000
- API: http://localhost:8088
- Swagger Docs: http://localhost:8089

### Using the Python SDK

1. Install the SDK:
```bash
pip install evalforge
# Or from source:
cd sdks/python && pip install -e .
```

2. Initialize in your application:
```python
from evalforge import EvalForge

# Initialize the client
ef = EvalForge(
    api_key="your-api-key",
    base_url="http://localhost:8088"  # Optional, defaults to cloud endpoint
)

# Track LLM calls
with ef.trace("chat-completion") as trace:
    # Your LLM call here
    response = openai.ChatCompletion.create(
        model="gpt-3.5-turbo",
        messages=[{"role": "user", "content": "Hello!"}]
    )
    
    # Log the interaction
    trace.log_llm_call(
        provider="openai",
        model="gpt-3.5-turbo",
        input_tokens=10,
        output_tokens=20,
        cost=0.0003,
        latency_ms=250
    )
```

## Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Application   │────▶│   EvalForge     │────▶│    Dashboard    │
│   with SDK      │     │      API        │     │   (React UI)    │
└─────────────────┘     └─────────────────┘     └─────────────────┘
                               │
                    ┌──────────┴──────────┐
                    │                     │
             ┌──────▼──────┐      ┌──────▼──────┐
             │  PostgreSQL  │      │  ClickHouse │
             │  (Metadata)  │      │   (Events)  │
             └──────────────┘      └─────────────┘
```

### Components

- **Backend (Go)**: High-performance API server with WebSocket support
- **Frontend (React)**: Modern dashboard with real-time updates
- **PostgreSQL**: Stores projects, evaluations, and metadata
- **ClickHouse**: Time-series storage for events and metrics
- **Redis**: Caching and session management
- **Mock LLM**: Local testing without API keys

## API Documentation

### Authentication
All API requests require a Bearer token:
```bash
curl -H "Authorization: Bearer YOUR_API_KEY" http://localhost:8088/api/projects
```

### Key Endpoints

#### Projects
- `GET /api/projects` - List all projects
- `POST /api/projects` - Create a new project
- `GET /api/projects/:id` - Get project details
- `DELETE /api/projects/:id` - Delete a project

#### Events
- `POST /api/events` - Ingest events (batch)
- `GET /api/projects/:id/events` - Get project events
- `GET /api/projects/:id/traces` - Get traces

#### Analytics
- `GET /api/projects/:id/analytics/summary` - Get analytics summary
- `GET /api/projects/:id/analytics/costs` - Cost breakdown
- `GET /api/projects/:id/analytics/latency` - Latency distribution
- `GET /api/projects/:id/analytics/errors` - Error analysis

#### Evaluations
- `POST /api/projects/:id/evaluations` - Create evaluation
- `GET /api/projects/:id/evaluations` - List evaluations
- `POST /api/evaluations/:id/run` - Run evaluation
- `GET /api/evaluations/:id/metrics` - Get evaluation metrics

## Configuration

### Environment Variables

Create a `.env` file with the following variables:

```bash
# Database
POSTGRES_USER=evalforge
POSTGRES_PASSWORD=your-secure-password
POSTGRES_DB=evalforge

# Redis
REDIS_PASSWORD=your-redis-password

# API Keys (optional)
ANTHROPIC_API_KEY=your-anthropic-key
OPENAI_API_KEY=your-openai-key

# Application
JWT_SECRET=your-jwt-secret
API_PORT=8088
FRONTEND_PORT=3000
```

### Docker Compose Services

The platform includes the following services:
- `backend`: Go API server
- `frontend`: React dashboard
- `postgres`: PostgreSQL database
- `clickhouse`: Time-series database
- `redis`: Cache and sessions
- `mock-llm`: Local LLM mock for testing
- `prometheus`: Metrics collection
- `grafana`: Metrics visualization
- `jaeger`: Distributed tracing

## Development

### Backend Development

```bash
cd backend
go mod download
go run main.go
```

### Frontend Development

```bash
cd frontend
npm install
npm run dev
```

### Running Tests

```bash
# Run production readiness tests
python test_production_readiness.py

# Run SDK tests
cd sdks/python
pytest tests/
```

## Deployment

### Docker Deployment

1. Build images:
```bash
docker-compose build
```

2. Start services:
```bash
docker-compose up -d
```

3. Check health:
```bash
curl http://localhost:8088/health
```

### Kubernetes Deployment

See [deployment/kubernetes/README.md](deployment/kubernetes/README.md) for Kubernetes deployment instructions.

### Cloud Deployment

- **AWS**: Use ECS or EKS with the provided Docker images
- **GCP**: Deploy to Cloud Run or GKE
- **Azure**: Use Container Instances or AKS

## Security Considerations

- Always use HTTPS in production
- Rotate API keys regularly
- Use strong passwords for databases
- Enable rate limiting for API endpoints
- Configure CORS appropriately
- Use environment variables for secrets
- Enable audit logging

## Monitoring

The platform includes built-in monitoring:
- Prometheus metrics at `/metrics`
- Health checks at `/health`
- Grafana dashboards at port 3001
- Jaeger tracing at port 16686

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) file for details

## Support

- Documentation: [docs.evalforge.ai](https://docs.evalforge.ai)
- Issues: [GitHub Issues](https://github.com/yourusername/evalforge/issues)
- Discord: [Join our community](https://discord.gg/evalforge)

## Roadmap

- [ ] Node.js SDK
- [ ] Java SDK
- [ ] Advanced A/B testing
- [ ] Custom evaluation metrics
- [ ] Multi-model comparison
- [ ] Cost optimization recommendations
- [ ] Export to common formats (CSV, Parquet)
- [ ] Slack/PagerDuty integrations