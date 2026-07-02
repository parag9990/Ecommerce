# Authentication

## Service Location

`backend/services/auth-service`.

## Responsibility

The auth service owns authentication accounts, password credentials, refresh tokens, OTP challenges, JWT issuance, JWKS publishing, logout, password reset, and internal role assignment APIs.

## Public HTTP APIs Found

Registered in `internal/transport/http/handler.go`:

| Method | Path | Purpose |
| --- | --- | --- |
| POST | `/api/v1/auth/login` | Login with credentials |
| POST | `/api/v1/auth/refresh` | Refresh token pair |
| POST | `/api/v1/auth/logout` | Revoke refresh token or logout all devices |
| POST | `/api/v1/auth/otp/send` | Create and deliver OTP challenge |
| POST | `/api/v1/auth/otp/verify` | Verify OTP challenge |
| POST | `/api/v1/auth/password/forgot` | Start password reset flow |
| GET | `/.well-known/jwks.json` | Publish JWT public keys |
| GET | `/healthz` | Liveness |
| GET | `/readyz` | Readiness |

`POST /api/v1/auth/password/reset` appears in the API catalog, but the visible auth handler registers internal reset APIs and password forgot publicly. Confirm before using it as a public route.

## Internal APIs Found

| Method | Path | Purpose |
| --- | --- | --- |
| POST | `/internal/v1/auth/credentials` | Credential management |
| POST | `/internal/v1/auth/password/verify` | Verify password |
| POST | `/internal/v1/auth/password/reset` | Reset password |
| POST | `/internal/v1/auth/tokens/issue` | Issue token pair |
| POST | `/internal/v1/auth/tokens/verify` | Verify token |
| GET | `/internal/v1/auth/roles` | List roles |
| POST | `/internal/v1/auth/roles/assign` | Assign role |
| POST | `/internal/v1/auth/roles/revoke` | Revoke role |

Role mutation routes are protected by auth middleware and configured role checks.

## Login Flow

1. Client posts login credentials.
2. Auth usecase verifies password through the configured password router.
3. Auth service issues an access token and refresh token.
4. Refresh token hash/session metadata is persisted.
5. Login session-link event is recorded through configured mode.
6. Response returns token/session metadata.

## Signup Flow

`POST /api/v1/auth/signup` is listed in `api/master-api.json` and used by `frontend/user-app`, but the current auth service HTTP router does not register a signup route.

Status: `Not found in current codebase`.

The IDE context referenced `backend/services/auth-service/internal/domain/signup.go`, but that file was not present in the current filesystem.

## JWT And Session Handling

| Item | Implementation found |
| --- | --- |
| Access token | RSA-signed JWT |
| Public keys | JWKS endpoint at `/.well-known/jwks.json` |
| Refresh token | Stored as hash in MySQL |
| Access TTL | `AUTH_ACCESS_TOKEN_TTL`, example `15m` |
| Refresh TTL | `AUTH_REFRESH_TOKEN_TTL`, example `720h` |
| Claims | User ID, session ID, roles, seller ID, tenant ID where available |
| Gateway validation | API Gateway verifies JWT using auth JWKS |

## Role-Based Access

Roles are modeled by the auth domain and role assignment tables. Roles found include buyer/seller/admin-oriented roles such as:

- `buyer`
- `seller`
- `seller_manager`
- `seller_catalog_editor`
- `seller_order_manager`
- `admin`
- `operations_admin`
- `finance_admin`
- `catalog_admin`
- `readonly_admin`
- `superadmin`

The API Gateway enforces route-level roles from `api/master-api.json`. Auth service role mutation internals enforce their own authorization checks.

## OTP Flow

1. OTP challenge is created and stored with hashed OTP data.
2. Redis-backed rate limiting/configuration is used by the service.
3. Notification service gRPC client delivers the OTP.
4. Verification checks the challenge and marks the OTP state.

OTP purposes include signup-related usage, but a public signup route was not found.

## Password Reset Flow

The service contains password forgot/reset usecases and public forgot route. Public reset route registration was not confirmed in the visible HTTP handler.

## Token Refresh And Logout

| Flow | Found behavior |
| --- | --- |
| Refresh | Uses refresh token to issue a new token pair |
| Logout | Revokes provided refresh token |
| Logout all devices | Supported by logout request shape/usecase |

## Middleware And Security

- JWT issuer/verifier.
- Password hashing supports Argon2id/bcrypt through password router.
- OTP generator/hasher.
- Pepper configuration for credentials and OTP.
- Redis integration.
- Session-link/outbox mode.

## Environment Variables

Important variables in `backend/services/auth-service/.env.example`:

- `AUTH_HTTP_ADDR`
- `AUTH_MYSQL_DSN`
- `AUTH_REDIS_ADDR`
- `AUTH_REDIS_DB`
- `AUTH_JWT_PRIVATE_KEY_PATH`
- `AUTH_JWT_PUBLIC_KEY_PATH`
- `AUTH_JWT_ISSUER`
- `AUTH_JWT_AUDIENCE`
- `AUTH_ACCESS_TOKEN_TTL`
- `AUTH_REFRESH_TOKEN_TTL`
- Password/OTP pepper values
- `AUTH_NOTIFICATION_GRPC_ADDR`
- `SESSION_LINK_MODE`
- `AUTH_EVENTS_TOPIC`
- `AUTH_EVENTS_PUBLISH_ENDPOINT`

## Data Stores

| Store | Usage |
| --- | --- |
| MySQL `auth_db` | Accounts, credentials, refresh tokens, OTP, roles, outbox |
| Redis | Rate limiting/session-related auth state |

## Run Command / Health

| Item | Found |
| --- | --- |
| Compose service | `auth-service` |
| Local host port | `8081` |
| Liveness | `GET /healthz` |
| Readiness | `GET /readyz` |

## Missing Information

- Public signup implementation: Not found in current codebase.
- Public password reset route confirmation: Not found in current codebase.
- Account registration end-to-end documentation: Not found in current codebase.
