-- 002_idempotency.sql: Implements the idempotency control layer.
-- Responsibility: Prevents duplicate processing of financial movements.

CREATE TABLE IF NOT EXISTS idempotency_keys (
    id_key UUID PRIMARY KEY,
    response_status INTEGER NOT NULL,
    response_body TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);