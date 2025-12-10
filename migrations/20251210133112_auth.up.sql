-- Auth Domain Database Schema
-- This migration creates all tables required for the authentication system

-- =============================================================================
-- USERS - Core user accounts
-- =============================================================================
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255),
    name VARCHAR(255),
    avatar_url VARCHAR(500),

    -- Email verification
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    email_verified_at TIMESTAMPTZ,

    -- Phone verification
    phone_number VARCHAR(20),
    phone_verified BOOLEAN NOT NULL DEFAULT FALSE,
    phone_verified_at TIMESTAMPTZ,

    -- Account status
    status VARCHAR(50) NOT NULL DEFAULT 'pending_verification',
    failed_login_attempts INT NOT NULL DEFAULT 0,
    locked_until TIMESTAMPTZ,

    -- Login tracking
    last_login_at TIMESTAMPTZ,
    last_login_ip VARCHAR(45),

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT users_email_unique UNIQUE (email),
    CONSTRAINT users_phone_unique UNIQUE (phone_number),
    CONSTRAINT users_status_check CHECK (status IN ('active', 'pending_verification', 'disabled', 'deleted'))
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status) WHERE status != 'deleted';
CREATE INDEX idx_users_phone ON users(phone_number) WHERE phone_number IS NOT NULL;

-- =============================================================================
-- REFRESH_TOKENS - JWT refresh token storage
-- =============================================================================
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL,

    -- Device/session info
    device_info VARCHAR(255),
    ip_address VARCHAR(45),
    user_agent TEXT,

    -- Token lifecycle
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT refresh_tokens_hash_unique UNIQUE (token_hash)
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
CREATE INDEX idx_refresh_tokens_expires ON refresh_tokens(expires_at) WHERE revoked_at IS NULL;

-- =============================================================================
-- EMAIL_VERIFICATIONS - Email verification tokens
-- =============================================================================
CREATE TABLE IF NOT EXISTS email_verifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    token_hash VARCHAR(64) NOT NULL,

    expires_at TIMESTAMPTZ NOT NULL,
    verified_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT email_verifications_hash_unique UNIQUE (token_hash)
);

CREATE INDEX idx_email_verifications_user_id ON email_verifications(user_id);
CREATE INDEX idx_email_verifications_token ON email_verifications(token_hash) WHERE verified_at IS NULL;

-- =============================================================================
-- PASSWORD_RESETS - Password reset tokens
-- =============================================================================
CREATE TABLE IF NOT EXISTS password_resets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL,

    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT password_resets_hash_unique UNIQUE (token_hash)
);

CREATE INDEX idx_password_resets_user_id ON password_resets(user_id);
CREATE INDEX idx_password_resets_token ON password_resets(token_hash) WHERE used_at IS NULL;

-- =============================================================================
-- MFA_TOTP - TOTP authenticator app secrets
-- =============================================================================
CREATE TABLE IF NOT EXISTS mfa_totp (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    secret_encrypted VARCHAR(255) NOT NULL,

    verified_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT mfa_totp_user_unique UNIQUE (user_id)
);

-- =============================================================================
-- MFA_SMS - SMS-based MFA configuration
-- =============================================================================
CREATE TABLE IF NOT EXISTS mfa_sms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    phone_number VARCHAR(20) NOT NULL,

    verified_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT mfa_sms_user_unique UNIQUE (user_id)
);

-- =============================================================================
-- RECOVERY_CODES - MFA backup recovery codes
-- =============================================================================
CREATE TABLE IF NOT EXISTS recovery_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash VARCHAR(64) NOT NULL,

    used_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_recovery_codes_user_id ON recovery_codes(user_id);
CREATE INDEX idx_recovery_codes_hash ON recovery_codes(code_hash) WHERE used_at IS NULL;

-- =============================================================================
-- OAUTH_CONNECTIONS - Social login provider links
-- =============================================================================
CREATE TABLE IF NOT EXISTS oauth_connections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    provider VARCHAR(50) NOT NULL,
    provider_user_id VARCHAR(255) NOT NULL,
    provider_email VARCHAR(255),
    provider_name VARCHAR(255),
    provider_avatar_url VARCHAR(500),

    access_token_encrypted TEXT,
    refresh_token_encrypted TEXT,
    token_expires_at TIMESTAMPTZ,

    connected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT oauth_connections_provider_user UNIQUE (provider, provider_user_id),
    CONSTRAINT oauth_connections_user_provider UNIQUE (user_id, provider),
    CONSTRAINT oauth_connections_provider_check CHECK (provider IN ('google', 'apple', 'facebook', 'github'))
);

CREATE INDEX idx_oauth_connections_user_id ON oauth_connections(user_id);
CREATE INDEX idx_oauth_connections_provider ON oauth_connections(provider, provider_user_id);

-- =============================================================================
-- PASSKEYS - WebAuthn/FIDO2 credentials
-- =============================================================================
CREATE TABLE IF NOT EXISTS passkeys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    credential_id BYTEA NOT NULL,
    public_key BYTEA NOT NULL,
    name VARCHAR(100) NOT NULL,

    -- WebAuthn metadata
    aaguid BYTEA,
    sign_count BIGINT NOT NULL DEFAULT 0,

    -- Credential properties
    transports TEXT[],
    device_type VARCHAR(20),
    backed_up BOOLEAN NOT NULL DEFAULT FALSE,

    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT passkeys_credential_unique UNIQUE (credential_id)
);

CREATE INDEX idx_passkeys_user_id ON passkeys(user_id);
CREATE INDEX idx_passkeys_credential ON passkeys(credential_id);

-- =============================================================================
-- MAGIC_LINK_SESSIONS - Passwordless login sessions
-- =============================================================================
CREATE TABLE IF NOT EXISTS magic_link_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    identifier VARCHAR(255) NOT NULL,
    identifier_type VARCHAR(10) NOT NULL,
    token_hash VARCHAR(64) NOT NULL,
    otp_hash VARCHAR(64),

    user_id UUID REFERENCES users(id) ON DELETE CASCADE,

    expires_at TIMESTAMPTZ NOT NULL,
    verified_at TIMESTAMPTZ,

    ip_address VARCHAR(45),
    user_agent TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT magic_link_sessions_token_unique UNIQUE (token_hash),
    CONSTRAINT magic_link_sessions_type_check CHECK (identifier_type IN ('email', 'phone'))
);

CREATE INDEX idx_magic_link_sessions_token ON magic_link_sessions(token_hash) WHERE verified_at IS NULL;
CREATE INDEX idx_magic_link_sessions_identifier ON magic_link_sessions(identifier, identifier_type) WHERE verified_at IS NULL;

-- =============================================================================
-- SECURITY_EVENTS - Audit log for security-related events
-- =============================================================================
CREATE TABLE IF NOT EXISTS security_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,

    event_type VARCHAR(50) NOT NULL,
    success BOOLEAN NOT NULL DEFAULT TRUE,
    risk_level VARCHAR(10) DEFAULT 'low',

    -- Request context
    ip_address VARCHAR(45),
    user_agent TEXT,

    -- Location (optional, from IP geolocation)
    location_city VARCHAR(100),
    location_region VARCHAR(100),
    location_country VARCHAR(100),
    location_country_code VARCHAR(2),

    -- Device info (parsed from user agent)
    device_type VARCHAR(20),
    device_os VARCHAR(50),
    device_browser VARCHAR(50),

    -- Event-specific details (JSONB for flexibility)
    details JSONB,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT security_events_risk_check CHECK (risk_level IN ('low', 'medium', 'high'))
);

CREATE INDEX idx_security_events_user_id ON security_events(user_id);
CREATE INDEX idx_security_events_type ON security_events(event_type);
CREATE INDEX idx_security_events_created ON security_events(created_at DESC);
CREATE INDEX idx_security_events_user_created ON security_events(user_id, created_at DESC) WHERE user_id IS NOT NULL;

-- =============================================================================
-- MFA_CHALLENGES - Temporary MFA challenge tokens during login
-- =============================================================================
CREATE TABLE IF NOT EXISTS mfa_challenges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    challenge_token_hash VARCHAR(64) NOT NULL,
    available_methods TEXT[] NOT NULL,

    expires_at TIMESTAMPTZ NOT NULL,
    completed_at TIMESTAMPTZ,

    ip_address VARCHAR(45),
    user_agent TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT mfa_challenges_token_unique UNIQUE (challenge_token_hash)
);

CREATE INDEX idx_mfa_challenges_token ON mfa_challenges(challenge_token_hash) WHERE completed_at IS NULL;

-- =============================================================================
-- ACCOUNT_RECOVERY_SESSIONS - Account recovery flow state
-- =============================================================================
CREATE TABLE IF NOT EXISTS account_recovery_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,

    identifier VARCHAR(255) NOT NULL,
    method VARCHAR(10) NOT NULL,

    otp_hash VARCHAR(64),
    recovery_token_hash VARCHAR(64),

    otp_verified_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ NOT NULL,
    completed_at TIMESTAMPTZ,

    ip_address VARCHAR(45),
    user_agent TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT recovery_sessions_method_check CHECK (method IN ('email', 'sms'))
);

CREATE INDEX idx_recovery_sessions_identifier ON account_recovery_sessions(identifier, method) WHERE completed_at IS NULL;
CREATE INDEX idx_recovery_sessions_recovery_token ON account_recovery_sessions(recovery_token_hash) WHERE completed_at IS NULL;

-- =============================================================================
-- EMAIL_CHANGE_REQUESTS - Email change verification
-- =============================================================================
CREATE TABLE IF NOT EXISTS email_change_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    new_email VARCHAR(255) NOT NULL,
    token_hash VARCHAR(64) NOT NULL,

    expires_at TIMESTAMPTZ NOT NULL,
    confirmed_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT email_change_token_unique UNIQUE (token_hash)
);

CREATE INDEX idx_email_change_user ON email_change_requests(user_id) WHERE confirmed_at IS NULL;
CREATE INDEX idx_email_change_token ON email_change_requests(token_hash) WHERE confirmed_at IS NULL;

-- =============================================================================
-- PHONE_CHANGE_REQUESTS - Phone change verification
-- =============================================================================
CREATE TABLE IF NOT EXISTS phone_change_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    new_phone_number VARCHAR(20) NOT NULL,
    otp_hash VARCHAR(64) NOT NULL,

    expires_at TIMESTAMPTZ NOT NULL,
    confirmed_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_phone_change_user ON phone_change_requests(user_id) WHERE confirmed_at IS NULL;

-- =============================================================================
-- ACCOUNT_DELETION_REQUESTS - Account deletion confirmation
-- =============================================================================
CREATE TABLE IF NOT EXISTS account_deletion_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    token_hash VARCHAR(64) NOT NULL,
    feedback TEXT,

    expires_at TIMESTAMPTZ NOT NULL,
    confirmed_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT deletion_request_token_unique UNIQUE (token_hash)
);

CREATE INDEX idx_deletion_request_user ON account_deletion_requests(user_id) WHERE confirmed_at IS NULL;
CREATE INDEX idx_deletion_request_token ON account_deletion_requests(token_hash) WHERE confirmed_at IS NULL;
