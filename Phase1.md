# DigiOrder Phase 1 - Complete Implementation Guide

## 🎯 Overview

This guide walks you through implementing the complete backend API and preparing for the React dashboard.

---

## 📋 Step-by-Step Implementation

### Step 1: Update Dependencies

```bash
# Update go.mod with JWT package
go get github.com/golang-jwt/jwt/v5
go mod tidy
```

### Step 2: Add Missing Handler Files

Create these new files in `internal/server/`:

1. **order_handlers.go** - Complete order management
2. **improved_products.go** - Enhanced product handlers with search
3. **auth_middleware.go** - JWT authentication
4. **user_handlers.go** - User management
5. **barcode_handler.go** - Barcode scanning

All the code is provided in the artifacts above.

### Step 3: Update Existing Files

Replace these files:

1. **internal/server/server.go** - Add new handler methods
2. **internal/server/routes.go** - Complete route registration with auth
3. **go.mod** - Add JWT dependency

### Step 4: Add Barcode SQL Queries

Create `internal/db/query/barcodes.sql`:

```sql
-- name: CreateProductBarcode :one
INSERT INTO product_barcodes (
    product_id, barcode, barcode_type
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetProductByBarcode :one
SELECT p.* FROM products p
JOIN product_barcodes pb ON p.id = pb.product_id
WHERE pb.barcode = $1 LIMIT 1;

-- name: GetProductBarcodes :many
SELECT * FROM product_barcodes
WHERE product_id = $1;

-- name: DeleteProductBarcode :exec
DELETE FROM product_barcodes WHERE id = $1;

-- name: SearchProductsByBarcode :many
SELECT p.* FROM products p
JOIN product_barcodes pb ON p.id = pb.product_id
WHERE pb.barcode ILIKE '%' || $1 || '%'
ORDER BY p.created_at DESC
LIMIT $2 OFFSET $3;
```

### Step 5: Regenerate SQLC Code

```bash
make sqlc
# or
sqlc generate
```

### Step 6: Update .env File

```bash
# Add strong JWT secret
JWT_SECRET=your_very_secure_random_secret_key_here_change_this_in_production
JWT_EXPIRY=24h
```

**⚠️ IMPORTANT:** Generate a secure secret:

```bash
# Linux/Mac
openssl rand -hex 32

# Or use online generator
# Never use the default in production!
```

### Step 7: Database Setup

```bash
# Start database
docker-compose up -d

# Run migrations
make migrate-up

# Verify tables created
docker exec -it digiorder-postgres psql -U postgres -d digiorder_db -c "\dt"
```

### Step 8: Create Initial Admin User

Create a script `scripts/create_admin.sh`:

```bash
#!/bin/bash

curl -X POST http://localhost:5582/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123456",
    "full_name": "System Administrator",
    "role_id": 1
  }'
```

Run it:

```bash
chmod +x scripts/create_admin.sh
./scripts/create_admin.sh
```

### Step 9: Test Authentication

```bash
# Login
TOKEN=$(curl -X POST http://localhost:5582/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123456"
  }' | jq -r '.data.token')

echo "Token: $TOKEN"

# Test protected endpoint
curl -X GET http://localhost:5582/api/v1/me \
  -H "Authorization: Bearer $TOKEN"
```

### Step 10: Seed Sample Data

Create `scripts/seed_data.sh`:

```bash
#!/bin/bash

TOKEN=$1

if [ -z "$TOKEN" ]; then
    echo "Usage: ./seed_data.sh <JWT_TOKEN>"
    exit 1
fi

BASE_URL="http://localhost:5582/api/v1"

# Create categories
echo "Creating categories..."
curl -X POST $BASE_URL/categories \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "دارویی"}'

curl -X POST $BASE_URL/categories \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "آرایشی"}'

# Create dosage forms
echo "Creating dosage forms..."
curl -X POST $BASE_URL/dosage_forms \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "قرص"}'

curl -X POST $BASE_URL/dosage_forms \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "شربت"}'

# Create products
echo "Creating products..."
curl -X POST $BASE_URL/products \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "آموکسی سیلین",
    "brand": "Bayer",
    "dosage_form_id": 1,
    "strength": "500mg",
    "unit": "بسته",
    "category_id": 1,
    "description": "آنتی بیوتیک گسترده الطیف"
  }'

echo "Sample data created!"
```

---

## 🧪 Testing Guide

### Manual Testing with cURL

```bash
# Get token
export TOKEN="your_jwt_token_here"

# 1. Search products
curl "$BASE_URL/products?q=amoxicillin&dosage_form_id=1" \
  -H "Authorization: Bearer $TOKEN"

# 2. Create order
curl -X POST $BASE_URL/orders \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "notes": "Weekly order",
    "items": [{
      "product_id": "PRODUCT_UUID_HERE",
      "requested_qty": 10,
      "unit": "بسته"
    }]
  }'

# 3. List orders
curl "$BASE_URL/orders?limit=10" \
  -H "Authorization: Bearer $TOKEN"

# 4. Update order status
curl -X PATCH "$BASE_URL/orders/ORDER_UUID_HERE/status" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status": "submitted"}'
```

### Automated Testing with Go

Create `internal/server/orders_test.go`:

```go
package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestCreateOrder(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`{
		"notes": "Test order",
		"items": [{
			"product_id": "123e4567-e89b-12d3-a456-426614174000",
			"requested_qty": 5,
			"unit": "pack"
		}]
	}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Test
	// handler := NewCreateOrderHandler(db, queries)
	// err := handler(c)

	// Assert
	// assert.NoError(t, err)
	// assert.Equal(t, http.StatusCreated, rec.Code)
}
```

---

## 📱 React Dashboard - Next Steps

### Tech Stack Recommendation

```json
{
  "framework": "React 18 with TypeScript",
  "routing": "React Router v6",
  "state": "React Query (TanStack Query) + Zustand",
  "ui": "Material-UI or Ant Design",
  "forms": "React Hook Form + Zod",
  "api": "Axios",
  "auth": "JWT in httpOnly cookies or localStorage",
  "build": "Vite"
}
```

### Dashboard Features (Priority Order)

#### 1. Authentication Module

- Login page
- JWT token management
- Role-based route protection
- Auto-refresh tokens

#### 2. Product Catalog

- Product list with filters (category, dosage form)
- Search with autocomplete
- Add/Edit/Delete products
- Barcode scanner integration (using device camera)
- Bulk product import (CSV)

#### 3. Order Management

- Create order interface (cart-style)
- Product search within order creation
- Real-time product availability check
- Order list with status filters
- Order details view
- Status update workflow

#### 4. Dashboard & Analytics

- Today's orders count
- Pending orders
- Most ordered products
- Order history charts
- Inventory alerts

#### 5. User Management (Admin only)

- User list
- Create/Edit users
- Role assignment
- Activity logs

### React Folder Structure

```
frontend/
├── src/
│   ├── api/
│   │   ├── axios.ts
│   │   ├── auth.ts
│   │   ├── products.ts
│   │   ├── orders.ts
│   │   └── users.ts
│   ├── components/
│   │   ├── common/
│   │   │   ├── Sidebar.tsx
│   │   │   ├── Header.tsx
│   │   │   └── Loading.tsx
│   │   ├── products/
│   │   │   ├── ProductList.tsx
│   │   │   ├── ProductForm.tsx
│   │   │   ├── ProductCard.tsx
│   │   │   └── BarcodeScanner.tsx
│   │   └── orders/
│   │       ├── OrderList.tsx
│   │       ├── CreateOrder.tsx
│   │       ├── OrderCart.tsx
│   │       └── OrderDetails.tsx
│   ├── pages/
│   │   ├── Login.tsx
│   │   ├── Dashboard.tsx
│   │   ├── Products.tsx
│   │   ├── Orders.tsx
│   │   └── Users.tsx
│   ├── hooks/
│   │   ├── useAuth.ts
│   │   ├── useProducts.ts
│   │   └── useOrders.ts
│   ├── store/
│   │   ├── authStore.ts
│   │   └── cartStore.ts
│   ├── types/
│   │   ├── product.ts
│   │   ├── order.ts
│   │   └── user.ts
│   ├── utils/
│   │   ├── formatters.ts
│   │   └── validators.ts
│   ├── App.tsx
│   └── main.tsx
├── package.json
└── vite.config.ts
```

### Sample React Component

```typescript
// src/components/orders/CreateOrder.tsx
import { useState } from "react";
import { useQuery, useMutation } from "@tanstack/react-query";
import { searchProducts, createOrder } from "@/api";

interface OrderItem {
  product_id: string;
  requested_qty: number;
  unit: string;
  note?: string;
}

export const CreateOrder = () => {
  const [cart, setCart] = useState<OrderItem[]>([]);
  const [search, setSearch] = useState("");

  const { data: products } = useQuery({
    queryKey: ["products", search],
    queryFn: () => searchProducts(search),
  });

  const createOrderMutation = useMutation({
    mutationFn: createOrder,
    onSuccess: () => {
      alert("Order created!");
      setCart([]);
    },
  });

  const addToCart = (product: Product) => {
    setCart([
      ...cart,
      {
        product_id: product.id,
        requested_qty: 1,
        unit: product.unit || "pack",
      },
    ]);
  };

  const handleSubmit = () => {
    createOrderMutation.mutate({
      notes: "",
      items: cart,
    });
  };

  return (
    <div>
      {/* Search products */}
      <input
        value={search}
        onChange={(e) => setSearch(e.target.value)}
        placeholder="Search products..."
      />

      {/* Product list */}
      {products?.map((product) => (
        <ProductCard
          key={product.id}
          product={product}
          onAdd={() => addToCart(product)}
        />
      ))}

      {/* Cart */}
      <h3>Order Items ({cart.length})</h3>
      {cart.map((item, idx) => (
        <CartItem key={idx} item={item} />
      ))}

      {/* Submit */}
      <button onClick={handleSubmit}>Create Order</button>
    </div>
  );
};
```

---

## 🔐 Security Checklist

- [x] JWT authentication implemented
- [x] Password hashing with bcrypt
- [x] Role-based access control
- [ ] Rate limiting (add later)
- [ ] Input sanitization
- [ ] SQL injection protection (using parameterized queries ✓)
- [ ] HTTPS in production
- [ ] CORS configuration
- [ ] API key for mobile apps (future)

---

## 🚀 Deployment Checklist

### Environment Variables (Production)

```bash
DB_HOST=your-db-host
DB_PORT=5432
DB_USER=secure_user
DB_PASSWORD=very_secure_password
DB_NAME=digiorder_prod
DB_SSLMODE=require

SERVER_PORT=5582
SERVER_HOST=0.0.0.0

JWT_SECRET=generate-64-char-random-secret
JWT_EXPIRY=24h
```

### Production Dockerfile

```dockerfile
# Build stage
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /digiorder cmd/main.go

# Runtime stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /digiorder .
EXPOSE 5582
CMD ["./digiorder"]
```

### docker-compose.prod.yml

```yaml
version: "3.8"

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: ${DB_USER}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_DB: ${DB_NAME}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: always

  api:
    build: .
    ports:
      - "5582:5582"
    environment:
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: ${DB_USER}
      DB_PASSWORD: ${DB_PASSWORD}
      DB_NAME: ${DB_NAME}
      JWT_SECRET: ${JWT_SECRET}
    depends_on:
      - postgres
    restart: always

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/nginx/ssl
    depends_on:
      - api
    restart: always

volumes:
  postgres_data:
```

---

## 📊 Monitoring & Logging

Add these packages:

```bash
go get github.com/sirupsen/logrus
go get github.com/prometheus/client_golang/prometheus
```

---

## 🎓 Training for Pharmacy Staff

### Quick Start Guide (Persian)

**برای کاربران داروخانه:**

1. **ورود به سیستم:** با نام کاربری و رمز عبور خود وارد شوید
2. **جستجوی محصول:** نام دارو را تایپ کنید و نوع آن را مشخص کنید (قرص، شربت، آمپول)
3. **ایجاد سفارش:** محصولات را به سبد اضافه کنید
4. **تعیین تعداد:** تعداد مورد نیاز را وارد کنید
5. **ثبت سفارش:** دکمه "ثبت سفارش" را بزنید

**مثال عملی:**

- قبلاً: "آموکسی سیلین 500: 10 بسته" (مبهم!)
- حالا: در سیستم جستجو کنید → "آموکسی سیلین 500mg قرص Bayer" را انتخاب کنید → 10 بسته → ثبت

---

## 🎯 Success Metrics

- ✅ Zero ambiguous orders
- ✅ 50% faster order creation
- ✅ 100% order traceability
- ✅ Real-time inventory awareness
- ✅ Reduced human error

---

## 📞 Support & Next Features

### Phase 2 Features:

- Supplier integration
- Inventory management
- Price tracking
- Expiry date management
- Reporting & analytics
- Mobile app (React Native)
- WhatsApp notifications
- PDF order generation

Would you like me to create the React dashboard starter code next?
