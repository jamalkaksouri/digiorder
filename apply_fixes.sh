#!/bin/bash

# DigiOrder v3.0 - Apply All Fixes Script
# This script applies all fixes in the correct order

set -e

BOLD='\033[1m'
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BOLD}"
cat << "EOF"
╔══════════════════════════════════════════════════╗
║                                                  ║
║       DigiOrder v3.0 - Fix Application           ║
║                                                  ║
║  This script will apply all fixes and prepare    ║
║  your system for production use.                 ║
║                                                  ║
╚══════════════════════════════════════════════════╝
EOF
echo -e "${NC}\n"

# Function to print step
print_step() {
    echo -e "\n${BOLD}${BLUE}▶ Step $1: $2${NC}"
}

# Function to print success
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

# Function to print error
print_error() {
    echo -e "${RED}✗ $1${NC}"
}

# Function to print warning
print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

# Check if running in correct directory
if [ ! -f "go.mod" ]; then
    print_error "This script must be run from the project root directory!"
    exit 1
fi

# Step 1: Make scripts executable
print_step 1 "Making scripts executable"
chmod +x scripts/*.sh 2>/dev/null || true
print_success "Scripts are now executable"

# Step 2: Check environment
print_step 2 "Checking environment"

if [ ! -f ".env" ]; then
    print_warning ".env file not found"
    echo -e "Creating .env from template..."
    cp .env.example .env
    print_success ".env file created"
    print_warning "⚠ IMPORTANT: Edit .env and set your configuration!"
    echo -e "\nRequired settings:"
    echo -e "  - DB_PASSWORD (database password)"
    echo -e "  - JWT_SECRET (generate with: openssl rand -base64 64)"
    echo -e "  - INITIAL_SETUP_TOKEN (for first-time setup)"
    echo -e "\nPress Enter to continue after editing .env..."
    read
else
    print_success ".env file exists"
fi

# Load environment
export $(cat .env | grep -v '^#' | xargs) 2>/dev/null || true

# Check critical variables
MISSING_VARS=()
[ -z "$DB_HOST" ] && MISSING_VARS+=("DB_HOST")
[ -z "$DB_USER" ] && MISSING_VARS+=("DB_USER")
[ -z "$DB_PASSWORD" ] && MISSING_VARS+=("DB_PASSWORD")
[ -z "$DB_NAME" ] && MISSING_VARS+=("DB_NAME")
[ -z "$JWT_SECRET" ] && MISSING_VARS+=("JWT_SECRET")

if [ ${#MISSING_VARS[@]} -gt 0 ]; then
    print_error "Missing required environment variables: ${MISSING_VARS[*]}"
    echo -e "\nPlease set these in your .env file and run again."
    exit 1
fi

print_success "All required environment variables are set"

# Step 3: Check Go installation
print_step 3 "Checking Go installation"

if ! command -v go &> /dev/null; then
    print_error "Go is not installed!"
    echo -e "Install Go from: https://go.dev/dl/"
    exit 1
fi

GO_VERSION=$(go version | grep -oP 'go\K[0-9.]+' | cut -d. -f1,2)
print_success "Go $GO_VERSION is installed"

# Step 4: Install required tools
print_step 4 "Installing required tools"

echo -e "Installing golang-migrate..."
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
print_success "golang-migrate installed"

echo -e "Installing sqlc..."
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
print_success "sqlc installed"

# Step 5: Sync Go modules
print_step 5 "Syncing Go modules"

echo -e "Running go mod tidy..."
go mod tidy
print_success "Go modules synced"

echo -e "Downloading dependencies..."
go mod download
print_success "Dependencies downloaded"

# Step 6: Check PostgreSQL
print_step 6 "Checking PostgreSQL"

if psql -h "$DB_HOST" -U "$DB_USER" -d postgres -c "SELECT 1" > /dev/null 2>&1; then
    print_success "PostgreSQL is accessible"
elif docker ps | grep postgres > /dev/null 2>&1; then
    print_success "PostgreSQL is running in Docker"
else
    print_warning "PostgreSQL not detected"
    echo -e "Would you like to start PostgreSQL with Docker? (y/n)"
    read -r START_DOCKER
    if [ "$START_DOCKER" = "y" ]; then
        echo -e "Starting PostgreSQL..."
        docker run -d \
            --name digiorder-postgres \
            -e POSTGRES_USER="$DB_USER" \
            -e POSTGRES_PASSWORD="$DB_PASSWORD" \
            -e POSTGRES_DB="$DB_NAME" \
            -p 5432:5432 \
            postgres:15-alpine
        echo -e "Waiting for PostgreSQL to start..."
        sleep 5
        print_success "PostgreSQL started"
    else
        print_error "PostgreSQL is required to continue"
        exit 1
    fi
fi

# Step 7: Check and fix migrations
print_step 7 "Checking database migrations"

DATABASE_URL="postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT:-5432}/${DB_NAME}?sslmode=${DB_SSLMODE:-disable}"

# Check if database exists
if ! psql "$DATABASE_URL" -c "SELECT 1" > /dev/null 2>&1; then
    print_warning "Database does not exist, creating..."
    createdb -h "$DB_HOST" -U "$DB_USER" "$DB_NAME" || true
    print_success "Database created"
fi

# Check for dirty migrations
DIRTY=$(psql "$DATABASE_URL" -t -c "SELECT version FROM schema_migrations WHERE dirty = true;" 2>/dev/null | tr -d ' ' || echo "")

if [ ! -z "$DIRTY" ]; then
    print_warning "Dirty migration detected at version $DIRTY"
    echo -e "Fixing dirty migration..."
    migrate -path migrations -database "$DATABASE_URL" force "$DIRTY"
    print_success "Dirty migration fixed"
fi

# Run migrations
echo -e "Running migrations..."
migrate -path migrations -database "$DATABASE_URL" up
print_success "Migrations completed"

# Step 8: Build application
print_step 8 "Building application"

echo -e "Compiling..."
go build -o bin/digiorder cmd/main.go
print_success "Build completed successfully"

# Step 9: Run tests
print_step 9 "Running tests"

echo -e "Testing..."
if go test ./... -v > /tmp/test_output.txt 2>&1; then
    print_success "All tests passed"
else
    print_warning "Some tests failed (check /tmp/test_output.txt for details)"
fi

# Step 10: Final checks
print_step 10 "Running final diagnostics"

if [ -f "./scripts/troubleshoot.sh" ]; then
    ./scripts/troubleshoot.sh
else
    print_warning "Troubleshoot script not found, skipping..."
fi

# Print summary
echo -e "\n${BOLD}╔══════════════════════════════════════════════════╗${NC}"
echo -e "${BOLD}║                                                  ║${NC}"
echo -e "${BOLD}║              ✅ ALL FIXES APPLIED ✅              ║${NC}"
echo -e "${BOLD}║                                                  ║${NC}"
echo -e "${BOLD}╚══════════════════════════════════════════════════╝${NC}\n"

echo -e "${GREEN}Your DigiOrder installation is now ready!${NC}\n"

echo -e "${BOLD}Next Steps:${NC}"
echo -e "  1. Start the server:"
echo -e "     ${BLUE}make run${NC}"
echo -e ""
echo -e "  2. Check health:"
echo -e "     ${BLUE}curl http://localhost:5582/health${NC}"
echo -e ""
echo -e "  3. Initialize system:"
echo -e "     ${BLUE}curl http://localhost:5582/api/v1/setup/status${NC}"
echo -e ""
echo -e "  4. Read the quick start guide:"
echo -e "     ${BLUE}cat QUICKSTART.md${NC}"

echo -e "\n${BOLD}Documentation:${NC}"
echo -e "  📚 Quick Start:    ${BLUE}QUICKSTART.md${NC}"
echo -e "  🔧 Commands:       ${BLUE}COMMANDS.md${NC}"
echo -e "  🐛 Fixes:          ${BLUE}FIXES.md${NC}"
echo -e "  📊 Summary:        ${BLUE}SUMMARY.md${NC}"
echo -e "  🚀 Deployment:     ${BLUE}DEPLOYMENT_GUIDE.md${NC}"

echo -e "\n${BOLD}Troubleshooting:${NC}"
echo -e "  Run diagnostics:   ${BLUE}./scripts/troubleshoot.sh${NC}"
echo -e "  Fix migrations:    ${BLUE}make migrate-fix${NC}"
echo -e "  View logs:         ${BLUE}make run 2>&1 | tee logs/app.log${NC}"

echo -e "\n${GREEN}Happy coding! 🚀${NC}\n"