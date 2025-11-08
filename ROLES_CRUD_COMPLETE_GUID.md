# Complete CRUD Guide for Roles

## Overview

Full CRUD (Create, Read, Update, Delete) operations are now available for role management.

---

## API Endpoints Summary

| Method | Endpoint          | Description         |
| ------ | ----------------- | ------------------- |
| POST   | /api/v1/roles     | Create a new role   |
| GET    | /api/v1/roles     | List all roles      |
| GET    | /api/v1/roles/:id | Get a specific role |
| PUT    | /api/v1/roles/:id | Update a role       |
| DELETE | /api/v1/roles/:id | Delete a role       |

---

## 1. CREATE - Add New Role

### Endpoint

```
POST /api/v1/roles
```

### Request Body

```json
{
  "name": "role_name"
}
```

### Example

```bash
curl -X POST http://localhost:5582/api/v1/roles \
  -H "Content-Type: application/json" \
  -d '{"name": "supervisor"}'
```

### Success Response (201 Created)

```json
{
  "data": {
    "id": 4,
    "name": "supervisor"
  }
}
```

### Error Responses

**Missing Name (400 Bad Request)**

```bash
curl -X POST http://localhost:5582/api/v1/roles \
  -H "Content-Type: application/json" \
  -d '{}'
```

Response:

```json
{
  "error": "validation_error",
  "details": "Key: 'CreateRoleReq.Name' Error:Field validation for 'Name' failed on the 'required' tag"
}
```

**Duplicate Name (500 Internal Server Error)**

```bash
curl -X POST http://localhost:5582/api/v1/roles \
  -H "Content-Type: application/json" \
  -d '{"name": "admin"}'
```

Response:

```json
{
  "error": "db_error",
  "details": "Failed to create role."
}
```

---

## 2. READ - Get Roles

### 2.1 List All Roles

#### Endpoint

```
GET /api/v1/roles
```

#### Example

```bash
curl http://localhost:5582/api/v1/roles
```

#### Success Response (200 OK)

```json
{
  "data": [
    {
      "id": 1,
      "name": "admin"
    },
    {
      "id": 2,
      "name": "pharmacist"
    },
    {
      "id": 3,
      "name": "clerk"
    },
    {
      "id": 4,
      "name": "supervisor"
    }
  ]
}
```

### 2.2 Get Single Role

#### Endpoint

```
GET /api/v1/roles/:id
```

#### Example

```bash
curl http://localhost:5582/api/v1/roles/4
```

#### Success Response (200 OK)

```json
{
  "data": {
    "id": 4,
    "name": "supervisor"
  }
}
```

#### Error Responses

**Invalid ID (400 Bad Request)**

```bash
curl http://localhost:5582/api/v1/roles/abc
```

Response:

```json
{
  "error": "invalid_id",
  "details": "The provided ID is not a valid number."
}
```

**Not Found (404 Not Found)**

```bash
curl http://localhost:5582/api/v1/roles/999
```

Response:

```json
{
  "error": "not_found",
  "details": "Role with the specified ID was not found."
}
```

---

## 3. UPDATE - Modify Role

### Endpoint

```
PUT /api/v1/roles/:id
```

### Request Body

```json
{
  "name": "new_role_name"
}
```

### Example

```bash
curl -X PUT http://localhost:5582/api/v1/roles/4 \
  -H "Content-Type: application/json" \
  -d '{"name": "senior_supervisor"}'
```

### Success Response (200 OK)

```json
{
  "data": {
    "id": 4,
    "name": "senior_supervisor"
  }
}
```

### Error Responses

**Missing Name (400 Bad Request)**

```bash
curl -X PUT http://localhost:5582/api/v1/roles/4 \
  -H "Content-Type: application/json" \
  -d '{}'
```

Response:

```json
{
  "error": "validation_error",
  "details": "Key: 'UpdateRoleReq.Name' Error:Field validation for 'Name' failed on the 'required' tag"
}
```

**Not Found (404 Not Found)**

```bash
curl -X PUT http://localhost:5582/api/v1/roles/999 \
  -H "Content-Type: application/json" \
  -d '{"name": "test"}'
```

Response:

```json
{
  "error": "not_found",
  "details": "Role with the specified ID was not found."
}
```

**Duplicate Name (500 Internal Server Error)**

```bash
curl -X PUT http://localhost:5582/api/v1/roles/4 \
  -H "Content-Type: application/json" \
  -d '{"name": "admin"}'
```

Response:

```json
{
  "error": "db_error",
  "details": "Failed to update role."
}
```

---

## 4. DELETE - Remove Role

### Endpoint

```
DELETE /api/v1/roles/:id
```

### Example

```bash
curl -X DELETE http://localhost:5582/api/v1/roles/4
```

### Success Response (204 No Content)

```
(Empty response body)
```

### Error Responses

**Invalid ID (400 Bad Request)**

```bash
curl -X DELETE http://localhost:5582/api/v1/roles/abc
```

Response:

```json
{
  "error": "invalid_id",
  "details": "The provided ID is not a valid number."
}
```

**Role In Use (500 Internal Server Error)**

```bash
# Try to delete a role that users are currently assigned to
curl -X DELETE http://localhost:5582/api/v1/roles/1
```

Response:

```json
{
  "error": "db_error",
  "details": "Failed to delete role. It may be in use by existing users."
}
```

---

## Complete Workflow Example

```bash
#!/bin/bash

BASE_URL="http://localhost:5582"

echo "=== Role CRUD Complete Workflow ==="

# 1. CREATE - Add new roles
echo -e "\n1. Creating new roles..."

curl -s -X POST $BASE_URL/api/v1/roles \
  -H "Content-Type: application/json" \
  -d '{"name": "supervisor"}' | jq
SUPERVISOR_ID=4

curl -s -X POST $BASE_URL/api/v1/roles \
  -H "Content-Type: application/json" \
  -d '{"name": "manager"}' | jq
MANAGER_ID=5

curl -s -X POST $BASE_URL/api/v1/roles \
  -H "Content-Type: application/json" \
  -d '{"name": "trainee"}' | jq
TRAINEE_ID=6

# 2. READ - List all roles
echo -e "\n2. Listing all roles..."
curl -s $BASE_URL/api/v1/roles | jq

# 3. READ - Get specific role
echo -e "\n3. Getting supervisor role details..."
curl -s $BASE_URL/api/v1/roles/$SUPERVISOR_ID | jq

# 4. UPDATE - Modify role name
echo -e "\n4. Updating supervisor to senior_supervisor..."
curl -s -X PUT $BASE_URL/api/v1/roles/$SUPERVISOR_ID \
  -H "Content-Type: application/json" \
  -d '{"name": "senior_supervisor"}' | jq

# 5. READ - Verify update
echo -e "\n5. Verifying update..."
curl -s $BASE_URL/api/v1/roles/$SUPERVISOR_ID | jq

# 6. UPDATE - Rename another role
echo -e "\n6. Updating trainee to junior_staff..."
curl -s -X PUT $BASE_URL/api/v1/roles/$TRAINEE_ID \
  -H "Content-Type: application/json" \
  -d '{"name": "junior_staff"}' | jq

# 7. READ - List all roles to see changes
echo -e "\n7. Final roles list..."
curl -s $BASE_URL/api/v1/roles | jq

# 8. DELETE - Remove a role
echo -e "\n8. Deleting junior_staff role..."
curl -s -X DELETE $BASE_URL/api/v1/roles/$TRAINEE_ID
echo "(No content - success)"

# 9. READ - Verify deletion
echo -e "\n9. Final roles list after deletion..."
curl -s $BASE_URL/api/v1/roles | jq

echo -e "\n=== Workflow Complete ==="
```

Save as `role_crud_workflow.sh`, make executable, and run:

```bash
chmod +x role_crud_workflow.sh
./role_crud_workflow.sh
```

---

## Real-World Use Cases

### 1. Renaming Roles for Clarity

```bash
# Initial role
curl -X POST http://localhost:5582/api/v1/roles \
  -H "Content-Type: application/json" \
  -d '{"name": "staff"}'

# Realize it needs to be more specific
curl -X PUT http://localhost:5582/api/v1/roles/4 \
  -H "Content-Type: application/json" \
  -d '{"name": "pharmacy_staff"}'
```

### 2. Organizational Restructuring

```bash
# Company reorganizes departments
curl -X PUT http://localhost:5582/api/v1/roles/5 \
  -H "Content-Type: application/json" \
  -d '{"name": "department_head"}'

curl -X PUT http://localhost:5582/api/v1/roles/6 \
  -H "Content-Type: application/json" \
  -d '{"name": "team_leader"}'
```

### 3. Removing Obsolete Roles

```bash
# Delete temporary or unused roles
curl -X DELETE http://localhost:5582/api/v1/roles/10

# Check remaining roles
curl http://localhost:5582/api/v1/roles
```

---

## Testing Error Conditions

### 1. Test Create with Empty Name

```bash
curl -X POST http://localhost:5582/api/v1/roles \
  -H "Content-Type: application/json" \
  -d '{"name": ""}' | jq
```

### 2. Test Update Non-Existent Role

```bash
curl -X PUT http://localhost:5582/api/v1/roles/999 \
  -H "Content-Type: application/json" \
  -d '{"name": "test"}' | jq
```

### 3. Test Delete Non-Existent Role

```bash
curl -X DELETE http://localhost:5582/api/v1/roles/999 | jq
```

### 4. Test Create Duplicate Role

```bash
curl -X POST http://localhost:5582/api/v1/roles \
  -H "Content-Type: application/json" \
  -d '{"name": "admin"}' | jq
```

### 5. Test Update to Duplicate Name

```bash
curl -X PUT http://localhost:5582/api/v1/roles/4 \
  -H "Content-Type: application/json" \
  -d '{"name": "admin"}' | jq
```

---

## Best Practices

### 1. Before Deleting a Role

Always check if users are assigned to that role:

```bash
# Get the role you want to delete
curl http://localhost:5582/api/v1/roles/4 | jq

# Check users (you would need to implement this query)
# curl "http://localhost:5582/api/v1/users?role_id=4" | jq

# If no users, safe to delete
curl -X DELETE http://localhost:5582/api/v1/roles/4
```

### 2. Role Naming Conventions

✅ **Good:**

- `senior_pharmacist`
- `inventory_manager`
- `shift_supervisor`

❌ **Bad:**

- `Senior Pharmacist` (spaces)
- `SeniorPharmacist` (camelCase)
- `sp` (unclear abbreviation)

### 3. Updating vs Deleting

**When to Update:**

- Minor name corrections
- Clarifying role names
- Rebranding

**When to Delete:**

- Role is truly obsolete
- No users assigned
- Duplicate role exists

---

## Integration with User Management

### Create Role and Assign to User

```bash
# 1. Create new role
RESPONSE=$(curl -s -X POST http://localhost:5582/api/v1/roles \
  -H "Content-Type: application/json" \
  -d '{"name": "inventory_manager"}')

# Extract role ID
ROLE_ID=$(echo $RESPONSE | jq -r '.data.id')
echo "Created role with ID: $ROLE_ID"

# 2. Create user with this role
curl -X POST http://localhost:5582/api/v1/users \
  -H "Content-Type: application/json" \
  -d "{
    \"username\": \"inv_manager1\",
    \"full_name\": \"مدیر انبار\",
    \"password\": \"secure123\",
    \"role_id\": $ROLE_ID
  }" | jq
```

### Update User's Role

```bash
# Update user to a different role
curl -X PUT http://localhost:5582/api/v1/users/{user_id} \
  -H "Content-Type: application/json" \
  -d '{"role_id": 5}' | jq
```

---

## Database Constraints

### Unique Constraint

```sql
CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL  -- ← Cannot have duplicate names
);
```

### Foreign Key Constraint

```sql
CREATE TABLE users (
    ...
    role_id INT REFERENCES roles(id)  -- ← Cannot delete role if users exist
);
```

---

## Advanced Scenarios

### Bulk Role Creation

```bash
#!/bin/bash

roles=(
  "pharmacy_director"
  "clinical_coordinator"
  "quality_assurance"
  "compliance_officer"
  "purchasing_agent"
  "reception_staff"
)

for role in "${roles[@]}"; do
  echo "Creating role: $role"
  curl -s -X POST http://localhost:5582/api/v1/roles \
    -H "Content-Type: application/json" \
    -d "{\"name\": \"$role\"}" | jq -r '.data.name'
done
```

### Role Migration

```bash
#!/bin/bash

# Scenario: Renaming all "clerk" references to "front_desk"

# 1. Create new role
curl -X POST http://localhost:5582/api/v1/roles \
  -H "Content-Type: application/json" \
  -d '{"name": "front_desk"}'

# 2. Update all users from clerk (id: 3) to front_desk (id: 7)
# (Would need to implement batch user update)

# 3. Delete old clerk role
# curl -X DELETE http://localhost:5582/api/v1/roles/3
```

---

## Summary Table

| Operation | Method | Endpoint          | Body Required | Success Code |
| --------- | ------ | ----------------- | ------------- | ------------ |
| Create    | POST   | /api/v1/roles     | Yes           | 201          |
| List All  | GET    | /api/v1/roles     | No            | 200          |
| Get One   | GET    | /api/v1/roles/:id | No            | 200          |
| Update    | PUT    | /api/v1/roles/:id | Yes           | 200          |
| Delete    | DELETE | /api/v1/roles/:id | No            | 204          |

---

## Quick Reference Commands

```bash
# CREATE
curl -X POST http://localhost:5582/api/v1/roles -H "Content-Type: application/json" -d '{"name": "role_name"}'

# READ ALL
curl http://localhost:5582/api/v1/roles

# READ ONE
curl http://localhost:5582/api/v1/roles/4

# UPDATE
curl -X PUT http://localhost:5582/api/v1/roles/4 -H "Content-Type: application/json" -d '{"name": "new_name"}'

# DELETE
curl -X DELETE http://localhost:5582/api/v1/roles/4
```

---

## Conclusion

You now have complete CRUD operations for roles:

- ✅ **Create** new roles as needed
- ✅ **Read** all roles or specific ones
- ✅ **Update** role names for clarity
- ✅ **Delete** obsolete roles

This provides full flexibility for managing your organization's role structure!
