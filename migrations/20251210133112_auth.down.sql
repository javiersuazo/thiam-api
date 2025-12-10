-- Auth Domain Database Schema - Rollback
-- Drop tables in reverse dependency order

DROP TABLE IF EXISTS account_deletion_requests;
DROP TABLE IF EXISTS phone_change_requests;
DROP TABLE IF EXISTS email_change_requests;
DROP TABLE IF EXISTS account_recovery_sessions;
DROP TABLE IF EXISTS mfa_challenges;
DROP TABLE IF EXISTS security_events;
DROP TABLE IF EXISTS magic_link_sessions;
DROP TABLE IF EXISTS passkeys;
DROP TABLE IF EXISTS oauth_connections;
DROP TABLE IF EXISTS recovery_codes;
DROP TABLE IF EXISTS mfa_sms;
DROP TABLE IF EXISTS mfa_totp;
DROP TABLE IF EXISTS password_resets;
DROP TABLE IF EXISTS email_verifications;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS users;
