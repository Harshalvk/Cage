CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS api_keys (
  id UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
  key_hash   TEXT NOT NULL UNIQUE,
  name       TEXT NOT NULL,
  create_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  revoked_at TIMESTAMPTZ
)