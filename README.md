# DigiOrder - Pharmacy Order Management System

A Go-based order management system for pharmacies built with Echo framework and PostgreSQL.

## Prerequisites

- Go 1.22 or higher
- PostgreSQL 15 or higher (or Docker)
- Make (optional but recommended)

## Quick Start

### 1. Clone the repository

```bash
git clone https://github.com/jamalkaksouri/DigiOrder.git
cd DigiOrder
```

### 2. Set up environment variables

```bash
cp .env.example .env
# Edit .env with your database credentials
```

### 3. Install dependencies

```bash
go mod download
```

### 4. Set up the database

#### Option A: Using Docker (Recommended)

```bash
# Start PostgreSQL and run migrations
docker-compose up -d

# Or using Make
make docker-up
make migrate-up
```

#### Option B: Using existing PostgreSQL

```bash
# Install migration tool
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run migrations
make migrate-up
```

### 5. Generate SQLC code

```bash
# Install SQLC
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Generate code
make sqlc
```

### 6. Run the application

```bash
# Run directly
go run cmd/main.go

# Or using Make
make run

# Or build and run
make build
./bin/digiorder
```

The server will start on `http://localhost:5582`

## API Endpoints

### Health Check

- `GET /health` - Check API health status

```bash
curl http://localhost:5582/health
```

### Products

- `POST /api/v1/products` - Create a new product
- `GET /api/v1/products` - List products with pagination
- `GET /api/v1/products/:id` - Get a specific product
- `PUT /api/v1/products/:id` - Update a product
- `DELETE /api/v1/products/:id` - Delete a product
- `GET /api/v1/products/search?q=query` - Search products

#### Examples:

```bash
# Create Product
curl -X POST http://localhost:5582/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "آسپرین",
    "brand": "Bayer",
    "dosage_form_id": 1,
    "strength": "500mg",
    "unit": "بسته",
    "category_id": 1,
    "description": "مسکن و ضد التهاب"
  }'

# List Products
curl "http://localhost:5582/api/v1/products?limit=10&offset=0"

# Get Product
curl http://localhost:5582/api/v1/products/{product_id}

# Update Product
curl -X PUT http://localhost:5582/api/v1/products/{product_id} \
  -H "Content-Type: application/json" \
  -d '{
    "name": "آسپرین 500",
    "brand": "Bayer"
  }'

# Delete Product
curl -X DELETE http://localhost:5582/api/v1/products/{product_id}

# Search Products
curl "http://localhost:5582/api/v1/products/search?q=آسپرین&limit=10"
```

### Categories

- `POST /api/v1/categories` - Create a new category
- `GET /api/v1/categories` - List all categories
- `GET /api/v1/categories/:id` - Get a specific category

#### Examples:

```bash
# Create Category
curl -X POST http://localhost:5582/api/v1/categories \
  -H "Content-Type: application/json" \
  -d '{"name": "داروهای قلبی"}'

# List Categories
curl http://localhost:5582/api/v1/categories

# Get Category
curl http://localhost:5582/api/v1/categories/1
```

### Dosage Forms

- `POST /api/v1/dosage_forms` - Create a new dosage form
- `GET /api/v1/dosage_forms` - List all dosage forms
- `GET /api/v1/dosage_forms/:id` - Get a specific dosage form

#### Examples:

```bash
# Create Dosage Form
curl -X POST http://localhost:5582/api/v1/dosage_forms \
  -H "Content-Type: application/json" \
  -d '{"name": "محلول"}'

# List Dosage Forms
curl http://localhost:5582/api/v1/dosage_forms

# Get Dosage Form
curl http://localhost:5582/api/v1/dosage_forms/1
```

### Orders

- `POST /api/v1/orders` - Create a new order
- `GET /api/v1/orders` - List orders with pagination
- `GET /api/v1/orders?user_id={uuid}` - List orders by user
- `GET /api/v1/orders/:id` - Get a specific order
- `PUT /api/v1/orders/:id/status` - Update order status
- `DELETE /api/v1/orders/:id` - Delete an order

#### Examples:

```bash
# Create Order
curl -X POST http://localhost:5582/api/v1/orders \
  -H "Content-Type: application/json" \
  -d '{
    "status": "draft",
    "notes": "سفارش تست"
  }'

# List Orders
curl "http://localhost:5582/api/v1/orders?limit=10&offset=0"

# List Orders by User
curl "http://localhost:5582/api/v1/orders?user_id={user_uuid}&limit=10"

# Get Order
curl http://localhost:5582/api/v1/orders/{order_id}

# Update Order Status
curl -X PUT http://localhost:5582/api/v1/orders/{order_id}/status \
  -H "Content-Type: application/json" \
  -d '{"status": "submitted"}'

# Delete Order
curl -X DELETE http://localhost:5582/api/v1/orders/{order_id}
```

### Order Items

- `POST /api/v1/orders/:order_id/items` - Add item to order
- `GET /api/v1/orders/:order_id/items` - Get all items in order
- `PUT /api/v1/order_items/:id` - Update an order item
- `DELETE /api/v1/order_items/:id` - Delete an order item

#### Examples:

```bash
# Add Item to Order
curl -X POST http://localhost:5582/api/v1/orders/{order_id}/items \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": "{product_uuid}",
    "requested_qty": 10,
    "unit": "بسته",
    "note": "فوری"
  }'

# Get Order Items
curl http://localhost:5582/api/v1/orders/{order_id}/items

# Update Order Item
curl -X PUT http://localhost:5582/api/v1/order_items/{item_id} \
  -H "Content-Type: application/json" \
  -d '{
    "requested_qty": 15,
    "unit": "بسته"
  }'

# Delete Order Item
curl -X DELETE http://localhost:5582/api/v1/order_items/{item_id}
```

### Users

- `POST /api/v1/users` - Create a new user
- `GET /api/v1/users` - List users with pagination
- `GET /api/v1/users/:id` - Get a specific user
- `PUT /api/v1/users/:id` - Update a user
- `DELETE /api/v1/users/:id` - Delete a user

#### Examples:

```bash
# Create User
curl -X POST http://localhost:5582/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "username": "pharmacist1",
    "full_name": "علی احمدی",
    "password": "secure123",
    "role_id": 2
  }'

# List Users
curl "http://localhost:5582/api/v1/users?limit=10&offset=0"

# Get User
curl http://localhost:5582/api/v1/users/{user_id}

# Update User
curl -X PUT http://localhost:5582/api/v1/users/{user_id} \
  -H "Content-Type: application/json" \
  -d '{
    "full_name": "علی احمدی نژاد",
    "role_id": 1
  }'

# Delete User
curl -X DELETE http://localhost:5582/api/v1/users/{user_id}
```

### Roles

- `POST /api/v1/roles` - Create a new role
- `GET /api/v1/roles` - List all roles
- `GET /api/v1/roles/:id` - Get a specific role
- `PUT /api/v1/roles/:id` - Update a role
- `DELETE /api/v1/roles/:id` - Delete a role

#### Examples:

```bash
# Create Role
curl -X POST http://localhost:5582/api/v1/roles \
  -H "Content-Type: application/json" \
  -d '{"name": "supervisor"}'

# List Roles
curl http://localhost:5582/api/v1/roles

# Get Role
curl http://localhost:5582/api/v1/roles/1

# Update Role
curl -X PUT http://localhost:5582/api/v1/roles/4 \
  -H "Content-Type: application/json" \
  -d '{"name": "senior_supervisor"}'

# Delete Role
curl -X DELETE http://localhost:5582/api/v1/roles/4
```

## Response Format

### Success Response

```json
{
  "data": {
    // Response data here
  }
}
```

### Error Response

```json
{
  "error": "error_code",
  "details": "Detailed error message"
}
```

## Development

### Available Make Commands

```bash
make help           # Show all available commands
make build          # Build the application
make run            # Run the application
make test           # Run tests
make migrate-up     # Run database migrations
make migrate-down   # Rollback migrations
make sqlc           # Generate SQLC code
make docker-up      # Start PostgreSQL in Docker
make docker-down    # Stop PostgreSQL container
make lint           # Run linter
make fmt            # Format code
make clean          # Clean build artifacts
```

### Project Structure

```
DigiOrder/
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   ├── db/
│   │   ├── connection.go       # Database connection
│   │   ├── db.go              # SQLC base code
│   │   ├── models.go          # SQLC models
│   │   ├── products.sql.go    # Product queries
│   │   ├── orders.sql.go      # Order queries
│   │   ├── users.sql.go       # User queries
│   │   ├── categories.sql.go  # Category queries
│   │   └── query/             # SQL definitions
│   └── server/
│       ├── server.go          # Server setup
│       ├── routes.go          # Route definitions
│       ├── response.go        # Response helpers
│       ├── products.go        # Product handlers
│       ├── orders.go          # Order handlers
│       ├── users.go           # User handlers
│       ├── categories.go      # Category handlers
│       ├── dosage_forms.go    # Dosage form handlers
│       └── roles.go           # Role handlers
├── migrations/                # Database migrations
├── docker-compose.yml         # Docker configuration
├── sqlc.yaml                  # SQLC configuration
├── Makefile                   # Build automation
└── README.md                  # This file
```

## Database Schema

The system includes the following main tables:

- **roles** - User roles (admin, pharmacist, clerk)
- **users** - System users
- **categories** - Product categories
- **dosage_forms** - Medicine forms (tablets, syrup, etc.)
- **products** - Medicine and product catalog
- **product_barcodes** - Product barcode information
- **orders** - Order management
- **order_items** - Order line items

## Testing

```bash
# Run all tests
make test

# Run tests with coverage
go test -v -cover ./...
```

## Error Handling

All endpoints return consistent error responses with proper HTTP status codes:

- `400` - Bad Request (validation errors, invalid input)
- `404` - Not Found (resource doesn't exist)
- `409` - Conflict (duplicate entries)
- `500` - Internal Server Error (database or server errors)

## Security

- Passwords are hashed using bcrypt
- Input validation on all endpoints
- SQL injection protection via SQLC
- CORS enabled for cross-origin requests

## Troubleshooting

### Database Connection Issues

1. Ensure PostgreSQL is running:

   ```bash
   docker ps  # If using Docker
   ```

2. Check database credentials in `.env` file

3. Verify database exists:
   ```bash
   psql -U postgres -h localhost -c "\l"
   ```

### SQLC Generation Issues

1. Ensure SQLC is installed:

   ```bash
   which sqlc
   ```

2. Verify `sqlc.yaml` configuration

3. Check SQL syntax in query files

### Migration Issues

1. Ensure migration tool is installed:

   ```bash
   which migrate
   ```

2. Check migration files syntax

3. Verify database connection string

## License

MIT

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.
