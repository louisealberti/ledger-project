-- migration: 003_transaction_and_idempotency

-- Table to store idempotency keys and cache responses
CREATE TABLE IF NOT EXISTS idempotency_keys (
    id_key UUID PRIMARY KEY,
    response_status INT NOT NULL,
    response_body TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Table to store the actual financial movements
CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY,
    account_from_id UUID REFERENCES accounts(id),
    account_to_id UUID REFERENCES accounts(id),
    amount BIGINT NOT NULL,
    description TEXT,
    idempotency_key UUID UNIQUE NOT NULL REFERENCES idempotency_keys(id_key),
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_transactions_from ON transactions(account_from_id);
CREATE INDEX idx_transactions_to ON transactions(account_to_id);