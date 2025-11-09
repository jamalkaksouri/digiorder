#!/bin/bash

# Rate Limiting Demonstration Script
# This script demonstrates how rate limiting works in the DigiOrder API

BASE_URL="http://localhost:5582"
BOLD='\033[1m'
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BOLD}====================================="
echo -e "DigiOrder Rate Limiting Demonstration"
echo -e "=====================================${NC}\n"

# Function to test rate limiting
test_rate_limit() {
    local endpoint=$1
    local requests=$2
    local description=$3
    
    echo -e "${BOLD}Testing: ${description}${NC}"
    echo -e "Endpoint: ${YELLOW}${endpoint}${NC}"
    echo -e "Requests: ${requests}\n"
    
    success_count=0
    rate_limited_count=0
    
    for ((i=1; i<=requests; i++)); do
        response=$(curl -s -w "%{http_code}" -o /dev/null ${BASE_URL}${endpoint})
        
        if [ "$response" == "200" ]; then
            echo -e "${GREEN}✓${NC} Request $i: Success (200)"
            ((success_count++))
        elif [ "$response" == "429" ]; then
            echo -e "${RED}✗${NC} Request $i: Rate Limited (429)"
            ((rate_limited_count++))
        else
            echo -e "${YELLOW}?${NC} Request $i: Other ($response)"
        fi
        
        # Small delay to show the flow
        sleep 0.01
    done
    
    echo -e "\n${BOLD}Results:${NC}"
    echo -e "  ${GREEN}Successful:${NC} $success_count"
    echo -e "  ${RED}Rate Limited:${NC} $rate_limited_count"
    echo -e "\n"
}

# 1. Test global rate limit (100 req/sec, burst 200)
echo -e "${BOLD}1. Global Rate Limit Test${NC}"
echo -e "   Limit: 100 requests/second, Burst: 200\n"
test_rate_limit "/health" 150 "Health Check Endpoint"

# Wait a bit
echo -e "Waiting 2 seconds to reset limiter...\n"
sleep 2

# 2. Test API rate limit for authenticated users
echo -e "${BOLD}2. API Key Rate Limit Test${NC}"
echo -e "   Limit: 1000 requests/minute for authenticated users\n"

# First, login to get token
echo -e "Logging in to get authentication token..."
TOKEN=$(curl -s -X POST ${BASE_URL}/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123456"
  }' | jq -r '.data.token')

if [ "$TOKEN" != "null" ] && [ -n "$TOKEN" ]; then
    echo -e "${GREEN}✓${NC} Login successful\n"
    
    # Test authenticated endpoint
    echo -e "${BOLD}Testing authenticated endpoint:${NC}"
    success=0
    failed=0
    
    for ((i=1; i<=50; i++)); do
        response=$(curl -s -w "%{http_code}" -o /dev/null \
            -H "Authorization: Bearer $TOKEN" \
            ${BASE_URL}/api/v1/products?limit=5)
        
        if [ "$response" == "200" ]; then
            echo -ne "${GREEN}✓${NC}"
            ((success++))
        else
            echo -ne "${RED}✗${NC}"
            ((failed++))
        fi
        
        if [ $((i % 10)) -eq 0 ]; then
            echo ""
        fi
    done
    
    echo -e "\n\n${BOLD}Results:${NC}"
    echo -e "  ${GREEN}Successful:${NC} $success"
    echo -e "  ${RED}Failed:${NC} $failed"
else
    echo -e "${RED}✗${NC} Login failed - skipping authenticated tests"
fi

echo -e "\n${BOLD}3. Demonstrating Rate Limit Recovery${NC}\n"

echo -e "Sending burst of requests..."
for ((i=1; i<=250; i++)); do
    curl -s -o /dev/null ${BASE_URL}/health &
done
wait

echo -e "${GREEN}✓${NC} Burst complete (250 requests)\n"

echo -e "Checking immediate status:"
response=$(curl -s -w "%{http_code}" -o /dev/null ${BASE_URL}/health)
if [ "$response" == "429" ]; then
    echo -e "${RED}✗${NC} Rate limited (as expected)\n"
else
    echo -e "${GREEN}✓${NC} Request succeeded\n"
fi

echo -e "Waiting 2 seconds for rate limiter to recover..."
sleep 2

echo -e "Checking status after recovery:"
response=$(curl -s -w "%{http_code}" -o /dev/null ${BASE_URL}/health)
if [ "$response" == "200" ]; then
    echo -e "${GREEN}✓${NC} Rate limit recovered - request successful\n"
else
    echo -e "${RED}✗${NC} Still rate limited\n"
fi

echo -e "${BOLD}4. Rate Limit Headers Demonstration${NC}\n"

echo -e "Making request to check rate limit headers:"
curl -s -i ${BASE_URL}/health | grep -i "x-rate" || echo -e "${YELLOW}Note: Rate limit headers not exposed in current implementation${NC}"

echo -e "\n${BOLD}====================================="
echo -e "Demonstration Complete"
echo -e "=====================================${NC}\n"

echo -e "${BOLD}Summary:${NC}"
echo -e "1. Global rate limit protects all endpoints"
echo -e "2. Authenticated users get higher limits"
echo -e "3. Rate limits automatically recover"
echo -e "4. 429 status code indicates rate limiting"
echo -e "\n${BOLD}Configuration:${NC}"
echo -e "- Global: 100 req/sec (burst 200)"
echo -e "- Authenticated: 1000 req/min"
echo -e "- Per-IP tracking with automatic cleanup"