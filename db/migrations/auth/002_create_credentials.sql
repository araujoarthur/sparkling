-- +goose Up
CREATE TYPE auth.credential_type AS ENUM ('password', 'service_token');

CREATE TABLE auth.credentials (
    id          UUID                 PRIMARY KEY DEFAULT gen_random_uuid(),
    identity_id UUID                 NOT NULL REFERENCES auth.identities(id) ON DELETE CASCADE,
    type        auth.credential_type NOT NULL,
    identifier  TEXT,
    secret_hash TEXT,
    created_at  TIMESTAMPTZ          NOT NULL DEFAULT now(),
    last_used_at TIMESTAMPTZ,

    UNIQUE (type, identifier)
);

CREATE INDEX idx_credentials_identity_id ON auth.credentials(identity_id);
CREATE INDEX idx_credentials_type_identifier ON auth.credentials(type, identifier);

-- +goose Down
DROP TABLE auth.credentials;
DROP TYPE auth.credential_type;