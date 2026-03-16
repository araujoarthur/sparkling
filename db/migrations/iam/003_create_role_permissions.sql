-- +goose Up
CREATE TABLE iam.role_permissions (
    role_id       UUID        NOT NULL REFERENCES iam.roles(id)       ON DELETE CASCADE,
    permission_id UUID        NOT NULL REFERENCES iam.permissions(id) ON DELETE CASCADE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (role_id, permission_id)
);

CREATE INDEX idx_role_permissions_role_id       ON iam.role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission_id ON iam.role_permissions(permission_id);

-- +goose Down
DROP TABLE iam.role_permissions;