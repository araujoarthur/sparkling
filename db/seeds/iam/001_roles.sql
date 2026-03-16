-- +goose Up
INSERT INTO iam.roles (name, description, is_system) VALUES
    ('superadmin', 'Full system access',      true),
    ('admin',      'Administrative access',   true),
    ('user',       'Standard user access',    true)
ON CONFLICT (name) DO NOTHING;

-- +goose Down
DELETE FROM iam.roles WHERE name IN ('superadmin', 'admin', 'user');