# Auth Service

## Overview

The auth service is responsible for identity and credential management.
It is the only service that issues tokens — all other services validate
tokens using the public key but never issue them.

## Responsibilities

- Managing identities (one per entity, referenced by IAM as external_id)
- Managing credentials (password, service_token) linked to identities
- Issuing access tokens and refresh tokens for human users
- Issuing and rotating service tokens for service accounts
- Validating credentials at login
- Creating an IAM principal automatically when a new identity is registered
- Running a daily background job to rotate all service tokens

## Token flows

### User login

  POST /api/v1/auth/login
      → validate username + password credential
      → issue access token (15 min) + refresh token (7 days)
      → return both to caller

### Token refresh

  POST /api/v1/auth/refresh
      → validate refresh token
      → issue new access token
      → return to caller

### Service token fetch

  POST /api/v1/auth/token
      → validate service_token credential
      → return current non-expiring service token

## Schema

Auth owns the following PostgreSQL schemas:

  auth.identities
      id              uuid PK
      created_at      timestamptz

  auth.credentials
      id              uuid PK
      identity_id     uuid FK → auth.identities
      type            credential_type enum ('password', 'service_token')
      secret_hash     text
      created_at      timestamptz
      last_used_at    timestamptz nullable

  auth.refresh_tokens
      id              uuid PK
      identity_id     uuid FK → auth.identities
      token_hash      text
      expires_at      timestamptz
      revoked_at      timestamptz nullable
      created_at      timestamptz

  auth.service_tokens
      id              uuid PK
      identity_id     uuid FK → auth.identities
      token           text
      issued_at       timestamptz
      revoked_at      timestamptz nullable

## Service structure

  services/auth/
  ├── contract/           request/response shapes
  ├── client/             HTTP client importable by other services
  ├── cmd/authd/          service entrypoint
  └── internal/
      ├── repository/     data access layer
      ├── domain/         business logic
      └── handler/rest/   HTTP handlers

## Business rules

- A new identity always triggers an IAM principal creation in the same transaction
- Credentials are always hashed before storage — never stored in plain text
- A revoked refresh token cannot be used to issue new access tokens
- Service tokens are rotated daily by a background job
- Service tokens can be manually rotated via inetbctl token rotate in emergencies
- Only one active service token exists per identity at any time —
  rotation revokes the previous token before issuing a new one
- An identity can have multiple credential types but only one of each type