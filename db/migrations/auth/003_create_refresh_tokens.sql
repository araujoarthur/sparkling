-- +goose Up
CREATE TABLE auth.refresh_tokens (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    identity_id UUID        NOT NULL REFERENCES auth.identities(id) ON DELETE CASCADE,
    token_hash  TEXT        NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_refresh_tokens_identity_id ON auth.refresh_tokens(identity_id);
CREATE INDEX idx_refresh_tokens_token_hash  ON auth.refresh_tokens(token_hash);

-- +goose Down
DROP TABLE auth.refresh_tokens;