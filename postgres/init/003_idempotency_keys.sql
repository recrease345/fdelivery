CREATE TABLE IF NOT EXISTS orders.idempotency_keys (
    key TEXT PRIMARY KEY,
    order_id TEXT NOT NULL,
    response_body JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL 
);

CREATE INDEX idx_idempotency_keys_created_at ON orders.idempotency_keys(created_at);