#!/bin/bash
# Backend development script with hot reloading

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Starting EvalForge Backend Development Server${NC}"
echo -e "${BLUE}===============================================${NC}"

# Check if Air is installed
if ! command -v air &> /dev/null; then
    echo -e "${YELLOW}⚠️  Air (hot reload) not found. Installing...${NC}"
    go install github.com/cosmtrek/air@latest
fi

# Create backend directory if it doesn't exist
mkdir -p backend

# Create .air.toml configuration if it doesn't exist
if [ ! -f backend/.air.toml ]; then
    echo -e "${BLUE}📝 Creating Air configuration...${NC}"
    cat > backend/.air.toml << 'EOF'
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ./cmd/api"
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata", "frontend", "dev"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html", "yaml", "yml", "json"]
  kill_delay = "0s"
  log = "build-errors.log"
  send_interrupt = false
  stop_on_error = true

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  time = false

[misc]
  clean_on_exit = false

[screen]
  clear_on_rebuild = false
EOF
fi

# Create basic backend structure if it doesn't exist
if [ ! -d backend/cmd/api ]; then
    echo -e "${BLUE}📁 Creating basic backend structure...${NC}"
    mkdir -p backend/{cmd/api,internal/{server,config,handlers},pkg,docs}
    
    # Create a basic main.go
    cat > backend/cmd/api/main.go << 'EOF'
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	r := mux.NewRouter()
	
	// Health check endpoint
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status": "healthy", "service": "evalforge-api", "timestamp": "%s"}`, 
			"2024-01-01T00:00:00Z")
	}).Methods("GET")

	// API v1 routes
	api := r.PathPrefix("/api/v1").Subrouter()
	
	// Events endpoint
	api.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message": "Event ingestion endpoint", "method": "%s"}`, r.Method)
	}).Methods("GET", "POST")

	// Projects endpoint
	api.HandleFunc("/projects", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"projects": [], "total": 0}`)
	}).Methods("GET")

	log.Printf("🚀 EvalForge API server starting on port %s", port)
	log.Printf("🏥 Health check: http://localhost:%s/health", port)
	log.Printf("📚 API endpoints: http://localhost:%s/api/v1/", port)
	
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
EOF

    # Create go.mod if it doesn't exist
    if [ ! -f backend/go.mod ]; then
        echo -e "${BLUE}📦 Initializing Go module...${NC}"
        cd backend
        go mod init evalforge/backend
        go get github.com/gorilla/mux
        cd ..
    fi
fi

# Set environment variables for development
export EVALFORGE_ENV=development
export EVALFORGE_LOG_LEVEL=debug
export EVALFORGE_MOCK_LLMS=true
export PORT=8000

# Change to backend directory and start Air
cd backend

echo -e "${GREEN}🔥 Starting hot-reload development server...${NC}"
echo -e "${YELLOW}💡 The server will automatically restart when you change Go files${NC}"
echo -e "${YELLOW}💡 Press Ctrl+C to stop the server${NC}"
echo ""

# Start Air with configuration
air -c .air.toml