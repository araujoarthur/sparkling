-- +goose Up
CREATE TABLE iam.permissions (
    id          UUID         PRIMARY KEY DEFAULT uuidv7(),
    name        VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_permissions_name ON iam.permissions(name);

-- +goose Down
DROP TABLE iam.permissions;