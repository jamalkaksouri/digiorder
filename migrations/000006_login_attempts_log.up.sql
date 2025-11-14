-- migrations/000006_login_attempts_log.up.sql
-- Comprehensive login attempts tracking with detailed user information

-- Login attempts log table
CREATE TABLE IF NOT EXISTS login_attempts_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username TEXT NOT NULL,
    ip_address TEXT NOT NULL,
    user_agent TEXT,
    attempt_time TIMESTAMPTZ DEFAULT NOW(),
    success BOOLEAN NOT NULL DEFAULT FALSE,
    failure_reason TEXT, -- 'invalid_credentials', 'rate_limited', 'account_locked', etc.
    rate_limited BOOLEAN DEFAULT FALSE,
    rate_limit_released_at TIMESTAMPTZ, -- When user was released from rate limit
    released_by TEXT, -- 'automatic', 'manual_admin', etc.
    session_id TEXT, -- Track session attempts
    country TEXT, -- Can be populated from IP geolocation
    city TEXT,
    device_info JSONB, -- Store parsed user agent details
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for fast queries
CREATE INDEX IF NOT EXISTS idx_login_attempts_username ON login_attempts_log(username);
CREATE INDEX IF NOT EXISTS idx_login_attempts_ip ON login_attempts_log(ip_address);
CREATE INDEX IF NOT EXISTS idx_login_attempts_time ON login_attempts_log(attempt_time DESC);
CREATE INDEX IF NOT EXISTS idx_login_attempts_rate_limited ON login_attempts_log(rate_limited) WHERE rate_limited = true;
CREATE INDEX IF NOT EXISTS idx_login_attempts_success ON login_attempts_log(success);
CREATE INDEX IF NOT EXISTS idx_login_attempts_cleanup ON login_attempts_log(created_at) WHERE created_at < NOW() - INTERVAL '90 days';

-- Rate limit release log table (separate tracking for releases)
CREATE TABLE IF NOT EXISTS rate_limit_releases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id TEXT NOT NULL,
    ip_address TEXT NOT NULL,
    username TEXT,
    blocked_at TIMESTAMPTZ NOT NULL,
    released_at TIMESTAMPTZ DEFAULT NOW(),
    released_by TEXT NOT NULL, -- 'automatic_expiry', 'admin_manual', 'system_reset'
    released_by_user_id UUID REFERENCES users(id), -- If manually released by admin
    block_duration INTERVAL,
    attempts_count INTEGER,
    release_reason TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_rate_limit_releases_client ON rate_limit_releases(client_id);
CREATE INDEX IF NOT EXISTS idx_rate_limit_releases_ip ON rate_limit_releases(ip_address);
CREATE INDEX IF NOT EXISTS idx_rate_limit_releases_time ON rate_limit_releases(released_at DESC);

-- Archive table for old rate limits (for historical analysis)
CREATE TABLE IF NOT EXISTS api_rate_limits_archive (
    id UUID,
    client_id TEXT NOT NULL,
    endpoint TEXT NOT NULL,
    requests_count INT NOT NULL,
    window_start TIMESTAMPTZ NOT NULL,
    archived_at TIMESTAMPTZ DEFAULT NOW(),
    original_created_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_rate_limits_archive_client ON api_rate_limits_archive(client_id);
CREATE INDEX IF NOT EXISTS idx_rate_limits_archive_archived ON api_rate_limits_archive(archived_at);

-- Function to archive old rate limits
CREATE OR REPLACE FUNCTION archive_old_rate_limits() RETURNS void AS $$
BEGIN
    -- Move records older than 7 days to archive
    INSERT INTO api_rate_limits_archive (id, client_id, endpoint, requests_count, window_start, original_created_at)
    SELECT id, client_id, endpoint, requests_count, window_start, created_at
    FROM api_rate_limits
    WHERE window_start < NOW() - INTERVAL '7 days';
    
    -- Delete archived records from main table
    DELETE FROM api_rate_limits
    WHERE window_start < NOW() - INTERVAL '7 days';
END;
$$ LANGUAGE plpgsql;

-- Function to clean up old login attempts (keep only 90 days)
CREATE OR REPLACE FUNCTION cleanup_old_login_attempts() RETURNS void AS $$
BEGIN
    DELETE FROM login_attempts_log
    WHERE created_at < NOW() - INTERVAL '90 days';
    
    DELETE FROM rate_limit_releases
    WHERE created_at < NOW() - INTERVAL '90 days';
END;
$$ LANGUAGE plpgsql;

-- Add exclusion for health and metrics endpoints in rate limits table
ALTER TABLE api_rate_limits ADD COLUMN IF NOT EXISTS exclude_from_tracking BOOLEAN DEFAULT FALSE;

-- Mark health and metrics endpoints to exclude from detailed tracking
UPDATE api_rate_limits 
SET exclude_from_tracking = TRUE 
WHERE endpoint IN ('/health', '/metrics', '/api/health', '/api/metrics');

-- Create a view for easy access to blocked IPs
CREATE OR REPLACE VIEW currently_blocked_ips AS
SELECT 
    client_id,
    endpoint,
    SUM(requests_count) as total_attempts,
    MAX(window_start) as last_attempt,
    COUNT(*) as block_windows
FROM api_rate_limits
WHERE endpoint = '/api/v1/auth/login'
  AND window_start >= NOW() - INTERVAL '5 minutes'
GROUP BY client_id, endpoint
HAVING SUM(requests_count) >= 5;

-- Create a view for login attempt statistics
CREATE OR REPLACE VIEW login_attempt_stats AS
SELECT 
    DATE_TRUNC('hour', attempt_time) as hour,
    COUNT(*) as total_attempts,
    COUNT(*) FILTER (WHERE success = true) as successful,
    COUNT(*) FILTER (WHERE success = false) as failed,
    COUNT(*) FILTER (WHERE rate_limited = true) as rate_limited_attempts,
    COUNT(DISTINCT ip_address) as unique_ips,
    COUNT(DISTINCT username) as unique_usernames
FROM login_attempts_log
WHERE attempt_time >= NOW() - INTERVAL '24 hours'
GROUP BY DATE_TRUNC('hour', attempt_time)
ORDER BY hour DESC;

-- Add comment explaining the purpose
COMMENT ON TABLE login_attempts_log IS 'Comprehensive logging of all login attempts with detailed user information for security auditing';
COMMENT ON TABLE rate_limit_releases IS 'Tracks when users are released from rate limiting, either automatically or manually';
COMMENT ON TABLE api_rate_limits_archive IS 'Archive of old rate limit records for historical analysis';
COMMENT ON FUNCTION archive_old_rate_limits() IS 'Archives rate limit records older than 7 days and removes them from the main table';
COMMENT ON FUNCTION cleanup_old_login_attempts() IS 'Deletes login attempt logs older than 90 days to maintain database performance';