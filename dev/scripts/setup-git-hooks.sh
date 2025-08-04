#!/bin/bash
# Setup Git hooks for EvalForge development

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔧 Setting up Git hooks for EvalForge...${NC}"
echo -e "${BLUE}======================================${NC}"

# Check if we're in a Git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo -e "${RED}❌ Not in a Git repository${NC}"
    echo -e "${YELLOW}💡 Run this script from the EvalForge project root${NC}"
    exit 1
fi

# Get the Git hooks directory
GIT_HOOKS_DIR=$(git rev-parse --git-dir)/hooks
PROJECT_HOOKS_DIR=".githooks"

echo -e "${BLUE}📁 Git hooks directory: $GIT_HOOKS_DIR${NC}"
echo -e "${BLUE}📁 Project hooks directory: $PROJECT_HOOKS_DIR${NC}"

# Check if .githooks directory exists
if [ ! -d "$PROJECT_HOOKS_DIR" ]; then
    echo -e "${RED}❌ .githooks directory not found${NC}"
    echo -e "${YELLOW}💡 Expected to find hooks in $PROJECT_HOOKS_DIR${NC}"
    exit 1
fi

# Function to install a hook
install_hook() {
    local hook_name=$1
    local source_file="$PROJECT_HOOKS_DIR/$hook_name"
    local target_file="$GIT_HOOKS_DIR/$hook_name"
    
    if [ -f "$source_file" ]; then
        echo -e "${BLUE}🔗 Installing $hook_name hook...${NC}"
        
        # Backup existing hook if it exists
        if [ -f "$target_file" ] && [ ! -L "$target_file" ]; then
            echo -e "${YELLOW}  💾 Backing up existing $hook_name hook${NC}"
            mv "$target_file" "$target_file.backup.$(date +%Y%m%d_%H%M%S)"
        fi
        
        # Create symlink to our hook
        ln -sf "../../$source_file" "$target_file"
        chmod +x "$target_file"
        
        echo -e "${GREEN}  ✅ $hook_name hook installed${NC}"
    else
        echo -e "${YELLOW}  ⚠️  $hook_name hook not found in $PROJECT_HOOKS_DIR${NC}"
    fi
}

# Install available hooks
echo -e "${BLUE}🔨 Installing hooks...${NC}"

# Pre-commit hook
install_hook "pre-commit"

# Pre-push hook (if it exists)
install_hook "pre-push"

# Commit-msg hook (if it exists)
install_hook "commit-msg"

# Post-merge hook (if it exists)
install_hook "post-merge"

echo ""
echo -e "${GREEN}🎉 Git hooks setup complete!${NC}"
echo ""
echo -e "${BLUE}📋 Installed hooks:${NC}"
for hook in pre-commit pre-push commit-msg post-merge; do
    if [ -f "$GIT_HOOKS_DIR/$hook" ]; then
        echo -e "${GREEN}  ✅ $hook${NC}"
    else
        echo -e "${YELLOW}  ➖ $hook (not available)${NC}"
    fi
done

echo ""
echo -e "${BLUE}💡 What these hooks do:${NC}"
echo -e "${CYAN}  • pre-commit:${NC} Runs code quality checks before each commit"
echo -e "${CYAN}  • pre-push:${NC}   Runs tests before pushing to remote"
echo -e "${CYAN}  • commit-msg:${NC} Validates commit message format"
echo -e "${CYAN}  • post-merge:${NC} Runs tasks after merging branches"

echo ""
echo -e "${YELLOW}🔧 To bypass hooks (not recommended):${NC}"
echo -e "${YELLOW}  • Skip pre-commit: git commit --no-verify${NC}"
echo -e "${YELLOW}  • Skip pre-push: git push --no-verify${NC}"

echo ""
echo -e "${GREEN}✨ Your development environment is now even better!${NC}"