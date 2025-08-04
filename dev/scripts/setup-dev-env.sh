#!/bin/bash
# EvalForge Development Environment Setup Script
# This script sets up everything needed for local development

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
NC='\033[0m' # No Color

# Configuration
MIN_GO_VERSION="1.21"
MIN_NODE_VERSION="18"
MIN_DOCKER_VERSION="20.0"

echo -e "${BLUE}"
cat << "EOF"
 ______            _______                      
|  ____|          |  _____|                     
| |____ __   ____ | |____  ___  _ __ __ _  ___  
|  ____|\\ \\ / / _\|| _____|/ _ \\| '_ \\/ _\\|/ _ \\ 
| |_____ \\ V / (_||| |    | (_) | |_) |  _|| __/ 
|_______| \\_/ \\____||_|     \\___/|  __/\\_| |\\___| 
                                |_|              
Development Environment Setup
EOF
echo -e "${NC}"

echo -e "${GREEN}🚀 Setting up your EvalForge development environment${NC}"
echo -e "${GREEN}This will install and configure everything you need!${NC}"
echo ""

# Function to compare versions
version_compare() {
    local version1=$1
    local version2=$2
    
    if [[ "$version1" == "$version2" ]]; then
        return 0
    fi
    
    local IFS=.
    local i ver1=($version1) ver2=($version2)
    
    # Fill empty fields in ver1 with zeros
    for ((i=${#ver1[@]}; i<${#ver2[@]}; i++)); do
        ver1[i]=0
    done
    
    for ((i=0; i<${#ver1[@]}; i++)); do
        if [[ -z ${ver2[i]} ]]; then
            ver2[i]=0
        fi
        if ((10#${ver1[i]} > 10#${ver2[i]})); then
            return 1
        fi
        if ((10#${ver1[i]} < 10#${ver2[i]})); then
            return 2
        fi
    done
    return 0
}

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to install a package on macOS
install_macos() {
    local package=$1
    local brew_name=$2
    
    if command_exists brew; then
        echo -e "${BLUE}📦 Installing $package via Homebrew...${NC}"
        brew install "$brew_name"
    else
        echo -e "${RED}❌ Homebrew not found. Please install Homebrew first:${NC}"
        echo -e "${YELLOW}/bin/bash -c \"\$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\"${NC}"
        return 1
    fi
}

# Function to install a package on Linux
install_linux() {
    local package=$1
    local apt_name=$2
    
    if command_exists apt-get; then
        echo -e "${BLUE}📦 Installing $package via apt-get...${NC}"
        sudo apt-get update && sudo apt-get install -y "$apt_name"
    elif command_exists yum; then
        echo -e "${BLUE}📦 Installing $package via yum...${NC}"
        sudo yum install -y "$apt_name"
    else
        echo -e "${RED}❌ Package manager not found. Please install $package manually.${NC}"
        return 1
    fi
}

# Function to install a package
install_package() {
    local package=$1
    local brew_name=$2
    local apt_name=${3:-$brew_name}
    
    case "$(uname)" in
        Darwin)
            install_macos "$package" "$brew_name"
            ;;
        Linux)
            install_linux "$package" "$apt_name"
            ;;
        *)
            echo -e "${RED}❌ Unsupported operating system. Please install $package manually.${NC}"
            return 1
            ;;
    esac
}

echo -e "${BLUE}🔍 Checking system requirements...${NC}"
echo -e "${BLUE}===================================${NC}"

# Check operating system
OS=$(uname)
echo -e "${CYAN}Operating System: $OS${NC}"

# Check architecture
ARCH=$(uname -m)
echo -e "${CYAN}Architecture: $ARCH${NC}"

echo ""

# Check required dependencies
DEPENDENCIES_OK=true

# Check Docker
echo -e "${BLUE}🐳 Checking Docker...${NC}"
if command_exists docker; then
    DOCKER_VERSION=$(docker --version | grep -oE '[0-9]+\.[0-9]+' | head -1)
    echo -e "${GREEN}✅ Docker $DOCKER_VERSION found${NC}"
    
    if version_compare "$DOCKER_VERSION" "$MIN_DOCKER_VERSION"; then
        case $? in
            2)
                echo -e "${YELLOW}⚠️  Docker version $DOCKER_VERSION is older than recommended $MIN_DOCKER_VERSION${NC}"
                ;;
        esac
    fi
    
    # Check if Docker is running
    if ! docker info >/dev/null 2>&1; then
        echo -e "${YELLOW}⚠️  Docker daemon is not running. Please start Docker.${NC}"
        DEPENDENCIES_OK=false
    fi
else
    echo -e "${RED}❌ Docker not found${NC}"
    echo -e "${YELLOW}💡 Install Docker from: https://docs.docker.com/get-docker/${NC}"
    DEPENDENCIES_OK=false
fi

# Check Docker Compose
echo -e "${BLUE}🐙 Checking Docker Compose...${NC}"
if command_exists docker-compose || docker compose version >/dev/null 2>&1; then
    if command_exists docker-compose; then
        COMPOSE_VERSION=$(docker-compose --version | grep -oE '[0-9]+\.[0-9]+' | head -1)
        echo -e "${GREEN}✅ Docker Compose $COMPOSE_VERSION found${NC}"
    else
        COMPOSE_VERSION=$(docker compose version --short 2>/dev/null || echo "integrated")
        echo -e "${GREEN}✅ Docker Compose (integrated) found${NC}"
    fi
else
    echo -e "${RED}❌ Docker Compose not found${NC}"
    echo -e "${YELLOW}💡 Install Docker Compose from: https://docs.docker.com/compose/install/${NC}"
    DEPENDENCIES_OK=false
fi

# Check Go
echo -e "${BLUE}🐹 Checking Go...${NC}"
if command_exists go; then
    GO_VERSION=$(go version | grep -oE 'go[0-9]+\.[0-9]+' | sed 's/go//')
    echo -e "${GREEN}✅ Go $GO_VERSION found${NC}"
    
    if version_compare "$GO_VERSION" "$MIN_GO_VERSION"; then
        case $? in
            2)
                echo -e "${YELLOW}⚠️  Go version $GO_VERSION is older than required $MIN_GO_VERSION${NC}"
                echo -e "${YELLOW}💡 Please upgrade Go from: https://golang.org/dl/${NC}"
                DEPENDENCIES_OK=false
                ;;
        esac
    fi
else
    echo -e "${RED}❌ Go not found${NC}"
    echo -e "${YELLOW}💡 Install Go from: https://golang.org/dl/${NC}"
    DEPENDENCIES_OK=false
fi

# Check Node.js
echo -e "${BLUE}🟢 Checking Node.js...${NC}"
if command_exists node; then
    NODE_VERSION=$(node --version | sed 's/v//')
    NODE_MAJOR=$(echo $NODE_VERSION | cut -d. -f1)
    echo -e "${GREEN}✅ Node.js $NODE_VERSION found${NC}"
    
    if [ "$NODE_MAJOR" -lt "$MIN_NODE_VERSION" ]; then
        echo -e "${YELLOW}⚠️  Node.js version $NODE_VERSION is older than required $MIN_NODE_VERSION${NC}"
        echo -e "${YELLOW}💡 Please upgrade Node.js from: https://nodejs.org/${NC}"
        DEPENDENCIES_OK=false
    fi
else
    echo -e "${RED}❌ Node.js not found${NC}"
    echo -e "${YELLOW}💡 Install Node.js from: https://nodejs.org/${NC}"
    DEPENDENCIES_OK=false
fi

# Check npm
if command_exists npm; then
    NPM_VERSION=$(npm --version)
    echo -e "${GREEN}✅ npm $NPM_VERSION found${NC}"
else
    echo -e "${RED}❌ npm not found (usually comes with Node.js)${NC}"
    DEPENDENCIES_OK=false
fi

echo ""

# Install optional but recommended tools
echo -e "${BLUE}🔧 Checking optional development tools...${NC}"
echo -e "${BLUE}=======================================${NC}"

# Air (Go hot reload)
if ! command_exists air; then
    echo -e "${YELLOW}📦 Installing Air (Go hot reload)...${NC}"
    go install github.com/cosmtrek/air@latest
    echo -e "${GREEN}✅ Air installed${NC}"
else
    echo -e "${GREEN}✅ Air already installed${NC}"
fi

# golangci-lint
if ! command_exists golangci-lint; then
    echo -e "${YELLOW}📦 Installing golangci-lint...${NC}"
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    echo -e "${GREEN}✅ golangci-lint installed${NC}"
else
    echo -e "${GREEN}✅ golangci-lint already installed${NC}"
fi

# goimports
if ! command_exists goimports; then
    echo -e "${YELLOW}📦 Installing goimports...${NC}"
    go install golang.org/x/tools/cmd/goimports@latest
    echo -e "${GREEN}✅ goimports installed${NC}"
else
    echo -e "${GREEN}✅ goimports already installed${NC}"
fi

# jq (JSON processor)
if ! command_exists jq; then
    echo -e "${YELLOW}📦 Installing jq...${NC}"
    if install_package "jq" "jq" "jq"; then
        echo -e "${GREEN}✅ jq installed${NC}"
    else
        echo -e "${YELLOW}⚠️  Could not install jq automatically${NC}"
    fi
else
    echo -e "${GREEN}✅ jq already installed${NC}"
fi

# curl (should be available on most systems)
if ! command_exists curl; then
    echo -e "${YELLOW}📦 Installing curl...${NC}"
    if install_package "curl" "curl" "curl"; then
        echo -e "${GREEN}✅ curl installed${NC}"
    else
        echo -e "${YELLOW}⚠️  Could not install curl automatically${NC}"
    fi
else
    echo -e "${GREEN}✅ curl already installed${NC}"
fi

echo ""

# Check if we can proceed
if [ "$DEPENDENCIES_OK" = false ]; then
    echo -e "${RED}❌ Some required dependencies are missing or outdated${NC}"
    echo -e "${YELLOW}💡 Please install the missing dependencies and run this script again${NC}"
    exit 1
fi

echo -e "${GREEN}🎉 All dependencies are satisfied!${NC}"
echo ""

# Setup development environment
echo -e "${BLUE}⚙️  Setting up development environment...${NC}"
echo -e "${BLUE}=======================================${NC}"

# Create necessary directories
echo -e "${CYAN}📁 Creating directory structure...${NC}"
mkdir -p backend/{cmd/{api,worker},internal,pkg,docs,bin,tmp}
mkdir -p frontend/{src,public,dist}
mkdir -p dev/{scripts,data,logs}
mkdir -p docs/{api,deployment,development}

# Initialize Go modules if not exists
if [ ! -f "backend/go.mod" ]; then
    echo -e "${CYAN}📦 Initializing Go module...${NC}"
    cd backend
    go mod init evalforge/backend
    cd ..
fi

# Initialize Node.js project if not exists
if [ ! -f "frontend/package.json" ] && [ -d "frontend" ]; then
    echo -e "${CYAN}📦 Setting up Node.js project...${NC}"
    # This will be handled by the frontend development script
fi

# Setup Git hooks
echo -e "${CYAN}🪝 Setting up Git hooks...${NC}"
if [ -f "dev/scripts/setup-git-hooks.sh" ]; then
    ./dev/scripts/setup-git-hooks.sh
else
    echo -e "${YELLOW}⚠️  Git hooks setup script not found${NC}"
fi

# Create environment files
echo -e "${CYAN}🔐 Creating environment configuration...${NC}"

# Backend .env file
if [ ! -f "backend/.env.development" ]; then
    cat > backend/.env.development << 'EOF'
# EvalForge Backend Development Configuration
EVALFORGE_ENV=development
EVALFORGE_LOG_LEVEL=debug
EVALFORGE_MOCK_LLMS=true

# Database URLs
POSTGRES_URL=postgres://evalforge:evalforge_dev@localhost:5432/evalforge?sslmode=disable
CLICKHOUSE_URL=clickhouse://evalforge:evalforge_dev@localhost:9000/evalforge
REDIS_URL=redis://localhost:6379/0

# API Configuration
PORT=8000
API_VERSION=v1

# External Services (Mock in development)
OPENAI_API_KEY=mock_key_for_development
ANTHROPIC_API_KEY=mock_key_for_development

# Observability
JAEGER_ENDPOINT=http://localhost:14268/api/traces
PROMETHEUS_ENABLED=true
EOF
    echo -e "${GREEN}✅ Backend environment file created${NC}"
fi

# Frontend .env file
if [ ! -f "frontend/.env.development" ]; then
    cat > frontend/.env.development << 'EOF'
# EvalForge Frontend Development Configuration
VITE_API_BASE_URL=http://localhost:8000
VITE_WS_URL=ws://localhost:8000
VITE_ENVIRONMENT=development
VITE_ENABLE_DEBUG=true
EOF
    echo -e "${GREEN}✅ Frontend environment file created${NC}"
fi

echo ""

# Final setup
echo -e "${BLUE}🎯 Final setup steps...${NC}"
echo -e "${BLUE}======================${NC}"

# Add helpful aliases to shell profile (optional)
echo -e "${CYAN}💡 Adding helpful aliases...${NC}"

SHELL_PROFILE=""
if [ -f "$HOME/.zshrc" ]; then
    SHELL_PROFILE="$HOME/.zshrc"
elif [ -f "$HOME/.bashrc" ]; then
    SHELL_PROFILE="$HOME/.bashrc"
elif [ -f "$HOME/.bash_profile" ]; then
    SHELL_PROFILE="$HOME/.bash_profile"
fi

if [ -n "$SHELL_PROFILE" ]; then
    # Check if aliases already exist
    if ! grep -q "# EvalForge Development Aliases" "$SHELL_PROFILE"; then
        cat >> "$SHELL_PROFILE" << 'EOF'

# EvalForge Development Aliases
alias ef-dev='make dev'
alias ef-logs='make logs'
alias ef-status='make status'
alias ef-reset='make dev-reset'
alias ef-test='make test'
alias ef-fmt='make fmt'
alias ef-lint='make lint'
EOF
        echo -e "${GREEN}✅ Development aliases added to $SHELL_PROFILE${NC}"
        echo -e "${YELLOW}💡 Run 'source $SHELL_PROFILE' or restart your terminal to use aliases${NC}"
    else
        echo -e "${GREEN}✅ Development aliases already exist${NC}"
    fi
fi

echo ""

# Success message
echo -e "${GREEN}"
cat << "EOF"
🎉 Development Environment Setup Complete! 🎉
EOF
echo -e "${NC}"

echo -e "${WHITE}Your EvalForge development environment is ready!${NC}"
echo ""
echo -e "${BLUE}🚀 Quick Start:${NC}"
echo -e "${CYAN}  1. Start the development environment:${NC}"
echo -e "${WHITE}     make dev${NC}"
echo ""
echo -e "${CYAN}  2. Access the applications:${NC}"
echo -e "${WHITE}     • Frontend:        http://localhost:3000${NC}"
echo -e "${WHITE}     • API Server:      http://localhost:8000${NC}"
echo -e "${WHITE}     • API Docs:        http://localhost:8089${NC}"
echo -e "${WHITE}     • ClickHouse:      http://localhost:8123${NC}"
echo -e "${WHITE}     • Grafana:         http://localhost:3001${NC}"
echo ""
echo -e "${CYAN}  3. Useful commands:${NC}"
echo -e "${WHITE}     • make status      # Check service status${NC}"
echo -e "${WHITE}     • make logs        # View service logs${NC}"
echo -e "${WHITE}     • make test        # Run tests${NC}"
echo -e "${WHITE}     • make fmt         # Format code${NC}"
echo -e "${WHITE}     • make help        # See all commands${NC}"
echo ""
echo -e "${BLUE}📚 Documentation:${NC}"
echo -e "${WHITE}   • Development Guide: local_development_guide.md${NC}"
echo -e "${WHITE}   • Architecture:      architecture.md${NC}"
echo -e "${WHITE}   • API Documentation: http://localhost:8089${NC}"
echo ""
echo -e "${YELLOW}💡 Pro Tips:${NC}"
echo -e "${YELLOW}   • Use 'make troubleshoot' if you encounter issues${NC}"
echo -e "${YELLOW}   • VS Code extensions will be suggested when you open the project${NC}"
echo -e "${YELLOW}   • Git hooks are automatically set up for code quality${NC}"
echo ""
echo -e "${GREEN}Happy coding! 🚀${NC}"