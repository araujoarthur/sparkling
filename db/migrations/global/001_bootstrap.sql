-- +goose Up
CREATE ROLE intranetbackend_auth WITH LOGIN PASSWORD 'default';
CREATE ROLE intranetbackend_iam  WITH LOGIN PASSWORD 'default';

CREATE SCHEMA global AUTHORIZATION intranetbackend_owner;
CREATE SCHEMA auth   AUTHORIZATION intranetbackend_owner;
CREATE SCHEMA iam    AUTHORIZATION intranetbackend_owner;

REVOKE ALL ON SCHEMA global FROM PUBLIC;
REVOKE ALL ON SCHEMA auth   FROM PUBLIC;
REVOKE ALL ON SCHEMA iam    FROM PUBLIC;

GRANT USAGE ON SCHEMA auth TO intranetbackend_auth;
GRANT USAGE ON SCHEMA iam  TO intranetbackend_iam;

ALTER DEFAULT PRIVILEGES IN SCHEMA auth
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO intranetbackend_auth;

ALTER DEFAULT PRIVILEGES IN SCHEMA iam
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO intranetbackend_iam;

-- +goose Down
DROP SCHEMA global CASCADE;
DROP SCHEMA auth   CASCADE;
DROP SCHEMA iam    CASCADE;
DROP ROLE intranetbackend_auth;
DROP ROLE intranetbackend_iam;