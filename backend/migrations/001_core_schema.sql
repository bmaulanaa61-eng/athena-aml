-- =============================================================================
-- ATHENA AML — Core Database Schema Migration v1.2
-- Target: PostgreSQL 14+ | 15 Core Tables with UUID PK & JSONB
-- =============================================================================

-- Enable pgcrypto for UUID generation if needed
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Drop existing tables in reverse dependency order
DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS webhook_deliveries CASCADE;
DROP TABLE IF EXISTS webhook_configs CASCADE;
DROP TABLE IF EXISTS jobs CASCADE;
DROP TABLE IF EXISTS investigation_resolutions CASCADE;
DROP TABLE IF EXISTS investigation_notes CASCADE;
DROP TABLE IF EXISTS alerts CASCADE;
DROP TABLE IF EXISTS risk_assessments CASCADE;
DROP TABLE IF EXISTS aml_rules CASCADE;
DROP TABLE IF EXISTS idempotency_keys CASCADE;
DROP TABLE IF EXISTS transactions CASCADE;
DROP TABLE IF EXISTS sessions CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS api_keys CASCADE;
DROP TABLE IF EXISTS tenants CASCADE;

-- -----------------------------------------------------------------------------
-- 1. TENANTS (Multi-Tenant Isolation)
-- -----------------------------------------------------------------------------
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE', -- 'ACTIVE', 'SUSPENDED', 'INACTIVE'
    plan VARCHAR(50) NOT NULL DEFAULT 'GROWTH',   -- 'STARTER', 'GROWTH', 'ENTERPRISE'
    settings JSONB NOT NULL DEFAULT '{
        "alert_policy": {
            "medium_threshold": 30,
            "high_threshold": 70,
            "create_alert_from": "MEDIUM"
        }
    }'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tenants_status ON tenants(status);

-- -----------------------------------------------------------------------------
-- 2. API KEYS (Machine-to-Machine Ingestion Auth)
-- -----------------------------------------------------------------------------
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    key_prefix VARCHAR(32) NOT NULL,              -- e.g. 'os_live_8F21' untuk display UI
    key_hash VARCHAR(255) NOT NULL,               -- bcrypt hash dari secret key
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE', -- 'ACTIVE', 'REVOKED'
    expires_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_api_keys_tenant_id ON api_keys(tenant_id);
CREATE INDEX idx_api_keys_key_prefix ON api_keys(key_prefix);

-- -----------------------------------------------------------------------------
-- 3. USERS (Compliance Team & Admin Users)
-- -----------------------------------------------------------------------------
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'COMPLIANCE', -- 'ADMIN', 'COMPLIANCE', 'VIEWER'
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',   -- 'ACTIVE', 'INACTIVE'
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_users_tenant_email UNIQUE(tenant_id, email)
);

CREATE INDEX idx_users_tenant_id ON users(tenant_id);
CREATE INDEX idx_users_email ON users(email);

-- -----------------------------------------------------------------------------
-- 4. SESSIONS (Dashboard User Server-Side Cookie Session)
-- -----------------------------------------------------------------------------
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ
);

CREATE INDEX idx_sessions_tenant_user ON sessions(tenant_id, user_id);
CREATE INDEX idx_sessions_session_hash ON sessions(session_hash);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- -----------------------------------------------------------------------------
-- 5. TRANSACTIONS (Financial Ingestion Data)
-- -----------------------------------------------------------------------------
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    external_transaction_id VARCHAR(100) NOT NULL, -- Client-provided TXN ID
    account_id VARCHAR(100) NOT NULL,              -- Sender Account/Phone
    counterparty_id VARCHAR(100) NOT NULL,         -- Receiver Account / Merchant
    amount NUMERIC(15, 2) NOT NULL CHECK(amount > 0),
    currency VARCHAR(10) NOT NULL DEFAULT 'IDR',
    transaction_type VARCHAR(50) NOT NULL,         -- 'PURCHASE', 'P2P_TRANSFER', 'BILL_PAYMENT', 'TOPUP', 'CASHOUT'
    occurred_at TIMESTAMPTZ NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,   -- Customer name, payment method, IP, location, device
    status VARCHAR(50) NOT NULL DEFAULT 'RECEIVED',-- 'RECEIVED', 'PROCESSING', 'ASSESSED', 'REVIEW', 'RESOLVED', 'FAILED'
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_transactions_tenant_external UNIQUE(tenant_id, external_transaction_id)
);

CREATE INDEX idx_transactions_tenant_id ON transactions(tenant_id);
CREATE INDEX idx_transactions_status ON transactions(status);
CREATE INDEX idx_transactions_occurred_at ON transactions(occurred_at);
CREATE INDEX idx_transactions_account_id ON transactions(account_id);
CREATE INDEX idx_transactions_created_at ON transactions(created_at);

-- -----------------------------------------------------------------------------
-- 6. IDEMPOTENCY KEYS (Duplicate Ingestion Prevention - TTL 24h)
-- -----------------------------------------------------------------------------
CREATE TABLE idempotency_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    idempotency_key VARCHAR(255) NOT NULL,
    request_hash VARCHAR(255) NOT NULL,
    transaction_id UUID REFERENCES transactions(id) ON DELETE SET NULL,
    response_snapshot JSONB,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_idempotency_tenant_key UNIQUE(tenant_id, idempotency_key)
);

CREATE INDEX idx_idempotency_tenant_key ON idempotency_keys(tenant_id, idempotency_key);
CREATE INDEX idx_idempotency_expires_at ON idempotency_keys(expires_at);

-- -----------------------------------------------------------------------------
-- 7. AML RULES (Configurable Compliance & Fraud Rules)
-- -----------------------------------------------------------------------------
CREATE TABLE aml_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    rule_code VARCHAR(100) NOT NULL,             -- e.g. 'AML-001', 'RULE_VELOCITY'
    name VARCHAR(255) NOT NULL,
    description TEXT,
    rule_type VARCHAR(100) NOT NULL,             -- 'AMOUNT_THRESHOLD', 'VELOCITY', 'FREQUENCY', 'NEW_COUNTERPARTY', 'CONFIGURABLE_PATTERN'
    configuration JSONB NOT NULL DEFAULT '{}'::jsonb,
    severity VARCHAR(50) NOT NULL DEFAULT 'MEDIUM', -- 'LOW', 'MEDIUM', 'HIGH'
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_aml_rules_tenant_code UNIQUE(tenant_id, rule_code)
);

CREATE INDEX idx_aml_rules_tenant_id ON aml_rules(tenant_id);
CREATE INDEX idx_aml_rules_enabled ON aml_rules(enabled);

-- -----------------------------------------------------------------------------
-- 8. RISK ASSESSMENTS (AI & Rule Engine Output)
-- -----------------------------------------------------------------------------
CREATE TABLE risk_assessments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    risk_score INTEGER NOT NULL CHECK(risk_score BETWEEN 0 AND 100),
    risk_level VARCHAR(50) NOT NULL,             -- 'LOW', 'MEDIUM', 'HIGH'
    rules_triggered JSONB NOT NULL DEFAULT '[]'::jsonb,
    risk_factors JSONB NOT NULL DEFAULT '[]'::jsonb,
    ai_verdict VARCHAR(50) NOT NULL DEFAULT 'CLEAR', -- 'REVIEW', 'CLEAR', 'ESCALATE'
    ai_confidence NUMERIC(4, 2) DEFAULT 0.0,
    ai_explanation TEXT,
    ai_model VARCHAR(100),
    prompt_version VARCHAR(50),
    assessed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_risk_assessments_tenant_id ON risk_assessments(tenant_id);
CREATE INDEX idx_risk_assessments_transaction_id ON risk_assessments(transaction_id);
CREATE INDEX idx_risk_assessments_risk_score ON risk_assessments(risk_score);
CREATE INDEX idx_risk_assessments_risk_level ON risk_assessments(risk_level);

-- -----------------------------------------------------------------------------
-- 9. ALERTS (Prioritized Items for Investigation)
-- -----------------------------------------------------------------------------
CREATE TABLE alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    risk_assessment_id UUID NOT NULL REFERENCES risk_assessments(id) ON DELETE CASCADE,
    risk_score INTEGER NOT NULL,
    risk_level VARCHAR(50) NOT NULL,
    priority VARCHAR(10) NOT NULL DEFAULT 'P2',  -- 'P1', 'P2'
    status VARCHAR(50) NOT NULL DEFAULT 'OPEN',  -- 'OPEN', 'IN_REVIEW', 'ESCALATED', 'RESOLVED', 'FALSE_POSITIVE'
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_alerts_tenant_id ON alerts(tenant_id);
CREATE INDEX idx_alerts_status ON alerts(status);
CREATE INDEX idx_alerts_priority ON alerts(priority);
CREATE INDEX idx_alerts_created_at ON alerts(created_at);

-- -----------------------------------------------------------------------------
-- 10. INVESTIGATION NOTES (Immutable Case Audit Trail)
-- -----------------------------------------------------------------------------
CREATE TABLE investigation_notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    alert_id UUID NOT NULL REFERENCES alerts(id) ON DELETE CASCADE,
    author_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    note TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_investigation_notes_alert_id ON investigation_notes(alert_id);

-- -----------------------------------------------------------------------------
-- 11. INVESTIGATION RESOLUTIONS (Case Closure History)
-- -----------------------------------------------------------------------------
CREATE TABLE investigation_resolutions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    alert_id UUID NOT NULL REFERENCES alerts(id) ON DELETE CASCADE,
    resolver_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    resolution VARCHAR(50) NOT NULL,             -- 'LEGITIMATE', 'FALSE_POSITIVE', 'SUSPICIOUS', 'ESCALATED'
    reason TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_investigation_resolutions_alert_id ON investigation_resolutions(alert_id);

-- -----------------------------------------------------------------------------
-- 12. JOBS (PostgreSQL Queue for Asynchronous Workers)
-- -----------------------------------------------------------------------------
CREATE TABLE jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    job_type VARCHAR(100) NOT NULL,              -- 'AML_ASSESSMENT', 'WEBHOOK_DELIVERY'
    payload JSONB NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING', -- 'PENDING', 'PROCESSING', 'COMPLETED', 'FAILED', 'RETRYING'
    attempts INTEGER NOT NULL DEFAULT 0,
    max_attempts INTEGER NOT NULL DEFAULT 5,
    available_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    locked_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_jobs_status_available ON jobs(status, available_at);
CREATE INDEX idx_jobs_tenant_id ON jobs(tenant_id);

-- -----------------------------------------------------------------------------
-- 13. WEBHOOK CONFIGS (Customer Endpoint Subscriptions)
-- -----------------------------------------------------------------------------
CREATE TABLE webhook_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    url VARCHAR(500) NOT NULL,
    secret_encrypted VARCHAR(255) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    subscribed_events JSONB NOT NULL DEFAULT '["AML_ALERT_CREATED", "AML_ALERT_UPDATED", "AML_ALERT_RESOLVED"]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_webhook_configs_tenant_id ON webhook_configs(tenant_id);

-- -----------------------------------------------------------------------------
-- 14. WEBHOOK DELIVERIES (Event Delivery Lifecycle & Replay Tracking)
-- -----------------------------------------------------------------------------
CREATE TABLE webhook_deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    event_id VARCHAR(100) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING', -- 'PENDING', 'DELIVERING', 'DELIVERED', 'RETRYING', 'FAILED'
    attempt INTEGER NOT NULL DEFAULT 0,
    response_code INTEGER,
    last_error TEXT,
    next_retry_at TIMESTAMPTZ,
    delivered_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_webhook_deliveries_tenant_event UNIQUE(tenant_id, event_id)
);

CREATE INDEX idx_webhook_deliveries_tenant_id ON webhook_deliveries(tenant_id);
CREATE INDEX idx_webhook_deliveries_status ON webhook_deliveries(status);

-- -----------------------------------------------------------------------------
-- 15. AUDIT LOGS (Immutable Append-Only Security Trail)
-- -----------------------------------------------------------------------------
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    actor_type VARCHAR(50) NOT NULL,             -- 'USER', 'SYSTEM', 'API_KEY'
    actor_id VARCHAR(100) NOT NULL,
    action VARCHAR(100) NOT NULL,                -- e.g. 'ALERT_RESOLVED', 'RULE_CREATED'
    resource_type VARCHAR(100) NOT NULL,         -- 'alert', 'rule', 'transaction'
    resource_id UUID,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_tenant_id ON audit_logs(tenant_id);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
