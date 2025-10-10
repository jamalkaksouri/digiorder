-- Roles
CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

-- Users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username TEXT UNIQUE NOT NULL,
    full_name TEXT,
    password_hash TEXT NOT NULL,
    role_id INT REFERENCES roles(id),
    created_at DATETIME DEFAULT GETDATE()
);

-- Dosage Forms
CREATE TABLE dosage_forms (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

-- Categories
CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

-- Products
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    brand TEXT,
    dosage_form_id INT NOT NULL REFERENCES dosage_forms(id),
    strength TEXT,
    unit TEXT,
    category_id INT NOT NULL REFERENCES categories(id),
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- Product Barcodes
CREATE TABLE product_barcodes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    barcode TEXT NOT NULL,
    barcode_type TEXT
);

-- Orders
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_by UUID REFERENCES users(id),
    status TEXT NOT NULL DEFAULT 'draft',
    created_at TIMESTAMPTZ DEFAULT now(),
    submitted_at TIMESTAMPTZ,
    notes TEXT
);

-- Order Items
CREATE TABLE order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id),
    requested_qty INT NOT NULL,
    unit TEXT,
    note TEXT
);
