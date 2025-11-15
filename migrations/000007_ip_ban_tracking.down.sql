-- migrations/000007_ip_ban_tracking.down.sql

DROP VIEW IF EXISTS ip_ban_stats;
DROP VIEW IF EXISTS active_ip_bans;
DROP FUNCTION IF EXISTS cleanup_old_ip_bans();
DROP FUNCTION IF EXISTS get_ip_ban_details(TEXT);
DROP FUNCTION IF EXISTS is_ip_banned(TEXT);
DROP FUNCTION IF EXISTS auto_release_expired_bans();
DROP TABLE IF EXISTS ip_ban_cleanup_log;
DROP TABLE IF EXISTS ip_bans;