-- +goose Up
CREATE TYPE iam.principal_type AS ENUM ('user', 'service');

CREATE TABLE iam.principals (
    id             UUID               PRIMARY KEY DEFAULT uuidv7(),
    external_id    UUID               NOT NULL,
    principal_type iam.principal_type NOT NULL,
    is_active      BOOLEAN            NOT NULL DEFAULT true,
    created_at     TIMESTAMPTZ        NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ        NOT NULL DEFAULT now(),

    UNIQUE (external_id, principal_type)
);

CREATE INDEX idx_principals_external_id ON iam.principals(external_id);

CREATE TRIGGER trg_principals_updated_at
    BEFORE UPDATE ON iam.principals
    FOR EACH ROW EXECUTE FUNCTION iam.set_updated_at();

-- +goose Down
DROP TRIGGER IF EXISTS trg_principals_updated_at ON iam.principals;
DROP TABLE iam.principals;
DROP TYPE iam.principal_type;