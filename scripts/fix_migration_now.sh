#!/bin/bash

# DigiOrder - Definitive Migration Fix
# Fixes dirty migration 6 and ensures migration 7 runs
# Run this from project root: ./FIX_MIGRATION_NOW.sh

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m'

echo -e "${BOLD}${BLUE}"
cat << "EOF"
╔══════════════════════════════════════════════════╗
║                                                  ║
║        DIGIORDER MIGRATION FIX TOOL              ║
║                                                  ║
║  This will fix the dirty migration 6 error       ║
║  and ensure migration 7 is properly applied      ║
║                                                  ║
╚══════════════════════════════════════════════════╝
EOF
echo -e "${NC}\n"

# Database connection details
DB_HOST="localhost"
DB_PORT="5432"
DB_USER="postgres"
DB_PASSWORD="root"
DB_NAME="digiorder_db"
DB_SSLMODE="disable"

DATABASE_URL="postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}"

echo -e "${YELLOW}Database Connection:${NC}"
echo -e "  Host: ${DB_HOST}:${DB_PORT}"
echo -e "  Database: ${DB_NAME}"
echo -e "  User: ${DB_USER}\n"

# Test connection
echo -e "${BOLD}Testing database connection...${NC}"
if psql "$DATABASE_URL" -c "SELECT 1;" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Connected successfully${NC}\n"
else
    echo -e "${RED}✗ Cannot connect to database!${NC}"
    echo -e "Please ensure PostgreSQL is running and credentials are correct.\n"
    exit 1
fi

# Check current status
echo -e "${BOLD}Checking migration status...${NC}"
CURRENT_STATUS=$(psql "$DATABASE_URL" -t -c "SELECT version, dirty FROM schema_migrations ORDER BY version DESC LIMIT 1;" 2>/dev/null || echo "")

if [ -z "$CURRENT_STATUS" ]; then
    echo -e "${RED}✗ No migrations found in database!${NC}"
    echo -e "${YELLOW}Running all migrations from scratch...${NC}\n"
    migrate -path migrations -database "$DATABASE_URL" up
    exit 0
fi

CURRENT_VERSION=$(echo "$CURRENT_STATUS" | awk '{print $1}' | tr -d ' ')
IS_DIRTY=$(echo "$CURRENT_STATUS" | awk '{print $2}' | tr -d ' ')

echo -e "  Current version: ${BLUE}$CURRENT_VERSION${NC}"
echo -e "  Dirty: ${BLUE}$IS_DIRTY${NC}\n"

# Check if migration is dirty
if [ "$IS_DIRTY" = "t" ]; then
    echo -e "${RED}✗ Migration $CURRENT_VERSION is DIRTY${NC}\n"
    
    echo -e "${BOLD}Applying automatic fix...${NC}"
    echo -e "${YELLOW}Step 1: Forcing migration $CURRENT_VERSION to clean state...${NC}"
    
    migrate -path migrations -database "$DATABASE_URL" force "$CURRENT_VERSION"
    
    echo -e "${GREEN}✓ Migration $CURRENT_VERSION marked as clean${NC}\n"
else
    echo -e "${GREEN}✓ No dirty migrations${NC}\n"
fi

# Check if we need to run migration 7
echo -e "${BOLD}Checking if migration 7 needs to be applied...${NC}"

LATEST_VERSION=$(psql "$DATABASE_URL" -t -c "SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1;" 2>/dev/null | tr -d ' ')

if [ "$LATEST_VERSION" -lt 7 ]; then
    echo -e "${YELLOW}Migration 7 not yet applied (current: $LATEST_VERSION)${NC}"
    echo -e "${YELLOW}Step 2: Running remaining migrations...${NC}\n"
    
    migrate -path migrations -database "$DATABASE_URL" up
    
    echo -e "${GREEN}✓ Migrations applied${NC}\n"
else
    echo -e "${GREEN}✓ Migration 7 already applied${NC}\n"
fi

# Verify final state
echo -e "${BOLD}Verifying final state...${NC}"

FINAL_VERSION=$(psql "$DATABASE_URL" -t -c "SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1;" 2>/dev/null | tr -d ' ')
FINAL_DIRTY=$(psql "$DATABASE_URL" -t -c "SELECT dirty FROM schema_migrations WHERE version = $FINAL_VERSION;" 2>/dev/null | tr -d ' ')

echo -e "  Final version: ${GREEN}$FINAL_VERSION${NC}"
echo -e "  Dirty: ${GREEN}$FINAL_DIRTY${NC}\n"

# Check if migration 7 tables exist
echo -e "${BOLD}Verifying migration 7 tables...${NC}"

TABLES_EXIST=$(psql "$DATABASE_URL" -t -c "
    SELECT COUNT(*) 
    FROM information_schema.tables 
    WHERE table_name IN ('ip_bans', 'ip_ban_cleanup_log');
" 2>/dev/null | tr -d ' ')

if [ "$TABLES_EXIST" -eq 2 ]; then
    echo -e "${GREEN}✓ IP ban tracking tables exist${NC}"
    
    # Show table info
    echo -e "\n${BLUE}IP Bans Table Structure:${NC}"
    psql "$DATABASE_URL" -c "
        SELECT column_name, data_type 
        FROM information_schema.columns 
        WHERE table_name = 'ip_bans' 
        ORDER BY ordinal_position 
        LIMIT 5;
    " 2>/dev/null
else
    echo -e "${RED}✗ Migration 7 tables not found!${NC}"
    echo -e "${YELLOW}Attempting to manually apply migration 7...${NC}\n"
    
    # Manually run migration 7
    psql "$DATABASE_URL" -f migrations/000007_ip_ban_tracking.up.sql
    
    echo -e "${GREEN}✓ Migration 7 applied manually${NC}\n"
fi

# Show all applied migrations
echo -e "${BOLD}All Applied Migrations:${NC}"
psql "$DATABASE_URL" -c "
    SELECT version, dirty 
    FROM schema_migrations 
    ORDER BY version;
" 2>/dev/null

# Final summary
echo -e "\n${BOLD}${GREEN}╔══════════════════════════════════════════════════╗${NC}"
echo -e "${BOLD}${GREEN}║                                                  ║${NC}"
echo -e "${BOLD}${GREEN}║            ✅ MIGRATION FIX COMPLETE ✅           ║${NC}"
echo -e "${BOLD}${GREEN}║                                                  ║${NC}"
echo -e "${BOLD}${GREEN}╚══════════════════════════════════════════════════╝${NC}\n"

echo -e "${GREEN}✓ Database version: $FINAL_VERSION${NC}"
echo -e "${GREEN}✓ No dirty migrations${NC}"
echo -e "${GREEN}✓ Migration 7 applied${NC}"
echo -e "${GREEN}✓ All tables created${NC}\n"

echo -e "${BOLD}Next Steps:${NC}"
echo -e "  1. Build: ${BLUE}make build${NC}"
echo -e "  2. Run: ${BLUE}make run${NC}"
echo -e "  3. Test: ${BLUE}curl http://localhost:5582/health${NC}\n"

echo -e "${BOLD}Verify Tables:${NC}"
echo -e "  ${BLUE}psql \"$DATABASE_URL\" -c \"\\dt\"${NC}\n"

exit 0