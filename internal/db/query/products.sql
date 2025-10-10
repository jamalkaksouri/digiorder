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
    created_at TIMESTAMPTZ DEFAULT NOW()
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
    dosage_form_id INT REFERENCES dosage_forms(id),
    strength TEXT,
    unit TEXT,
    category_id INT REFERENCES categories(id),
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Product Barcodes
CREATE TABLE product_barcodes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID REFERENCES products(id) ON DELETE CASCADE,
    barcode TEXT NOT NULL UNIQUE,
    barcode_type TEXT
);

-- Orders
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_by UUID REFERENCES users(id),
    status TEXT NOT NULL DEFAULT 'draft',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    submitted_at TIMESTAMPTZ,
    notes TEXT
);

-- Order Items
CREATE TABLE order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID REFERENCES orders(id) ON DELETE CASCADE,
    product_id UUID REFERENCES products(id),
    requested_qty INT NOT NULL,
    unit TEXT,
    note TEXT
);

-- Create indexes for better performance
CREATE INDEX idx_products_name ON products(name);
CREATE INDEX idx_products_category ON products(category_id);
CREATE INDEX idx_products_dosage_form ON products(dosage_form_id);
CREATE INDEX idx_product_barcodes_barcode ON product_barcodes(barcode);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created_by ON orders(created_by);

-- Insert default data
INSERT INTO roles (name) VALUES 
    ('admin'),
    ('pharmacist'),
    ('clerk');

INSERT INTO categories (name) VALUES 
    ('دارویی'),
    ('آرایشی'),
    ('بهداشتی'),
    ('مکمل');

INSERT INTO dosage_forms (name) VALUES 
    ('قرص'),
    ('کپسول'),
    ('شربت'),
    ('آمپول'),
    ('قطره'),
    ('پماد'),
    ('ژل'),
    ('اسپری');