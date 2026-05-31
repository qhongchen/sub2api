CREATE TABLE IF NOT EXISTS request_records (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ NULL,

    request_id TEXT NOT NULL,
    client_request_id TEXT NULL,

    user_id BIGINT NULL,
    api_key_id BIGINT NULL,
    account_id BIGINT NULL,
    group_id BIGINT NULL,

    session_id TEXT NULL,
    session_source TEXT NULL,
    client_session_id TEXT NULL,

    platform TEXT NULL,
    model TEXT NULL,
    requested_model TEXT NULL,
    upstream_model TEXT NULL,

    request_type SMALLINT NULL,
    stream BOOLEAN NOT NULL DEFAULT FALSE,
    inbound_endpoint TEXT NULL,
    upstream_endpoint TEXT NULL,

    outcome TEXT NOT NULL DEFAULT 'pending',
    status_code INTEGER NULL,
    duration_ms INTEGER NULL,
    first_token_ms INTEGER NULL,

    billable BOOLEAN NOT NULL DEFAULT FALSE,
    input_tokens INTEGER NULL,
    output_tokens INTEGER NULL,
    total_cost NUMERIC NULL,
    actual_cost NUMERIC NULL,

    ip_address TEXT NULL,
    user_agent TEXT NULL,
    error_message TEXT NULL
);

CREATE INDEX IF NOT EXISTS idx_request_records_created_at_id
    ON request_records (created_at DESC, id DESC);

CREATE UNIQUE INDEX IF NOT EXISTS idx_request_records_request_id_unique
    ON request_records (request_id);

CREATE INDEX IF NOT EXISTS idx_request_records_client_request_id
    ON request_records (client_request_id);

CREATE INDEX IF NOT EXISTS idx_request_records_session_created_at
    ON request_records (session_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_request_records_api_key_created_at
    ON request_records (api_key_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_request_records_account_created_at
    ON request_records (account_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_request_records_user_created_at
    ON request_records (user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_request_records_outcome_created_at
    ON request_records (outcome, created_at DESC);
