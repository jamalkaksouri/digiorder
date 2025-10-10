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
sqlc generate
# Or using Make
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

## API Endpoints

### Products

- `POST /api/v1/products` - Create a new product
- `GET /api/v1/products?limit=50&offset=0` - List products with pagination

#### Create Product Example

```bash
curl -X POST http://localhost:8080/api/v1/products \
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
```

#### List Products Example

```bash
curl http://localhost:8080/api/v1/products?limit=10&offset=0
```

## Development

### Available Make Commands

```bash
make help           # Show all available commands
make build          # Build the application
make run           # Run the application
make test          # Run tests
make migrate-up    # Run database migrations
make migrate-down  # Rollback migrations
make sqlc          # Generate SQLC code
make docker-up     # Start PostgreSQL in Docker
make docker-down   # Stop PostgreSQL container
make lint          # Run linter
make fmt           # Format code
```

### Project Structure

```
DigiOrder/
├── cmd/
│   └── main.go              # Application entry point
├── internal/
│   ├── db/
│   │   ├── connection.go    # Database connection logic
│   │   ├── db.go           # SQLC generated base code
│   │   ├── models.go       # SQLC generated models
│   │   ├── products.sql.go # SQLC generated product queries
│   │   └── query/          # SQL query definitions
│   │       ├── products.sql
│   │       ├── users.sql
│   │       ├── orders.sql
│   │       └── categories.sql
│   └── handlers/
│       ├── products.go     # Product handlers
│       └── response.go     # Response utilities
├── migrations/             # Database migrations
├── docker-compose.yml      # Docker configuration
├── sqlc.yaml              # SQLC configuration
├── Makefile               # Build automation
├── go.mod                 # Go module definition
└── README.md              # This file
```

## Testing

```bash
# Run all tests
make test

# Run tests with coverage
go test -v -cover ./...
```

## Database Schema

The system includes the following main tables:

- **users** - System users (pharmacists, clerks, admins)
- **products** - Medicine and product catalog
- **orders** - Order management
- **order_items** - Order line items
- **categories** - Product categories
- **dosage_forms** - Medicine forms (tablets, syrup, etc.)
- **roles** - User roles

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
