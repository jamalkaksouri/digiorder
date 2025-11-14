-- internal/db/query/login_attempts.sql

-- name: LogLoginAttempt :one
INSERT INTO login_attempts_log (
    username, ip_address, user_agent, success, failure_reason, 
    rate_limited, session_id, device_info
)
VALUES (
    sqlc.arg('username'), 
    sqlc.arg('ip_address'), 
    sqlc.arg('user_agent'), 
    sqlc.arg('success'), 
    sqlc.arg('failure_reason'),
    sqlc.arg('rate_limited'), 
    sqlc.arg('session_id'), 
    sqlc.arg('device_info')
)
RETURNING *;

-- name: LogRateLimitRelease :one
INSERT INTO rate_limit_releases (
    client_id, ip_address, username, blocked_at, released_by, 
    released_by_user_id, block_duration, attempts_count, release_reason
)
VALUES (
    sqlc.arg('client_id'), 
    sqlc.arg('ip_address'), 
    sqlc.arg('username'), 
    sqlc.arg('blocked_at'),
    sqlc.arg('released_by'), 
    sqlc.arg('released_by_user_id'), 
    sqlc.arg('block_duration'), 
    sqlc.arg('attempts_count'),
    sqlc.arg('release_reason')
)
RETURNING *;

-- name: UpdateLoginAttemptRelease :exec
UPDATE login_attempts_log
SET 
    rate_limit_released_at = NOW(),
    released_by = sqlc.arg('released_by')
WHERE ip_address = sqlc.arg('ip_address')
  AND rate_limited = true
  AND rate_limit_released_at IS NULL;

-- name: GetRecentLoginAttempts :many
SELECT * FROM login_attempts_log
WHERE ip_address = sqlc.arg('ip_address')
  AND attempt_time >= sqlc.arg('since')
ORDER BY attempt_time DESC
LIMIT sqlc.arg('limit');

-- name: GetLoginAttemptsByUsername :many
SELECT * FROM login_attempts_log
WHERE username = sqlc.arg('username')
  AND attempt_time >= sqlc.arg('since')
ORDER BY attempt_time DESC
LIMIT sqlc.arg('limit');

-- name: GetRateLimitedAttempts :many
SELECT * FROM login_attempts_log
WHERE rate_limited = true
  AND attempt_time >= NOW() - INTERVAL '24 hours'
ORDER BY attempt_time DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetCurrentlyBlockedIPs :many
SELECT * FROM currently_blocked_ips;

-- name: GetLoginAttemptStats :many
SELECT * FROM login_attempt_stats;

-- name: CountFailedAttempts :one
SELECT COUNT(*) FROM login_attempts_log
WHERE ip_address = sqlc.arg('ip_address')
  AND attempt_time >= sqlc.arg('since')
  AND success = false;

-- name: GetRateLimitReleases :many
SELECT * FROM rate_limit_releases
WHERE ip_address = sqlc.arg('ip_address')
ORDER BY released_at DESC
LIMIT sqlc.arg('limit');

-- name: ArchiveOldRateLimits :exec
SELECT archive_old_rate_limits();

-- name: CleanupOldLoginAttempts :exec
SELECT cleanup_old_login_attempts();

-- name: GetRateLimitWithExclusion :many
SELECT * FROM api_rate_limits
WHERE client_id = sqlc.arg('client_id')
  AND endpoint = sqlc.arg('endpoint')
  AND window_start >= sqlc.arg('window_start')
  AND (exclude_from_tracking = false OR exclude_from_tracking IS NULL);

-- name: DeleteOldRateLimitsExcludingHealthMetrics :exec
DELETE FROM api_rate_limits
WHERE window_start < sqlc.arg('cutoff')
  AND endpoint NOT IN ('/health', '/metrics', '/api/health', '/api/metrics');

-- name: GetLoginSecurityReport :many
SELECT 
    ip_address,
    COUNT(*) as total_attempts,
    COUNT(*) FILTER (WHERE success = false) as failed_attempts,
    COUNT(*) FILTER (WHERE rate_limited = true) as rate_limited_count,
    COUNT(DISTINCT username) as unique_usernames_tried,
    MIN(attempt_time) as first_attempt,
    MAX(attempt_time) as last_attempt,
    ARRAY_AGG(DISTINCT username) as attempted_usernames
FROM login_attempts_log
WHERE attempt_time >= NOW() - INTERVAL '24 hours'
GROUP BY ip_address
HAVING COUNT(*) FILTER (WHERE success = false) >= 3
ORDER BY failed_attempts DESC
LIMIT sqlc.arg('limit');

-- name: GetUserLoginHistory :many
SELECT 
    attempt_time,
    ip_address,
    user_agent,
    success,
    failure_reason,
    rate_limited
FROM login_attempts_log
WHERE username = sqlc.arg('username')
ORDER BY attempt_time DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: ManuallyReleaseRateLimit :exec
DELETE FROM api_rate_limits
WHERE client_id = sqlc.arg('client_id')
  AND endpoint = '/api/v1/auth/login';