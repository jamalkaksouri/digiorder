#!/bin/bash

# Barcode Support Demonstration Script
# Shows complete barcode lifecycle: generation, scanning, management

BASE_URL="http://localhost:5582"
BOLD='\033[1m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${BOLD}======================================"
echo -e "DigiOrder Barcode System Demonstration"
echo -e "======================================${NC}\n"

# Login first
echo -e "${BOLD}Step 1: Authentication${NC}"
TOKEN=$(curl -s -X POST ${BASE_URL}/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123456"
  }' | jq -r '.data.token')

if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
    echo -e "${RED}✗ Login failed${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Authenticated successfully${NC}\n"

# Create a test product
echo -e "${BOLD}Step 2: Creating Test Product${NC}"
PRODUCT_RESPONSE=$(curl -s -X POST ${BASE_URL}/api/v1/products \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "اموکسی سیلین 500mg",
    "brand": "Bayer",
    "dosage_form_id": 1,
    "strength": "500mg",
    "unit": "قرص",
    "category_id": 1,
    "description": "آنتی بیوتیک"
  }')

PRODUCT_ID=$(echo $PRODUCT_RESPONSE | jq -r '.data.id')
echo -e "${GREEN}✓ Product created: $PRODUCT_ID${NC}"
echo -e "  Name: $(echo $PRODUCT_RESPONSE | jq -r '.data.name')"
echo -e "\n"

# Generate and add barcodes
echo -e "${BOLD}Step 3: Adding Barcodes to Product${NC}\n"

# EAN-13 barcode (European Article Number)
echo -e "${BLUE}3.1 Adding EAN-13 Barcode${NC}"
BARCODE1=$(curl -s -X POST ${BASE_URL}/api/v1/barcodes \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"product_id\": \"$PRODUCT_ID\",
    \"barcode\": \"5901234123457\",
    \"barcode_type\": \"EAN-13\"
  }")
echo -e "${GREEN}✓ EAN-13 barcode added${NC}"
echo $BARCODE1 | jq '{id, barcode, barcode_type}'
echo ""

# UPC-A barcode (Universal Product Code)
echo -e "${BLUE}3.2 Adding UPC-A Barcode${NC}"
BARCODE2=$(curl -s -X POST ${BASE_URL}/api/v1/barcodes \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"product_id\": \"$PRODUCT_ID\",
    \"barcode\": \"012345678905\",
    \"barcode_type\": \"UPC-A\"
  }")
echo -e "${GREEN}✓ UPC-A barcode added${NC}"
echo $BARCODE2 | jq '{id, barcode, barcode_type}'
echo ""

# Code128 barcode (alphanumeric)
echo -e "${BLUE}3.3 Adding Code128 Barcode${NC}"
BARCODE3=$(curl -s -X POST ${BASE_URL}/api/v1/barcodes \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"product_id\": \"$PRODUCT_ID\",
    \"barcode\": \"AMOX500-BAYER-2024\",
    \"barcode_type\": \"Code128\"
  }")
echo -e "${GREEN}✓ Code128 barcode added${NC}"
echo $BARCODE3 | jq '{id, barcode, barcode_type}'
echo ""

# List all barcodes for product
echo -e "${BOLD}Step 4: Listing Product Barcodes${NC}"
BARCODES=$(curl -s -H "Authorization: Bearer $TOKEN" \
  ${BASE_URL}/api/v1/products/${PRODUCT_ID}/barcodes)
echo -e "Product has $(echo $BARCODES | jq '.data | length') barcodes:\n"
echo $BARCODES | jq -r '.data[] | "  - \(.barcode_type): \(.barcode)"'
echo ""

# Demonstrate barcode scanning
echo -e "${BOLD}Step 5: Barcode Scanning Demo${NC}\n"

scan_barcode() {
    local barcode=$1
    local barcode_type=$2
    
    echo -e "${BLUE}Scanning: $barcode ($barcode_type)${NC}"
    RESULT=$(curl -s -H "Authorization: Bearer $TOKEN" \
      ${BASE_URL}/api/v1/products/barcode/${barcode})
    
    if echo $RESULT | jq -e '.data' > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Product found!${NC}"
        echo $RESULT | jq -r '.data | "  Product: \(.name)\n  Brand: \(.brand)\n  Strength: \(.strength)"'
    else
        echo -e "${YELLOW}✗ Product not found${NC}"
    fi
    echo ""
}

scan_barcode "5901234123457" "EAN-13"
scan_barcode "012345678905" "UPC-A"
scan_barcode "AMOX500-BAYER-2024" "Code128"

# Test invalid barcode
echo -e "${BLUE}Scanning invalid barcode: 9999999999999${NC}"
INVALID=$(curl -s -w "%{http_code}" -o /dev/null \
  -H "Authorization: Bearer $TOKEN" \
  ${BASE_URL}/api/v1/products/barcode/9999999999999)
if [ "$INVALID" == "404" ]; then
    echo -e "${GREEN}✓ Correctly returned 404 for invalid barcode${NC}\n"
fi

# Update a barcode
echo -e "${BOLD}Step 6: Updating Barcode${NC}"
BARCODE1_ID=$(echo $BARCODE1 | jq -r '.data.id')
UPDATED=$(curl -s -X PUT ${BASE_URL}/api/v1/barcodes/${BARCODE1_ID} \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "barcode": "5901234123458",
    "barcode_type": "EAN-13"
  }')
echo -e "${GREEN}✓ Barcode updated${NC}"
echo $UPDATED | jq '{id, barcode, barcode_type}'
echo ""

# Search barcodes
echo -e "${BOLD}Step 7: Barcode Search${NC}"
echo -e "Searching for barcodes containing 'AMOX'..."
SEARCH=$(curl -s -H "Authorization: Bearer $TOKEN" \
  "${BASE_URL}/api/v1/barcodes/search?q=AMOX&limit=10")
echo -e "Found $(echo $SEARCH | jq '.data | length') results\n"

# Practical use case simulation
echo -e "${BOLD}Step 8: Practical Use Case - Quick Order${NC}"
echo -e "Simulating pharmacy staff scanning products for order:\n"

# Simulate scanning 3 different products
echo -e "${BLUE}Cashier scans:${NC}"
echo -e "  1. Barcode: 5901234123458"
scan_result=$(curl -s -H "Authorization: Bearer $TOKEN" \
  ${BASE_URL}/api/v1/products/barcode/5901234123458)
product_id=$(echo $scan_result | jq -r '.data.id')
echo -e "     ${GREEN}→ Found:${NC} $(echo $scan_result | jq -r '.data.name')"

# Add to order
echo -e "\n${BLUE}Adding to order:${NC}"
ORDER=$(curl -s -X POST ${BASE_URL}/api/v1/orders \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "draft",
    "notes": "Barcode scan order"
  }')
ORDER_ID=$(echo $ORDER | jq -r '.data.id')

curl -s -X POST ${BASE_URL}/api/v1/orders/${ORDER_ID}/items \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"product_id\": \"$product_id\",
    \"requested_qty\": 2,
    \"unit\": \"boxes\",
    \"note\": \"Scanned via barcode\"
  }" > /dev/null

echo -e "${GREEN}✓ Item added to order${NC}\n"

# Display order
echo -e "${BOLD}Order Summary:${NC}"
curl -s -H "Authorization: Bearer $TOKEN" \
  ${BASE_URL}/api/v1/orders/${ORDER_ID}/items | jq -r '.data[] | "  - Product ID: \(.product_id)\n    Quantity: \(.requested_qty) \(.unit)"'

echo -e "\n${BOLD}Step 9: Cleanup${NC}"
echo -e "Deleting test barcode..."
curl -s -X DELETE ${BASE_URL}/api/v1/barcodes/${BARCODE1_ID} \
  -H "Authorization: Bearer $TOKEN" > /dev/null
echo -e "${GREEN}✓ Cleanup complete${NC}\n"

echo -e "${BOLD}======================================"
echo -e "Demonstration Complete"
echo -e "======================================${NC}\n"

echo -e "${BOLD}Key Features Demonstrated:${NC}"
echo -e "1. ${GREEN}✓${NC} Multiple barcode types (EAN-13, UPC-A, Code128)"
echo -e "2. ${GREEN}✓${NC} Add/Update/Delete barcodes"
echo -e "3. ${GREEN}✓${NC} Quick product lookup by barcode"
echo -e "4. ${GREEN}✓${NC} Search barcode database"
echo -e "5. ${GREEN}✓${NC} Practical ordering workflow"
echo -e "\n${BOLD}Benefits:${NC}"
echo -e "- Fast product identification"
echo -e "- Reduced manual entry errors"
echo -e "- Support for multiple barcode standards"
echo -e "- Easy integration with barcode scanners"