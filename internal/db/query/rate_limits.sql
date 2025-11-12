-- internal/db/query/rate_limits.sql

-- name: GetOrCreateRateLimit :one
INSERT INTO api_rate_limits (client_id, endpoint, requests_count, window_start)
VALUES ($1, $2, 1, $3)
ON CONFLICT (client_id, endpoint, window_start) 
DO UPDATE SET 
    requests_count = api_rate_limits.requests_count + 1,
    created_at = NOW()
RETURNING *;

-- name: GetRateLimitByWindow :one
SELECT * FROM api_rate_limits
WHERE client_id = $1 
  AND endpoint = $2 
  AND window_start = $3
LIMIT 1;

-- name: DeleteOldRateLimits :exec
DELETE FROM api_rate_limits
WHERE window_start < $1;

-- name: CountLoginAttempts :one
SELECT COALESCE(SUM(requests_count), 0)::bigint as count
FROM api_rate_limits
WHERE client_id = $1 
  AND endpoint = '/api/v1/auth/login'
  AND window_start >= $2;

-- name: RecordLoginAttempt :one
INSERT INTO api_rate_limits (client_id, endpoint, requests_count, window_start)
VALUES ($1, '/api/v1/auth/login', 1, NOW())
ON CONFLICT (client_id, endpoint, window_start) 
DO UPDATE SET requests_count = api_rate_limits.requests_count + 1
RETURNING *;

-- name: GetRateLimitStats :many
SELECT 
    client_id,
    endpoint,
    SUM(requests_count) as total_requests,
    MAX(window_start) as last_request
FROM api_rate_limits
WHERE window_start >= NOW() - INTERVAL '1 hour'
GROUP BY client_id, endpoint
ORDER BY total_requests DESC
LIMIT $1;

-- name: GetTopRateLimitedIPs :many
SELECT 
    client_id,
    COUNT(*) as violation_count,
    MAX(window_start) as last_violation
FROM api_rate_limits
WHERE requests_count > $1
  AND window_start >= NOW() - INTERVAL '24 hours'
GROUP BY client_id
ORDER BY violation_count DESC
LIMIT $2;