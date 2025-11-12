-- migrations/000003_add_permissions.up.sql

-- Permissions table
CREATE TABLE IF NOT EXISTS permissions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(resource, action)
);

-- Role-Permission junction table
CREATE TABLE IF NOT EXISTS role_permissions (
    id SERIAL PRIMARY KEY,
    role_id INT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id INT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(role_id, permission_id)
);

-- Add indexes
CREATE INDEX idx_permissions_resource ON permissions(resource);
CREATE INDEX idx_role_permissions_role ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission ON role_permissions(permission_id);

-- Insert default permissions
INSERT INTO permissions (name, resource, action, description) VALUES
    -- Product permissions
    ('view_products', 'products', 'read', 'View products'),
    ('create_products', 'products', 'create', 'Create products'),
    ('update_products', 'products', 'update', 'Update products'),
    ('delete_products', 'products', 'delete', 'Delete products'),

    -- Order permissions
    ('view_orders', 'orders', 'read', 'View orders'),
    ('create_orders', 'orders', 'create', 'Create orders'),
    ('update_orders', 'orders', 'update', 'Update orders'),
    ('delete_orders', 'orders', 'delete', 'Delete orders'),

    -- User permissions
    ('view_users', 'users', 'read', 'View users'),
    ('create_users', 'users', 'create', 'Create users'),
    ('update_users', 'users', 'update', 'Update users'),
    ('delete_users', 'users', 'delete', 'Delete users'),

    -- Audit permissions
    ('view_audit_logs', 'audit', 'read', 'View audit logs'),

    -- System permissions
    ('manage_permissions', 'permissions', 'manage', 'Manage permissions'),
    ('manage_roles', 'roles', 'manage', 'Manage roles');

-- Assign all permissions to admin role
INSERT INTO role_permissions (role_id, permission_id)
SELECT 1, id FROM permissions;

-- Assign limited permissions to pharmacist
INSERT INTO role_permissions (role_id, permission_id)
SELECT 2, id FROM permissions
WHERE name IN ('view_products', 'create_products', 'update_products',
               'view_orders', 'create_orders', 'update_orders');

-- Assign read-only permissions to clerk
INSERT INTO role_permissions (role_id, permission_id)
SELECT 3, id FROM permissions
WHERE action = 'read';

-- Create primary admin user (protected from deletion)
INSERT INTO users (id, username, full_name, password_hash, role_id)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'admin',
    'Primary Administrator',
    '$2a$10$Zu7yVNJ0e9Fn9vwUy9vRbO5CqPQZMB8l5k8hEWnGvhkrFUKqj9iEW',
    1
) ON CONFLICT (id) DO NOTHING;