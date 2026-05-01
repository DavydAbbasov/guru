-- +goose Up
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS products (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_products_active_created
    ON products (created_at DESC, id DESC)
    WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS products;
