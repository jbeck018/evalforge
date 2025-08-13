#!/bin/bash

# EvalForge Integration Test Runner
# This script runs the comprehensive SDK integration test

set -e

echo "🚀 EvalForge Integration Test Runner"
echo "======================================"

# Check if we're in the right directory
if [ ! -f "docker-compose.yml" ]; then
    echo "❌ Error: Please run this script from the EvalForge root directory"
    echo "   cd /Users/jacob/projects/evalforge && ./test_scripts/run_integration_test.sh"
    exit 1
fi

# Check if development environment is running
if ! docker-compose ps | grep -q "Up"; then
    echo "⚠️  Development environment not running. Starting it now..."
    echo "   Running: make dev"
    make dev
    
    echo "⏳ Waiting for services to be ready..."
    sleep 30
    
    # Check if services are healthy
    if ! curl -s http://localhost:8000/health > /dev/null; then
        echo "❌ Backend service not responding. Please check the logs:"
        echo "   docker-compose logs backend"
        exit 1
    fi
    
    echo "✅ Services are ready!"
else
    echo "✅ Development environment is already running"
fi

# Run the integration test
echo ""
echo "🧪 Running SDK Integration Test..."
echo "=================================="

cd "$(dirname "$0")"
if python3 test_real_sdk_integration.py; then
    echo ""
    echo "🎉 Integration test completed successfully!"
    echo ""
    echo "📋 Next Steps:"
    echo "1. Check the test output above for your API key and Project ID"
    echo "2. Use those credentials to run performance tests:"
    echo "   python test_scripts/test_batch_ingestion.py <api_key> <project_id>"
    echo "3. Run error scenario tests:"
    echo "   python test_scripts/test_error_scenarios.py <api_key> <project_id>"
    echo "4. View the data in your database:"
    echo "   docker-compose exec postgres psql -U evalforge -d evalforge"
    echo "   SELECT COUNT(*) FROM trace_events;"
    echo ""
    echo "🌐 You can also check the frontend at: http://localhost:3000"
    echo "   (Log in with the test user credentials from the test output)"
    
else
    echo ""
    echo "❌ Integration test failed!"
    echo ""
    echo "🔍 Troubleshooting steps:"
    echo "1. Check backend logs: docker-compose logs backend"
    echo "2. Check database logs: docker-compose logs postgres"
    echo "3. Verify all services are running: docker-compose ps"
    echo "4. Check API health: curl http://localhost:8000/health"
    echo ""
    exit 1
fi