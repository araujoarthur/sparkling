-- +goose Up
CREATE TABLE iam.principal_roles (
    principal_id UUID        NOT NULL REFERENCES iam.principals(id) ON DELETE CASCADE,
    role_id      UUID        NOT NULL REFERENCES iam.roles(id)      ON DELETE CASCADE,
    granted_by   UUID        NOT NULL REFERENCES iam.principals(id) ON DELETE RESTRICT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (principal_id, role_id)
);

CREATE INDEX idx_principal_roles_principal_id ON iam.principal_roles(principal_id);
CREATE INDEX idx_principal_roles_role_id      ON iam.principal_roles(role_id);

-- +goose Down
DROP TABLE iam.principal_roles;