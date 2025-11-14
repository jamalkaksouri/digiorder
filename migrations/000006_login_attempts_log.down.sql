-- migrations/000006_login_attempts_log.down.sql

DROP VIEW IF EXISTS login_attempt_stats;
DROP VIEW IF EXISTS currently_blocked_ips;
DROP FUNCTION IF EXISTS cleanup_old_login_attempts();
DROP FUNCTION IF EXISTS archive_old_rate_limits();
DROP TABLE IF EXISTS api_rate_limits_archive;
DROP TABLE IF EXISTS rate_limit_releases;
DROP TABLE IF EXISTS login_attempts_log;
ALTER TABLE api_rate_limits DROP COLUMN IF EXISTS exclude_from_tracking;