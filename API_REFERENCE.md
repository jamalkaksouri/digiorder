# DigiOrder API Reference

## Base URL

```
Production: https://api.digiorder.com
Development: http://localhost:5582
```

## Table of Contents

1. [Authentication](#authentication)
2. [Products](#products)
3. [Orders](#orders)
4. [Order Items](#order-items)
5. [Users](#users)
6. [Roles](#roles)
7. [Permissions](#permissions)
8. [Categories](#categories)
9. [Dosage Forms](#dosage-forms)
10. [Barcodes](#barcodes)
11. [Audit Logs](#audit-logs)
12. [Error Codes](#error-codes)

---

## Authentication

All endpoints except `/health` and `/api/v1/auth/login` require authentication using JWT bearer tokens.

### Include Token in Requests

```http
Authorization: Bearer <your_jwt_token>
```

---

### POST /api/v1/auth/login

Authenticate user and receive JWT token.

**Request Body:**

```json
{
  "username": "string (required)",
  "password": "string (required, min 8 chars)"
}
```

**Response:** `200 OK`

```json
{
  "data": [
    {
      "id": "a50e8400-e29b-41d4-a716-446655440005",
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "username": "admin",
      "action": "create",
      "entity_type": "product",
      "entity_id": "550e8400-e29b-41d4-a716-446655440000",
      "old_values": null,
      "new_values": { "name": "Amoxicillin 500mg", "brand": "Bayer" },
      "ip_address": "192.168.1.100",
      "user_agent": "Mozilla/5.0...",
      "created_at": "2025-11-10T10:30:00Z"
    }
  ]
}
```

**Example:**

```bash
curl -X GET "http://localhost:5582/api/v1/audit-logs?limit=20&user_id=550e8400..." \
  -H "Authorization: Bearer $TOKEN"
```

---

### GET /api/v1/audit-logs/:id

Get specific audit log.

**Authentication:** Required  
**Roles:** admin

**Path Parameters:**

- `id` (required) - Audit log UUID

**Response:** `200 OK`

```json
{
  "data": {
    "id": "a50e8400-e29b-41d4-a716-446655440005",
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "admin",
    "action": "create",
    "entity_type": "product",
    "entity_id": "550e8400-e29b-41d4-a716-446655440000",
    "old_values": null,
    "new_values": { "name": "Amoxicillin 500mg" },
    "ip_address": "192.168.1.100",
    "user_agent": "Mozilla/5.0...",
    "created_at": "2025-11-10T10:30:00Z"
  }
}
```

---

### GET /api/v1/audit-logs/entity/:type/:id

Get entity history.

**Authentication:** Required  
**Roles:** admin

**Path Parameters:**

- `type` (required) - Entity type (product, order, user)
- `id` (required) - Entity ID

**Query Parameters:**

- `limit` (optional, default: 50)
- `offset` (optional, default: 0)

**Response:** `200 OK`

```json
{
  "data": [
    {
      "id": "a50e8400-e29b-41d4-a716-446655440005",
      "action": "create",
      "username": "admin",
      "old_values": null,
      "new_values": { "name": "Amoxicillin 500mg" },
      "ip_address": "192.168.1.100",
      "created_at": "2025-11-10T10:30:00Z"
    },
    {
      "id": "a50e8400-e29b-41d4-a716-446655440006",
      "action": "update",
      "username": "pharmacist1",
      "old_values": { "strength": "500mg" },
      "new_values": { "strength": "1000mg" },
      "ip_address": "192.168.1.101",
      "created_at": "2025-11-10T11:00:00Z"
    }
  ]
}
```

**Example:**

```bash
curl -X GET "http://localhost:5582/api/v1/audit-logs/entity/product/550e8400..." \
  -H "Authorization: Bearer $TOKEN"
```

---

### GET /api/v1/audit-logs/stats

Get audit statistics.

**Authentication:** Required  
**Roles:** admin

**Response:** `200 OK`

```json
{
  "data": {
    "total_logs": 1543,
    "unique_users": 12,
    "unique_entities": 5
  }
}
```

---

### GET /api/v1/users/:user_id/activity

Get user activity logs.

**Authentication:** Required  
**Roles:** admin

**Path Parameters:**

- `user_id` (required) - User UUID

**Query Parameters:**

- `limit` (optional, default: 50)
- `offset` (optional, default: 0)

**Response:** `200 OK`

```json
{
  "data": [
    {
      "id": "a50e8400-e29b-41d4-a716-446655440005",
      "action": "create",
      "entity_type": "product",
      "entity_id": "550e8400-e29b-41d4-a716-446655440000",
      "ip_address": "192.168.1.100",
      "created_at": "2025-11-10T10:30:00Z"
    }
  ]
}
```

---

## Error Codes

### HTTP Status Codes

| Code | Meaning               | Description                                  |
| ---- | --------------------- | -------------------------------------------- |
| 200  | OK                    | Request successful                           |
| 201  | Created               | Resource created successfully                |
| 204  | No Content            | Request successful, no content to return     |
| 400  | Bad Request           | Invalid request format or validation error   |
| 401  | Unauthorized          | Missing or invalid authentication token      |
| 403  | Forbidden             | Insufficient permissions                     |
| 404  | Not Found             | Resource not found                           |
| 409  | Conflict              | Resource conflict (e.g., duplicate username) |
| 429  | Too Many Requests     | Rate limit exceeded                          |
| 500  | Internal Server Error | Server error                                 |

---

### Error Response Format

```json
{
  "error": "error_code",
  "details": "Human-readable error message"
}
```

### Common Error Codes

| Error Code                 | Description                     | Solution                          |
| -------------------------- | ------------------------------- | --------------------------------- |
| `invalid_request`          | Request body is malformed       | Check JSON syntax                 |
| `validation_error`         | Field validation failed         | Check required fields and formats |
| `invalid_id`               | UUID format is invalid          | Use valid UUID format             |
| `invalid_credentials`      | Username or password incorrect  | Verify credentials                |
| `invalid_token`            | JWT token is invalid or expired | Re-login or refresh token         |
| `insufficient_permissions` | User lacks required permissions | Contact admin for access          |
| `not_found`                | Resource doesn't exist          | Check resource ID                 |
| `duplicate_username`       | Username already exists         | Choose different username         |
| `protected_user`           | Cannot modify protected user    | Primary admin cannot be deleted   |
| `last_admin`               | Cannot delete last admin        | At least one admin must exist     |
| `db_error`                 | Database operation failed       | Contact support                   |
| `rate_limit_exceeded`      | Too many requests               | Wait and retry                    |

---

### Example Error Responses

**Validation Error:**

```json
{
  "error": "validation_error",
  "details": "Key: 'CreateProductReq.Name' Error:Field validation for 'Name' failed on the 'required' tag"
}
```

**Authentication Error:**

```json
{
  "error": "invalid_credentials",
  "details": "Invalid username or password."
}
```

**Authorization Error:**

```json
{
  "error": "insufficient_permissions",
  "details": "Only administrators can create users."
}
```

**Not Found Error:**

```json
{
  "error": "not_found",
  "details": "Product with the specified ID was not found."
}
```

**Rate Limit Error:**

```json
{
  "error": "rate_limit_exceeded",
  "details": "Rate limit exceeded"
}
```

---

## Rate Limiting

### Limits

- **Global:** 100 requests/second (burst: 200)
- **Authenticated:** 1,000 requests/minute
- **Per IP:** Individual tracking per client

### Rate Limit Headers

When rate limited, the API returns:

- **Status Code:** `429 Too Many Requests`
- **Retry-After:** Time to wait before retrying (if available)

### Best Practices

1. **Implement exponential backoff** when receiving 429 responses
2. **Cache responses** when possible
3. **Use pagination** for large datasets
4. **Batch operations** where supported

---

## Pagination

All list endpoints support pagination using `limit` and `offset` parameters.

### Parameters

- `limit` - Number of results per page (default: 50, max: 100)
- `offset` - Number of results to skip (default: 0)

### Example

```bash
# Get first page (items 1-50)
curl "http://localhost:5582/api/v1/products?limit=50&offset=0"

# Get second page (items 51-100)
curl "http://localhost:5582/api/v1/products?limit=50&offset=50"

# Get third page (items 101-150)
curl "http://localhost:5582/api/v1/products?limit=50&offset=100"
```

### Response

Pagination metadata is not currently included in responses. Calculate total pages based on:

- Empty array `[]` indicates no more results
- Partial results indicate last page

---

## Filtering

### Query Parameters

Most list endpoints support filtering via query parameters:

```bash
# Filter orders by user
GET /api/v1/orders?user_id=550e8400-e29b-41d4-a716-446655440000

# Filter permissions by resource
GET /api/v1/permissions?resource=products

# Search products
GET /api/v1/products/search?q=amoxicillin

# Filter audit logs by action
GET /api/v1/audit-logs?action=create
```

---

## Sorting

Results are sorted by default:

| Endpoint    | Default Sort               |
| ----------- | -------------------------- |
| Products    | `created_at DESC`          |
| Orders      | `created_at DESC`          |
| Users       | `created_at DESC`          |
| Audit Logs  | `created_at DESC`          |
| Permissions | `resource ASC, action ASC` |

Custom sorting is not currently supported.

---

## Caching

The API implements response caching with the following configuration:

- **TTL:** 5 minutes
- **Cache Key:** Based on method, path, query params, and user ID
- **Cache Headers:**
  - `X-Cache: HIT` - Response served from cache
  - `X-Cache: MISS` - Fresh response from database
  - `X-Cache-Age: <seconds>` - Age of cached response

### Cache Invalidation

Cache is automatically cleared on:

- `POST` requests (creates)
- `PUT` requests (updates)
- `PATCH` requests (partial updates)
- `DELETE` requests (deletes)

---

## Versioning

The API uses URL versioning:

- **Current Version:** `v1`
- **Base Path:** `/api/v1`

Future versions will use `/api/v2`, `/api/v3`, etc.

---

## Request ID Tracing

Every request receives unique identifiers for tracing:

**Response Headers:**

```http
X-Request-ID: 550e8400-e29b-41d4-a716-446655440000
X-Trace-ID: 650e8400-e29b-41d4-a716-446655440001
X-Span-ID: 750e8400-e29b-41d4-a716-446655440002
```

Include these IDs when reporting issues for easier debugging.

---

## CORS

The API supports Cross-Origin Resource Sharing (CORS) with the following configuration:

**Allowed Origins:** Configured per environment
**Allowed Methods:** `GET, POST, PUT, DELETE, OPTIONS, PATCH`
**Allowed Headers:** `Accept, Authorization, Content-Type, X-CSRF-Token, X-Request-ID`
**Exposed Headers:** `X-Request-ID, X-Trace-ID, X-Cache, X-Cache-Age`

---

## Health Check

### GET /health

Check API health status (no authentication required).

**Response:** `200 OK`

```json
{
  "status": "healthy",
  "service": "DigiOrder API",
  "database": "connected",
  "version": "3.0.0"
}
```

**Response:** `503 Service Unavailable` (if unhealthy)

```json
{
  "status": "unhealthy",
  "service": "DigiOrder API",
  "database": "disconnected",
  "error": "connection refused"
}
```

**Example:**

```bash
curl http://localhost:5582/health
```

---

## Metrics

### GET /metrics

Get Prometheus-compatible metrics (no authentication required).

**Response:** `200 OK` (Plain text Prometheus format)

```
# HELP http_requests_total Total number of HTTP requests
# TYPE http_requests_total counter
http_requests_total{method="GET",endpoint="/api/v1/products",status="200"} 1543

# HELP http_request_duration_seconds HTTP request duration
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{method="GET",endpoint="/api/v1/products",status="200",le="0.005"} 123
...
```

**Example:**

```bash
curl http://localhost:5582/metrics
```

---

## WebSocket Support

WebSocket connections are **not currently supported**. The API uses RESTful HTTP only.

For real-time updates, implement polling or use Server-Sent Events (SSE) in future versions.

---

## Best Practices

### 1. Authentication

```bash
# Store token securely
TOKEN=$(curl -s -X POST http://localhost:5582/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123456"}' \
  | jq -r '.data.token')

# Use token in subsequent requests
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:5582/api/v1/products
```

### 2. Error Handling

```javascript
try {
  const response = await fetch("http://localhost:5582/api/v1/products", {
    headers: { Authorization: `Bearer ${token}` },
  });

  if (!response.ok) {
    const error = await response.json();
    console.error(`Error ${response.status}:`, error.details);

    if (response.status === 401) {
      // Token expired, re-authenticate
      await refreshToken();
    }
  }

  const data = await response.json();
  return data.data;
} catch (err) {
  console.error("Network error:", err);
}
```

### 3. Pagination Loop

```javascript
async function getAllProducts() {
  const allProducts = [];
  let offset = 0;
  const limit = 50;

  while (true) {
    const response = await fetch(
      `http://localhost:5582/api/v1/products?limit=${limit}&offset=${offset}`,
      { headers: { Authorization: `Bearer ${token}` } }
    );

    const data = await response.json();

    if (data.data.length === 0) break;

    allProducts.push(...data.data);
    offset += limit;
  }

  return allProducts;
}
```

### 4. Rate Limit Handling

```javascript
async function fetchWithRetry(url, options, maxRetries = 3) {
  for (let i = 0; i < maxRetries; i++) {
    const response = await fetch(url, options);

    if (response.status !== 429) {
      return response;
    }

    // Exponential backoff
    const delay = Math.pow(2, i) * 1000;
    console.log(`Rate limited, retrying in ${delay}ms...`);
    await new Promise((resolve) => setTimeout(resolve, delay));
  }

  throw new Error("Max retries exceeded");
}
```

### 5. Batch Operations

```javascript
// Instead of creating products one by one
async function createProductsBatch(products) {
  const promises = products.map((product) =>
    fetch("http://localhost:5582/api/v1/products", {
      method: "POST",
      headers: {
        Authorization: `Bearer ${token}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify(product),
    })
  );

  // Use Promise.all for parallel requests
  // Be mindful of rate limits!
  return await Promise.all(promises);
}
```

---

## Code Examples

### cURL Examples

**Login:**

```bash
curl -X POST http://localhost:5582/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123456"}'
```

**Create Product:**

```bash
curl -X POST http://localhost:5582/api/v1/products \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Amoxicillin 500mg",
    "brand": "Bayer",
    "dosage_form_id": 1,
    "strength": "500mg",
    "unit": "tablet",
    "category_id": 1,
    "description": "Broad-spectrum antibiotic"
  }'
```

**Search Products:**

```bash
curl "http://localhost:5582/api/v1/products/search?q=amoxicillin&limit=10" \
  -H "Authorization: Bearer $TOKEN"
```

### JavaScript Examples

**Using Fetch API:**

```javascript
// Login
const loginResponse = await fetch("http://localhost:5582/api/v1/auth/login", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({
    username: "admin",
    password: "admin123456",
  }),
});

const { data } = await loginResponse.json();
const token = data.token;

// Get products
const productsResponse = await fetch("http://localhost:5582/api/v1/products", {
  headers: { Authorization: `Bearer ${token}` },
});

const products = await productsResponse.json();
console.log(products.data);
```

**Using Axios:**

```javascript
import axios from "axios";

const api = axios.create({
  baseURL: "http://localhost:5582/api/v1",
  headers: { "Content-Type": "application/json" },
});

// Login
const { data: loginData } = await api.post("/auth/login", {
  username: "admin",
  password: "admin123456",
});

const token = loginData.data.token;

// Set token for future requests
api.defaults.headers.common["Authorization"] = `Bearer ${token}`;

// Get products
const { data: productsData } = await api.get("/products");
console.log(productsData.data);

// Create product
const { data: newProduct } = await api.post("/products", {
  name: "Amoxicillin 500mg",
  brand: "Bayer",
  dosage_form_id: 1,
  strength: "500mg",
  unit: "tablet",
  category_id: 1,
  description: "Broad-spectrum antibiotic",
});
```

### Python Examples

**Using Requests:**

```python
import requests

BASE_URL = 'http://localhost:5582/api/v1'

# Login
response = requests.post(f'{BASE_URL}/auth/login', json={
    'username': 'admin',
    'password': 'admin123456'
})
token = response.json()['data']['token']

# Get products
headers = {'Authorization': f'Bearer {token}'}
response = requests.get(f'{BASE_URL}/products', headers=headers)
products = response.json()['data']

# Create product
new_product = {
    'name': 'Amoxicillin 500mg',
    'brand': 'Bayer',
    'dosage_form_id': 1,
    'strength': '500mg',
    'unit': 'tablet',
    'category_id': 1,
    'description': 'Broad-spectrum antibiotic'
}
response = requests.post(f'{BASE_URL}/products',
                        json=new_product,
                        headers=headers)
created_product = response.json()['data']
```

### Go Examples

**Using net/http:**

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

const baseURL = "http://localhost:5582/api/v1"

type LoginRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

func main() {
    // Login
    loginReq := LoginRequest{
        Username: "admin",
        Password: "admin123456",
    }

    body, _ := json.Marshal(loginReq)
    resp, _ := http.Post(baseURL+"/auth/login", "application/json", bytes.NewBuffer(body))

    var loginResp struct {
        Data struct {
            Token string `json:"token"`
        } `json:"data"`
    }
    json.NewDecoder(resp.Body).Decode(&loginResp)
    token := loginResp.Data.Token

    // Get products
    req, _ := http.NewRequest("GET", baseURL+"/products", nil)
    req.Header.Set("Authorization", "Bearer "+token)

    client := &http.Client{}
    resp, _ = client.Do(req)

    var productsResp struct {
        Data []map[string]interface{} `json:"data"`
    }
    json.NewDecoder(resp.Body).Decode(&productsResp)

    fmt.Println(productsResp.Data)
}
```

---

## Postman Collection

Import this collection URL into Postman:

```
Coming soon...
```

Or manually create requests using the examples in this documentation.

---

## Changelog

### v3.0.0 (2025-11-10)

- Added permission management system
- Added audit logging for all operations
- Added admin user protection
- Enhanced observability with Prometheus/Grafana
- Added distributed tracing
- Improved rate limiting
- Added barcode scanning support

### v2.0.0 (2025-11-01)

- Added JWT authentication
- Added role-based access control
- Added rate limiting
- Added caching layer
- Added soft deletes

### v1.0.0 (2025-10-15)

- Initial release
- Basic CRUD for products, orders, users
- PostgreSQL database
- RESTful API design

---

## Support

- **Documentation:** See `docs/` directory
- **Issues:** Report bugs on GitHub
- **Email:** support@digiorder.com

---

## License

MIT License

---

**API Version:** 3.0.0  
**Last Updated:** November 10, 2025  
**Base URL:** `http://localhost:5582` (development)
"data": {
"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
"expires_in": "24h",
"user": {
"id": "550e8400-e29b-41d4-a716-446655440000",
"username": "admin",
"full_name": "System Administrator",
"role_id": 1,
"role_name": "admin"
}
}
}

````

**Example:**
```bash
curl -X POST http://localhost:5582/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123456"
  }'
````

**Errors:**

- `401 Unauthorized` - Invalid credentials
- `400 Bad Request` - Invalid request format

---

### POST /api/v1/auth/refresh

Refresh JWT token.

**Authentication:** Required

**Request Body:**

```json
{
  "token": "string (required, current valid token)"
}
```

**Response:** `200 OK`

```json
{
  "data": {
    "token": "new_jwt_token_here",
    "expires_in": "24h"
  }
}
```

**Example:**

```bash
curl -X POST http://localhost:5582/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"token": "eyJ..."}'
```

---

### GET /api/v1/auth/profile

Get current user profile.

**Authentication:** Required

**Response:** `200 OK`

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "admin",
    "full_name": "System Administrator",
    "role_id": 1,
    "role_name": "admin"
  }
}
```

**Example:**

```bash
curl -X GET http://localhost:5582/api/v1/auth/profile \
  -H "Authorization: Bearer $TOKEN"
```

---

### PUT /api/v1/auth/password

Change user password.

**Authentication:** Required

**Request Body:**

```json
{
  "old_password": "string (required)",
  "new_password": "string (required, min 8 chars)"
}
```

**Response:** `200 OK`

```json
{
  "data": {
    "message": "Password updated successfully"
  }
}
```

**Errors:**

- `401 Unauthorized` - Old password incorrect
- `400 Bad Request` - Validation error

---

### GET /api/v1/auth/check-permission

Check if current user has specific permission.

**Authentication:** Required

**Query Parameters:**

- `resource` (required) - Resource name (e.g., "products")
- `action` (required) - Action name (e.g., "create")

**Response:** `200 OK`

```json
{
  "data": {
    "has_permission": true,
    "resource": "products",
    "action": "create"
  }
}
```

**Example:**

```bash
curl -X GET "http://localhost:5582/api/v1/auth/check-permission?resource=products&action=create" \
  -H "Authorization: Bearer $TOKEN"
```

---

## Products

### POST /api/v1/products

Create a new product.

**Authentication:** Required  
**Roles:** admin, pharmacist

**Request Body:**

```json
{
  "name": "string (required)",
  "brand": "string (optional)",
  "dosage_form_id": "integer (required, >0)",
  "strength": "string (optional)",
  "unit": "string (optional)",
  "category_id": "integer (required, >0)",
  "description": "string (optional)"
}
```

**Response:** `201 Created`

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Amoxicillin 500mg",
    "brand": "Bayer",
    "dosage_form_id": 1,
    "strength": "500mg",
    "unit": "tablet",
    "category_id": 1,
    "description": "Broad-spectrum antibiotic",
    "created_at": "2025-11-10T10:30:00Z",
    "deleted_at": null
  }
}
```

**Example:**

```bash
curl -X POST http://localhost:5582/api/v1/products \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Amoxicillin 500mg",
    "brand": "Bayer",
    "dosage_form_id": 1,
    "strength": "500mg",
    "unit": "tablet",
    "category_id": 1,
    "description": "Broad-spectrum antibiotic"
  }'
```

---

### GET /api/v1/products

List all products with pagination.

**Authentication:** Required

**Query Parameters:**

- `limit` (optional, default: 50) - Number of results
- `offset` (optional, default: 0) - Pagination offset

**Response:** `200 OK`

```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Amoxicillin 500mg",
      "brand": "Bayer",
      "dosage_form_id": 1,
      "strength": "500mg",
      "unit": "tablet",
      "category_id": 1,
      "description": "Broad-spectrum antibiotic",
      "created_at": "2025-11-10T10:30:00Z",
      "deleted_at": null
    }
  ]
}
```

**Example:**

```bash
curl -X GET "http://localhost:5582/api/v1/products?limit=20&offset=0" \
  -H "Authorization: Bearer $TOKEN"
```

---

### GET /api/v1/products/:id

Get specific product by ID.

**Authentication:** Required

**Path Parameters:**

- `id` (required) - Product UUID

**Response:** `200 OK`

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Amoxicillin 500mg",
    "brand": "Bayer",
    "dosage_form_id": 1,
    "strength": "500mg",
    "unit": "tablet",
    "category_id": 1,
    "description": "Broad-spectrum antibiotic",
    "created_at": "2025-11-10T10:30:00Z",
    "deleted_at": null
  }
}
```

**Errors:**

- `404 Not Found` - Product doesn't exist
- `400 Bad Request` - Invalid UUID format

---

### PUT /api/v1/products/:id

Update existing product.

**Authentication:** Required  
**Roles:** admin, pharmacist

**Path Parameters:**

- `id` (required) - Product UUID

**Request Body:**

```json
{
  "name": "string (optional)",
  "brand": "string (optional)",
  "dosage_form_id": "integer (optional)",
  "strength": "string (optional)",
  "unit": "string (optional)",
  "category_id": "integer (optional)",
  "description": "string (optional)"
}
```

**Response:** `200 OK`

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Amoxicillin 1000mg",
    "brand": "Bayer",
    ...
  }
}
```

**Example:**

```bash
curl -X PUT http://localhost:5582/api/v1/products/550e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"strength": "1000mg"}'
```

---

### DELETE /api/v1/products/:id

Delete product (soft delete).

**Authentication:** Required  
**Roles:** admin

**Path Parameters:**

- `id` (required) - Product UUID

**Response:** `204 No Content`

**Example:**

```bash
curl -X DELETE http://localhost:5582/api/v1/products/550e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer $TOKEN"
```

---

### GET /api/v1/products/search

Search products by name or brand.

**Authentication:** Required

**Query Parameters:**

- `q` (required) - Search query
- `limit` (optional, default: 50)
- `offset` (optional, default: 0)

**Response:** `200 OK`

```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Amoxicillin 500mg",
      "brand": "Bayer",
      ...
    }
  ]
}
```

**Example:**

```bash
curl -X GET "http://localhost:5582/api/v1/products/search?q=amoxicillin&limit=10" \
  -H "Authorization: Bearer $TOKEN"
```

---

### GET /api/v1/products/barcode/:barcode

Search product by barcode.

**Authentication:** Required

**Path Parameters:**

- `barcode` (required) - Barcode string

**Response:** `200 OK`

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Amoxicillin 500mg",
    "brand": "Bayer",
    ...
  }
}
```

**Errors:**

- `404 Not Found` - Barcode not found

**Example:**

```bash
curl -X GET http://localhost:5582/api/v1/products/barcode/5901234123457 \
  -H "Authorization: Bearer $TOKEN"
```

---

## Orders

### POST /api/v1/orders

Create new order.

**Authentication:** Required

**Request Body:**

```json
{
  "created_by": "uuid (optional)",
  "status": "string (required: draft|submitted|processing|completed|cancelled)",
  "notes": "string (optional)"
}
```

**Response:** `201 Created`

```json
{
  "data": {
    "id": "650e8400-e29b-41d4-a716-446655440001",
    "created_by": "550e8400-e29b-41d4-a716-446655440000",
    "status": "draft",
    "created_at": "2025-11-10T10:30:00Z",
    "submitted_at": null,
    "notes": "Weekly order",
    "deleted_at": null
  }
}
```

**Example:**

```bash
curl -X POST http://localhost:5582/api/v1/orders \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "draft",
    "notes": "Weekly order"
  }'
```

---

### GET /api/v1/orders

List all orders with pagination.

**Authentication:** Required

**Query Parameters:**

- `limit` (optional, default: 50)
- `offset` (optional, default: 0)
- `user_id` (optional) - Filter by user UUID

**Response:** `200 OK`

```json
{
  "data": [
    {
      "id": "650e8400-e29b-41d4-a716-446655440001",
      "created_by": "550e8400-e29b-41d4-a716-446655440000",
      "status": "submitted",
      "created_at": "2025-11-10T10:30:00Z",
      "submitted_at": "2025-11-10T11:00:00Z",
      "notes": "Weekly order",
      "deleted_at": null
    }
  ]
}
```

---

### GET /api/v1/orders/:id

Get specific order.

**Authentication:** Required

**Path Parameters:**

- `id` (required) - Order UUID

**Response:** `200 OK`

```json
{
  "data": {
    "id": "650e8400-e29b-41d4-a716-446655440001",
    "created_by": "550e8400-e29b-41d4-a716-446655440000",
    "status": "submitted",
    "created_at": "2025-11-10T10:30:00Z",
    "submitted_at": "2025-11-10T11:00:00Z",
    "notes": "Weekly order",
    "deleted_at": null
  }
}
```

---

### PUT /api/v1/orders/:id/status

Update order status.

**Authentication:** Required

**Path Parameters:**

- `id` (required) - Order UUID

**Request Body:**

```json
{
  "status": "string (required: draft|submitted|processing|completed|cancelled)"
}
```

**Response:** `200 OK`

```json
{
  "data": {
    "id": "650e8400-e29b-41d4-a716-446655440001",
    "status": "submitted",
    "submitted_at": "2025-11-10T11:00:00Z",
    ...
  }
}
```

---

### DELETE /api/v1/orders/:id

Delete order (soft delete).

**Authentication:** Required  
**Roles:** admin

**Path Parameters:**

- `id` (required) - Order UUID

**Response:** `204 No Content`

---

## Order Items

### POST /api/v1/orders/:order_id/items

Add item to order.

**Authentication:** Required

**Path Parameters:**

- `order_id` (required) - Order UUID

**Request Body:**

```json
{
  "product_id": "uuid (required)",
  "requested_qty": "integer (required, >0)",
  "unit": "string (optional)",
  "note": "string (optional)"
}
```

**Response:** `201 Created`

```json
{
  "data": {
    "id": "750e8400-e29b-41d4-a716-446655440002",
    "order_id": "650e8400-e29b-41d4-a716-446655440001",
    "product_id": "550e8400-e29b-41d4-a716-446655440000",
    "requested_qty": 100,
    "unit": "tablets",
    "note": "Urgent"
  }
}
```

**Example:**

```bash
curl -X POST http://localhost:5582/api/v1/orders/650e8400.../items \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": "550e8400-e29b-41d4-a716-446655440000",
    "requested_qty": 100,
    "unit": "tablets",
    "note": "Urgent"
  }'
```

---

### GET /api/v1/orders/:order_id/items

Get all items in order.

**Authentication:** Required

**Path Parameters:**

- `order_id` (required) - Order UUID

**Response:** `200 OK`

```json
{
  "data": [
    {
      "id": "750e8400-e29b-41d4-a716-446655440002",
      "order_id": "650e8400-e29b-41d4-a716-446655440001",
      "product_id": "550e8400-e29b-41d4-a716-446655440000",
      "requested_qty": 100,
      "unit": "tablets",
      "note": "Urgent"
    }
  ]
}
```

---

### PUT /api/v1/order_items/:id

Update order item.

**Authentication:** Required

**Path Parameters:**

- `id` (required) - Order item UUID

**Request Body:**

```json
{
  "requested_qty": "integer (required, >0)",
  "unit": "string (optional)",
  "note": "string (optional)"
}
```

**Response:** `200 OK`

```json
{
  "data": {
    "id": "750e8400-e29b-41d4-a716-446655440002",
    "requested_qty": 150,
    "unit": "tablets",
    "note": "Updated quantity"
  }
}
```

---

### DELETE /api/v1/order_items/:id

Delete order item.

**Authentication:** Required

**Path Parameters:**

- `id` (required) - Order item UUID

**Response:** `204 No Content`

---

## Users

### POST /api/v1/users

Create new user (admin only).

**Authentication:** Required  
**Roles:** admin

**Request Body:**

```json
{
  "username": "string (required, min 3, max 50)",
  "full_name": "string (optional)",
  "password": "string (required, min 8)",
  "role_id": "integer (required, >0)"
}
```

**Response:** `201 Created`

```json
{
  "data": {
    "id": "850e8400-e29b-41d4-a716-446655440003",
    "username": "pharmacist1",
    "full_name": "John Doe",
    "role_id": 2,
    "created_at": "2025-11-10T10:30:00Z",
    "deleted_at": null
  }
}
```

**Errors:**

- `409 Conflict` - Username already exists
- `403 Forbidden` - Non-admin user

---

### GET /api/v1/users

List all users.

**Authentication:** Required  
**Roles:** admin

**Query Parameters:**

- `limit` (optional, default: 50)
- `offset` (optional, default: 0)

**Response:** `200 OK`

```json
{
  "data": [
    {
      "id": "850e8400-e29b-41d4-a716-446655440003",
      "username": "pharmacist1",
      "full_name": "John Doe",
      "role_id": 2,
      "role_name": "pharmacist",
      "created_at": "2025-11-10T10:30:00Z"
    }
  ]
}
```

---

### GET /api/v1/users/:id

Get specific user.

**Authentication:** Required  
**Roles:** admin

**Path Parameters:**

- `id` (required) - User UUID

**Response:** `200 OK`

```json
{
  "data": {
    "id": "850e8400-e29b-41d4-a716-446655440003",
    "username": "pharmacist1",
    "full_name": "John Doe",
    "role_id": 2,
    "role_name": "pharmacist",
    "created_at": "2025-11-10T10:30:00Z"
  }
}
```

---

### PUT /api/v1/users/:id

Update user.

**Authentication:** Required  
**Roles:** admin

**Path Parameters:**

- `id` (required) - User UUID

**Request Body:**

```json
{
  "full_name": "string (optional)",
  "role_id": "integer (optional)"
}
```

**Response:** `200 OK`

```json
{
  "data": {
    "id": "850e8400-e29b-41d4-a716-446655440003",
    "username": "pharmacist1",
    "full_name": "John Smith",
    "role_id": 2,
    ...
  }
}
```

---

### DELETE /api/v1/users/:id

Delete user (soft delete).

**Authentication:** Required  
**Roles:** admin

**Path Parameters:**

- `id` (required) - User UUID

**Response:** `204 No Content`

**Errors:**

- `403 Forbidden` - Cannot delete primary admin
- `403 Forbidden` - Cannot delete last admin

---

## Roles

### POST /api/v1/roles

Create new role.

**Authentication:** Required  
**Roles:** admin

**Request Body:**

```json
{
  "name": "string (required)"
}
```

**Response:** `201 Created`

```json
{
  "data": {
    "id": 4,
    "name": "supervisor"
  }
}
```

---

### GET /api/v1/roles

List all roles.

**Authentication:** Required  
**Roles:** admin

**Response:** `200 OK`

```json
{
  "data": [
    { "id": 1, "name": "admin" },
    { "id": 2, "name": "pharmacist" },
    { "id": 3, "name": "clerk" }
  ]
}
```

---

### GET /api/v1/roles/:id

Get specific role.

**Authentication:** Required  
**Roles:** admin

**Path Parameters:**

- `id` (required) - Role ID

**Response:** `200 OK`

```json
{
  "data": {
    "id": 1,
    "name": "admin"
  }
}
```

---

### PUT /api/v1/roles/:id

Update role.

**Authentication:** Required  
**Roles:** admin

**Path Parameters:**

- `id` (required) - Role ID

**Request Body:**

```json
{
  "name": "string (required)"
}
```

**Response:** `200 OK`

```json
{
  "data": {
    "id": 4,
    "name": "senior_supervisor"
  }
}
```

---

### DELETE /api/v1/roles/:id

Delete role.

**Authentication:** Required  
**Roles:** admin

**Path Parameters:**

- `id` (required) - Role ID

**Response:** `204 No Content`

**Errors:**

- `500 Internal Server Error` - Role in use by users

---

## Permissions

### POST /api/v1/permissions

Create new permission.

**Authentication:** Required  
**Roles:** admin

**Request Body:**

```json
{
  "name": "string (required, min 3, max 100)",
  "resource": "string (required)",
  "action": "string (required)",
  "description": "string (optional)"
}
```

**Response:** `201 Created`

```json
{
  "data": {
    "id": 15,
    "name": "export_reports",
    "resource": "reports",
    "action": "export",
    "description": "Export system reports",
    "created_at": "2025-11-10T10:30:00Z"
  }
}
```

---

### GET /api/v1/permissions

List all permissions.

**Authentication:** Required  
**Roles:** admin

**Query Parameters:**

- `limit` (optional, default: 100)
- `offset` (optional, default: 0)
- `resource` (optional) - Filter by resource

**Response:** `200 OK`

```json
{
  "data": [
    {
      "id": 1,
      "name": "view_products",
      "resource": "products",
      "action": "read",
      "description": "View products",
      "created_at": "2025-11-10T10:30:00Z"
    }
  ]
}
```

---

### POST /api/v1/roles/:role_id/permissions

Assign permission to role.

**Authentication:** Required  
**Roles:** admin

**Path Parameters:**

- `role_id` (required) - Role ID

**Request Body:**

```json
{
  "permission_id": "integer (required)"
}
```

**Response:** `201 Created`

```json
{
  "data": {
    "id": 25,
    "role_id": 2,
    "permission_id": 15,
    "created_at": "2025-11-10T10:30:00Z"
  }
}
```

---

### GET /api/v1/roles/:role_id/permissions

Get role permissions.

**Authentication:** Required

**Path Parameters:**

- `role_id` (required) - Role ID

**Response:** `200 OK`

```json
{
  "data": [
    {
      "id": 1,
      "name": "view_products",
      "resource": "products",
      "action": "read",
      "description": "View products",
      "created_at": "2025-11-10T10:30:00Z"
    }
  ]
}
```

---

## Categories

### POST /api/v1/categories

Create new category.

**Authentication:** Required  
**Roles:** admin

**Request Body:**

```json
{
  "name": "string (required)"
}
```

**Response:** `201 Created`

```json
{
  "data": {
    "id": 5,
    "name": "Cardiovascular"
  }
}
```

---

### GET /api/v1/categories

List all categories.

**Authentication:** Required

**Response:** `200 OK`

```json
{
  "data": [
    { "id": 1, "name": "دارویی" },
    { "id": 2, "name": "آرایشی" }
  ]
}
```

---

### GET /api/v1/categories/:id

Get specific category.

**Authentication:** Required

**Path Parameters:**

- `id` (required) - Category ID

**Response:** `200 OK`

```json
{
  "data": {
    "id": 1,
    "name": "دارویی"
  }
}
```

---

## Dosage Forms

### POST /api/v1/dosage_forms

Create new dosage form.

**Authentication:** Required  
**Roles:** admin

**Request Body:**

```json
{
  "name": "string (required)"
}
```

**Response:** `201 Created`

```json
{
  "data": {
    "id": 9,
    "name": "Injection"
  }
}
```

---

### GET /api/v1/dosage_forms

List all dosage forms.

**Authentication:** Required

**Response:** `200 OK`

```json
{
  "data": [
    { "id": 1, "name": "قرص" },
    { "id": 2, "name": "کپسول" }
  ]
}
```

---

### GET /api/v1/dosage_forms/:id

Get specific dosage form.

**Authentication:** Required

**Path Parameters:**

- `id` (required) - Dosage form ID

**Response:** `200 OK`

```json
{
  "data": {
    "id": 1,
    "name": "قرص"
  }
}
```

---

## Barcodes

### POST /api/v1/barcodes

Create new barcode for product.

**Authentication:** Required  
**Roles:** admin, pharmacist

**Request Body:**

```json
{
  "product_id": "uuid (required)",
  "barcode": "string (required)",
  "barcode_type": "string (optional: EAN-13|UPC-A|Code128)"
}
```

**Response:** `201 Created`

```json
{
  "data": {
    "id": "950e8400-e29b-41d4-a716-446655440004",
    "product_id": "550e8400-e29b-41d4-a716-446655440000",
    "barcode": "5901234123457",
    "barcode_type": "EAN-13",
    "created_at": "2025-11-10T10:30:00Z"
  }
}
```

---

### GET /api/v1/products/:product_id/barcodes

Get all barcodes for product.

**Authentication:** Required

**Path Parameters:**

- `product_id` (required) - Product UUID

**Response:** `200 OK`

```json
{
  "data": [
    {
      "id": "950e8400-e29b-41d4-a716-446655440004",
      "product_id": "550e8400-e29b-41d4-a716-446655440000",
      "barcode": "5901234123457",
      "barcode_type": "EAN-13",
      "created_at": "2025-11-10T10:30:00Z"
    }
  ]
}
```

---

### PUT /api/v1/barcodes/:id

Update barcode.

**Authentication:** Required  
**Roles:** admin, pharmacist

**Path Parameters:**

- `id` (required) - Barcode UUID

**Request Body:**

```json
{
  "barcode": "string (optional)",
  "barcode_type": "string (optional)"
}
```

**Response:** `200 OK`

```json
{
  "data": {
    "id": "950e8400-e29b-41d4-a716-446655440004",
    "barcode": "5901234123458",
    "barcode_type": "EAN-13",
    ...
  }
}
```

---

### DELETE /api/v1/barcodes/:id

Delete barcode.

**Authentication:** Required  
**Roles:** admin

**Path Parameters:**

- `id` (required) - Barcode UUID

**Response:** `204 No Content`

---

## Audit Logs

### GET /api/v1/audit-logs

List audit logs.

**Authentication:** Required  
**Roles:** admin

**Query Parameters:**

- `limit` (optional, default: 50)
- `offset` (optional, default: 0)
- `user_id` (optional) - Filter by user
- `entity_type` (optional) - Filter by entity type
- `entity_id` (optional) - Filter by entity ID
- `action` (optional) - Filter by action

**Response:** `200 OK`

```json
{
```
