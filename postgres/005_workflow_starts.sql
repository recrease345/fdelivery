CREATE TABLE workflow_starts (
    event_id UUID PRIMARY KEY,
    order_id UUID NOT NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT now()
);