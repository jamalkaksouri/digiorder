-- migrations/000007_ip_ban_tracking.up.sql
-- Enhanced IP ban tracking with automatic cleanup

-- Table for tracking currently banned IPs
CREATE TABLE IF NOT EXISTS ip_bans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ip_address TEXT NOT NULL UNIQUE,
    banned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    banned_until TIMESTAMPTZ NOT NULL,
    reason TEXT NOT NULL,
    failed_attempts INTEGER DEFAULT 0,
    endpoint TEXT,
    banned_by TEXT DEFAULT 'system', -- 'system' or 'admin_manual'
    released_at TIMESTAMPTZ,
    released_by TEXT,
    auto_released BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for fast lookups
CREATE INDEX idx_ip_bans_ip ON ip_bans(ip_address);
CREATE INDEX idx_ip_bans_active ON ip_bans(banned_until) WHERE released_at IS NULL;
CREATE INDEX idx_ip_bans_cleanup ON ip_bans(banned_until) WHERE released_at IS NULL AND banned_until < NOW();

-- Function to automatically release expired bans
CREATE OR REPLACE FUNCTION auto_release_expired_bans() RETURNS void AS $$
BEGIN
    UPDATE ip_bans
    SET 
        released_at = NOW(),
        auto_released = TRUE,
        released_by = 'auto_expiry'
    WHERE released_at IS NULL
      AND banned_until < NOW();
END;
$$ LANGUAGE plpgsql;

-- Function to check if IP is currently banned
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

-- Function to get ban details for an IP
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
        EXTRACT(EPOCH FROM (ib.banned_until - NOW()))::INTEGER / 60 as minutes_remaining
    FROM ip_bans ib
    WHERE ib.ip_address = check_ip
      AND ib.released_at IS NULL
      AND ib.banned_until > NOW()
    ORDER BY ib.banned_at DESC
    LIMIT 1;
END;
$$ LANGUAGE plpgsql;

-- View for currently banned IPs with details
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

-- View for ban statistics
CREATE OR REPLACE VIEW ip_ban_stats AS
SELECT 
    DATE_TRUNC('hour', banned_at) as hour,
    COUNT(*) as total_bans,
    COUNT(DISTINCT ip_address) as unique_ips,
    AVG(failed_attempts) as avg_attempts,
    SUM(CASE WHEN auto_released = true THEN 1 ELSE 0 END) as auto_released_count,
    SUM(CASE WHEN banned_by = 'admin_manual' THEN 1 ELSE 0 END) as manual_bans
FROM ip_bans
WHERE banned_at >= NOW() - INTERVAL '24 hours'
GROUP BY DATE_TRUNC('hour', banned_at)
ORDER BY hour DESC;

-- Trigger to automatically cleanup old records (older than 30 days)
CREATE OR REPLACE FUNCTION cleanup_old_ip_bans() RETURNS void AS $$
BEGIN
    DELETE FROM ip_bans
    WHERE created_at < NOW() - INTERVAL '30 days';
END;
$$ LANGUAGE plpgsql;

-- Add comment explaining the purpose
COMMENT ON TABLE ip_bans IS 'Tracks temporarily banned IPs with automatic expiry and cleanup. Records are automatically removed after ban expires and retained for 30 days for auditing.';
COMMENT ON FUNCTION auto_release_expired_bans() IS 'Automatically releases IPs whose ban period has expired';
COMMENT ON FUNCTION is_ip_banned(TEXT) IS 'Quick check if an IP is currently banned';
COMMENT ON FUNCTION get_ip_ban_details(TEXT) IS 'Returns detailed ban information for an IP including time remaining';
COMMENT ON VIEW active_ip_bans IS 'Shows all currently active IP bans with time remaining';
COMMENT ON VIEW ip_ban_stats IS 'Hourly statistics of IP bans for the last 24 hours';

-- Insert initial cleanup job marker (for reference)
CREATE TABLE IF NOT EXISTS ip_ban_cleanup_log (
    id SERIAL PRIMARY KEY,
    last_cleanup TIMESTAMPTZ DEFAULT NOW(),
    records_cleaned INTEGER DEFAULT 0
);

INSERT INTO ip_ban_cleanup_log (last_cleanup, records_cleaned) 
VALUES (NOW(), 0);