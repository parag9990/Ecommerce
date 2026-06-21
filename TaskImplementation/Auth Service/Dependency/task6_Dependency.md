# Project Dependency & Setup Guide

Input implementation file:

```text
TaskImplementation/Auth Service/task6.md
```

Generated dependency file:

```text
TaskImplementation/Auth Service/task6_Dependency.md
```

This guide is for **Auth Service - Task 6: RBAC Middleware**.

Simple goal: beginner developer ko clearly samajh aaye ki RBAC role checks chalane ke liye kaunsi dependency, database migration, env variable, token setup, local testing flow, and security precautions chahiye.

Important reuse rule followed:

- Previous Auth Service dependency files were inspected first.
- Common clone, Go install, MySQL install, Redis install, Docker setup, JWT keys, OTP setup, and full local run flow already documented hai.
- Is file me wahi content repeat nahi kiya gaya. Sirf Task 6 RBAC-specific setup and differences detail me explain kiye gaye hain.

Previous dependency files inspected:

```text
TaskImplementation/Auth Service/task1_Dependency.md
TaskImplementation/Auth Service/task2_Dependency.md
TaskImplementation/Auth Service/task3_Dependency.md
TaskImplementation/Auth Service/task4_Dependency.md
TaskImplementation/Auth Service/task5_Dependency.md
```

Current repository note:

- `task6.md` is a design/implementation guide.
- Current repo already contains RBAC code under `backend/services/auth-service/internal/authorization/`, `internal/domain/role.go`, `internal/usecase/role.go`, `internal/repository/mysql_role_repository.go`, `internal/authctx/`, and HTTP middleware.
- `task6.md` mentions future Gateway/gRPC examples and packages like `chi`, `grpc`, and `golang-jwt/jwt/v5`. Current Auth Service code does **not** import those packages.

---

## 1. Project Overview

Task 6 adds centralized RBAC setup.

RBAC ka full form hai **Role-Based Access Control**. Simple Hinglish me: user logged in hai ya nahi ye authentication hai, aur user ko specific route/action karne ka right hai ya nahi ye authorization/RBAC hai.

Task 6 runtime responsibility:

| Area | Current repo status | Why it matters |
|---|---|---|
| Role constants | Implemented in `internal/domain/role.go` | Typos like `super_admin` vs `superadmin` avoid hote hain |
| Auth level mapping | Implemented in `internal/authorization/rbac.go` | `buyer`, `seller`, `admin`, `superadmin` route rules central hote hain |
| Permission mapping | Implemented in `internal/authorization/permission.go` | Fine-grained actions like refund review or platform setting update control hote hain |
| JWT claims to auth context | Implemented in `internal/authctx/context.go` | Verified token claims request context me safely available hote hain |
| HTTP middleware | Implemented in `internal/transport/http/middleware.go` | Protected role APIs bearer token + role check karte hain |
| Role management usecase | Implemented in `internal/usecase/role.go` | Assign, revoke, get roles, and fresh role lookup rules |
| MySQL role repository | Implemented in `internal/repository/mysql_role_repository.go` | `role_assignments` table source of truth hai |
| Security tests | Present in `internal/authorization/rbac_security_test.go` and `internal/usecase/role_security_test.go` | Role bypass and escalation cases verify hote hain |

Task 6 does **not** introduce:

- New database server
- New Redis feature
- Kafka/RabbitMQ/NATS
- New Docker container
- New public port
- New third-party SaaS credentials
- New external Go package in current code

Baseline full service setup already exists in:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Sections:
3. Required Software
4. Dependency Management
5. Database Setup
6. Redis / Queue / External Services
7. Environment Variables
8. Docker Setup
9. Local Development Setup
10. Running the Project
```

---

## 2. Tech Stack

### Task 6 specific technologies

| Technology | What it is | Why Task 6 uses it | Required? | Beginner explanation |
|---|---|---|---:|---|
| Go | Backend programming language | RBAC middleware, role usecase, repository, and tests Go me hain | Yes | Go ek compiled backend language hai. Is Auth Service ka runtime Go code se chalta hai. |
| Go standard `net/http` | Built-in HTTP server/middleware package | `AuthRequired` and `RequireRoles` middleware expose karne ke liye | Yes | Current Auth Service Gin/Fiber/chi use nahi karta. Plain Go HTTP use ho raha hai. |
| Go `context` | Request-scoped data carrier | Verified JWT claims ko request/usecase context me pass karne ke liye | Yes | Context ek safe request bag jaisa hai jisme auth claims pass hote hain. |
| `internal/authctx` | Local Auth Service auth context helper | Claims normalize, context me store, metadata convert karne ke liye | Yes | Ye local package user id, session id, seller id, tenant id, roles ko typed format me rakhta hai. |
| `internal/authorization` | Local RBAC helper package | Auth levels, permissions, seller scope, role checks central karne ke liye | Yes | Ye package decide karta hai kaun se roles allowed hain. |
| JWT access tokens | Signed token with roles claim | Middleware token verify karke roles context me load karta hai | Yes for HTTP RBAC | JWT me `roles`, `seller_id`, `session_id` jaise claims aate hain. |
| MySQL `role_assignments` | Role source-of-truth table | Active roles store, assign, revoke, fresh lookup karne ke liye | Yes for real role flow | DB me roles durable way me store hote hain. |
| MySQL transaction + row lock | DB consistency feature | Duplicate role assignment and race condition avoid karne ke liye | Yes | Assign/revoke ke time account row lock hoti hai. |
| Generated MySQL columns | MySQL computed columns | Active role uniqueness enforce karne ke liye | Yes after migration `004` | MySQL `NULL` unique behavior tricky hota hai, generated columns usko fix karte hain. |
| Go `log/slog` | Structured logging | Role assign/revoke audit-style logs ke liye | Yes | Logs JSON/text fields me actor, role, target details record karte hain. |

### Current code vs `task6.md` design packages

`task6.md` includes future-oriented examples. Current repository implementation differs like this:

| Item mentioned in `task6.md` | Current repository status | What beginner should do |
|---|---|---|
| `github.com/go-chi/chi/v5` | Not imported | Do not install unless Gateway/router code later imports it |
| `google.golang.org/grpc` | Not imported in Auth Service | Do not install for current Task 6 tests |
| `google.golang.org/grpc/metadata` | Not imported | Current repo has generic metadata helpers in `internal/authctx/metadata.go` |
| `github.com/golang-jwt/jwt/v5` | Not imported | Current JWT code uses Go standard library implementation from Task 4 |
| API Gateway service code | Not present under `backend/services/` | Current executable is Auth Service only; Gateway RBAC is documented as architecture/future integration |
| `github.com/go-sql-driver/mysql` | Imported | Required for MySQL role repository |

Do **not** run `go get` for unused packages only because the markdown design guide names them. Extra packages create unnecessary `go.mod` and `go.sum` churn.

### Already documented baseline technologies

| Baseline topic | Reuse documentation |
|---|---|
| Go `1.26.3`, Go modules, Go workspace | `TaskImplementation/Auth Service/task1_Dependency.md`, section `4. Dependency Management` |
| MySQL 8+ installation, Docker, DSN | `TaskImplementation/Auth Service/task1_Dependency.md`, section `5. Database Setup` |
| Base auth schema and `role_assignments` table | `TaskImplementation/Auth Service/task2_Dependency.md`, section `5. Database Setup` |
| JWT issuing, roles claim, JWKS | `TaskImplementation/Auth Service/task4_Dependency.md`, sections `2`, `6`, `7` |
| Redis setup for full server startup | `TaskImplementation/Auth Service/task1_Dependency.md`, section `6. Redis / Queue / External Services` |
| Full local `.env` | `TaskImplementation/Auth Service/task1_Dependency.md`, section `7. Environment Variables` |
| Docker dependency containers | `TaskImplementation/Auth Service/task1_Dependency.md`, section `8. Docker Setup` |

---

## 3. Required Software

### For RBAC unit/security tests only

| Software | Required? | Purpose | Verify command |
|---|---:|---|---|
| Go | Yes | Compile and run RBAC tests | `go version` |
| Internet access or existing Go module cache | Usually yes | Download Go dependencies if not cached | `go mod download` |

RBAC focused tests do **not** need MySQL, Redis, Docker, JWT key files, or Notification Service.

### For full RBAC HTTP flow

| Software/service | Required? | Why | Setup reference |
|---|---:|---|---|
| Go | Yes | Run Auth Service | `task1_Dependency.md`, section `3` |
| MySQL 8+ | Yes | `role_assignments`, `auth_accounts`, `refresh_tokens` | `task1_Dependency.md`, section `5` |
| Redis | Required for current full server startup | RBAC itself does not use Redis, but `main.go` pings Redis for OTP rate limiting | `task1_Dependency.md`, section `6` |
| JWT RSA keys | Yes for full server startup | `AuthRequired` verifies bearer access token before RBAC | `task4_Dependency.md`, section `7` |
| curl | Recommended | Call health, JWKS, token, and role endpoints | `task1_Dependency.md`, section `3` |
| Docker | Optional but recommended | Run MySQL/Redis locally | `task1_Dependency.md`, section `8` |

Task 6 readiness, simple version:

- Go dependencies installed.
- MySQL running.
- Redis running for full server startup.
- Migrations `001` to `005` applied for current full service, especially migration `004` for RBAC hardening.
- JWT key env variables configured.
- `AUTH_ROLE_REASON_MAX_LENGTH` valid.
- Bearer token used for protected role endpoints.

---

## 4. Dependency Management

### Language-specific dependency system: Go modules

Full beginner explanation of:

- `go.mod`
- `go.sum`
- Go modules
- `go mod download`
- `go mod tidy`
- `go build`
- `go run`
- common Go module issues

already exists in:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 4. Dependency Management
```

### Task 6 current dependency impact

Current `backend/services/auth-service/go.mod` direct dependencies are:

```text
github.com/go-sql-driver/mysql v1.9.3
golang.org/x/crypto v0.51.0
```

Task 6 RBAC implementation itself uses mostly local packages and Go standard library. It does **not** add a new external dependency in the current repository.

Task 6 important package areas:

| Package/file | Why used |
|---|---|
| `internal/domain/role.go` | Role constants, seller/admin role classification |
| `internal/authorization/rbac.go` | Auth levels and role gates |
| `internal/authorization/permission.go` | Role to permission mapping |
| `internal/authctx/context.go` | JWT claims to request context |
| `internal/authctx/metadata.go` | Future internal metadata propagation |
| `internal/usecase/role.go` | Assign/revoke/get/fresh role logic |
| `internal/repository/mysql_role_repository.go` | MySQL role persistence |
| `internal/transport/http/middleware.go` | HTTP bearer token + RBAC middleware |

### Commands for Task 6 verification

Run RBAC focused tests:

```bash
cd backend/services/auth-service
go test ./internal/authorization ./internal/usecase
```

Run HTTP middleware/security tests too:

```bash
cd backend/services/auth-service
go test ./internal/authorization ./internal/usecase ./internal/transport/http
```

Run all Auth Service tests:

```bash
cd backend/services/auth-service
go test ./...
```

If Go cache is not writable in your environment:

```bash
cd backend/services/auth-service
GOCACHE=/tmp/auth-go-cache go test ./...
```

### Common dependency confusion

| Symptom | Likely reason | Fix |
|---|---|---|
| Developer expects `chi` in `go.mod` | `task6.md` design examples mention chi, current code uses `net/http` | Do not add `chi` unless code imports it |
| Developer expects `grpc` dependency | Gateway-to-service gRPC is architecture guidance, not current Auth Service import | Do not add gRPC unless actual code is added |
| Developer expects `golang-jwt/jwt/v5` | Task 4/6 docs mention JWT libraries, current implementation uses standard library | Do not add it unless JWT code is refactored |
| `go: go.mod requires go >= 1.26.3` | Local Go version old hai | Install compatible Go version |
| `package not in std` | Command wrong folder se run hua | Run from `backend/services/auth-service` |

---

## 5. Database Setup

### A. Database detected

Task 6 uses the existing Auth Service MySQL database:

```text
auth_db
```

Main RBAC table:

```text
role_assignments
```

Migration files involved:

```text
backend/services/auth-service/migrations/001_create_auth_tables.up.sql
backend/services/auth-service/migrations/004_harden_role_assignments.up.sql
```

### B. What MySQL is

MySQL ek relational database hai. Data tables, rows, columns me store hota hai. RBAC me roles sensitive security data hain, isliye durable DB storage and transaction safety important hai.

Detailed MySQL install, Docker setup, start commands, verify commands, default port, DSN format, and credential placement already exist in:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 5. Database Setup

TaskImplementation/Auth Service/task2_Dependency.md
Section: 5. Database Setup
```

### C. Why Task 6 uses MySQL

Task 6 MySQL ko role source-of-truth ke liye use karta hai:

| DB object/field | Why needed |
|---|---|
| `role_assignments.account_id` | Role valid auth account se linked rahe |
| `role_assignments.role` | User ke active roles store karne ke liye |
| `scope_type` and `scope_id` | Seller scoped roles ko specific seller account se bind karne ke liye |
| `assigned_by` | Role kis actor ne assign kiya track karne ke liye |
| `reason` | Role change ka audit reason store karne ke liye |
| `assigned_at` | Assignment time |
| `revoked_at` | Soft revoke; history preserve hoti hai |
| `uk_role_assignments_active` | Same active role duplicate na ho |
| `idx_role_assignments_scope` | Seller/admin scoped lookup fast ho |

Hinglish: Role data sirf JWT me rakhna enough nahi hai. JWT short-lived cache jaisa hai, but actual source of truth MySQL `role_assignments` table hai.

### D. Required or optional

| Scenario | MySQL needed? | Notes |
|---|---:|---|
| `go test ./internal/authorization` | No | Pure role/auth helper tests |
| `go test ./internal/usecase` role tests | No | Security role tests use fake repository |
| Full Auth Service startup | Yes | `main.go` pings MySQL |
| Real `/internal/v1/auth/roles` | Yes | Reads `role_assignments` |
| Real `/internal/v1/auth/roles/assign` | Yes | Inserts into `role_assignments` |
| Real `/internal/v1/auth/roles/revoke` | Yes | Updates `revoked_at` |
| Login/refresh token role claims | Yes | Token usecase loads active roles from DB |

### E. Local installation

Do not duplicate MySQL install steps here.

Refer:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 5. Database Setup
Subsections: D. Local installation, E. Docker setup, F. Start commands
```

### F. Docker setup

Task 6 does not change MySQL Docker settings.

Use the same MySQL container setup from:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 8. Docker Setup
```

Same important settings:

| Setting | Value |
|---|---|
| MySQL port | `3306` |
| Database | `auth_db` |
| Suggested user | `auth_user` |
| Persistent volume | `ecommerce-auth-mysql-data` |

### G. Apply Task 6 related migration

For current full Auth Service, easiest is to apply all migrations in order as described in `task1_Dependency.md`.

Task 6 specifically needs migration `004`:

```text
backend/services/auth-service/migrations/004_harden_role_assignments.up.sql
```

This migration adds:

| Added item | Why |
|---|---|
| `reason` column | Role mutation audit reason |
| `active_role_key` generated column | Active-only uniqueness |
| `scope_type_key` generated column | Normalize nullable scope type for unique key |
| `scope_id_key` generated column | Normalize nullable scope id for unique key |
| `uk_role_assignments_active` | Prevent duplicate active role assignment |
| `idx_role_assignments_scope` | Fast scope-based role lookup |

Run from repo root if earlier migrations are already applied:

```bash
mysql -u root -p < backend/services/auth-service/migrations/004_harden_role_assignments.up.sql
```

If you are setting up the current full service from scratch, apply all migrations in order from `backend/services/auth-service`:

```bash
cd backend/services/auth-service
mysql -h 127.0.0.1 -P 3306 -u root -p < migrations/001_create_auth_tables.up.sql
mysql -h 127.0.0.1 -P 3306 -u root -p < migrations/002_add_token_claim_metadata.up.sql
mysql -h 127.0.0.1 -P 3306 -u root -p < migrations/003_add_otp_account_fk.up.sql
mysql -h 127.0.0.1 -P 3306 -u root -p < migrations/004_harden_role_assignments.up.sql
mysql -h 127.0.0.1 -P 3306 -u root -p < migrations/005_create_auth_outbox_events.up.sql
```

Warning: Do not rerun migration `004` blindly on the same database. If columns/indexes already exist, MySQL will throw duplicate column/index errors.

### H. Verify Task 6 schema

```bash
mysql -u root -p auth_db
```

Inside MySQL:

```sql
SHOW TABLES LIKE 'role_assignments';
SHOW COLUMNS FROM role_assignments;
SHOW INDEX FROM role_assignments;

SELECT CONSTRAINT_NAME, TABLE_NAME, REFERENCED_TABLE_NAME
FROM information_schema.KEY_COLUMN_USAGE
WHERE TABLE_SCHEMA = 'auth_db'
  AND TABLE_NAME = 'role_assignments'
  AND REFERENCED_TABLE_NAME IS NOT NULL;
```

Expected important columns after migration `004`:

| Column | Expected |
|---|---|
| `reason` | Present |
| `active_role_key` | Present, generated |
| `scope_type_key` | Present, generated |
| `scope_id_key` | Present, generated |

Expected important indexes:

| Index | Purpose |
|---|---|
| `idx_role_assignments_account` | Find roles by account |
| `idx_role_assignments_role` | Find assignments by role |
| `uk_role_assignments_active` | Prevent duplicate active assignments |
| `idx_role_assignments_scope` | Find roles by seller/admin scope |

Expected foreign key:

```text
fk_role_assignments_account
```

### I. Rollback Task 6 migration

Only rollback if you intentionally want to undo RBAC hardening:

```bash
mysql -u root -p < backend/services/auth-service/migrations/004_harden_role_assignments.down.sql
```

Warning: rollback removes RBAC hardening columns/indexes from `role_assignments`. Duplicate active roles may become possible again because MySQL treats `NULL` values in unique indexes as distinct.

### J. Connection string and credentials

Connection string format is unchanged:

```text
AUTH_MYSQL_DSN='auth_user:auth_password@tcp(127.0.0.1:3306)/auth_db?parseTime=true&charset=utf8mb4&loc=UTC'
```

Credential placement is already explained in:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 5. Database Setup
Subsection: J. Where to place credentials

TaskImplementation/Auth Service/task1_Dependency.md
Section: 7. Environment Variables
```

Important for RBAC:

- Role assignments are stored in MySQL, not `.env`.
- DB password goes inside `AUTH_MYSQL_DSN`, not Go code.
- Keep `parseTime=true`; role code scans MySQL timestamps into Go `time.Time`.
- Use least-privilege DB user in production.

---

## 6. Redis / Queue / External Services

### Redis

Task 6 RBAC logic does **not** use Redis directly.

However, current full Auth Service startup still requires Redis because `main.go` initializes and pings Redis for OTP rate limiting. Agar Redis down hai, full server start nahi hoga even if you only want to test RBAC HTTP endpoints.

Reuse Redis setup from:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 6. Redis / Queue / External Services
Subsection: Redis
```

### JWT / JWKS

RBAC HTTP middleware depends on valid access tokens.

Flow:

```text
Authorization: Bearer <access_token>
  -> AuthRequired verifies JWT
  -> authctx stores user_id/session_id/roles/seller_id/tenant_id
  -> RequireRoles or RequireAuthLevel checks roles
```

JWT/JWKS setup is already documented in:

```text
TaskImplementation/Auth Service/task4_Dependency.md
Sections:
6. Redis / Queue / External Services - JWKS endpoint
7. Environment Variables
```

Task 6 specific JWT note:

- Access token must contain at least one role.
- Seller-scoped actions need `seller_id` claim.
- Gateway/services must validate same `JWT_ISSUER`, `JWT_AUDIENCE`, `JWT_SIGNING_ALG`, and `kid`.
- Revoked roles may remain inside already-issued JWTs until access token expiry. Use fresh DB role lookup for high-risk actions.

### API Gateway / gRPC status

`task6.md` describes Gateway + service-layer RBAC architecture. Current repo does not currently include a real `backend/services/api-gateway` implementation.

Current Auth Service contains reusable metadata helpers:

```text
backend/services/auth-service/internal/authctx/metadata.go
```

Metadata keys:

| Key | Value |
|---|---|
| `x-user-id` | Verified user id |
| `x-session-id` | Verified session id |
| `x-roles` | Comma-separated roles |
| `x-seller-id` | Seller context if present |
| `x-tenant-id` | Tenant context if present |

If a Gateway service is added later, it should:

- Verify JWT using Auth JWKS.
- Apply route-level RBAC from `api/master-api.json`.
- Forward only safe verified claims, not the full JWT.
- Keep internal service metadata trusted only on private network.

### Kafka/RabbitMQ/NATS status

Task 6 does not introduce Kafka, RabbitMQ, or NATS.

Session outbox/event setup is separate and already covered in:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 6. Redis / Queue / External Services
Subsection: Session Link / Outbox

TaskImplementation/Auth Service/task4_Dependency.md
Section: 6. Redis / Queue / External Services
Subsection: Session link / outbox
```

### Notification Service status

Task 6 does not use Notification Service directly.

Full server startup still validates Notification config because OTP usecase is initialized. Reuse:

```text
TaskImplementation/Auth Service/task5_Dependency.md
Section: 6. Redis / Queue / External Services
Subsection: Notification Service
```

---

## 7. Environment Variables

Full local `.env` already exists in:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 7. Environment Variables
```

Task 6 introduces only one RBAC-specific env variable in current code.

### Task 6 `.env` snippet

```bash
# RBAC
AUTH_ROLE_REASON_MAX_LENGTH=512
```

### Task 6 env variable explanation

| Variable | Required? | Default/current behavior | Why Task 6 needs it |
|---|---:|---|---|
| `AUTH_ROLE_REASON_MAX_LENGTH` | Optional | `512` | Role assign/revoke reason max length. Must be greater than zero. |

### Existing variables RBAC depends on

Do not duplicate the full `.env` here. Verify these existing variables from earlier docs when running the full service:

| Variable | Why RBAC flow needs it | Already documented in |
|---|---|---|
| `AUTH_HTTP_ADDR` | Auth Service HTTP port | `task1_Dependency.md`, section `7` |
| `AUTH_MYSQL_DSN` | Role repository connects to MySQL | `task1_Dependency.md`, section `7` |
| `AUTH_REDIS_ADDR` | Full server startup pings Redis | `task1_Dependency.md`, section `7` |
| `JWT_ISSUER` | Access token verification | `task4_Dependency.md`, section `7` |
| `JWT_AUDIENCE` | Access token verification | `task4_Dependency.md`, section `7` |
| `JWT_ACCESS_TTL` | Limits stale-role window | `task4_Dependency.md`, section `7` |
| `JWT_SIGNING_ALG` | Must be `RS256` | `task4_Dependency.md`, section `7` |
| `JWT_KEY_ID` | JWT `kid` / JWKS lookup | `task4_Dependency.md`, section `7` |
| `JWT_PRIVATE_KEY_PEM_PATH` | Token signing and verifier setup | `task4_Dependency.md`, section `7` |
| `JWT_PUBLIC_KEY_PEM_PATH` | JWKS explicit public key | `task4_Dependency.md`, section `7` |
| `REFRESH_TOKEN_PEPPER` | Full token usecase startup | `task4_Dependency.md`, section `7` |
| `OTP_HASH_PEPPER` | Full server startup initializes OTP hasher | `task5_Dependency.md`, section `7` |
| `OTP_RATE_LIMIT_PEPPER` | Full server startup initializes OTP rate repo | `task5_Dependency.md`, section `7` |
| `SESSION_LINK_MODE` | Can be `disabled` for simple local runs | `task1_Dependency.md`, section `7` |
| `SESSION_EVENT_PEPPER` | Required when `SESSION_LINK_MODE=outbox` | `task1_Dependency.md`, section `7` |

### Credentials placement

Local development:

```bash
cd backend/services/auth-service
set -a
source .env
set +a
```

Production:

| Config/secret | Recommended placement |
|---|---|
| `AUTH_MYSQL_DSN` | Secret manager or Kubernetes Secret |
| `JWT_PRIVATE_KEY_PEM_PATH` file content | Mounted secret file |
| `REFRESH_TOKEN_PEPPER` | Secret manager |
| `OTP_HASH_PEPPER` | Secret manager |
| `OTP_RATE_LIMIT_PEPPER` | Secret manager |
| `SESSION_EVENT_PEPPER` | Secret manager |
| `AUTH_ROLE_REASON_MAX_LENGTH` | Normal config/env var, not a secret |

Warning: Current working tree contains local `.env` and `secrets/` files. Do not commit real env files or JWT private keys.

---

## 8. Docker Setup

Task 6 does not add a Dockerfile or docker-compose file.

Current repo status:

| File | Status |
|---|---|
| `backend/services/auth-service/Dockerfile` | Not present |
| root `docker-compose.yml` | Not present |
| MySQL/Redis Docker examples | Already documented in previous dependency files |

Refer:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 8. Docker Setup
Subsection: Docker Compose example for dependencies
```

Task 6 uses the same dependency containers:

| Container/service | Port | Why |
|---|---:|---|
| MySQL | `3306` | Stores `role_assignments` |
| Redis | `6379` | Required by current full server startup |
| Auth Service | `8081` default | Exposes role HTTP endpoints |

If running Auth Service itself in Docker later, make sure:

- `AUTH_MYSQL_DSN` uses Docker network host `mysql`, not `127.0.0.1`.
- `AUTH_REDIS_ADDR` uses `redis:6379`.
- JWT key files are mounted into the container.
- Secrets are injected at runtime, not baked into image.
- Internal routes like `/internal/v1/auth/tokens/issue` and `/internal/v1/auth/roles/*` are not exposed publicly.

---

## 9. Local Development Setup

### Flow A: Run RBAC tests only

This is the fastest Task 6 check and does not need MySQL, Redis, Docker, JWT keys, or Notification Service.

```bash
cd backend/services/auth-service
go test ./internal/authorization ./internal/usecase
```

What this verifies:

- Buyer cannot access seller/admin/superadmin auth levels.
- Seller cannot access admin auth level.
- Admin cannot access superadmin auth level.
- Superadmin can access superadmin auth level.
- Seller cannot access another seller scope.
- Seller order manager cannot write catalog.
- Finance admin can review refunds.
- Buyer cannot escalate to admin.
- Seller manager cannot assign seller role for a different seller.
- Fresh role lookup ignores revoked roles.

### Flow B: Prepare DB for real RBAC HTTP flow

Follow baseline MySQL setup first:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 9. Local Development Setup
Steps 4 and 5
```

Then verify migration `004` is applied:

```bash
mysql -u root -p auth_db
```

Inside MySQL:

```sql
SHOW COLUMNS FROM role_assignments LIKE 'reason';
SHOW INDEX FROM role_assignments WHERE Key_name = 'uk_role_assignments_active';
```

### Flow C: Prepare Redis

RBAC does not use Redis, but full server startup does.

Refer:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 6. Redis / Queue / External Services
Section: 8. Docker Setup
```

Verify:

```bash
redis-cli PING
```

Expected:

```text
PONG
```

### Flow D: Prepare JWT config

RBAC protected endpoints need bearer access tokens.

Refer:

```text
TaskImplementation/Auth Service/task4_Dependency.md
Section: 7. Environment Variables
Section: 9. Local Development Setup
```

Minimum checks:

```bash
cd backend/services/auth-service
test -n "$JWT_KEY_ID"
test -n "$JWT_PRIVATE_KEY_PEM_PATH"
test -f "$JWT_PRIVATE_KEY_PEM_PATH"
```

### Flow E: Run current full Auth Service

Use full baseline instructions:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 9. Local Development Setup
Section: 10. Running the Project
```

Quick version:

```bash
cd backend/services/auth-service
set -a
source .env
set +a
go run ./cmd/server
```

Verify:

```bash
curl http://localhost:8081/healthz
curl http://localhost:8081/.well-known/jwks.json
```

Expected health:

```json
{"status":"ok"}
```

### Flow F: Manual RBAC endpoint test

Important: These are internal endpoints. Use only on local/dev network.

1. Issue a local token with privileged role for testing:

```bash
curl -s -X POST http://localhost:8081/internal/v1/auth/tokens/issue \
  -H 'Content-Type: application/json' \
  -d '{"account_id":"acct_admin_001","user_id":"user_admin_001","session_id":"","roles":["superadmin"],"seller_id":"","tenant_id":""}'
```

Copy `tokens.access_token` from the response.

2. Assign a buyer role to a target user:

```bash
curl -i -X POST http://localhost:8081/internal/v1/auth/roles/assign \
  -H "Authorization: Bearer PASTE_ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -H 'X-Request-ID: req_local_rbac_001' \
  -d '{"user_id":"user_target_001","role":"buyer","reason":"local RBAC setup test"}'
```

Expected success:

```json
{
  "success": true,
  "role": {
    "role": "buyer",
    "assigned_at": "..."
  }
}
```

3. Assign a seller-scoped role:

```bash
curl -i -X POST http://localhost:8081/internal/v1/auth/roles/assign \
  -H "Authorization: Bearer PASTE_ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -H 'X-Request-ID: req_local_rbac_002' \
  -d '{"user_id":"user_target_001","role":"seller_catalog_editor","scope_type":"seller","scope_id":"seller_001","reason":"local seller catalog access test"}'
```

4. Read roles:

```bash
curl -i "http://localhost:8081/internal/v1/auth/roles?user_id=user_target_001" \
  -H "Authorization: Bearer PASTE_ACCESS_TOKEN"
```

5. Revoke role:

```bash
curl -i -X POST http://localhost:8081/internal/v1/auth/roles/revoke \
  -H "Authorization: Bearer PASTE_ACCESS_TOKEN" \
  -H 'Content-Type: application/json' \
  -H 'X-Request-ID: req_local_rbac_003' \
  -d '{"user_id":"user_target_001","role":"seller_catalog_editor","scope_type":"seller","scope_id":"seller_001","reason":"local cleanup"}'
```

Note: `/internal/v1/auth/tokens/issue` accepts role input directly and is not protected by `AuthRequired` in current code. Never expose it to public internet.

---

## 10. Running the Project

### Normal run

Use the existing full run guide:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 10. Running the Project
```

Quick command:

```bash
cd backend/services/auth-service
set -a
source .env
set +a
go run ./cmd/server
```

### Task 6 endpoints

| Method | Endpoint | Purpose | Auth |
|---|---|---|---|
| `GET` | `/internal/v1/auth/roles` | Get current or target user roles | Bearer token required |
| `POST` | `/internal/v1/auth/roles/assign` | Assign role | Bearer token + privileged role required |
| `POST` | `/internal/v1/auth/roles/revoke` | Revoke role | Bearer token + privileged role required |
| `GET` | `/.well-known/jwks.json` | Public JWT verification keys | Public |
| `POST` | `/internal/v1/auth/tokens/issue` | Issue token pair for internal/local flow | Internal only; not role-protected currently |
| `POST` | `/internal/v1/auth/tokens/verify` | Verify access token | Internal only |

### Ports and networking

Task 6 introduces no new port.

| Service | Port | Purpose | Task 6 status |
|---|---:|---|---|
| Auth Service HTTP | `8081` default | Role endpoints and JWKS | Reused |
| MySQL | `3306` | `role_assignments` table | Reused |
| Redis | `6379` | Required by full server startup | Reused, not RBAC-specific |
| Notification Service | `8084` example | Required for OTP send runtime | Reused, not RBAC-specific |
| API Gateway | Not present in current repo | Future route-level RBAC | Architecture-only for now |

Port conflict guidance already exists in:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 11. Ports & Networking
```

---

## 11. Common Errors & Fixes

General Go/MySQL/Redis/Docker/JWT startup errors already exist in:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 12. Common Errors & Fixes

TaskImplementation/Auth Service/task4_Dependency.md
Section: 11. Common Errors & Fixes
```

Only Task 6-specific issues are listed here.

| Error/symptom | Likely reason | Fix |
|---|---|---|
| `invalid rbac config: AUTH_ROLE_REASON_MAX_LENGTH must be greater than zero` | Env set to `0` or negative value | Set `AUTH_ROLE_REASON_MAX_LENGTH=512` or remove override |
| `/internal/v1/auth/roles` returns `401 AUTHENTICATION_REQUIRED` | Missing `Authorization: Bearer ...` header | Issue/login token and pass bearer token |
| `/internal/v1/auth/roles/assign` returns `403 PERMISSION_DENIED` | Actor token lacks allowed role or tries privilege escalation | Use valid admin/superadmin/seller-manager actor depending on role |
| Role assign returns `VALIDATION_ERROR: role change reason is required` | `reason` missing or blank | Add a clear `reason` field |
| Seller scoped role returns invalid scope error | Missing `scope_type=seller` or `scope_id` | For seller staff roles, pass both fields |
| Global role returns invalid scope error | Scope sent for non-seller role | Remove `scope_type` and `scope_id` for buyer/admin roles |
| `ROLE_ALREADY_ASSIGNED` | Same active role already exists | Do not reassign; query roles first or revoke old role |
| `NOT_FOUND` | Target `user_id`/`account_id` not found in `auth_accounts` | Seed/create auth account first |
| `Unknown column 'reason' in 'field list'` | Migration `004` not applied | Apply `004_harden_role_assignments.up.sql` |
| Duplicate active global role appears | Migration `004` missing, old nullable unique key behavior | Apply migration `004`; clean duplicate test data manually |
| Role revoked but token still allows route | Existing JWT still carries old role until expiry | Use short `JWT_ACCESS_TTL`; use fresh role lookup for high-risk actions |
| `go get chi/grpc/jwt` confusion | Design doc mentions future packages | Do not install unless current code imports them |

---

## 12. Security & Best Practices

### Task 6 security checklist

- Keep Auth Service `role_assignments` as source of truth.
- Keep role constants centralized in `internal/domain/role.go`.
- Do not hardcode role strings in random files.
- Do not rely only on frontend role guards.
- Do not rely only on Gateway route checks for high-risk actions.
- Service/usecase layer must still check seller scope and permissions.
- Keep JWT access token TTL short, for example `15m`.
- Use fresh DB role lookup for admin role changes, refund review, and platform settings.
- Always require a role change reason.
- Audit every role assign/revoke action.
- Never log JWT access token, refresh token, password, OTP, or raw secret values.
- Keep internal endpoints private behind network/service mesh/API Gateway rules.

### Beginner best practices

| Practice | Why |
|---|---|
| Test role bypass cases | Auth bugs are high-risk; tests catch escalation mistakes |
| Use `superadmin` only for local bootstrap/highest privilege | Overusing superadmin hides missing role rules |
| Use seller scoped roles with `scope_type=seller` and exact `scope_id` | Prevent one seller from accessing another seller's resources |
| Use `X-Request-ID` while testing | Role logs become easier to trace |
| Query roles before assign/revoke | Avoid duplicate assignment or not-found confusion |
| Keep `.env` out of Git | It contains DB, JWT, and pepper secrets |

### Security audit from current implementation

| Finding | Risk | Suggested fix |
|---|---|---|
| `/internal/v1/auth/tokens/issue` is not protected by `AuthRequired` | If publicly exposed, attacker could request tokens with arbitrary roles | Keep route internal only; add service-to-service auth or remove from public router before production |
| `/internal/v1/auth/credentials` and password reset internal endpoints are also internal-style routes | Public exposure can become account compromise risk | Protect internal routes at network and middleware level |
| Role assign/revoke logs are structured but no immutable audit table exists yet | Logs can be lost or rotated; compliance trail weak | Add dedicated audit table/outbox event for role changes |
| Assign role persists `reason`, but revoke does not persist `revoked_by` or revoke reason in `role_assignments` | Revoke audit detail incomplete | Add `revoked_by`, `revoke_reason`, or audit log table |
| Gateway service code is not present | Architecture says Gateway RBAC, but current repo only enforces Auth Service role endpoints | Implement Gateway RBAC before exposing full platform routes |
| Stale JWT roles after revoke | User may keep revoked role until token expiry | Short access token TTL plus fresh lookup for high-risk actions |
| Dockerfile, Compose, and Kubernetes base are present | Repeatable deployment is available | Keep image, Compose, and Kustomize validation in CI |
| `/healthz` reports liveness and `/readyz` checks MySQL/Redis | Probes now distinguish process health from dependency readiness | Keep readiness dependency checks fail-closed and free of secret details |
| JWT secret paths are ignored and runtime-mounted | Private key material stays outside source | Use a secret manager and rotation policy in production |

---

## 13. Missing or Misconfigured Things

| Missing/misconfigured item | Why it matters | Recommendation |
|---|---|---|
| API Gateway exists but its downstream bridge remains incomplete | Task 6 architecture expects Gateway + service checks | Complete Gateway forwarding and retain Auth Service authorization checks |
| gRPC interceptors not implemented in current Auth Service | Task 6 docs mention gRPC metadata propagation | Implement only when gRPC services are added |
| Internal endpoint protection incomplete | Internal routes can be dangerous if exposed | Add internal auth middleware, mTLS, private network, or API Gateway restrictions |
| Immutable role audit storage missing | Role changes need compliance-grade history | Add `auth_role_audit_logs` or shared audit service event |
| Sanitized `.env.example` is committed | Local configuration has a safe template | Keep real secret values out of source control |
| App Dockerfile is present | Service image deployment is standardized | Keep the multi-stage build current with the Go module |
| Root Compose stack is present | Local MySQL/Redis and service startup are automated | Keep dependency health ordering current |
| Compose migration job is present | SQL migrations run before Auth Service startup | Preserve ordered, idempotent migration execution |

---

## 14. References to Previous Dependency Files

| Topic | Reuse this documentation |
|---|---|
| Full clone/install/onboarding flow | `TaskImplementation/Auth Service/task1_Dependency.md`, section `9. Local Development Setup` |
| Running the service | `TaskImplementation/Auth Service/task1_Dependency.md`, section `10. Running the Project` |
| Required software installation | `TaskImplementation/Auth Service/task1_Dependency.md`, section `3. Required Software` |
| Go modules, `go.mod`, `go.sum`, Go commands | `TaskImplementation/Auth Service/task1_Dependency.md`, section `4. Dependency Management` |
| MySQL installation and base DB setup | `TaskImplementation/Auth Service/task1_Dependency.md`, section `5. Database Setup` |
| Base `role_assignments` schema | `TaskImplementation/Auth Service/task2_Dependency.md`, section `5. Database Setup` |
| MySQL nullable unique issue on role assignments | `TaskImplementation/Auth Service/task2_Dependency.md`, sections `11` and `12` |
| Password/hash dependency context | `TaskImplementation/Auth Service/task3_Dependency.md`, sections `4`, `7`, `11`, `12` |
| JWT roles claim and JWKS setup | `TaskImplementation/Auth Service/task4_Dependency.md`, sections `2`, `6`, `7`, `9` |
| Redis setup | `TaskImplementation/Auth Service/task1_Dependency.md`, section `6. Redis / Queue / External Services` |
| Notification Service setup | `TaskImplementation/Auth Service/task5_Dependency.md`, section `6. Redis / Queue / External Services` |
| Full `.env` | `TaskImplementation/Auth Service/task1_Dependency.md`, section `7. Environment Variables` |
| Docker dependency setup | `TaskImplementation/Auth Service/task1_Dependency.md`, section `8. Docker Setup` |
| Ports and networking | `TaskImplementation/Auth Service/task1_Dependency.md`, section `11. Ports & Networking` |
| General common errors | `TaskImplementation/Auth Service/task1_Dependency.md`, section `12. Common Errors & Fixes` |
| Baseline security checklist | `TaskImplementation/Auth Service/task1_Dependency.md`, section `13. Security & Best Practices` |

---

## 15. Final Checklist

### For RBAC tests

- [ ] Go version is compatible with `backend/services/auth-service/go.mod`.
- [ ] Go dependencies are downloaded.
- [ ] `go test ./internal/authorization ./internal/usecase` passes.
- [ ] No unused `chi`, `grpc`, or `golang-jwt/jwt/v5` dependency added accidentally.

### For full RBAC HTTP flow

- [ ] MySQL is running.
- [ ] Redis is running because full server startup requires it.
- [ ] `AUTH_MYSQL_DSN` points to `auth_db`.
- [ ] `AUTH_MYSQL_DSN` includes `parseTime=true&charset=utf8mb4&loc=UTC`.
- [ ] Migrations `001` to `005` are applied for current full service.
- [ ] Migration `004_harden_role_assignments.up.sql` is applied.
- [ ] `role_assignments.reason` column exists.
- [ ] `uk_role_assignments_active` index exists.
- [ ] JWT RSA keys exist and env paths are correct.
- [ ] `JWT_KEY_ID` is set.
- [ ] `JWT_SIGNING_ALG=RS256`.
- [ ] `AUTH_ROLE_REASON_MAX_LENGTH` is unset or greater than zero.
- [ ] Local `.env` is sourced before `go run ./cmd/server`.
- [ ] `/healthz` returns `{"status":"ok"}`.
- [ ] `/.well-known/jwks.json` returns public keys.
- [ ] Role mutation requests include `Authorization: Bearer ...`.
- [ ] Role mutation requests include a non-empty `reason`.
- [ ] Seller scoped roles include `scope_type=seller` and a `scope_id`.
- [ ] Internal endpoints are not exposed publicly in any deployment.

### Quick Task 6 commands

```bash
cd backend/services/auth-service
go test ./internal/authorization ./internal/usecase
```

```bash
cd backend/services/auth-service
mysql -h 127.0.0.1 -P 3306 -u root -p < migrations/004_harden_role_assignments.up.sql
```

```bash
cd backend/services/auth-service
set -a
source .env
set +a
go run ./cmd/server
```

```bash
curl http://localhost:8081/healthz
curl http://localhost:8081/.well-known/jwks.json
```
