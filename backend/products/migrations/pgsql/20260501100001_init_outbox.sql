-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS outbox (
    id            UUID PRIMARY KEY,
    aggregate_id  UUID NOT NULL,
    event_type    TEXT NOT NULL,
    payload       BYTEA NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    sent_at       TIMESTAMPTZ,
    attempts      INT NOT NULL DEFAULT 0,
    next_retry_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_error    TEXT
);

CREATE INDEX IF NOT EXISTS idx_outbox_pending_due
    ON outbox (next_retry_at)
    WHERE sent_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS outbox;
-- +goose StatementEnd
