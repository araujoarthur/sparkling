-- +goose Up
CREATE TABLE auth.service_tokens (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    identity_id UUID        NOT NULL REFERENCES auth.identities(id) ON DELETE CASCADE,
    token       TEXT        NOT NULL UNIQUE,
    issued_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at  TIMESTAMPTZ
);

CREATE INDEX idx_service_tokens_identity_id ON auth.service_tokens(identity_id);
CREATE INDEX idx_service_tokens_token       ON auth.service_tokens(token);

-- +goose Down
DROP TABLE auth.service_tokens;