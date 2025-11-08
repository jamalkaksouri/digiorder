# Authentication & Authorization Guide

## Overview

DigiOrder v2.0 implements JWT-based authentication with role-based access control (RBAC). All API endpoints (except login and health check) require authentication.

---

## Authentication Flow

### 1. Login

**Endpoint:** `POST /api/v1/auth/login`

**Request:**

```json
{
  "username": "admin",
  "password": "your_password"
}
```

**Response:**

```json
{
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": "24h",
    "user": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "username": "admin",
      "full_name": "مدیر سیستم",
      "role_id": 1,
      "role_name": "admin"
    }
  }
}
```

**Example:**

```bash
# Login and save token
TOKEN=$(curl -X POST http://localhost:5582/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123456"
  }' | jq -r '.data.token')

echo "Token: $TOKEN"
```

### 2. Using the Token

Include the token in the `Authorization` header for all subsequent requests:

```bash
curl http://localhost:5582/api/v1/products \
  -H "Authorization: Bearer $TOKEN"
```

### 3. Token Refresh

**Endpoint:** `POST /api/v1/auth/refresh`

**Request:**

```json
{
  "token": "your_current_token"
}
```

**Response:**

```json
{
  "data": {
    "token": "new_token_here",
    "expires_in": "24h"
  }
}
```

---

## Protected Endpoints

### Get User Profile

**Endpoint:** `GET /api/v1/auth/profile`

**Headers:**

```
Authorization: Bearer <token>
```

**Response:**

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "admin",
    "full_name": "مدیر سیستم",
    "role_id": 1,
    "role_name": "admin"
  }
}
```

### Change Password

**Endpoint:** `PUT /api/v1/auth/password`

**Request:**

```json
{
  "old_password": "current_password",
  "new_password": "new_password_here"
}
```

**Response:**

```json
{
  "data": {
    "message": "Password updated successfully"
  }
}
```

---

## Role-Based Access Control (RBAC)

### Default Roles

| Role       | ID  | Permissions                           |
| ---------- | --- | ------------------------------------- |
| admin      | 1   | Full access to all endpoints          |
| pharmacist | 2   | Can manage products, orders, barcodes |
| clerk      | 3   | Can view and create orders            |

### Endpoint Permissions

| Endpoint             | Admin | Pharmacist | Clerk |
| -------------------- | ----- | ---------- | ----- |
| POST /products       | ✅    | ✅         | ❌    |
| GET /products        | ✅    | ✅         | ✅    |
| PUT /products/:id    | ✅    | ✅         | ❌    |
| DELETE /products/:id | ✅    | ❌         | ❌    |
| POST /orders         | ✅    | ✅         | ✅    |
| GET /orders          | ✅    | ✅         | ✅    |
| DELETE /orders/:id   | ✅    | ❌         | ❌    |
| POST /users          | ✅    | ❌         | ❌    |
| POST /roles          | ✅    | ❌         | ❌    |
| POST /barcodes       | ✅    | ✅         | ❌    |

---

## Error Responses

### 401 Unauthorized

Missing or invalid token:

```json
{
  "error": "missing authorization header",
  "details": ""
}
```

```json
{
  "error": "invalid or expired token",
  "details": ""
}
```

### 403 Forbidden

Insufficient permissions:

```json
{
  "error": "insufficient permissions",
  "details": ""
}
```

---

## JWT Token Details

### Token Structure

The JWT token contains:

```json
{
  "user_id": "uuid",
  "username": "string",
  "role_id": 1,
  "role_name": "admin",
  "exp": 1234567890,
  "iat": 1234567890,
  "nbf": 1234567890
}
```

### Token Expiry

- Default: 24 hours
- Configurable via `JWT_EXPIRY` environment variable
- Format: Go duration string (e.g., "24h", "30m", "2h30m")

---

## Security Best Practices

### 1. Store Tokens Securely

**✅ Good:**

- Store in memory (for SPAs)
- Use HttpOnly cookies
- Store in secure storage (mobile apps)

**❌ Bad:**

- LocalStorage (vulnerable to XSS)
- SessionStorage without encryption
- Exposed in URL parameters

### 2. Token Management

```javascript
// Example: Axios interceptor
axios.interceptors.request.use((config) => {
  const token = getSecureToken();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Handle 401 responses
axios.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (error.response?.status === 401) {
      // Token expired, redirect to login
      window.location.href = "/login";
    }
    return Promise.reject(error);
  }
);
```

### 3. Password Requirements

- Minimum 6 characters
- Bcrypt hashed with cost factor 10
- Recommend: 12+ characters with mixed case, numbers, symbols

---

## Complete Authentication Workflow

```bash
#!/bin/bash

BASE_URL="http://localhost:5582"

echo "=== DigiOrder Authentication Workflow ==="

# 1. First user (admin) needs to be created via DB or migration
echo -e "\n1. Creating initial admin user..."
# (This would typically be done via migration or direct DB insert)

# 2. Login
echo -e "\n2. Logging in..."
LOGIN_RESPONSE=$(curl -s -X POST $BASE_URL/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123456"
  }')

TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.data.token')
echo "Token obtained: ${TOKEN:0:50}..."

# 3. Get profile
echo -e "\n3. Getting user profile..."
curl -s $BASE_URL/api/v1/auth/profile \
  -H "Authorization: Bearer $TOKEN" | jq

# 4. Access protected resource
echo -e "\n4. Accessing products (protected)..."
curl -s $BASE_URL/api/v1/products \
  -H "Authorization: Bearer $TOKEN" | jq '.data[0:2]'

# 5. Create product (requires pharmacist or admin role)
echo -e "\n5. Creating product (admin only)..."
curl -s -X POST $BASE_URL/api/v1/products \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Medicine",
    "brand": "Test Brand",
    "dosage_form_id": 1,
    "strength": "100mg",
    "unit": "tablet",
    "category_id": 1,
    "description": "Test"
  }' | jq

# 6. Change password
echo -e "\n6. Changing password..."
curl -s -X PUT $BASE_URL/api/v1/auth/password \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "old_password": "admin123456",
    "new_password": "new_admin_password"
  }' | jq

# 7. Refresh token
echo -e "\n7. Refreshing token..."
NEW_TOKEN_RESPONSE=$(curl -s -X POST $BASE_URL/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{\"token\": \"$TOKEN\"}")

NEW_TOKEN=$(echo $NEW_TOKEN_RESPONSE | jq -r '.data.token')
echo "New token: ${NEW_TOKEN:0:50}..."

echo -e "\n=== Workflow Complete ==="
```

---

## Multi-User Scenario

```bash
# Admin creates pharmacist user
curl -X POST http://localhost:5582/api/v1/users \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "pharmacist1",
    "full_name": "داروساز احمد",
    "password": "pharm123456",
    "role_id": 2
  }'

# Pharmacist logs in
PHARM_TOKEN=$(curl -s -X POST http://localhost:5582/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "pharmacist1",
    "password": "pharm123456"
  }' | jq -r '.data.token')

# Pharmacist can create products
curl -X POST http://localhost:5582/api/v1/products \
  -H "Authorization: Bearer $PHARM_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{ ... }'

# But cannot create users (403 Forbidden)
curl -X POST http://localhost:5582/api/v1/users \
  -H "Authorization: Bearer $PHARM_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{ ... }'
```

---

## Testing Authentication

### Postman Collection

1. Create environment variables:

   - `BASE_URL`: http://localhost:5582
   - `TOKEN`: (will be set automatically)

2. Pre-request script for login:

```javascript
pm.sendRequest(
  {
    url: pm.environment.get("BASE_URL") + "/api/v1/auth/login",
    method: "POST",
    header: {
      "Content-Type": "application/json",
    },
    body: {
      mode: "raw",
      raw: JSON.stringify({
        username: "admin",
        password: "admin123456",
      }),
    },
  },
  (err, res) => {
    if (!err) {
      const token = res.json().data.token;
      pm.environment.set("TOKEN", token);
    }
  }
);
```

3. Authorization header for all protected requests:

```
Bearer {{TOKEN}}
```

---

## Troubleshooting

### Token Not Working

1. Check token format:

   ```bash
   echo $TOKEN | cut -d'.' -f1 | base64 -d
   ```

2. Verify expiry:

   ```bash
   echo $TOKEN | cut -d'.' -f2 | base64 -d | jq
   ```

3. Check server logs for detailed error messages

### 401 Errors

- Verify `Authorization` header is set
- Check token hasn't expired
- Ensure `Bearer ` prefix is included
- Verify JWT_SECRET matches between login and validation

### 403 Errors

- Check user's role in token claims
- Verify endpoint permissions
- Ensure role hasn't changed (may need to re-login)

---

## Security Configuration

### Environment Variables

```bash
# JWT Configuration
JWT_SECRET=your_very_long_secret_key_minimum_64_characters_recommended
JWT_EXPIRY=24h

# Password Requirements (configured in validator)
MIN_PASSWORD_LENGTH=6
```

### Recommended JWT Secret Generation

```bash
# Generate random secret
openssl rand -base64 64

# Or using pwgen
pwgen -s 64 1
```

---

## Summary

- ✅ JWT-based authentication
- ✅ Role-based access control
- ✅ Secure password hashing (bcrypt)
- ✅ Token refresh mechanism
- ✅ Profile management
- ✅ Password change functionality
- ✅ Comprehensive error handling

All endpoints except `/health` and `/api/v1/auth/login` require authentication!
