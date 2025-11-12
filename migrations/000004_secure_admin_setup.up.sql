-- migrations/000004_secure_admin_setup.up.sql
-- This migration removes hardcoded admin password and requires manual setup

-- Remove any existing test admin users with hardcoded passwords
DELETE FROM users WHERE username = 'admin' AND id = '00000000-0000-0000-0000-000000000001';

-- Add a table to track initial setup status
CREATE TABLE IF NOT EXISTS system_setup (
    id SERIAL PRIMARY KEY,
    admin_created BOOLEAN DEFAULT FALSE,
    setup_completed_at TIMESTAMPTZ,
    setup_by_ip TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Insert initial setup record
INSERT INTO system_setup (admin_created, setup_completed_at, setup_by_ip)
VALUES (FALSE, NULL, NULL);

-- Add index on login attempts for faster queries
CREATE INDEX IF NOT EXISTS idx_api_rate_limits_login 
ON api_rate_limits(client_id, endpoint) 
WHERE endpoint = '/api/v1/auth/login';

-- Add index for cleanup queries
CREATE INDEX IF NOT EXISTS idx_api_rate_limits_cleanup 
ON api_rate_limits(window_start);

-- Create function to check if initial admin exists
CREATE OR REPLACE FUNCTION has_admin_user() RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS (SELECT 1 FROM users WHERE role_id = 1 AND deleted_at IS NULL LIMIT 1);
END;
$$ LANGUAGE plpgsql;

-- Add comment explaining security improvement
COMMENT ON TABLE system_setup IS 'Tracks system initialization. Admin user must be created via secure setup endpoint with strong password.';