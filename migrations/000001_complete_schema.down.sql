-- ============================================================================
-- DigiOrder Complete Database Schema - Rollback Migration
-- Version: 3.0.1
-- Description: Drops all tables, indexes, functions, and views in reverse order
-- ============================================================================

-- Drop views
DROP VIEW IF EXISTS ip_ban_stats CASCADE;
DROP VIEW IF EXISTS active_ip_bans CASCADE;
DROP VIEW IF EXISTS login_attempt_stats CASCADE;
DROP VIEW IF EXISTS currently_blocked_ips CASCADE;

-- Drop functions
DROP FUNCTION IF EXISTS cleanup_old_ip_bans() CASCADE;
DROP FUNCTION IF EXISTS get_ip_ban_details(TEXT) CASCADE;
DROP FUNCTION IF EXISTS is_ip_banned(TEXT) CASCADE;
DROP FUNCTION IF EXISTS auto_release_expired_bans() CASCADE;
DROP FUNCTION IF EXISTS cleanup_old_login_attempts() CASCADE;
DROP FUNCTION IF EXISTS archive_old_rate_limits() CASCADE;
DROP FUNCTION IF EXISTS has_admin_user() CASCADE;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS ip_ban_cleanup_log CASCADE;
DROP TABLE IF EXISTS ip_bans CASCADE;
DROP TABLE IF EXISTS api_rate_limits_archive CASCADE;
DROP TABLE IF EXISTS rate_limit_releases CASCADE;
DROP TABLE IF EXISTS login_attempts_log CASCADE;
DROP TABLE IF EXISTS system_setup CASCADE;
DROP TABLE IF EXISTS role_permissions CASCADE;
DROP TABLE IF EXISTS permissions CASCADE;
DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS api_rate_limits CASCADE;
DROP TABLE IF EXISTS order_items CASCADE;
DROP TABLE IF EXISTS orders CASCADE;
DROP TABLE IF EXISTS product_barcodes CASCADE;
DROP TABLE IF EXISTS products CASCADE;
DROP TABLE IF EXISTS categories CASCADE;
DROP TABLE IF EXISTS dosage_forms CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS roles CASCADE;