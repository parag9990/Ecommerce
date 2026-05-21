# Project Dependency & Setup Guide

Input implementation file:

```text
TaskImplementation/Auth Service/task4.md
```

Generated dependency file:

```text
TaskImplementation/Auth Service/task4_Dependency.md
```

This guide is intentionally incremental. Previous Auth Service dependency files already explain the common setup in detail, so this file focuses on Task 4 specific setup: JWT issuing, RSA keys, refresh token rotation, token metadata columns, and token-related environment variables.

Previous files inspected before writing this guide:

```text
TaskImplementation/Auth Service/task1_Dependency.md
TaskImplementation/Auth Service/task2_Dependency.md
TaskImplementation/Auth Service/task3_Dependency.md
```

---

## 1. Project Overview

Task 4 ka main goal hai Auth Service se secure token pair issue karna:

- Short-lived JWT access token
- Long-lived opaque refresh token
- Refresh token rotation
- JWT claims me `user_id`, `session_id`, `roles`, optional `seller_id`, optional `tenant_id`
- Public JWKS endpoint so Gateway/other services JWT verify kar sakein

Beginner mental model:

```text
Login/signup success
  -> Auth Service access token sign karta hai
  -> Auth Service refresh token generate karta hai
  -> Plain refresh token client ko milta hai
  -> Sirf hashed refresh token MySQL me store hota hai
  -> Gateway access token ko JWKS public key se verify karta hai
```

Important reuse note:

- Clone, Go install, MySQL install, Redis install, Docker setup, complete `.env`, and full local run flow already documented hai.
- Refer `TaskImplementation/Auth Service/task1_Dependency.md` for baseline setup.
- This file sirf Task 4 ke new or important setup differences ko explain karta hai.

---

## 2. Tech Stack

### Task 4 specific technologies

| Technology | What it is | Why Task 4 uses it | Required? | Beginner explanation |
|---|---|---|---:|---|
| JWT access token | Signed JSON token format | API requests me authenticated user/session/role claims carry karne ke liye | Yes | JWT ek signed token hota hai. Client isko `Authorization: Bearer ...` me bhejta hai. |
| RS256 | RSA SHA-256 JWT signing algorithm | Auth private key se sign, Gateway public key se verify | Yes | RS256 me private key secret hoti hai, public key safely share ho sakti hai. |
| RSA private key | PEM file containing signing key | Access token sign karne ke liye | Yes | Ye sabse sensitive key hai. Isko Git me commit nahi karna. |
| RSA public key | PEM file or derived public key | JWKS me publish/verify karne ke liye | Recommended | Public key Gateway ko token verify karne me help karti hai. |
| JWKS | JSON Web Key Set endpoint | Public verification keys expose karne ke liye | Yes | `/.well-known/jwks.json` se Gateway key fetch kar sakta hai. |
| `kid` header | Key ID in JWT header | Key rotation ke time correct public key choose karne ke liye | Yes | `kid` ek label hai, jaise `auth-local-1`. |
| Go `crypto/rand` | Secure random generator | Refresh token, session id, token id generate karne ke liye | Yes | Randomness predictable nahi honi chahiye, warna token guess ho sakta hai. |
| Go `crypto/hmac` + `crypto/sha256` | Standard keyed hash utilities | Refresh token ko DB me irreversible hash form me store karne ke liye | Yes | DB leak hone par plain refresh tokens expose nahi hote. |
| Go `crypto/rsa` + `crypto/x509` | Standard RSA key/signature packages | RS256 sign/verify and PEM parsing ke liye | Yes | Current code JWT signing manually standard library se karta hai. |
| Base64URL | URL-safe encoding format | JWT sections and random token strings encode karne ke liye | Yes | JWT ke 3 parts Base64URL encoded hote hain. |
| MySQL transaction support | DB atomic operation support | Old refresh token revoke + new refresh token insert ek transaction me karne ke liye | Yes | Agar refresh rotate halfway fail ho, token state corrupt nahi honi chahiye. |
| OpenSSL | CLI crypto tool | Local RSA key pair generate karne ke liye | Recommended | Local development me keys banane ka simple tool. |

### Important dependency status

`task4.md` mentions:

```text
github.com/golang-jwt/jwt/v5
```

But current `backend/services/auth-service/go.mod` does not include this module, and current implementation signs/verifies JWT using Go standard library packages.

So for the current codebase:

- Do not install `github.com/golang-jwt/jwt/v5` just because Task 4 markdown mentions it.
- Install it only if the implementation is later refactored to import that package.
- If future code imports it, then run:

```bash
cd backend/services/auth-service
go get github.com/golang-jwt/jwt/v5
go mod tidy
```

### Already documented baseline technologies

| Baseline topic | Reuse documentation |
|---|---|
| Go `1.26.3`, Go modules, Go workspace | `TaskImplementation/Auth Service/task1_Dependency.md`, section `4. Dependency Management` |
| `net/http` server | `TaskImplementation/Auth Service/task1_Dependency.md`, section `2. Tech Stack` |
| MySQL 8+ installation and Docker | `TaskImplementation/Auth Service/task1_Dependency.md`, section `5. Database Setup` |
| Redis setup | `TaskImplementation/Auth Service/task1_Dependency.md`, section `6. Redis / Queue / External Services` |
| Docker dependency containers | `TaskImplementation/Auth Service/task1_Dependency.md`, section `8. Docker Setup` |
| Full local `.env` | `TaskImplementation/Auth Service/task1_Dependency.md`, section `7. Environment Variables` |

---

## 3. Required Software

Task 4 does not introduce a new database, cache, queue, or container service. It does require token-specific tooling/config.

| Software/tool | Needed for Task 4? | Why | Setup source |
|---|---:|---|---|
| Go | Yes | Build and test token code | `task1_Dependency.md`, section `3. Required Software` |
| MySQL client/server | Yes for full token flow | `refresh_tokens` and account claim metadata live in MySQL | `task1_Dependency.md`, section `5. Database Setup` |
| Redis | Required for current full server startup | Not Task 4 token logic, but `main.go` initializes Redis for OTP rate limits | `task1_Dependency.md`, section `6. Redis / Queue / External Services` |
| OpenSSL | Recommended | Generate local JWT RSA key pair | `task1_Dependency.md`, section `6. Redis / Queue / External Services`, subsection `OpenSSL / JWT Keys` |
| curl | Recommended | Verify JWKS, token issue, refresh, and health endpoints | `task1_Dependency.md`, section `3. Required Software` |
| Docker | Optional but recommended | Run local MySQL/Redis containers | `task1_Dependency.md`, section `8. Docker Setup` |

Task 4 readiness, simple version:

- MySQL running
- Migration `001` applied
- Migration `002` applied
- RSA private key available
- `JWT_KEY_ID`, `JWT_PRIVATE_KEY_PEM_PATH`, and `REFRESH_TOKEN_PEPPER` set
- Full server run also needs Redis and other baseline env variables from Task 1 docs

---

## 4. Dependency Management

### Language-specific dependency system: Go modules

The complete beginner explanation of:

- `go.mod`
- `go.sum`
- Go modules
- `go mod download`
- `go mod tidy`
- `go build`
- `go run`
- common Go dependency failures

already exists in:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 4. Dependency Management
```

### Task 4 current dependency impact

Current `go.mod` direct dependencies are still:

```text
github.com/go-sql-driver/mysql
golang.org/x/crypto
```

Task 4 token implementation mainly uses Go standard library packages, so no new external module is required in the current code.

Task 4 standard-library dependencies:

| Package area | Why used |
|---|---|
| `crypto/rand` | Secure refresh token/session/token id randomness |
| `crypto/rsa` | RS256 signing and verification |
| `crypto/sha256` | JWT signature digest and refresh token hash |
| `crypto/hmac` | Pepper-based refresh token hashing |
| `crypto/x509` + `encoding/pem` | RSA PEM key parsing |
| `encoding/base64` | JWT and token-safe encoding |
| `encoding/json` | JWT header/claim and JWKS JSON |

### Commands

Use baseline commands from `task1_Dependency.md`. Task 4 focused verification commands:

```bash
cd backend/services/auth-service
go mod download
go test ./internal/security/token ./internal/usecase
go test ./internal/transport/http
```

Full service test:

```bash
cd backend/services/auth-service
go test ./...
```

If your environment has read-only Go cache issues:

```bash
cd backend/services/auth-service
GOCACHE=/tmp/auth-go-cache go test ./...
```

### Common dependency confusion

| Symptom | Likely reason | Fix |
|---|---|---|
| Developer expects `github.com/golang-jwt/jwt/v5` in `go.mod` | `task4.md` design text mentions it, but current code does not import it | Do not add it unless code is refactored to use it |
| `go: go.mod requires go >= 1.26.3` | Local Go version old hai | Install Go version compatible with `backend/services/auth-service/go.mod` |
| `package not in std` | Command wrong directory se run hua | Run from `backend/services/auth-service` or use backend workspace command |

---

## 5. Database Setup

### A. Database detected

Task 4 uses the existing Auth Service MySQL database:

```text
auth_db
```

No new database server is introduced.

### B. What MySQL is

MySQL ek relational database hai. Isme data rows and columns ke form me tables me store hota hai. Auth Service me token, account, role, and credential data sensitive hai, isliye transactional relational DB useful hai.

Detailed MySQL installation and Docker setup already exists in:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 5. Database Setup
```

### C. Why Task 4 uses MySQL

Task 4 MySQL ko mainly 2 reasons ke liye use karta hai:

| DB object | Why Task 4 needs it |
|---|---|
| `refresh_tokens` table | Refresh token hash, expiry, revoked state, session id, and rotation chain store karne ke liye |
| `auth_accounts.user_id`, `seller_id`, `tenant_id` | JWT claims issue/refresh ke time account se user/seller/tenant context load karne ke liye |

Important: Refresh token plain text DB me store nahi hota. DB me only HMAC-SHA256 hash store hota hai.

### D. Required or optional

| Scenario | MySQL required? | Notes |
|---|---:|---|
| Token package unit tests only | No | Pure crypto/JWT tests DB ke bina run ho sakte hain |
| Refresh token repository tests/manual verification | Yes | `refresh_tokens` table needed |
| Full Auth Service startup | Yes | `main.go` MySQL ping karta hai |
| Login/refresh/logout HTTP flow | Yes | Account, roles, refresh token rows needed |

### E. Local installation

Do not duplicate MySQL install instructions here.

Refer:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 5. Database Setup
Subsections: D. Local installation, E. Docker setup, F. Start commands
```

### F. Docker setup

No Task 4 specific MySQL Docker change.

Use the same MySQL container:

| Setting | Value |
|---|---|
| Image | `mysql:8.4` or MySQL 8 compatible |
| Host port | `3306` |
| Database | `auth_db` |
| Suggested local app user | `auth_user` |
| Persistent volume | `ecommerce-auth-mysql-data` |

Refer:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 8. Docker Setup
```

### G. Task 4 migration requirement

Task 4 depends on migration `002`:

```text
backend/services/auth-service/migrations/002_add_token_claim_metadata.up.sql
```

This migration adds claim metadata columns to `auth_accounts`:

| Column | Purpose |
|---|---|
| `user_id` | JWT `sub` claim source |
| `seller_id` | Optional seller context claim |
| `tenant_id` | Optional tenant/store context claim |

Before `002`, apply base migration `001` from Task 2:

```text
backend/services/auth-service/migrations/001_create_auth_tables.up.sql
```

If you already followed `task1_Dependency.md` full migration flow, then migration `002` is already included. Do not rerun it on the same DB unless you intentionally reset the DB.

### H. Apply Task 4 migration

Run from Auth Service directory:

```bash
cd backend/services/auth-service
mysql -h 127.0.0.1 -P 3306 -u root -p < migrations/002_add_token_claim_metadata.up.sql
```

If this is a fresh DB, run `001` first:

```bash
cd backend/services/auth-service
mysql -h 127.0.0.1 -P 3306 -u root -p < migrations/001_create_auth_tables.up.sql
mysql -h 127.0.0.1 -P 3306 -u root -p < migrations/002_add_token_claim_metadata.up.sql
```

### I. Verify Task 4 migration

Check columns:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -p -e "SHOW COLUMNS FROM auth_db.auth_accounts LIKE 'user_id';"
mysql -h 127.0.0.1 -P 3306 -u root -p -e "SHOW COLUMNS FROM auth_db.auth_accounts LIKE 'seller_id';"
mysql -h 127.0.0.1 -P 3306 -u root -p -e "SHOW COLUMNS FROM auth_db.auth_accounts LIKE 'tenant_id';"
```

Check indexes:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -p -e "SHOW INDEX FROM auth_db.auth_accounts;"
```

Expected important indexes:

```text
uk_auth_accounts_user_id
idx_auth_accounts_seller_id
idx_auth_accounts_tenant_id
```

### J. Rollback Task 4 migration

Rollback file:

```text
backend/services/auth-service/migrations/002_add_token_claim_metadata.down.sql
```

Run only when you intentionally want to remove Task 4 claim columns:

```bash
cd backend/services/auth-service
mysql -h 127.0.0.1 -P 3306 -u root -p < migrations/002_add_token_claim_metadata.down.sql
```

Warning:

- Rollback removes `user_id`, `seller_id`, and `tenant_id`.
- Existing local data in those columns will be lost.
- Production rollback needs backup and approval.

### K. Connection string and credentials

No new DSN format is introduced by Task 4.

Use the existing Auth Service MySQL DSN:

```env
AUTH_MYSQL_DSN='auth_user:auth_password@tcp(127.0.0.1:3306)/auth_db?parseTime=true&charset=utf8mb4&loc=UTC'
```

This is already explained in:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 5. Database Setup
Subsection: I. Connection string format
Section: 7. Environment Variables
```

Credentials placement:

- Local: `backend/services/auth-service/.env`
- Production: secret manager or Kubernetes Secret
- Never commit real DB password

---

## 6. Redis / Queue / External Services

### Redis

Task 4 token logic itself does not introduce Redis usage.

However, current full Auth Service startup still requires Redis because `main.go` initializes Redis for OTP rate limiting. Agar Redis down hai, full server start nahi hoga.

Do not duplicate Redis setup here.

Refer:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 6. Redis / Queue / External Services
Subsection: Redis
```

### JWKS endpoint

JWKS external service nahi hai. Ye Auth Service ka HTTP endpoint hai:

```text
GET /.well-known/jwks.json
```

Purpose:

- Gateway public key fetch karega
- Access token signature verify karega
- Auth DB ko har request pe hit karne ki zarurat kam hogi

Verify locally after server starts:

```bash
curl -i http://localhost:8081/.well-known/jwks.json
```

Expected response has:

```json
{
  "keys": [
    {
      "kty": "RSA",
      "kid": "auth-local-1",
      "use": "sig",
      "alg": "RS256"
    }
  ]
}
```

Do not paste the full real JWKS response into logs if your environment treats infrastructure data as sensitive. Public keys are shareable, but logs should still stay clean.

### Session link / outbox

Task 4 refresh-token reuse detection can record session-link events in current code. The full outbox setup is already documented.

Refer:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 6. Redis / Queue / External Services
Subsection: Session Link / Outbox
```

Beginner local recommendation remains:

```env
SESSION_LINK_MODE=disabled
```

Integrated mode requires `SESSION_EVENT_PEPPER` and optional `AUTH_EVENTS_PUBLISH_ENDPOINT`.

### Kafka/RabbitMQ status

No direct Kafka, RabbitMQ, NATS, or broker client is introduced by Task 4 current code.

The current event publishing path is:

```text
MySQL auth_outbox_events -> optional HTTP publisher endpoint
```

If your platform later uses Kafka, it should usually sit behind the event-ingress service, not as a direct Task 4 dependency.

---

## 7. Environment Variables

Full `.env` is already documented in:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 7. Environment Variables
```

This section lists only Task 4 token-specific variables.

### Task 4 `.env` snippet

Add or verify these values in:

```text
backend/services/auth-service/.env
```

```env
# JWT access tokens
JWT_ISSUER=ecommerce-auth
JWT_AUDIENCE=ecommerce-api
JWT_ACCESS_TTL=15m
JWT_CLOCK_SKEW=30s
JWT_SIGNING_ALG=RS256
JWT_KEY_ID=auth-local-1
JWT_PRIVATE_KEY_PEM_PATH=secrets/jwt-private.pem
JWT_PUBLIC_KEY_PEM_PATH=secrets/jwt-public.pem

# Refresh tokens
REFRESH_TOKEN_TTL=720h
REFRESH_TOKEN_PEPPER=local_refresh_token_pepper_change_me_32_chars
```

### Variable explanation

| Variable | Required by current config? | Default if empty | Purpose | Beginner note |
|---|---:|---|---|---|
| `JWT_ISSUER` | No | `ecommerce-auth` | JWT `iss` claim | Gateway verifier must expect same issuer. |
| `JWT_AUDIENCE` | No | `ecommerce-api` | JWT `aud` claim | Gateway/services must validate same audience. |
| `JWT_ACCESS_TTL` | No | `15m` | Access token lifetime | Short rakho. Leaked access token expiry tak valid hota hai. |
| `JWT_CLOCK_SKEW` | No | `30s` | Small time drift allowance | Keep small, for example `30s` or `60s`. |
| `JWT_SIGNING_ALG` | No, but constrained | `RS256` | Signing algorithm | Current code accepts only `RS256`. |
| `JWT_KEY_ID` | Yes | none | JWT header `kid` | Key rotation and JWKS key lookup ke liye required. |
| `JWT_PRIVATE_KEY_PEM_PATH` | Yes | none | RSA private key file path | Private key se token sign hota hai. Never commit. |
| `JWT_PUBLIC_KEY_PEM_PATH` | No, recommended | derived from private key | RSA public key file path | JWKS ke liye explicit public key better hai. |
| `REFRESH_TOKEN_TTL` | No | `720h` | Refresh token expiry | `720h` means 30 days. |
| `REFRESH_TOKEN_PEPPER` | Yes | none | HMAC secret for refresh token hashes | Change karne se old refresh tokens invalid ho sakte hain. |

### Credentials placement

Local development:

- Keep values in `backend/services/auth-service/.env`.
- Load before running:

```bash
cd backend/services/auth-service
set -a
source .env
set +a
```

Production:

- Put `JWT_PRIVATE_KEY_PEM_PATH` file content in a mounted secret or secret volume.
- Put `REFRESH_TOKEN_PEPPER` in a secret manager.
- Keep `JWT_KEY_ID` coordinated with the public key published through JWKS.
- Do not bake private keys or peppers into Docker images.

### RSA key generation

Do not duplicate the full key-generation flow here.

Refer:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 6. Redis / Queue / External Services
Subsection: OpenSSL / JWT Keys
```

Expected local files after following that section:

```text
backend/services/auth-service/secrets/jwt-private.pem
backend/services/auth-service/secrets/jwt-public.pem
```

Security note:

- `secrets/jwt-private.pem` must not be committed.
- `.env` must not be committed.
- This repo currently has no `.gitignore` detected, so be extra careful until one is added.

---

## 8. Docker Setup

Task 4 does not add a Dockerfile, docker-compose file, or new container.

Current repo status remains:

| Item | Status |
|---|---|
| Auth Service Dockerfile | Not found |
| docker-compose.yml | Not found |
| MySQL container | Use previous docs |
| Redis container | Use previous docs |
| JWT key secret mount | Needed when app is containerized |

Reuse:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 8. Docker Setup
```

### Containerization note for future Dockerfile

If Auth Service later runs inside a container, remember:

- Mount private key file as a secret volume.
- Set `JWT_PRIVATE_KEY_PEM_PATH` to the in-container path.
- Do not copy `secrets/jwt-private.pem` into the image.
- Use Docker network service names for dependencies:

```env
AUTH_MYSQL_DSN='auth_user:auth_password@tcp(mysql:3306)/auth_db?parseTime=true&charset=utf8mb4&loc=UTC'
AUTH_REDIS_ADDR=redis:6379
JWT_PRIVATE_KEY_PEM_PATH=/run/secrets/auth/jwt-private.pem
JWT_PUBLIC_KEY_PEM_PATH=/run/secrets/auth/jwt-public.pem
```

### Volumes and networks

No new Task 4 volume is needed for MySQL/Redis.

Future app container should use:

| Mount/network | Purpose |
|---|---|
| Secret volume for JWT private key | Secure runtime key access |
| Secret/env for `REFRESH_TOKEN_PEPPER` | Secure refresh token hashing |
| Internal Docker network | App connects to `mysql`, `redis`, event ingress |
| No public MySQL/Redis exposure in prod | DB/cache should stay private |

---

## 9. Local Development Setup

This is the Task 4 incremental setup. For full beginner onboarding from clone, use:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 9. Local Development Setup
```

### Flow A: Verify Task 4 token code only

This checks token and refresh logic tests without starting the full server:

```bash
cd backend/services/auth-service
go mod download
go test ./internal/security/token ./internal/usecase
```

If Go cache is not writable:

```bash
cd backend/services/auth-service
GOCACHE=/tmp/auth-go-cache go test ./internal/security/token ./internal/usecase
```

### Flow B: Prepare DB for Task 4 manual testing

Start MySQL using previous Docker/native setup.

Apply migrations:

```bash
cd backend/services/auth-service
mysql -h 127.0.0.1 -P 3306 -u root -p < migrations/001_create_auth_tables.up.sql
mysql -h 127.0.0.1 -P 3306 -u root -p < migrations/002_add_token_claim_metadata.up.sql
```

If full migrations from Task 1 are already applied, skip this step.

Verify:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -p -e "DESCRIBE auth_db.refresh_tokens;"
mysql -h 127.0.0.1 -P 3306 -u root -p -e "DESCRIBE auth_db.auth_accounts;"
```

### Flow C: Prepare local JWT config

Follow the key setup already documented in:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 6. Redis / Queue / External Services
Subsection: OpenSSL / JWT Keys
```

Then verify Task 4 env values exist in `.env`:

```bash
cd backend/services/auth-service
set -a
source .env
set +a
```

Quick env sanity checks:

```bash
test -n "$JWT_KEY_ID"
test -n "$JWT_PRIVATE_KEY_PEM_PATH"
test -n "$REFRESH_TOKEN_PEPPER"
test -f "$JWT_PRIVATE_KEY_PEM_PATH"
```

### Flow D: Run current full Auth Service

Full server startup still needs MySQL, Redis, JWT keys, OTP peppers, and baseline env.

Refer:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 9. Local Development Setup
Section: 10. Running the Project
```

Task 4 specific quick run:

```bash
cd backend/services/auth-service
set -a
source .env
set +a
go run ./cmd/server
```

Verify in another terminal:

```bash
curl -i http://localhost:8081/healthz
curl -i http://localhost:8081/.well-known/jwks.json
```

### Flow E: Manual token issue test

This requires an existing account row because `refresh_tokens.account_id` references `auth_accounts.account_id`.

For local bootstrap user creation, refer:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 9. Local Development Setup
Subsection: Step 12: Optional local bootstrap user
```

After account and role exist, issue token pair through internal endpoint:

```bash
curl -i -X POST http://localhost:8081/internal/v1/auth/tokens/issue \
  -H "Content-Type: application/json" \
  -d '{"account_id":"acct_dev_001","user_id":"user_dev_001","session_id":"","roles":["buyer"],"seller_id":"","tenant_id":""}'
```

Expected:

- Response status `201`
- `tokens.access_token` present
- `tokens.refresh_token` present
- `tokens.expires_in` usually `900` when `JWT_ACCESS_TTL=15m`
- New row in `auth_db.refresh_tokens`
- DB row stores `token_hash`, not the plain refresh token

Refresh example using the returned refresh token:

```bash
curl -i -X POST http://localhost:8081/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"PASTE_LOCAL_REFRESH_TOKEN_HERE"}'
```

Do not paste real tokens into shared docs, tickets, commits, or logs.

---

## 10. Running the Project

### Normal run

Use the normal local run already documented:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 10. Running the Project
```

Task 4 related endpoints:

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/.well-known/jwks.json` | Public JWT verification keys |
| `POST` | `/internal/v1/auth/tokens/issue` | Internal token pair issue endpoint |
| `POST` | `/internal/v1/auth/tokens/verify` | Internal access token verification endpoint |
| `POST` | `/api/v1/auth/refresh` | Rotate refresh token and issue new token pair |
| `POST` | `/api/v1/auth/logout` | Revoke current/all refresh tokens |
| `POST` | `/api/v1/auth/login` | Login flow that returns Task 4 token pair |

### Ports and networking

No new port is introduced by Task 4.

| Service | Port | Purpose | New in Task 4? |
|---|---:|---|---:|
| Auth Service HTTP | `8081` by default | Login, refresh, JWKS, token APIs | No |
| MySQL | `3306` | `refresh_tokens` and account claim metadata | No |
| Redis | `6379` | Required by current full server startup for OTP rate limits | No |
| Notification Service | `8084` example | OTP delivery integration | No |
| Event publish HTTP endpoint | No fixed port | Optional outbox publishing | No |

Port conflict guidance already exists in:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 11. Ports & Networking
```

---

## 11. Common Errors & Fixes

General Go/MySQL/Redis/Docker/common startup errors already exist in:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 12. Common Errors & Fixes
```

Task 4 specific issues:

| Error/symptom | Likely reason | Fix |
|---|---|---|
| `JWT_KEY_ID cannot be empty` | `JWT_KEY_ID` missing | Set `JWT_KEY_ID=auth-local-1` in `.env` and source it |
| `JWT_PRIVATE_KEY_PEM_PATH cannot be empty` | Private key path missing | Generate keys and set `JWT_PRIVATE_KEY_PEM_PATH` |
| `read jwt private key` | File path wrong or command run from wrong directory | Use correct relative path from `backend/services/auth-service` or absolute path |
| `invalid rsa private key pem` | File is not RSA private key or is corrupted | Regenerate RSA private key with OpenSSL |
| `JWT_SIGNING_ALG must be RS256` | Env set to unsupported algorithm | Set `JWT_SIGNING_ALG=RS256` |
| `REFRESH_TOKEN_PEPPER cannot be empty` | Missing pepper secret | Set a long random `REFRESH_TOKEN_PEPPER` |
| Refresh returns `INVALID_REFRESH_TOKEN` right after config change | Pepper changed, token expired, token already rotated, or old token reused | Login/issue fresh token; keep pepper stable per environment |
| Refresh token row insert fails with foreign key error | `account_id` does not exist in `auth_accounts` | Seed/create account before issuing token pair |
| JWT verify fails in Gateway | `JWT_ISSUER`, `JWT_AUDIENCE`, `kid`, or JWKS key mismatch | Make Auth and Gateway config match exactly |
| JWKS endpoint returns key but Gateway still rejects token | Gateway cached old JWKS or expects different `kid` | Clear cache or rotate with compatible old+new key support |
| Migration `002` fails duplicate column/index | Migration already applied | Do not rerun; verify columns instead |
| Token issue returns role/subject validation error | Roles empty or account lacks claim data | Add active role and ensure `user_id` exists |
| Developer installs `github.com/golang-jwt/jwt/v5` but nothing changes | Current code does not import it | Leave `go.mod` unchanged unless implementation is refactored |

---

## 12. Security & Best Practices

### Task 4 security checklist

- Keep access token TTL short, for example `15m`.
- Store refresh tokens only as HMAC-SHA256 hashes.
- Keep `REFRESH_TOKEN_PEPPER` in secret manager or `.env` locally.
- Do not rotate `REFRESH_TOKEN_PEPPER` casually because old refresh tokens become unverifiable.
- Never log:
  - access token
  - refresh token
  - refresh token hash
  - Authorization header
  - JWT private key
  - refresh token pepper
- Use `RS256` only unless code explicitly supports another algorithm safely.
- Keep `JWT_PRIVATE_KEY_PEM_PATH` private and never commit private key files.
- Publish only public key material through JWKS.
- Keep `JWT_KEY_ID` stable for an active key.
- During key rotation, keep old public keys available until old access tokens expire.
- Validate `iss`, `aud`, `exp`, `nbf`, `token_type`, and `kid`.
- Treat `/internal/v1/auth/tokens/issue` and `/internal/v1/auth/tokens/verify` as internal-only endpoints.
- Use HTTPS in any environment with real user tokens.

### Beginner best practices

| Practice | Why it matters |
|---|---|
| Use separate local and production keys | Local keys should never sign production tokens |
| Keep one `.env` per environment | Mismatched issuer/audience/key id creates confusing invalid-token bugs |
| Use clear `kid` names | Example: `auth-local-1`, `auth-staging-2026-05`, `auth-prod-2026-05` |
| Verify JWKS after startup | Gateway depends on this endpoint for JWT validation |
| Test refresh twice | First refresh should work, old refresh token reuse should fail |
| Keep internal endpoints private | Token issue endpoint can mint tokens if exposed incorrectly |

---

## 13. Missing or Misconfigured Things

These findings are from inspecting Task 4 documentation and current Auth Service files.

| Finding | Impact | Suggested fix |
|---|---|---|
| `task4.md` mentions `github.com/golang-jwt/jwt/v5`, but current `go.mod` does not include it | Dependency docs and implementation can confuse beginners | Either update Task 4 docs to say standard library JWT implementation, or refactor code to use the library and add it to `go.mod` |
| No `.gitignore` detected | `.env` and `secrets/*.pem` can be accidentally committed | Add `.gitignore` with `.env`, `secrets/`, `*.pem`, build binaries |
| `.env` exists in working tree | Local secrets may leak if committed | Keep untracked and create sanitized `.env.example` |
| `secrets/` exists in working tree | Private JWT key may leak if committed | Keep untracked, add ignore rule, use secret manager in production |
| No Dockerfile found | Auth Service app cannot be containerized directly yet | Add `backend/services/auth-service/Dockerfile` before deployment |
| No docker-compose file found | Local dependency setup remains manual | Add `docker-compose.local.yml` for MySQL/Redis and optional notification/event mocks |
| No migration runner/tracking | Manual SQL order can be skipped or repeated | Add Makefile or migration tool such as Goose/golang-migrate |
| Single active key model in current config | Smooth production key rotation may need old+new public keys | Add multi-key JWKS/key ring support before complex rotation |
| Internal token endpoints are on same HTTP server | Risk if public routing exposes `/internal/*` | Protect with gateway rules, network policy, mTLS, or service auth |
| `/healthz` only returns process health | It does not prove MySQL/Redis/key readiness | Add readiness endpoint checking DB, Redis, and key load status |
| `JWT_PUBLIC_KEY_PEM_PATH` optional | If omitted, code derives public key from private key; okay locally but less explicit | Set public key path explicitly in shared environments |
| Broker integration not direct | Architecture may mention Kafka/RabbitMQ, but current code uses HTTP publisher | Keep docs clear: broker is behind event-ingress unless code changes |

---

## 14. References to Previous Dependency Files

Use these references instead of duplicating already written setup.

| Need | Refer |
|---|---|
| Full clone/install/onboarding flow | `TaskImplementation/Auth Service/task1_Dependency.md`, section `9. Local Development Setup` |
| Running the service | `TaskImplementation/Auth Service/task1_Dependency.md`, section `10. Running the Project` |
| Required software installation | `TaskImplementation/Auth Service/task1_Dependency.md`, section `3. Required Software` |
| Go modules, `go.mod`, `go.sum`, Go commands | `TaskImplementation/Auth Service/task1_Dependency.md`, section `4. Dependency Management` |
| MySQL installation and base DB setup | `TaskImplementation/Auth Service/task1_Dependency.md`, section `5. Database Setup` |
| Base schema migration `001` | `TaskImplementation/Auth Service/task2_Dependency.md`, section `5. Database Setup` |
| Refresh token base table explanation | `TaskImplementation/Auth Service/task2_Dependency.md`, section `5. Database Setup` |
| Password/hash dependency context | `TaskImplementation/Auth Service/task3_Dependency.md`, sections `4`, `7`, `11`, `12` |
| Redis setup | `TaskImplementation/Auth Service/task1_Dependency.md`, section `6. Redis / Queue / External Services` |
| OpenSSL/JWT key generation | `TaskImplementation/Auth Service/task1_Dependency.md`, section `6. Redis / Queue / External Services`, subsection `OpenSSL / JWT Keys` |
| Full `.env` | `TaskImplementation/Auth Service/task1_Dependency.md`, section `7. Environment Variables` |
| Docker dependency setup | `TaskImplementation/Auth Service/task1_Dependency.md`, section `8. Docker Setup` |
| Ports and networking | `TaskImplementation/Auth Service/task1_Dependency.md`, section `11. Ports & Networking` |
| General common errors | `TaskImplementation/Auth Service/task1_Dependency.md`, section `12. Common Errors & Fixes` |
| Baseline security checklist | `TaskImplementation/Auth Service/task1_Dependency.md`, section `13. Security & Best Practices` |

---

## 15. Final Checklist

Task 4 setup checklist:

- [ ] Previous Auth Service setup docs reviewed.
- [ ] Go version is compatible with `backend/services/auth-service/go.mod`.
- [ ] You understand current code does not require `github.com/golang-jwt/jwt/v5`.
- [ ] MySQL is running.
- [ ] Migration `001_create_auth_tables.up.sql` is applied.
- [ ] Migration `002_add_token_claim_metadata.up.sql` is applied.
- [ ] `auth_accounts.user_id` column exists.
- [ ] `auth_accounts.seller_id` column exists.
- [ ] `auth_accounts.tenant_id` column exists.
- [ ] `refresh_tokens` table exists.
- [ ] Local RSA private/public keys are generated or provided by secure secret storage.
- [ ] `JWT_KEY_ID` is set.
- [ ] `JWT_PRIVATE_KEY_PEM_PATH` points to an existing private key file.
- [ ] `JWT_SIGNING_ALG=RS256`.
- [ ] `REFRESH_TOKEN_PEPPER` is set to a long secret.
- [ ] Full local `.env` from `task1_Dependency.md` is present if running the server.
- [ ] `.env` is sourced before `go run`.
- [ ] Redis is running before full server startup.
- [ ] `go test ./internal/security/token ./internal/usecase` passes.
- [ ] Auth Service starts successfully.
- [ ] `GET /healthz` returns OK.
- [ ] `GET /.well-known/jwks.json` returns an RSA key with expected `kid`.
- [ ] Token issue endpoint works with a seeded local account.
- [ ] Refresh endpoint rotates token and rejects old refresh token reuse.
- [ ] `.env`, `secrets/`, and private PEM files are not committed.
