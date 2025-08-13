# EvalForge Test Scripts

This directory contains comprehensive test scripts for validating the EvalForge platform functionality.

## Quick Start

1. **Start the development environment:**
   ```bash
   cd /Users/jacob/projects/evalforge
   make dev
   ```

2. **Run the basic integration test:**
   ```bash
   python test_scripts/test_real_sdk_integration.py
   ```

## Test Scripts Overview

### Core Integration Tests

- **`test_real_sdk_integration.py`** - Comprehensive SDK integration test
  - Creates test project and API key
  - Tests basic event ingestion
  - Tests batch processing
  - Validates error handling
  - Confirms data persistence

### Performance Tests

- **`test_batch_ingestion.py`** - High-volume and performance testing
  - High-volume event ingestion (1000+ events)
  - Concurrent client testing
  - Burst ingestion patterns
  - Performance metrics collection

### Error Handling Tests

- **`test_error_scenarios.py`** - Comprehensive error scenario testing
  - Network failure handling
  - Invalid credential handling
  - Malformed data processing
  - Rate limiting behavior

## Test Execution Order

1. **Phase 1: Basic Validation**
   ```bash
   python test_scripts/test_real_sdk_integration.py
   ```

2. **Phase 2: Performance Testing** (use API key from Phase 1)
   ```bash
   python test_scripts/test_batch_ingestion.py <api_key> <project_id>
   ```

3. **Phase 3: Error Scenarios** (use API key from Phase 1)
   ```bash
   python test_scripts/test_error_scenarios.py <api_key> <project_id>
   ```

## Expected Results

### Successful Integration Test Output:
```
=== EvalForge SDK Integration Test Suite ===
Testing against: http://localhost:8000
✓ Created test project 1 with API key

=== Testing Basic Event Ingestion ===
✓ Sent event 1: llm_completion (trace_id: ...)
✓ Sent event 2: embedding_generation (trace_id: ...)
✓ Sent event 3: classification (trace_id: ...)
✓ Successfully sent 3 events

=== Testing Batch Ingestion ===
✓ Sent batch of events up to index 9
✓ Sent batch of events up to index 19
✓ Successfully sent batch of 25 events

=== Testing Error Handling ===
✓ Invalid API key test completed (expected to fail gracefully)
✓ Invalid project ID test completed

=== Validating Data Persistence ===
✓ Access control working (validator user correctly denied access)

🎉 All tests completed successfully!
📋 Test Project ID: 1
🔑 API Key: ef_...
```

## Troubleshooting

### Common Issues

1. **Connection Refused**
   - Ensure development environment is running: `make dev`
   - Check if backend is running on port 8000
   - Verify Docker containers are healthy: `docker-compose ps`

2. **Database Connection Errors**
   - Check PostgreSQL container status
   - Verify database schema is initialized
   - Check logs: `docker-compose logs postgres`

3. **SDK Import Errors**
   - Ensure Python SDK dependencies are installed
   - Check if the SDK path is correct
   - Install required packages: `cd sdks/python && pip install -r requirements.txt`

### Debugging

Enable debug mode in any test by setting environment variable:
```bash
EVALFORGE_DEBUG=true python test_scripts/test_real_sdk_integration.py
```

Or add debug output to the SDK client:
```python
client = evalforge.EvalForge(
    # ... other params
    debug=True
)
```

## Test Data Validation

### Database Verification

Connect to PostgreSQL to verify data:
```bash
docker-compose exec postgres psql -U evalforge -d evalforge
```

Query test events:
```sql
SELECT COUNT(*) FROM trace_events;
SELECT operation_type, status, created_at FROM trace_events ORDER BY created_at DESC LIMIT 10;
```

### API Verification

Test API endpoints directly:
```bash
# Get project events (requires authentication)
curl -H "Authorization: Bearer <token>" http://localhost:8000/api/projects/<project_id>/events
```

## Integration with CI/CD

These tests can be integrated into a CI/CD pipeline:

```yaml
# Example GitHub Actions step
- name: Run Integration Tests
  run: |
    make dev-start
    sleep 30  # Wait for services to start
    python test_scripts/test_real_sdk_integration.py
    API_KEY=$(grep "API Key:" test_output | cut -d' ' -f3)
    PROJECT_ID=$(grep "Test Project ID:" test_output | cut -d' ' -f4)
    python test_scripts/test_batch_ingestion.py $API_KEY $PROJECT_ID
```

## Test Coverage

The test scripts cover:

- ✅ **SDK Functionality**: Event creation, batching, flushing
- ✅ **API Integration**: Authentication, project management, event ingestion
- ✅ **Error Handling**: Network failures, invalid credentials, malformed data
- ✅ **Performance**: High-volume ingestion, concurrent clients, burst patterns
- ✅ **Data Persistence**: Database storage, query validation
- ✅ **Rate Limiting**: SDK and API rate limiting behavior

## Adding New Tests

To add new test scenarios:

1. Create a new test class inheriting from a base tester
2. Implement test methods following the naming convention `test_*`
3. Update the main execution flow
4. Add documentation to this README

Example:
```python
def test_new_scenario(self):
    """Test description."""
    print("\n=== Testing New Scenario ===")
    # Test implementation
    self.results["new_scenario"] = {"status": "completed"}
    print("✓ New scenario test completed")
```