-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION iam.set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE iam.roles (
    id          UUID        PRIMARY KEY DEFAULT uuidv7(),
    name        VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    is_system   BOOLEAN     NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- +goose StatementEnd

CREATE INDEX idx_roles_name ON iam.roles(name);

CREATE TRIGGER trg_roles_updated_at
    BEFORE UPDATE ON iam.roles
    FOR EACH ROW EXECUTE FUNCTION iam.set_updated_at();

-- +goose Down
DROP TRIGGER IF EXISTS trg_roles_updated_at ON iam.roles;
DROP TABLE iam.roles;
DROP FUNCTION IF EXISTS iam.set_updated_at();