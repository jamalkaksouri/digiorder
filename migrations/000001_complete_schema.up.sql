-- ============================================================================
-- DigiOrder Complete Database Schema - Single Migration File
-- Version: 3.0.1
-- Description: Comprehensive schema with all tables, indexes, functions, and views
-- ============================================================================


-- ============================================================================
-- CORE TABLES
-- ============================================================================

CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username TEXT UNIQUE NOT NULL,
    full_name TEXT,
    password_hash TEXT NOT NULL,
    role_id INT REFERENCES roles(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS dosage_forms (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS categories (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    brand TEXT,
    dosage_form_id INT REFERENCES dosage_forms(id),
    strength TEXT,
    unit TEXT,
    category_id INT REFERENCES categories(id),
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS product_barcodes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID REFERENCES products(id) ON DELETE CASCADE,
    barcode TEXT NOT NULL UNIQUE,
    barcode_type TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_by UUID REFERENCES users(id),
    status TEXT NOT NULL DEFAULT 'draft',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    submitted_at TIMESTAMPTZ,
    notes TEXT,
    deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID REFERENCES orders(id) ON DELETE CASCADE,
    product_id UUID REFERENCES products(id),
    requested_qty INT NOT NULL,
    unit TEXT,
    note TEXT
);


-- ============================================================================
-- SECURITY & AUDIT
-- ============================================================================

CREATE TABLE IF NOT EXISTS api_rate_limits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id TEXT NOT NULL,
    endpoint TEXT NOT NULL,
    requests_count INT NOT NULL DEFAULT 0,
    window_start TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    exclude_from_tracking BOOLEAN DEFAULT FALSE,
    UNIQUE(client_id, endpoint, window_start)
);

CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    action TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    old_values JSONB,
    new_values JSONB,
    ip_address TEXT,
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS permissions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    resource VARCHAR(100) NOT NULL,
    action VARCHAR(50) NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(resource, action)
);

CREATE TABLE IF NOT EXISTS role_permissions (
    id SERIAL PRIMARY KEY,
    role_id INT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id INT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(role_id, permission_id)
);

CREATE TABLE IF NOT EXISTS system_setup (
    id SERIAL PRIMARY KEY,
    admin_created BOOLEAN DEFAULT FALSE,
    setup_completed_at TIMESTAMPTZ,
    setup_by_ip TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS login_attempts_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username TEXT NOT NULL,
    ip_address TEXT NOT NULL,
    user_agent TEXT,
    attempt_time TIMESTAMPTZ DEFAULT NOW(),
    success BOOLEAN NOT NULL DEFAULT FALSE,
    failure_reason TEXT,
    rate_limited BOOLEAN DEFAULT FALSE,
    rate_limit_released_at TIMESTAMPTZ,
    released_by TEXT,
    session_id TEXT,
    country TEXT,
    city TEXT,
    device_info JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS rate_limit_releases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id TEXT NOT NULL,
    ip_address TEXT NOT NULL,
    username TEXT,
    blocked_at TIMESTAMPTZ NOT NULL,
    released_at TIMESTAMPTZ DEFAULT NOW(),
    released_by TEXT NOT NULL,
    released_by_user_id UUID REFERENCES users(id),
    block_duration INTERVAL,
    attempts_count INTEGER,
    release_reason TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS api_rate_limits_archive (
    id UUID,
    client_id TEXT NOT NULL,
    endpoint TEXT NOT NULL,
    requests_count INT NOT NULL,
    window_start TIMESTAMPTZ NOT NULL,
    archived_at TIMESTAMPTZ DEFAULT NOW(),
    original_created_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS ip_bans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ip_address TEXT NOT NULL UNIQUE,
    banned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    banned_until TIMESTAMPTZ NOT NULL,
    reason TEXT NOT NULL,
    failed_attempts INTEGER DEFAULT 0,
    endpoint TEXT,
    banned_by TEXT DEFAULT 'system',
    released_at TIMESTAMPTZ,
    released_by TEXT,
    auto_released BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ip_ban_cleanup_log (
    id SERIAL PRIMARY KEY,
    last_cleanup TIMESTAMPTZ DEFAULT NOW(),
    records_cleaned INTEGER DEFAULT 0
);


-- ============================================================================
-- INDEXES (fixed)
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_products_name ON products(name);
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category_id);
CREATE INDEX IF NOT EXISTS idx_products_dosage_form ON products(dosage_form_id);
CREATE INDEX IF NOT EXISTS idx_products_deleted_at ON products(deleted_at);

CREATE INDEX IF NOT EXISTS idx_product_barcodes_barcode ON product_barcodes(barcode);
CREATE INDEX IF NOT EXISTS idx_product_barcodes_product_id ON product_barcodes(product_id);

CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_created_by ON orders(created_by);
CREATE INDEX IF NOT EXISTS idx_orders_deleted_at ON orders(deleted_at);

CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_permissions_resource ON permissions(resource);
CREATE INDEX IF NOT EXISTS idx_role_permissions_role ON role_permissions(role_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_permission ON role_permissions(permission_id);

CREATE INDEX IF NOT EXISTS idx_api_rate_limits_login 
ON api_rate_limits(client_id, endpoint)
WHERE endpoint = '/api/v1/auth/login';

CREATE INDEX IF NOT EXISTS idx_api_rate_limits_cleanup 
ON api_rate_limits(window_start);

CREATE INDEX IF NOT EXISTS idx_login_attempts_username ON login_attempts_log(username);
CREATE INDEX IF NOT EXISTS idx_login_attempts_ip ON login_attempts_log(ip_address);
CREATE INDEX IF NOT EXISTS idx_login_attempts_time ON login_attempts_log(attempt_time DESC);
CREATE INDEX IF NOT EXISTS idx_login_attempts_rate_limited 
ON login_attempts_log(rate_limited)
WHERE rate_limited = true;

CREATE INDEX IF NOT EXISTS idx_rate_limit_releases_client ON rate_limit_releases(client_id);
CREATE INDEX IF NOT EXISTS idx_rate_limit_releases_ip ON rate_limit_releases(ip_address);
CREATE INDEX IF NOT EXISTS idx_rate_limit_releases_time ON rate_limit_releases(released_at DESC);

CREATE INDEX IF NOT EXISTS idx_rate_limits_archive_client ON api_rate_limits_archive(client_id);
CREATE INDEX IF NOT EXISTS idx_rate_limits_archive_archived ON api_rate_limits_archive(archived_at);

CREATE INDEX IF NOT EXISTS idx_ip_bans_ip ON ip_bans(ip_address);
CREATE INDEX IF NOT EXISTS idx_ip_bans_active 
ON ip_bans(banned_until)
WHERE released_at IS NULL;


-- ============================================================================
-- FUNCTIONS
-- ============================================================================

CREATE OR REPLACE FUNCTION has_admin_user() RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS (SELECT 1 FROM users WHERE role_id = 1 AND deleted_at IS NULL LIMIT 1);
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION archive_old_rate_limits() RETURNS void AS $$
BEGIN
    INSERT INTO api_rate_limits_archive (id, client_id, endpoint, requests_count, window_start, original_created_at)
    SELECT id, client_id, endpoint, requests_count, window_start, created_at
    FROM api_rate_limits
    WHERE window_start < NOW() - INTERVAL '7 days';

    DELETE FROM api_rate_limits
    WHERE window_start < NOW() - INTERVAL '7 days';
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION cleanup_old_login_attempts() RETURNS void AS $$
BEGIN
    DELETE FROM login_attempts_log
    WHERE created_at < NOW() - INTERVAL '90 days';

    DELETE FROM rate_limit_releases
    WHERE created_at < NOW() - INTERVAL '90 days';
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION auto_release_expired_bans() RETURNS void AS $$
BEGIN
    UPDATE ip_bans
    SET released_at = NOW(),
        auto_released = TRUE,
        released_by = 'auto_expiry'
    WHERE released_at IS NULL
      AND banned_until < NOW();
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION is_ip_banned(check_ip TEXT) RETURNS BOOLEAN AS $$
DECLARE
    ban_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO ban_count
    FROM ip_bans
    WHERE ip_address = check_ip
      AND released_at IS NULL
      AND banned_until > NOW();

    RETURN ban_count > 0;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION get_ip_ban_details(check_ip TEXT)
RETURNS TABLE (
    banned_until TIMESTAMPTZ,
    reason TEXT,
    failed_attempts INTEGER,
    minutes_remaining INTEGER
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        ib.banned_until,
        ib.reason,
        ib.failed_attempts,
        EXTRACT(EPOCH FROM (ib.banned_until - NOW()))::INTEGER / 60
    FROM ip_bans ib
    WHERE ib.ip_address = check_ip
      AND ib.released_at IS NULL
      AND ib.banned_until > NOW()
    ORDER BY ib.banned_at DESC
    LIMIT 1;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION cleanup_old_ip_bans() RETURNS void AS $$
BEGIN
    DELETE FROM ip_bans
    WHERE created_at < NOW() - INTERVAL '30 days';
END;
$$ LANGUAGE plpgsql;


-- ============================================================================
-- VIEWS
-- ============================================================================

CREATE OR REPLACE VIEW currently_blocked_ips AS
SELECT 
    client_id,
    endpoint,
    SUM(requests_count) AS total_attempts,
    MAX(window_start) AS last_attempt,
    COUNT(*) AS block_windows
FROM api_rate_limits
WHERE endpoint = '/api/v1/auth/login'
  AND window_start >= NOW() - INTERVAL '5 minutes'
GROUP BY client_id, endpoint
HAVING SUM(requests_count) >= 5;

CREATE OR REPLACE VIEW login_attempt_stats AS
SELECT 
    DATE_TRUNC('hour', attempt_time) AS hour,
    COUNT(*) AS total_attempts,
    COUNT(*) FILTER (WHERE success = true) AS successful,
    COUNT(*) FILTER (WHERE success = false) AS failed,
    COUNT(*) FILTER (WHERE rate_limited = true) AS rate_limited_attempts,
    COUNT(DISTINCT ip_address) AS unique_ips,
    COUNT(DISTINCT username) AS unique_usernames
FROM login_attempts_log
WHERE attempt_time >= NOW() - INTERVAL '24 hours'
GROUP BY DATE_TRUNC('hour', attempt_time)
ORDER BY hour DESC;

CREATE OR REPLACE VIEW active_ip_bans AS
SELECT 
    ip_address,
    banned_at,
    banned_until,
    reason,
    failed_attempts,
    endpoint,
    EXTRACT(EPOCH FROM (banned_until - NOW()))::INTEGER AS seconds_remaining,
    EXTRACT(EPOCH FROM (banned_until - NOW()))::INTEGER / 60 AS minutes_remaining
FROM ip_bans
WHERE released_at IS NULL
  AND banned_until > NOW()
ORDER BY banned_until DESC;

CREATE OR REPLACE VIEW ip_ban_stats AS
SELECT 
    DATE_TRUNC('hour', banned_at) AS hour,
    COUNT(*) AS total_bans,
    COUNT(DISTINCT ip_address) AS unique_ips,
    AVG(failed_attempts) AS avg_attempts,
    SUM(CASE WHEN auto_released = true THEN 1 ELSE 0 END) AS auto_released_count,
    SUM(CASE WHEN banned_by = 'admin_manual' THEN 1 ELSE 0 END) AS manual_bans
FROM ip_bans
WHERE banned_at >= NOW() - INTERVAL '24 hours'
GROUP BY DATE_TRUNC('hour', banned_at)
ORDER BY hour DESC;


-- ============================================================================
-- DEFAULT DATA
-- ============================================================================

INSERT INTO roles (name) VALUES
    ('admin'),
    ('pharmacist'),
    ('clerk')
ON CONFLICT (name) DO NOTHING;

INSERT INTO categories (name) VALUES
    ('دارویی'),
    ('آرایشی'),
    ('بهداشتی'),
    ('مکمل')
ON CONFLICT (name) DO NOTHING;

INSERT INTO dosage_forms (name) VALUES
    ('قرص'),
    ('کپسول'),
    ('شربت'),
    ('آمپول'),
    ('قطره'),
    ('پماد'),
    ('ژل'),
    ('اسپری')
ON CONFLICT (name) DO NOTHING;

INSERT INTO permissions (name, resource, action, description) VALUES
    ('view_products', 'products', 'read', 'View products'),
    ('create_products', 'products', 'create', 'Create products'),
    ('update_products', 'products', 'update', 'Update products'),
    ('delete_products', 'products', 'delete', 'Delete products'),
    ('view_orders', 'orders', 'read', 'View orders'),
    ('create_orders', 'orders', 'create', 'Create orders'),
    ('update_orders', 'orders', 'update', 'Update orders'),
    ('delete_orders', 'orders', 'delete', 'Delete orders'),
    ('view_users', 'users', 'read', 'View users'),
    ('create_users', 'users', 'create', 'Create users'),
    ('update_users', 'users', 'update', 'Update users'),
    ('delete_users', 'users', 'delete', 'Delete users'),
    ('view_audit_logs', 'audit', 'read', 'View audit logs'),
    ('manage_permissions', 'permissions', 'manage', 'Manage permissions'),
    ('manage_roles', 'roles', 'manage', 'Manage roles')
ON CONFLICT (resource, action) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT 1, id FROM permissions
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT 2, id FROM permissions
WHERE name IN (
    'view_products', 'create_products', 'update_products',
    'view_orders', 'create_orders', 'update_orders'
)
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT 3, id FROM permissions
WHERE action = 'read'
ON CONFLICT DO NOTHING;

INSERT INTO system_setup (admin_created, setup_completed_at, setup_by_ip)
VALUES (FALSE, NULL, NULL)
ON CONFLICT DO NOTHING;

INSERT INTO ip_ban_cleanup_log (last_cleanup, records_cleaned)
VALUES (NOW(), 0)
ON CONFLICT DO NOTHING;


-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON TABLE system_setup IS 'Tracks system initialization.';
COMMENT ON TABLE login_attempts_log IS 'Logs all login attempts.';
COMMENT ON TABLE rate_limit_releases IS 'Tracks rate limit releases.';
COMMENT ON TABLE api_rate_limits_archive IS 'Archive of old rate limits.';
COMMENT ON TABLE ip_bans IS 'Tracks temporarily banned IPs with auto-expiry.';

COMMENT ON FUNCTION archive_old_rate_limits() IS 'Archives 7-day-old rate limit records.';
COMMENT ON FUNCTION cleanup_old_login_attempts() IS 'Deletes logs older than 90 days.';
COMMENT ON FUNCTION auto_release_expired_bans() IS 'Auto-releases expired bans.';
COMMENT ON FUNCTION is_ip_banned(TEXT) IS 'Checks if IP is currently banned.';
COMMENT ON FUNCTION get_ip_ban_details(TEXT) IS 'Gets detailed ban info for an IP.';

COMMENT ON VIEW currently_blocked_ips IS 'Shows login-blocked IPs.';
COMMENT ON VIEW login_attempt_stats IS 'Hourly login attempt statistics.';
COMMENT ON VIEW active_ip_bans IS 'Shows active IP bans.';
COMMENT ON VIEW ip_ban_stats IS 'Hourly IP ban statistics.';
