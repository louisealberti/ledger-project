-- 001_initial_schema.sql: Sets up the core ledger entities.
-- Responsibility: Defines identity, balance, and history tables.

CREATE TABLE IF NOT EXISTS accounts (
    id UUID PRIMARY KEY,
    owner_id UUID NOT NULL,
    balance BIGINT NOT NULL DEFAULT 0,
    active BOOLEAN NOT NULL DEFAULT true,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY,
    account_from_id UUID REFERENCES accounts(id),
    account_to_id UUID REFERENCES accounts(id),
    amount BIGINT NOT NULL,
    description TEXT,
    idempotency_key UUID UNIQUE NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);