# EvalForge Quick Start & Testing Guide

## 🚀 Quick Start (5 Minutes)

### 1. Prerequisites
- Docker and Docker Compose installed
- Python 3.8+ installed
- Node.js 18+ installed
- 8GB RAM minimum

### 2. Clone and Setup
```bash
# Clone the repository
git clone <your-repo-url> evalforge
cd evalforge

# Run the setup script
./dev/scripts/setup-dev-env.sh

# Start all services
make dev
```

### 3. Verify Services
```bash
# Check all services are running
make status

# Expected output:
# ✅ Backend API: Running
# ✅ Frontend: Running
# ✅ PostgreSQL: Running
# ✅ ClickHouse: Running
# ✅ Redis: Running
```

### 4. Access the Application
- **Frontend Dashboard**: http://localhost:3000
- **API Documentation**: http://localhost:8080/swagger
- **Grafana Monitoring**: http://localhost:3001
- **ClickHouse UI**: http://localhost:8123

## 🧪 Testing the Application

### Automated E2E Testing
```bash
# Run the comprehensive E2E test suite
python test_evalforge_e2e.py

# This will test:
# - Service health checks
# - User registration and authentication
# - Project creation
# - SDK integration
# - Event ingestion
# - Analytics queries
# - Performance under load
# - Error handling
# - Data persistence
```

### Manual Testing Workflow

#### 1. Test User Registration
```bash
# Register a new user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "TestPassword123!",
    "name": "Test User"
  }'
```

#### 2. Test SDK Integration

**Python SDK:**
```python
from evalforge import EvalForge

# Initialize the SDK
ef = EvalForge(api_key="your-api-key")

# Trace an LLM call
with ef.trace("test_completion") as trace:
    # Your LLM call here
    response = "This is a test response"
    
    trace.log_input({"prompt": "What is 2+2?"})
    trace.log_output({"response": response})
    trace.log_metrics({
        "latency_ms": 150,
        "tokens_used": 10,
        "cost": 0.0002
    })
```

**JavaScript SDK:**
```javascript
const { EvalForge } = require('@evalforge/sdk');

// Initialize the SDK
const ef = new EvalForge({ apiKey: 'your-api-key' });

// Trace an LLM call
const trace = ef.trace('test_completion');
try {
  // Your LLM call here
  const response = "This is a test response";
  
  trace.logInput({ prompt: "What is 2+2?" });
  trace.logOutput({ response });
  trace.logMetrics({
    latency_ms: 150,
    tokens_used: 10,
    cost: 0.0002
  });
} finally {
  await trace.end();
}
```

#### 3. Test Dashboard Features
1. Navigate to http://localhost:3000
2. Login with test credentials
3. Create a new project
4. View real-time metrics
5. Test filtering and search
6. Export data as CSV

### Performance Testing
```bash
# Run load test (1000 events)
make perf-test

# Monitor performance metrics
# - Check Grafana dashboards
# - View API response times
# - Monitor database performance
```

### Testing Checklist

#### ✅ Core Functionality
- [ ] User can register and login
- [ ] Projects can be created and managed
- [ ] SDK successfully sends events
- [ ] Events appear in dashboard
- [ ] Analytics queries work correctly
- [ ] Data exports properly

#### ✅ Performance
- [ ] API responds in <100ms (P95)
- [ ] Dashboard loads in <2 seconds
- [ ] Can handle 1000+ events/second
- [ ] No memory leaks after extended use

#### ✅ Error Handling
- [ ] Invalid auth returns 401
- [ ] Rate limiting works (429)
- [ ] Malformed requests return 400
- [ ] Graceful degradation on service failure

#### ✅ Developer Experience
- [ ] Hot reload works for both frontend and backend
- [ ] SDK integration takes <5 minutes
- [ ] Mock LLM services work offline
- [ ] Documentation is clear and helpful

## 🔧 Troubleshooting

### Common Issues

#### Services Won't Start
```bash
# Check Docker is running
docker info

# Clean and restart
make clean
make dev

# Check logs
make logs
```

#### Database Connection Issues
```bash
# Reset databases
make db-reset

# Check connections
make db-status
```

#### SDK Not Sending Events
```bash
# Check API key is valid
curl -H "Authorization: Bearer YOUR_API_KEY" \
  http://localhost:8080/api/v1/projects

# Check network connectivity
curl http://localhost:8080/health
```

#### Performance Issues
```bash
# Check resource usage
docker stats

# Increase Docker resources
# Docker Desktop > Preferences > Resources
# Recommended: 4 CPUs, 8GB RAM
```

## 📊 Monitoring

### Real-time Monitoring
- **Grafana Dashboards**: http://localhost:3001
  - API Performance
  - Database Metrics
  - System Resources
  - Business Metrics

### Logs
```bash
# View all logs
make logs

# View specific service
make logs service=backend
make logs service=frontend

# Follow logs
make logs-follow
```

### Health Checks
```bash
# Check all services
make health-check

# Check specific service
curl http://localhost:8080/health
curl http://localhost:3000
```

## 🚦 Next Steps

1. **Run E2E Tests**: Validate the entire system works
2. **Try SDK Integration**: Integrate with a sample app
3. **Explore Dashboard**: Test all analytics features
4. **Load Testing**: Verify performance at scale
5. **Review Documentation**: Check API docs and guides

## 💡 Tips for Testing

1. **Use Mock LLM Service**: For consistent testing without API costs
2. **Reset Data**: Use `make db-reset` for clean test runs
3. **Monitor Resources**: Keep an eye on Docker stats
4. **Check Logs**: Always check logs when debugging
5. **Test Incrementally**: Start with basic features, then test advanced

## 📞 Getting Help

- **Documentation**: `/docs` directory
- **API Reference**: http://localhost:8080/swagger
- **Logs**: `make logs` for debugging
- **Health Checks**: `make health-check` for service status

Remember: The entire application runs locally without external dependencies, making testing fast and reliable!