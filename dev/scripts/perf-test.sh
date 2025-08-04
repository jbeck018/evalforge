#!/bin/bash
# Performance testing script for EvalForge

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
API_BASE_URL="http://localhost:8000"
CONCURRENT_USERS=${1:-10}
TEST_DURATION=${2:-30}
RAMP_UP_TIME=${3:-5}

echo -e "${BLUE}⚡ EvalForge Performance Test${NC}"
echo -e "${BLUE}============================${NC}"
echo -e "${CYAN}API Base URL: $API_BASE_URL${NC}"
echo -e "${CYAN}Concurrent Users: $CONCURRENT_USERS${NC}"
echo -e "${CYAN}Test Duration: ${TEST_DURATION}s${NC}"
echo -e "${CYAN}Ramp-up Time: ${RAMP_UP_TIME}s${NC}"
echo ""

# Check if required tools are available
check_dependency() {
    local tool=$1
    local install_hint=$2
    
    if ! command -v "$tool" &> /dev/null; then
        echo -e "${RED}❌ $tool is required but not installed${NC}"
        echo -e "${YELLOW}💡 Install with: $install_hint${NC}"
        return 1
    fi
    return 0
}

echo -e "${BLUE}🔍 Checking dependencies...${NC}"

DEPENDENCIES_OK=true
check_dependency "curl" "built-in on most systems" || DEPENDENCIES_OK=false
check_dependency "jq" "brew install jq (macOS) or apt-get install jq (Ubuntu)" || DEPENDENCIES_OK=false

# Optional but recommended
if ! command -v "ab" &> /dev/null && ! command -v "wrk" &> /dev/null; then
    echo -e "${YELLOW}⚠️  Neither 'ab' nor 'wrk' found. Installing basic load testing...${NC}"
    echo -e "${YELLOW}💡 For better performance testing, install: brew install wrk${NC}"
fi

if [ "$DEPENDENCIES_OK" = false ]; then
    echo -e "${RED}❌ Missing required dependencies${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Dependencies check passed${NC}"
echo ""

# Function to check API health
check_api_health() {
    echo -e "${BLUE}🏥 Checking API health...${NC}"
    
    if curl -s -f "$API_BASE_URL/health" > /dev/null; then
        echo -e "${GREEN}✅ API is healthy${NC}"
        return 0
    else
        echo -e "${RED}❌ API is not responding${NC}"
        echo -e "${YELLOW}💡 Make sure the development environment is running: make dev${NC}"
        return 1
    fi
}

# Function to run a simple load test
run_simple_load_test() {
    local endpoint=$1
    local method=${2:-GET}
    local payload=${3:-""}
    local test_name=$4
    
    echo -e "${BLUE}🚀 Running $test_name...${NC}"
    
    local start_time=$(date +%s)
    local successful_requests=0
    local failed_requests=0
    local total_response_time=0
    
    # Run concurrent requests
    for ((i=1; i<=CONCURRENT_USERS; i++)); do
        (
            local request_start=$(date +%s.%3N)
            
            if [ "$method" = "POST" ] && [ -n "$payload" ]; then
                response=$(curl -s -w "%{http_code},%{time_total}" -X POST \
                    -H "Content-Type: application/json" \
                    -H "Authorization: Bearer ef_dev_test_key" \
                    -d "$payload" \
                    "$API_BASE_URL$endpoint" 2>/dev/null || echo "000,0")
            else
                response=$(curl -s -w "%{http_code},%{time_total}" \
                    "$API_BASE_URL$endpoint" 2>/dev/null || echo "000,0")
            fi
            
            local http_code=$(echo "$response" | tail -c 10 | cut -d',' -f1)
            local response_time=$(echo "$response" | tail -c 10 | cut -d',' -f2)
            
            echo "$http_code,$response_time"
        ) &
        
        # Small delay to simulate ramp-up
        sleep $(echo "scale=2; $RAMP_UP_TIME / $CONCURRENT_USERS" | bc)
    done
    
    # Wait for all background jobs to complete
    wait
    
    echo -e "${GREEN}✅ $test_name completed${NC}"
}

# Function to run Apache Bench test (if available)
run_ab_test() {
    local endpoint=$1
    local test_name=$2
    
    if command -v ab &> /dev/null; then
        echo -e "${BLUE}🔥 Running Apache Bench test: $test_name${NC}"
        
        ab -n $((CONCURRENT_USERS * 10)) -c $CONCURRENT_USERS "$API_BASE_URL$endpoint" 2>/dev/null | \
        grep -E "(Requests per second|Time per request|Transfer rate)" | \
        while read line; do
            echo -e "${CYAN}  $line${NC}"
        done
    fi
}

# Function to run wrk test (if available)
run_wrk_test() {
    local endpoint=$1
    local test_name=$2
    
    if command -v wrk &> /dev/null; then
        echo -e "${BLUE}⚡ Running wrk test: $test_name${NC}"
        
        wrk -t4 -c$CONCURRENT_USERS -d${TEST_DURATION}s --latency "$API_BASE_URL$endpoint" | \
        grep -E "(Requests/sec|Latency|Transfer/sec)" | \
        while read line; do
            echo -e "${CYAN}  $line${NC}"
        done
    fi
}

# Function to monitor system resources during test
monitor_resources() {
    echo -e "${BLUE}📊 System Resource Monitoring${NC}"
    echo -e "${BLUE}=============================${NC}"
    
    # Get Docker container stats
    if command -v docker &> /dev/null; then
        echo -e "${PURPLE}Docker Container Resources:${NC}"
        docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}" \
            evalforge_postgres evalforge_clickhouse evalforge_redis 2>/dev/null || \
            echo -e "${YELLOW}⚠️  Docker containers not running${NC}"
    fi
    
    echo ""
    
    # System resources
    echo -e "${PURPLE}System Resources:${NC}"
    if command -v htop &> /dev/null; then
        echo -e "${CYAN}  CPU & Memory: Run 'htop' for detailed view${NC}"
    fi
    
    # Memory usage
    if [ "$(uname)" = "Darwin" ]; then
        # macOS
        echo -e "${CYAN}  Memory Usage:${NC}"
        vm_stat | grep -E "(Pages free|Pages active|Pages inactive|Pages wired)" | \
        while read line; do
            echo -e "${CYAN}    $line${NC}"
        done
    else
        # Linux
        echo -e "${CYAN}  Memory Usage:${NC}"
        free -h | while read line; do
            echo -e "${CYAN}    $line${NC}"
        done
    fi
}

# Main performance test execution
main() {
    # Check API health first
    if ! check_api_health; then
        exit 1
    fi
    
    echo ""
    echo -e "${BLUE}🎯 Starting Performance Tests${NC}"
    echo -e "${BLUE}=============================${NC}"
    
    # Test 1: Health check endpoint
    echo -e "${YELLOW}Test 1: Health Check Endpoint${NC}"
    run_ab_test "/health" "Health Check"
    run_wrk_test "/health" "Health Check"
    echo ""
    
    # Test 2: Event ingestion (single event)
    echo -e "${YELLOW}Test 2: Single Event Ingestion${NC}"
    EVENT_PAYLOAD='{
        "trace_id": "perf_test_trace_'$(date +%s)'",
        "span_id": "perf_test_span_'$(date +%s)'",
        "project_id": "p1p1p1p1-p1p1-p1p1-p1p1-p1p1p1p1p1p1",
        "model": "gpt-4",
        "tokens": 150,
        "latency_ms": 523,
        "cost_cents": 45,
        "timestamp": "'$(date -u +%Y-%m-%dT%H:%M:%S.%3NZ)'"
    }'
    
    # Use curl for POST requests in a loop
    echo -e "${CYAN}  Running $CONCURRENT_USERS concurrent POST requests...${NC}"
    
    start_time=$(date +%s.%3N)
    successful_requests=0
    failed_requests=0
    
    for ((i=1; i<=CONCURRENT_USERS; i++)); do
        (
            response=$(curl -s -w "%{http_code}" -X POST \
                -H "Content-Type: application/json" \
                -H "Authorization: Bearer ef_dev_test_key" \
                -d "$EVENT_PAYLOAD" \
                "$API_BASE_URL/api/v1/events" 2>/dev/null)
            
            if [[ "$response" =~ 2[0-9][0-9]$ ]]; then
                echo "SUCCESS"
            else
                echo "FAILED:$response"
            fi
        ) &
    done
    
    # Wait for all requests to complete
    wait
    
    end_time=$(date +%s.%3N)
    duration=$(echo "$end_time - $start_time" | bc)
    
    echo -e "${CYAN}  Test completed in ${duration}s${NC}"
    echo ""
    
    # Test 3: Batch event ingestion
    echo -e "${YELLOW}Test 3: Batch Event Ingestion${NC}"
    BATCH_PAYLOAD='{
        "events": ['
    
    # Generate 10 events for batch test
    for ((i=1; i<=10; i++)); do
        if [ $i -gt 1 ]; then
            BATCH_PAYLOAD="$BATCH_PAYLOAD,"
        fi
        BATCH_PAYLOAD="$BATCH_PAYLOAD"'{
            "trace_id": "batch_trace_'$i'_'$(date +%s)'",
            "span_id": "batch_span_'$i'_'$(date +%s)'",
            "project_id": "p1p1p1p1-p1p1-p1p1-p1p1-p1p1p1p1p1p1",
            "model": "gpt-4",
            "tokens": '$((100 + i * 10))',
            "latency_ms": '$((400 + i * 50))',
            "cost_cents": '$((30 + i * 5))'
        }'
    done
    
    BATCH_PAYLOAD="$BATCH_PAYLOAD]}"
    
    echo -e "${CYAN}  Testing batch ingestion with 10 events per request...${NC}"
    
    # Test batch endpoint if it exists
    curl -s -X POST \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer ef_dev_test_key" \
        -d "$BATCH_PAYLOAD" \
        "$API_BASE_URL/api/v1/events/batch" > /dev/null 2>&1 && \
        echo -e "${GREEN}  ✅ Batch ingestion test successful${NC}" || \
        echo -e "${YELLOW}  ⚠️  Batch endpoint not available or failed${NC}"
    
    echo ""
    
    # Monitor system resources
    monitor_resources
    
    echo ""
    echo -e "${GREEN}🎉 Performance testing completed!${NC}"
    echo ""
    echo -e "${BLUE}📋 Summary:${NC}"
    echo -e "${CYAN}  • Health check endpoint tested${NC}"
    echo -e "${CYAN}  • Single event ingestion tested${NC}"
    echo -e "${CYAN}  • Batch event ingestion tested${NC}"
    echo -e "${CYAN}  • System resources monitored${NC}"
    echo ""
    echo -e "${YELLOW}💡 Tips for better performance:${NC}"
    echo -e "${YELLOW}  • Increase connection pooling${NC}"
    echo -e "${YELLOW}  • Optimize database queries${NC}"
    echo -e "${YELLOW}  • Use batch processing for high throughput${NC}"
    echo -e "${YELLOW}  • Monitor and tune memory usage${NC}"
    echo ""
    echo -e "${BLUE}🔍 For detailed profiling, run:${NC}"
    echo -e "${BLUE}  make profile-cpu    # CPU profiling${NC}"
    echo -e "${BLUE}  make profile-mem    # Memory profiling${NC}"
    echo -e "${BLUE}  make profile-trace  # Execution tracing${NC}"
}

# Handle script arguments
case "$1" in
    --help|-h)
        echo "Usage: $0 [concurrent_users] [duration_seconds] [ramp_up_seconds]"
        echo ""
        echo "Examples:"
        echo "  $0                # Default: 10 users, 30s duration, 5s ramp-up"
        echo "  $0 20 60 10       # 20 users, 60s duration, 10s ramp-up"
        echo "  $0 --help         # Show this help"
        exit 0
        ;;
    --monitor)
        monitor_resources
        exit 0
        ;;
    *)
        main
        ;;
esac