#!/bin/bash

# DigiOrder Troubleshooting Tool
# Diagnoses common issues and provides solutions

set -e

BOLD='\033[1m'
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BOLD}╔════════════════════════════════════════════╗${NC}"
echo -e "${BOLD}║  DigiOrder v3.0 - Troubleshooting Tool    ║${NC}"
echo -e "${BOLD}╚════════════════════════════════════════════╝${NC}\n"

ISSUES_FOUND=0
WARNINGS_FOUND=0

# Function to check and report
check_item() {
    local name=$1
    local command=$2
    local fix=$3
    
    echo -ne "${BLUE}▶${NC} Checking $name... "
    
    if eval "$command" > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC}"
        return 0
    else
        echo -e "${RED}✗${NC}"
        echo -e "  ${YELLOW}Fix:${NC} $fix"
        ((ISSUES_FOUND++))
        return 1
    fi
}

# 1. Check Go Installation
echo -e "${BOLD}1. Environment Checks${NC}"
check_item "Go installation" \
    "go version" \
    "Install Go 1.25+: https://go.dev/dl/"

check_item "Go version >= 1.25" \
    "[ \$(go version | grep -oP 'go\K[0-9.]+' | cut -d. -f1,2 | awk '{print (\$1 >= 1.25)}') -eq 1 ]" \
    "Update Go to version 1.25 or higher"

# 2. Check PostgreSQL
echo -e "\n${BOLD}2. Database Checks${NC}"
check_item "PostgreSQL connection" \
    "psql -h localhost -U postgres -c 'SELECT 1' > /dev/null 2>&1 || docker ps | grep postgres" \
    "Start PostgreSQL: make docker-up OR install locally"

# 3. Check Environment File
echo -e "\n${BOLD}3. Configuration Checks${NC}"
check_item ".env file exists" \
    "[ -f .env ]" \
    "Copy template: cp .env.example .env"

if [ -f .env ]; then
    source .env 2>/dev/null || true
    
    check_item "JWT_SECRET is set" \
        "[ ! -z \"\$JWT_SECRET\" ]" \
        "Generate: openssl rand -base64 64 >> .env"
    
    check_item "JWT_SECRET length >= 32" \
        "[ \${#JWT_SECRET} -ge 32 ]" \
        "Generate stronger secret: openssl rand -base64 64"
    
    check_item "DB_HOST is set" \
        "[ ! -z \"\$DB_HOST\" ]" \
        "Add DB_HOST=localhost to .env"
    
    check_item "DB_PASSWORD is set" \
        "[ ! -z \"\$DB_PASSWORD\" ]" \
        "Add DB_PASSWORD=your_password to .env"
fi

# 4. Check Dependencies
echo -e "\n${BOLD}4. Dependencies${NC}"
check_item "golang-migrate installed" \
    "migrate -version" \
    "Install: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"

check_item "sqlc installed" \
    "sqlc version" \
    "Install: go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest"

# 5. Check Database State
echo -e "\n${BOLD}5. Database State${NC}"

if [ ! -z "$DB_HOST" ] && [ ! -z "$DB_USER" ] && [ ! -z "$DB_PASSWORD" ] && [ ! -z "$DB_NAME" ]; then
    DATABASE_URL="postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT:-5432}/${DB_NAME}?sslmode=${DB_SSLMODE:-disable}"
    
    check_item "Database exists" \
        "psql '$DATABASE_URL' -c 'SELECT 1'" \
        "Create database: createdb -h $DB_HOST -U $DB_USER $DB_NAME"
    
    if psql "$DATABASE_URL" -c 'SELECT 1' > /dev/null 2>&1; then
        # Check for dirty migrations
        DIRTY=$(psql "$DATABASE_URL" -t -c "SELECT version FROM schema_migrations WHERE dirty = true;" 2>/dev/null | tr -d ' ')
        
        if [ ! -z "$DIRTY" ]; then
            echo -e "${RED}✗${NC} Migration is dirty at version $DIRTY"
            echo -e "  ${YELLOW}Fix:${NC} Run: make migrate-fix"
            ((ISSUES_FOUND++))
        else
            echo -e "${GREEN}✓${NC} No dirty migrations"
        fi
        
        # Check migration version
        CURRENT_VERSION=$(psql "$DATABASE_URL" -t -c "SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1;" 2>/dev/null | tr -d ' ')
        LATEST_FILE=$(ls migrations/*.up.sql 2>/dev/null | tail -1 | grep -oP '\d{6}' || echo "0")
        
        if [ ! -z "$CURRENT_VERSION" ]; then
            if [ "$CURRENT_VERSION" -lt "$LATEST_FILE" ]; then
                echo -e "${YELLOW}⚠${NC} Migrations pending (current: $CURRENT_VERSION, latest: $LATEST_FILE)"
                echo -e "  ${YELLOW}Fix:${NC} Run: make migrate-up"
                ((WARNINGS_FOUND++))
            else
                echo -e "${GREEN}✓${NC} Migrations up to date (version: $CURRENT_VERSION)"
            fi
        fi
    fi
else
    echo -e "${YELLOW}⚠${NC} Cannot check database (missing connection details)"
fi

# 6. Check Build
echo -e "\n${BOLD}6. Build Checks${NC}"
check_item "Go modules synced" \
    "go mod verify" \
    "Run: go mod tidy && go mod download"

echo -ne "${BLUE}▶${NC} Checking if code compiles... "
if go build -o /tmp/digiorder_test ./cmd/main.go 2>/dev/null; then
    echo -e "${GREEN}✓${NC}"
    rm -f /tmp/digiorder_test
else
    echo -e "${RED}✗${NC}"
    echo -e "  ${YELLOW}Fix:${NC} Check build errors with: go build ./cmd/main.go"
    ((ISSUES_FOUND++))
fi

# 7. Check Required Files
echo -e "\n${BOLD}7. Required Files${NC}"
check_item "migrations directory" \
    "[ -d migrations ]" \
    "Ensure migrations directory exists"

check_item "internal/middleware/jwt.go" \
    "[ -f internal/middleware/jwt.go ]" \
    "File is missing - should be created by this fix"

check_item "internal/server/security.go" \
    "[ -f internal/server/security.go ]" \
    "File is missing - should be created by this fix"

# 8. Test API (if running)
echo -e "\n${BOLD}8. API Health Check${NC}"
if curl -s http://localhost:5582/health > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} API is running and responding"
    
    HEALTH=$(curl -s http://localhost:5582/health)
    echo -e "  $(echo $HEALTH | jq -r '\"Status: \" + .status + \", Version: \" + .version' 2>/dev/null || echo $HEALTH)"
else
    echo -e "${YELLOW}⚠${NC} API is not running"
    echo -e "  ${YELLOW}Note:${NC} This is normal if you haven't started the server yet"
    echo -e "  ${YELLOW}Start with:${NC} make run"
fi

# Summary
echo -e "\n${BOLD}═══════════════════════════════════════════${NC}"
echo -e "${BOLD}Summary${NC}"
echo -e "${BOLD}═══════════════════════════════════════════${NC}"

if [ $ISSUES_FOUND -eq 0 ] && [ $WARNINGS_FOUND -eq 0 ]; then
    echo -e "${GREEN}✓ All checks passed! System is ready.${NC}\n"
    echo -e "${BOLD}Next steps:${NC}"
    echo -e "  1. Run: ${BLUE}make migrate-up${NC}"
    echo -e "  2. Run: ${BLUE}make run${NC}"
    echo -e "  3. Initialize system via API"
elif [ $ISSUES_FOUND -eq 0 ]; then
    echo -e "${YELLOW}⚠ ${WARNINGS_FOUND} warning(s) found (non-critical)${NC}\n"
    echo -e "System should work but review warnings above."
else
    echo -e "${RED}✗ ${ISSUES_FOUND} issue(s) found${NC}"
    echo -e "${YELLOW}⚠ ${WARNINGS_FOUND} warning(s) found${NC}\n"
    echo -e "${BOLD}Please fix the issues above before proceeding.${NC}"
    exit 1
fi

echo -e "\n${BOLD}Quick Fix Commands:${NC}"
echo -e "  ${BLUE}make migrate-fix${NC}      - Fix dirty migrations"
echo -e "  ${BLUE}make migrate-up${NC}       - Run pending migrations"
echo -e "  ${BLUE}go mod tidy${NC}           - Sync Go dependencies"
echo -e "  ${BLUE}make build${NC}            - Build the application"
echo -e "  ${BLUE}make run${NC}              - Start the server"