# API Testing Guide for DigiOrder

This guide provides comprehensive testing examples for all API endpoints.

## Prerequisites

- Server running on `http://localhost:5582`
- `curl` or Postman installed
- `jq` for JSON formatting (optional)

## Environment Setup

```bash
# Set base URL
export BASE_URL="http://localhost:5582"

# Function to pretty print JSON responses (requires jq)
alias api_get='curl -s "$BASE_URL$1" | jq'
alias api_post='curl -s -X POST -H "Content-Type: application/json" "$BASE_URL$1" -d "$2" | jq'
```

## Test Scenarios

### 1. Health Check

```bash
# Check if API is running
curl $BASE_URL/health
```

Expected Response:

```json
{
  "status": "healthy",
  "service": "DigiOrder API"
}
```

---

### 2. Categories Management

#### Create Categories

```bash
# Create دارویی category
curl -X POST $BASE_URL/api/v1/categories \
  -H "Content-Type: application/json" \
  -d '{"name": "داروهای قلبی عروقی"}'

# Create آرایشی category
curl -X POST $BASE_URL/api/v1/categories \
  -H "Content-Type: application/json" \
  -d '{"name": "مکمل های غذایی"}'
```

#### List All Categories

```bash
curl $BASE_URL/api/v1/categories | jq
```

#### Get Specific Category

```bash
curl $BASE_URL/api/v1/categories/1 | jq
```

---

### 3. Dosage Forms Management

#### Create Dosage Forms

```bash
# Create different dosage forms
curl -X POST $BASE_URL/api/v1/dosage_forms \
  -H "Content-Type: application/json" \
  -d '{"name": "محلول تزریقی"}'

curl -X POST $BASE_URL/api/v1/dosage_forms \
  -H "Content-Type: application/json" \
  -d '{"name": "کرم"}'
```

#### List All Dosage Forms

```bash
curl $BASE_URL/api/v1/dosage_forms | jq
```

---

### 4. Roles Management

#### Create Custom Roles

```bash
# Create supervisor role
curl -X POST $BASE_URL/api/v1/roles \
  -H "Content-Type: application/json" \
  -d '{"name": "supervisor"}' | jq

# Save the role ID
export ROLE_ID=4

# Create manager role
curl -X POST $BASE_URL/api/v1/roles \
  -H "Content-Type: application/json" \
  -d '{"name": "manager"}' | jq

# Create warehouse_staff role
curl -X POST $BASE_URL/api/v1/roles \
  -H "Content-Type: application/json" \
  -d '{"name": "warehouse_staff"}' | jq
```

#### List All Roles

```bash
curl $BASE_URL/api/v1/roles | jq
```

This will show all roles including the default ones (admin, pharmacist, clerk) and any custom roles you created.

#### Get Specific Role

```bash
curl $BASE_URL/api/v1/roles/1 | jq
```

#### Update Role

```bash
# Rename supervisor to senior_supervisor
curl -X PUT $BASE_URL/api/v1/roles/$ROLE_ID \
  -H "Content-Type: application/json" \
  -d '{"name": "senior_supervisor"}' | jq

# Verify the update
curl $BASE_URL/api/v1/roles/$ROLE_ID | jq
```

#### Delete Role

```bash
# Delete a role (only works if no users are assigned to it)
curl -X DELETE $BASE_URL/api/v1/roles/$ROLE_ID

# Verify deletion
curl $BASE_URL/api/v1/roles | jq
```

---

### 5. Users Management

#### Create a User

```bash
# Create admin user
curl -X POST $BASE_URL/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "full_name": "مدیر سیستم",
    "password": "admin123456",
    "role_id": 1
  }' | jq

# Save the returned user ID for later use
export USER_ID="<paste-user-id-here>"
```

#### Create Multiple Users

```bash
# Create pharmacist
curl -X POST $BASE_URL/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "username": "pharmacist1",
    "full_name": "داروساز احمد رضایی",
    "password": "pharm123456",
    "role_id": 2
  }' | jq

# Create clerk
curl -X POST $BASE_URL/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "username": "clerk1",
    "full_name": "کارمند فاطمه محمدی",
    "password": "clerk123456",
    "role_id": 3
  }' | jq
```

#### List Users

```bash
curl "$BASE_URL/api/v1/users?limit=10&offset=0" | jq
```

#### Get User by ID

```bash
curl $BASE_URL/api/v1/users/$USER_ID | jq
```

#### Update User

```bash
curl -X PUT $BASE_URL/api/v1/users/$USER_ID \
  -H "Content-Type: application/json" \
  -d '{
    "full_name": "مدیر ارشد سیستم",
    "role_id": 1
  }' | jq
```

---

### 6. Products Management

#### Create Products

```bash
# Create product 1
curl -X POST $BASE_URL/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "آسپرین",
    "brand": "Bayer",
    "dosage_form_id": 1,
    "strength": "500mg",
    "unit": "قرص",
    "category_id": 1,
    "description": "مسکن و ضد التهاب"
  }' | jq

# Save product ID
export PRODUCT_ID_1="<paste-product-id-here>"

# Create product 2
curl -X POST $BASE_URL/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "ویتامین D3",
    "brand": "Nature Made",
    "dosage_form_id": 2,
    "strength": "1000IU",
    "unit": "کپسول",
    "category_id": 4,
    "description": "مکمل ویتامین D"
  }' | jq

export PRODUCT_ID_2="<paste-product-id-here>"

# Create product 3
curl -X POST $BASE_URL/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "پانادول",
    "brand": "GSK",
    "dosage_form_id": 1,
    "strength": "500mg",
    "unit": "قرص",
    "category_id": 1,
    "description": "مسکن و تب بر"
  }' | jq

export PRODUCT_ID_3="<paste-product-id-here>"
```

#### List Products

```bash
# List first 10 products
curl "$BASE_URL/api/v1/products?limit=10&offset=0" | jq

# List next 10 products
curl "$BASE_URL/api/v1/products?limit=10&offset=10" | jq
```

#### Get Single Product

```bash
curl $BASE_URL/api/v1/products/$PRODUCT_ID_1 | jq
```

#### Update Product

```bash
curl -X PUT $BASE_URL/api/v1/products/$PRODUCT_ID_1 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "آسپرین 500",
    "strength": "500mg",
    "description": "مسکن قوی و ضد التهاب"
  }' | jq
```

#### Search Products

```bash
# Search by name
curl "$BASE_URL/api/v1/products/search?q=آسپرین" | jq

# Search by brand
curl "$BASE_URL/api/v1/products/search?q=Bayer" | jq

# Search with pagination
curl "$BASE_URL/api/v1/products/search?q=ویتامین&limit=5&offset=0" | jq
```

#### Delete Product

```bash
curl -X DELETE $BASE_URL/api/v1/products/$PRODUCT_ID_3
```

---

### 7. Orders Management

#### Create Order

```bash
# Create draft order
curl -X POST $BASE_URL/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "created_by": "'$USER_ID'",
    "status": "draft",
    "notes": "سفارش هفتگی"
  }' | jq

# Save order ID
export ORDER_ID="<paste-order-id-here>"

# Create another order without user
curl -X POST $BASE_URL/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "status": "draft",
    "notes": "سفارش فوری"
  }' | jq

export ORDER_ID_2="<paste-order-id-here>"
```

#### List Orders

```bash
# List all orders
curl "$BASE_URL/api/v1/orders?limit=10&offset=0" | jq

# List orders by specific user
curl "$BASE_URL/api/v1/orders?user_id=$USER_ID&limit=10" | jq
```

#### Get Order

```bash
curl $BASE_URL/api/v1/orders/$ORDER_ID | jq
```

#### Update Order Status

```bash
# Change to submitted
curl -X PUT $BASE_URL/api/v1/orders/$ORDER_ID/status \
  -H "Content-Type: application/json" \
  -d '{"status": "submitted"}' | jq

# Change to processing
curl -X PUT $BASE_URL/api/v1/orders/$ORDER_ID/status \
  -H "Content-Type: application/json" \
  -d '{"status": "processing"}' | jq

# Change to completed
curl -X PUT $BASE_URL/api/v1/orders/$ORDER_ID/status \
  -H "Content-Type: application/json" \
  -d '{"status": "completed"}' | jq
```

---

### 8. Order Items Management

#### Add Items to Order

```bash
# Add first item
curl -X POST $BASE_URL/api/v1/orders/$ORDER_ID/items \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": "'$PRODUCT_ID_1'",
    "requested_qty": 100,
    "unit": "قرص",
    "note": "فوری"
  }' | jq

export ORDER_ITEM_ID_1="<paste-item-id-here>"

# Add second item
curl -X POST $BASE_URL/api/v1/orders/$ORDER_ID/items \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": "'$PRODUCT_ID_2'",
    "requested_qty": 50,
    "unit": "کپسول"
  }' | jq

export ORDER_ITEM_ID_2="<paste-item-id-here>"
```

#### Get All Items in Order

```bash
curl $BASE_URL/api/v1/orders/$ORDER_ID/items | jq
```

#### Update Order Item

```bash
curl -X PUT $BASE_URL/api/v1/order_items/$ORDER_ITEM_ID_1 \
  -H "Content-Type: application/json" \
  -d '{
    "requested_qty": 150,
    "unit": "قرص",
    "note": "افزایش تعداد"
  }' | jq
```

#### Delete Order Item

```bash
curl -X DELETE $BASE_URL/api/v1/order_items/$ORDER_ITEM_ID_2
```

---

### 9. Complete Workflow Test

Here's a complete workflow from creating a user to completing an order:

```bash
#!/bin/bash

BASE_URL="http://localhost:5582"

echo "=== DigiOrder Complete Workflow Test ==="

# 1. Create user
echo -e "\n1. Creating user..."
USER_RESPONSE=$(curl -s -X POST $BASE_URL/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "username": "test_pharmacist",
    "full_name": "داروساز تست",
    "password": "test123456",
    "role_id": 2
  }')
USER_ID=$(echo $USER_RESPONSE | jq -r '.data.id')
echo "User ID: $USER_ID"

# 2. Create product
echo -e "\n2. Creating product..."
PRODUCT_RESPONSE=$(curl -s -X POST $BASE_URL/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Medicine",
    "brand": "Test Brand",
    "dosage_form_id": 1,
    "strength": "100mg",
    "unit": "tablet",
    "category_id": 1,
    "description": "Test product"
  }')
PRODUCT_ID=$(echo $PRODUCT_RESPONSE | jq -r '.data.id')
echo "Product ID: $PRODUCT_ID"

# 3. Create order
echo -e "\n3. Creating order..."
ORDER_RESPONSE=$(curl -s -X POST $BASE_URL/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "created_by": "'$USER_ID'",
    "status": "draft",
    "notes": "Test order"
  }')
ORDER_ID=$(echo $ORDER_RESPONSE | jq -r '.data.id')
echo "Order ID: $ORDER_ID"

# 4. Add item to order
echo -e "\n4. Adding item to order..."
curl -s -X POST $BASE_URL/api/v1/orders/$ORDER_ID/items \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": "'$PRODUCT_ID'",
    "requested_qty": 100,
    "unit": "tablet",
    "note": "Test item"
  }' | jq

# 5. Get order with items
echo -e "\n5. Fetching order details..."
curl -s $BASE_URL/api/v1/orders/$ORDER_ID | jq

# 6. Get order items
echo -e "\n6. Fetching order items..."
curl -s $BASE_URL/api/v1/orders/$ORDER_ID/items | jq

# 7. Submit order
echo -e "\n7. Submitting order..."
curl -s -X PUT $BASE_URL/api/v1/orders/$ORDER_ID/status \
  -H "Content-Type: application/json" \
  -d '{"status": "submitted"}' | jq

echo -e "\n=== Test Complete ==="
```

Save this as `test_workflow.sh`, make it executable (`chmod +x test_workflow.sh`), and run it.

---

## Error Testing

### Test Validation Errors

```bash
# Missing required field
curl -X POST $BASE_URL/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{
    "brand": "Test"
  }' | jq

# Invalid UUID
curl $BASE_URL/api/v1/products/invalid-uuid | jq

# Invalid number
curl $BASE_URL/api/v1/categories/abc | jq
```

### Test Not Found Errors

```bash
# Non-existent product
curl $BASE_URL/api/v1/products/00000000-0000-0000-0000-000000000000 | jq

# Non-existent category
curl $BASE_URL/api/v1/categories/999 | jq
```

### Test Duplicate Entries

```bash
# Try to create user with existing username
curl -X POST $BASE_URL/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "test123",
    "role_id": 1
  }' | jq
```

---

## Performance Testing

### Bulk Create Products

```bash
#!/bin/bash
for i in {1..100}; do
  curl -s -X POST $BASE_URL/api/v1/products \
    -H "Content-Type: application/json" \
    -d '{
      "name": "Product '$i'",
      "brand": "Brand '$i'",
      "dosage_form_id": 1,
      "strength": "'$i'mg",
      "unit": "tablet",
      "category_id": 1,
      "description": "Test product '$i'"
    }' > /dev/null
  echo "Created product $i"
done
```

### Test Pagination

```bash
# Test different page sizes
for limit in 10 50 100; do
  echo "Testing limit=$limit"
  time curl -s "$BASE_URL/api/v1/products?limit=$limit&offset=0" > /dev/null
done
```

---

## Cleanup

```bash
# Note: You'll need to delete items manually using their IDs
# There's no bulk delete endpoint for safety reasons

# Delete test data (replace IDs with actual values)
curl -X DELETE $BASE_URL/api/v1/order_items/$ORDER_ITEM_ID
curl -X DELETE $BASE_URL/api/v1/orders/$ORDER_ID
curl -X DELETE $BASE_URL/api/v1/products/$PRODUCT_ID
curl -X DELETE $BASE_URL/api/v1/users/$USER_ID
```

---

## Tips

1. **Use jq for readable output**: Install jq (`sudo apt install jq` on Ubuntu)
2. **Save IDs in environment variables**: Makes testing easier
3. **Create test scripts**: Automate repetitive testing
4. **Use Postman**: Import the endpoints as a collection
5. **Check logs**: Run server with `make run` to see detailed logs

## Common Issues

1. **Connection refused**: Make sure server is running on port 5582
2. **Invalid UUID**: UUIDs must be in proper format (xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx)
3. **Validation errors**: Check required fields in error response
4. **Foreign key errors**: Ensure referenced IDs exist (category_id, dosage_form_id, etc.)
