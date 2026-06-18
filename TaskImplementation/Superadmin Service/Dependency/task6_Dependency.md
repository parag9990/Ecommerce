# Project Dependency & Setup Guide

> Incremental beginner-friendly dependency handbook for session visibility. Shared setup is referenced from previous dependency guides so junior developers do not have to read duplicated Go, MySQL, Docker, and `.env` instructions again and again.

## Document Variables

| Variable | Meaning |
|---|---|
| `SERVICE_NAME` | Service folder value supplied by the task request |
| `TASK_FILE_NAME` | Current implementation task filename supplied by the task request |
| `INPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | Replace `.md` in `TASK_FILE_NAME` with `_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}` |

## 1. Project Overview

`INPUT_FILE_PATH` defines session visibility access control. Simple meaning: analytics dashboard ka session data Session Management Service own karega, but kaunsa admin dashboard, journey, funnel, heatmap, ya suspicious activity dekh sakta hai ye current service ka RBAC decide karega.

### Current implementation reality

The current repository has a runnable Go HTTP implementation for session visibility authorization:

```text
backend/services/superadmin-service/
├── internal/domain/session_visibility.go
├── internal/usecase/session_visibility.go
├── internal/transport/http/session_visibility_handler.go
├── internal/rbac/admin_context.go
├── internal/config/config.go
└── migrations/
```

Current flow:

```text
Session Analytics Dashboard / API Gateway
        |
        | trusted admin headers
        v
GET /api/v1/admin/session-analytics/access
POST /api/v1/admin/session-analytics/authorize
        |
        v
MySQL RBAC permission check
        |
        v
Authorization decision + downstream metadata headers
        |
        v
Gateway forwards allowed analytics request to Session Management Service
```

Important boundary:

- Current service does **not** store session events.
- Current service does **not** read MongoDB, Redis, heatmaps, journeys, or funnels directly.
- Current service does **not** call Session Management Service directly.
- Current service only gives a permission decision and masking metadata for the gateway or dashboard flow.

Beginner note: Is task me "session visibility" ka matlab session data banana nahi hai. Iska matlab sensitive session analytics ko dekhne ka permission gate banana hai.

## 2. Tech Stack

### Task-specific technologies

| Technology / component | Required? | Why used | Simple Hinglish explanation |
|---|---|---|---|
| Go `net/http` session visibility routes | Required | Exposes access and authorization APIs | Current code external framework ke bina Go standard HTTP server use karta hai. |
| RBAC permission service | Required | Checks `sessions:read` and `sessions:risk:read` from admin role mapping | RBAC role ko permission group banata hai. Final access exact permission se decide hota hai. |
| MySQL RBAC tables | Required for functional local/prod run | Stores active admin role and role-permission mappings | MySQL se pata chalta hai admin active hai ya nahi, aur uske role ko session permissions hain ya nahi. |
| Audit recorder | Required for risk analytics access when DB is configured | Records `include_risk=true` session analytics access | Sensitive risk data access ka trace hona chahiye. |
| API Gateway | Required for secure production integration | Validates admin identity and forwards trusted headers | Public client ke headers par direct trust unsafe hai. Gateway pehle JWT/session verify karega. |
| Session Management Service | Required for complete dashboard data | Owns live sessions, journeys, funnels, heatmaps, and suspicious activity data | Current service sirf access decision deta hai; actual analytics data Session Service se aata hai. |

### Already documented shared technologies

Go installation, Go modules, MySQL driver, MySQL setup, Docker-for-MySQL, HTTP server basics, migrations, request headers, and generic troubleshooting are already explained.

Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`2. Tech Stack`, `3. Required Software`, `4. Dependency Management`, `5. Database Setup`, `7. Environment Variables`, and `12. Common Errors & Fixes`

Also refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/Go_Modules.md`

Section:
`Go module setup`

### Technologies mentioned in broader design but not current direct dependencies

| Technology | Current status in this service | Setup action for this task |
|---|---|---|
| MongoDB | Not imported or configured here | Do not set up for Superadmin-only authorization tests |
| Redis | Not imported or configured here | Do not set up for current service |
| Kafka / RabbitMQ / NATS | Not used by current session visibility implementation | No broker setup required |
| gRPC / protobuf | Not present in current `go.mod` | Do not install unless transport implementation changes |
| Session Service client package | No direct client exists in current service | Gateway/Session Service integration must be configured outside this service |

## 3. Required Software

No new local software is required beyond the previous guides for running the current service.

| Software | Required? | Why | Status |
|---|---|---|---|
| Go `1.26.3` compatible toolchain | Yes | Build, test, run current Go service | Reused |
| MySQL `8.x` and MySQL client | Yes for functional authorization | RBAC and audit tables | Reused |
| curl or similar HTTP client | Recommended | Verify access and authorize routes | Reused |
| Docker Engine / Docker Desktop | Optional | Simple local MySQL container | Reused |
| Session Management Service runtime | Required only for end-to-end analytics data | Provides actual session analytics responses | External to current service |
| MongoDB / Redis | Required only by Session Management Service if implementing it locally | Session event store and active session/cache layer in broader design | Not a current service dependency |

For install commands, refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`3. Required Software`

## 4. Dependency Management

### New dependency result

No new Go module is required for this task.

Current direct dependency remains:

```text
github.com/go-sql-driver/mysql v1.10.0
```

Current implementation uses:

- Go standard library `net/http`
- Go standard library `database/sql`
- Existing internal RBAC, audit, config, and HTTP packages
- Existing MySQL migrations

Do not add unused packages like gRPC, Gin, Redis clients, MongoDB drivers, or Kafka clients just because broader architecture documents mention them. Agar code import nahi karta, dependency install karna unnecessary churn hai.

### Task-specific verification

From the service directory:

```bash
go test ./internal/domain ./internal/rbac ./internal/usecase ./internal/transport/http
go mod verify
```

These tests verify session route policies, filter validation, role visibility, masking decisions, risk denial, and handler responses. They do not verify a real Session Management Service or gateway forwarding.

For full Go module setup and common Go module errors, refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`4. Dependency Management`

## 5. Database Setup

### Database ownership

MySQL setup is reused, but ownership is important:

| Data / decision | Owner |
|---|---|
| Active admin identity and DB role | Current service MySQL |
| `sessions:read` and `sessions:risk:read` role mapping | Current service MySQL |
| Risk analytics access audit log | Current service MySQL `admin_audit_logs` |
| Session events, active sessions, journeys, funnels, heatmaps | Session Management Service |
| MongoDB / Redis session storage | Session Management Service, not current service |

### Reused MySQL installation and credentials

Do not repeat MySQL installation, Docker, database creation, DSN, local admin seed, or complete migration instructions here.

Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`5. Database Setup`, `Recommended Docker setup`, `Run migrations`, `Seed a local admin`, and `Verify database`

Also refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/MySQL.md`

Section:
`Local MySQL setup`

### Task-relevant migrations

Apply the complete migration sequence exactly once, but these files matter most for this task:

| Migration | Why it matters for session visibility |
|---|---|
| `001_create_superadmin_rbac.up.sql` | Creates `admin_users`, `admin_permissions`, and `admin_role_permissions` |
| `002_seed_admin_rbac.up.sql` | Seeds `sessions:read` and `sessions:risk:read` permissions and role mappings |
| `006_create_admin_audit_logs.up.sql` | Stores audit records for risk session analytics access |

Task-specific permission seed:

| Permission | Risk level in code | Meaning |
|---|---|---|
| `sessions:read` | Medium | Admin can view general session analytics dashboard |
| `sessions:risk:read` | High | Admin can view suspicious/risk session details |

Current seeded role behavior:

| Role | `sessions:read` | `sessions:risk:read` | Result |
|---|---:|---:|---|
| `superadmin` | Yes | Yes | Full session dashboard and risk visibility |
| `operations_admin` | Yes | Yes | Full session dashboard and risk visibility |
| `readonly_admin` | Yes | No | Dashboard visible, PII/risk details masked |
| `finance_admin` | No | No | Session dashboard denied |
| `catalog_admin` | No | No | Session dashboard denied |

### Task-specific database verification

After migrations and local admin provisioning:

```bash
mysql -h 127.0.0.1 -P 3306 -u superadmin -p -D superadmin_db -e \
  "SELECT permission_key, description FROM admin_permissions WHERE permission_key LIKE 'sessions:%' ORDER BY permission_key;"

mysql -h 127.0.0.1 -P 3306 -u superadmin -p -D superadmin_db -e \
  "SELECT role, permission_key FROM admin_role_permissions WHERE permission_key LIKE 'sessions:%' ORDER BY role, permission_key;"

mysql -h 127.0.0.1 -P 3306 -u superadmin -p -D superadmin_db -e \
  "SHOW TABLES LIKE 'admin_audit_logs';"
```

Expected:

- `sessions:read` exists.
- `sessions:risk:read` exists.
- `superadmin` and `operations_admin` have both permissions.
- `readonly_admin` has only `sessions:read`.
- `admin_audit_logs` table exists if migration `006` is applied.

> Important: If `SUPERADMIN_DATABASE_DSN` is empty, the runtime uses an empty static permission repository and protected session visibility checks will deny access. Functional verification needs MySQL.

## 6. Redis / Queue / External Services

### Session Management Service

Session Management Service is mandatory for the complete dashboard experience, but current service does not have a direct `SESSION_SERVICE_*` env variable or HTTP/gRPC client.

Expected integration shape:

```text
Dashboard/Gateway
  -> current service /api/v1/admin/session-analytics/authorize
  -> if allowed, Gateway forwards original analytics request to Session Management Service
  -> Gateway includes returned downstream headers
  -> Session Service applies masking and risk scope
```

Analytics routes that need authorization decision:

| Analytics route key | Intended Session Service route | Required permission | Extra validation |
|---|---|---|---|
| `analytics.live` | `GET /api/v1/analytics/live` | `sessions:read` | Normal dashboard access |
| `analytics.sessions` | `GET /api/v1/analytics/sessions` | `sessions:read` | Optional risk filter requires risk permission |
| `analytics.journey` | `GET /api/v1/analytics/sessions/{session_id}/journey` | `sessions:read` | `session_id` is required |
| `analytics.funnels` | `GET /api/v1/analytics/funnels` | `sessions:read` | Normal aggregate access |
| `analytics.heatmaps` | `GET /api/v1/analytics/heatmaps` | `sessions:read` | `path` is required and must start with `/` |

Returned downstream metadata headers:

| Header | Meaning |
|---|---|
| `x-admin-id` | Verified admin ID from trusted request context |
| `x-admin-roles` | Admin roles from trusted request context |
| `x-request-id` | Request trace ID |
| `x-session-id` | Admin session ID |
| `x-admin-permission` | Granted base permission, currently `sessions:read` |
| `x-admin-risk-allowed` | `true` when admin has `sessions:risk:read` |
| `x-admin-mask-pii` | `true` when Session Service should mask PII/risk fields |

Session Service should treat these headers as trusted only if they come from the gateway/internal network, not from public clients.

### API Gateway

API Gateway is required for production security because current service trusts incoming admin headers. Gateway should:

- Validate admin JWT/session/MFA using Auth Service.
- Strip any client-supplied `X-Admin-*`, `X-Session-Id`, `X-Request-Id`, and `X-MFA-Verified` headers.
- Inject verified admin context headers.
- Call the current service's authorize route before forwarding analytics requests.
- Forward only allowed requests to Session Management Service.
- Pass returned downstream metadata to Session Service over a trusted internal channel.

For shared gateway security warning, refer:
`TaskImplementation/{SERVICE_NAME}/task3_Dependency.md`

Section:
`6. Redis / Queue / External Services` -> `Auth Service and API Gateway`

### Redis, MongoDB, and queues

| Service | Current service status | When needed |
|---|---|---|
| Redis | Not used directly | Needed only when running/implementing Session Management Service active-session cache |
| MongoDB | Not used directly | Needed only when running/implementing Session Management Service event store |
| Kafka / RabbitMQ / NATS | Not used directly | Only if Session Service event pipeline is implemented |

No Redis, MongoDB, or queue credential is required to verify current service authorization APIs.

## 7. Environment Variables

### Reused environment setup

General `.env` creation, sourcing, DSN, HTTP address, timeout, and downstream-service env behavior are already documented.

Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`7. Environment Variables`

Also refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/Environment.md`

Sections:
`Required Environment Variables` and `Common Errors And Fixes`

### Task-specific existing variables

These variables already exist in the current config and are already mentioned in shared environment docs. This section explains why they matter for `INPUT_FILE_PATH`.

```env
# Session visibility limits
SUPERADMIN_SESSION_ANALYTICS_MAX_RANGE=720h
SUPERADMIN_SESSION_ANALYTICS_MAX_PAGE_SIZE=100
```

| Variable | Default | Required? | Purpose | Security / setup note |
|---|---|---|---|---|
| `SUPERADMIN_SESSION_ANALYTICS_MAX_RANGE` | `720h` | Optional | Maximum allowed `from` to `to` date range for session analytics authorization | Use Go duration format like `24h`, `168h`, `720h`; too-large ranges can create heavy Session Service queries |
| `SUPERADMIN_SESSION_ANALYTICS_MAX_PAGE_SIZE` | `100` | Optional | Maximum accepted page size in authorize payload | Keep small enough to avoid expensive dashboard queries and data exposure |

Common mistakes:

- `SUPERADMIN_SESSION_ANALYTICS_MAX_RANGE=30d` is invalid because Go duration does not support `d`. Use `720h`.
- `SUPERADMIN_SESSION_ANALYTICS_MAX_PAGE_SIZE=0` fails config validation because value must be greater than zero.
- `page_size` in request body cannot exceed `SUPERADMIN_SESSION_ANALYTICS_MAX_PAGE_SIZE`.

### Variables that do not exist in current service

Do not invent these inside the current service unless implementation changes:

| Variable idea | Current status |
|---|---|
| `SESSION_SERVICE_ADMIN_BASE_URL` | Not read by current service |
| `SESSION_SERVICE_TIMEOUT` | Not read by current service |
| `MONGO_URI` | Not read by current service |
| `REDIS_URL` | Not read by current service |
| `KAFKA_BROKERS` | Not read by current session visibility code |

If end-to-end analytics forwarding is added later, gateway or Session Management Service should own those settings, or a new current-service client should be implemented and documented.

## 8. Docker Setup

No new Docker container, image, volume, network, health check, or restart policy is introduced by current session visibility authorization.

Reuse existing MySQL Docker setup:

`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`Recommended Docker setup`, `Docker Compose example for MySQL`, and `9. Docker Setup`

Current repository status for this task:

| Docker item | Present? | Task impact |
|---|---|---|
| Service Dockerfile | No | Run Go locally for now |
| Full local Compose stack | No | Start MySQL separately |
| Session Service Compose service | No | End-to-end analytics data flow cannot be launched from current folder |
| MongoDB/Redis containers | No for this service | Only needed for actual Session Management Service runtime |
| MySQL local container command | Documented previously | Reuse for RBAC and audit tables |

### Ports & networking

| Service / component | Port | Purpose | Status |
|---|---:|---|---|
| Current backend HTTP server | `8088` via `HTTP_ADDR` | Session visibility access and authorize APIs | Reused |
| MySQL | `3306` | RBAC and risk access audit logs | Reused |
| Session Management Service | Not configured in current service | Actual analytics data APIs | External, gateway-owned |
| MongoDB for Session Service | Common default `27017` | Session event storage if Session Service is run locally | Not current service |
| Redis for Session Service | Common default `6379` | Active sessions/cache if Session Service is run locally | Not current service |
| API Gateway | Not configured in current service | Authenticates admin and forwards requests | External |

For port conflicts, firewall guidance, and Docker networking basics, refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`8. Ports & Networking`

Production warning: Current service port should stay private. Public access would allow forged admin headers unless a trusted gateway protects it.

## 9. Local Development Setup

### Step 1: Read reused setup first

Follow these before this task-specific section:

- `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`
- `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md`
- `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md`
- `TaskImplementation/{SERVICE_NAME}/Dependency/Environment.md`
- `TaskImplementation/{SERVICE_NAME}/Dependency/HTTP_API.md`

### Step 2: Enter service directory

```bash
cd backend/services/superadmin-service
```

### Step 3: Install dependencies

No new dependency is required. Use the existing Go module flow:

```bash
go mod download
go mod verify
```

### Step 4: Start MySQL and apply migrations

Use the reused MySQL setup and apply all `.up.sql` files in order. Confirm migrations `001`, `002`, and `006` are applied for this task.

### Step 5: Provision active local admins

For task-specific verification, create or use disposable local admins with these DB roles:

| Local role to test | Why |
|---|---|
| `superadmin` | Should allow session dashboard and risk access |
| `operations_admin` | Should allow session dashboard and risk access |
| `readonly_admin` | Should allow dashboard but deny `include_risk=true` |

Important: Changing only `X-Admin-Roles` header is not enough in functional DB mode. Exact permission is loaded through `X-Admin-Id` from MySQL.

### Step 6: Configure task-specific limits if needed

Use defaults for normal local testing:

```bash
export SUPERADMIN_SESSION_ANALYTICS_MAX_RANGE=720h
export SUPERADMIN_SESSION_ANALYTICS_MAX_PAGE_SIZE=100
```

Keep the existing DSN and HTTP variables from previous dependency docs.

### Step 7: Run tests and build

```bash
go test ./internal/domain ./internal/rbac ./internal/usecase ./internal/transport/http
go test ./...
go build ./cmd/server
```

### Step 8: Start backend service

```bash
APP_ENV=local HTTP_ADDR=127.0.0.1:8088 go run ./cmd/server
```

If you use `.env`, remember current code does not auto-load it. Source it first:

```bash
set -a
source .env
set +a
go run ./cmd/server
```

### Step 9: Verify health and readiness

```bash
curl -i http://127.0.0.1:8088/healthz
curl -i http://127.0.0.1:8088/readyz
```

`/readyz` depends on MySQL when DSN is configured.

### Step 10: Verify dashboard access

Use an admin whose DB role has `sessions:read`, for example `readonly_admin` or `operations_admin`.

```bash
curl -i \
  -H 'X-Admin-Id: admin_local_readonly' \
  -H 'X-User-Id: user_local_readonly' \
  -H 'X-Admin-Roles: readonly_admin' \
  -H 'X-Session-Id: session_local_001' \
  -H 'X-Request-Id: request_session_access_001' \
  http://127.0.0.1:8088/api/v1/admin/session-analytics/access
```

Expected:

- HTTP `200`
- `can_view_dashboard: true`
- `risk_allowed: false` for `readonly_admin`
- `mask_pii: true` for `readonly_admin`
- suspicious/risk modules not visible for `readonly_admin`

### Step 11: Verify analytics authorization

```bash
curl -i \
  -X POST http://127.0.0.1:8088/api/v1/admin/session-analytics/authorize \
  -H 'Content-Type: application/json' \
  -H 'X-Admin-Id: admin_local_readonly' \
  -H 'X-User-Id: user_local_readonly' \
  -H 'X-Admin-Roles: readonly_admin' \
  -H 'X-Session-Id: session_local_002' \
  -H 'X-Request-Id: request_session_authorize_001' \
  -d '{
    "route": "analytics.sessions",
    "from": "2026-06-01",
    "to": "2026-06-08",
    "page": 1,
    "page_size": 20
  }'
```

Expected:

- HTTP `200`
- `allowed: true`
- `permission: "sessions:read"`
- `risk_allowed: false`
- `mask_pii: true`
- `downstream_headers.x-admin-mask-pii: "true"`

### Step 12: Verify risk denial for readonly admin

```bash
curl -i \
  -X POST http://127.0.0.1:8088/api/v1/admin/session-analytics/authorize \
  -H 'Content-Type: application/json' \
  -H 'X-Admin-Id: admin_local_readonly' \
  -H 'X-User-Id: user_local_readonly' \
  -H 'X-Admin-Roles: readonly_admin' \
  -H 'X-Session-Id: session_local_003' \
  -H 'X-Request-Id: request_session_authorize_002' \
  -d '{
    "route": "analytics.sessions",
    "include_risk": true
  }'
```

Expected:

- HTTP `403`
- `code: "FORBIDDEN"`
- `required_permission: "sessions:risk:read"`

### Step 13: Verify risk allow for operations or superadmin

```bash
curl -i \
  -X POST http://127.0.0.1:8088/api/v1/admin/session-analytics/authorize \
  -H 'Content-Type: application/json' \
  -H 'X-Admin-Id: admin_local_operations' \
  -H 'X-User-Id: user_local_operations' \
  -H 'X-Admin-Roles: operations_admin' \
  -H 'X-Session-Id: session_local_004' \
  -H 'X-Request-Id: request_session_authorize_003' \
  -d '{
    "route": "analytics.sessions",
    "include_risk": true,
    "page_size": 20
  }'
```

Expected:

- HTTP `200`
- `risk_allowed: true`
- `mask_pii: false`
- `downstream_headers.x-admin-risk-allowed: "true"`
- An audit record is attempted with action `session.analytics.access`

## 10. Running the Project

No new start command is introduced.

Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`10. Local Development Setup`, `11. Running the Project`, and `Quick Command Reference`

### Task-specific success criteria

```text
MySQL is reachable.
Migrations 001, 002, and 006 are applied.
Active local admins exist for readonly and risk-allowed roles.
GET /api/v1/admin/session-analytics/access returns expected module visibility.
POST /api/v1/admin/session-analytics/authorize returns downstream headers.
Readonly admin gets 403 for include_risk=true.
Operations or superadmin gets risk_allowed=true for include_risk=true.
No MongoDB, Redis, Kafka, or direct Session Service client is required for this authorization-only verification.
```

### End-to-end dashboard run expectation

For a real dashboard demo, current service alone is not enough. You also need:

- API Gateway route that calls current service before Session Service.
- Session Management Service analytics endpoints.
- Session Service storage stack such as MongoDB and Redis if that service implementation uses them.
- Frontend dashboard wired to the gateway.

Those pieces are outside current service setup and should be documented in their own service dependency guides.

## 11. Common Errors & Fixes

Generic Go, MySQL, Docker, port, `.env`, migration, and trusted-header issues are already documented in previous guides.

Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`12. Common Errors & Fixes`

Only task-specific errors are listed here.

| Error / symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `403 FORBIDDEN` with `required_permission: sessions:read` | Admin DB role does not have dashboard permission, admin is inactive, or DSN is empty | Use active `superadmin`, `operations_admin`, or `readonly_admin`; verify MySQL seed and DSN | Add role-permission verification to onboarding |
| `403 FORBIDDEN` with `required_permission: sessions:risk:read` | Request has `include_risk=true`, but role lacks risk permission | Use `superadmin` or `operations_admin` for risk views | Keep risk module hidden for readonly admins |
| `400 VALIDATION_FAILED` with unknown route | Payload route is not one of known route keys | Use `analytics.live`, `analytics.sessions`, `analytics.journey`, `analytics.funnels`, or `analytics.heatmaps` | Keep frontend route enum synced with backend |
| `400 VALIDATION_FAILED` for export | Payload set `export=true` | Remove export flag; export is not included in this task | Add separate reviewed export permission and implementation later |
| `400 VALIDATION_FAILED` for date range | `from` after `to`, only one date supplied, or range exceeds max | Send both values in `YYYY-MM-DD` or RFC3339 and keep range within max | Configure sane `SUPERADMIN_SESSION_ANALYTICS_MAX_RANGE` |
| `400 VALIDATION_FAILED` for page size | `page_size` exceeds configured max | Reduce page size or adjust env after review | Keep UI pagination under backend limit |
| `400 VALIDATION_FAILED` for journey | Route `analytics.journey` missing `session_id` | Include valid `session_id` | Frontend should require session selection before journey call |
| `400 VALIDATION_FAILED` for heatmaps | Route `analytics.heatmaps` missing `path`, or path does not start with `/` | Send path like `/products/123` | Normalize dashboard path filters |
| `400 VALIDATION_FAILED` for device type | Device type is not `desktop`, `mobile`, or `tablet` | Send one of the accepted values or omit it | Use a controlled UI select |
| `500 INTERNAL` during risk access | Audit insert failed, commonly because migration `006` is missing or DB schema is wrong | Apply complete migrations and inspect `admin_audit_logs` | Include audit migration check in setup |
| Dashboard says allowed but analytics data still fails | Gateway or Session Service forwarding is missing/misconfigured | Configure gateway to call Session Service with returned headers | Add end-to-end contract test |
| Session Service ignores masking | Session Service does not respect `x-admin-mask-pii` or risk headers | Update Session Service response masking contract | Test readonly and risk roles together |

## 12. Security & Best Practices

### Task-specific security audit

| Severity | Finding | Recommended fix |
|---|---|---|
| Critical for public deployment | Current service trusts admin headers and does not validate JWT itself | Keep service private behind trusted API Gateway and strip forged client headers |
| High | Session analytics can contain user behavior, device, IP, and risk signals | Keep `sessions:risk:read` separate from `sessions:read`; mask PII by default |
| High | Gateway-to-Session-Service handoff is not implemented in this repository | Add gateway route policy, internal auth, and tests before production dashboard use |
| High | No direct Session Service readiness check exists in current service | Gateway or deployment stack should monitor Session Service separately |
| High | Returned downstream headers are plain metadata | Send them only over trusted internal network; consider signed internal context if services cross trust boundaries |
| Medium | Risk access audit depends on migration `006` and working MySQL insert | Verify audit table and include risk-access tests |
| Medium | No export support exists, but analytics export would be sensitive | Keep export disabled until a separate permission and audit-heavy workflow exists |
| Medium | Session Service masking behavior is only a contract expectation here | Add consumer-driven tests between gateway/current service and Session Service |
| Medium | `SUPERADMIN_SESSION_ANALYTICS_MAX_RANGE` and page size can be set too high | Keep environment values reviewed and environment-specific |

### Task-specific best practices

- Least privilege follow karo: readonly admins ko dashboard mile, but risk details nahi.
- Risk data ke liye separate UI module, separate permission, and audit trail rakho.
- `include_risk=true` request only tab bhejo jab user ne suspicious/risk view open kiya ho.
- Session Service ko raw identifiers return karne se pehle `x-admin-mask-pii` respect karna chahiye.
- Date ranges and page sizes small rakho; analytics queries expensive ho sakti hain.
- Gateway and Session Service logs me request ID preserve karo.
- Raw IP, user agent, fingerprint, or PII logs me mat dump karo.
- Permission changes ko Go code, migration, tests, and docs ke saath deploy karo.
- End-to-end dashboard tests me readonly and operations roles dono cover karo.

## 13. Missing or Misconfigured Things

These are not blockers for current authorization route tests, but they matter for full product readiness:

- [ ] No API Gateway route implementation in this repository for pre-authorizing analytics routes
- [ ] No direct Session Management Service client or `SESSION_SERVICE_*` environment variables in current service
- [ ] No Session Management Service Docker/Compose setup in current service folder
- [ ] No MongoDB/Redis setup for Session Management Service in current service folder
- [ ] No end-to-end test proving gateway forwards allowed analytics requests with downstream metadata
- [ ] No contract test proving Session Service masks PII when `x-admin-mask-pii=true`
- [ ] No Session Service readiness/health dependency check in current service
- [ ] No signed internal admin-context metadata
- [ ] No export permission or export workflow for session analytics
- [ ] No production gateway header-sanitization config in current service folder
- [ ] No service Dockerfile or complete local Compose stack

## 14. References to Previous Dependency Files

| Previous dependency file | Section / topic reused | Why reused |
|---|---|---|
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Tech stack and required software | Same Go, MySQL, HTTP, Docker, and curl setup |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Dependency management | No new Go module required |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | MySQL installation, Docker setup, DSN, migrations, and local admin seed | Same database setup and migration process |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Environment variables and `.env` loading | Same env loading behavior; task-specific vars already listed there |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Ports, Docker, local run, health/readiness, generic troubleshooting | Same runtime and networking basics |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | MySQL ownership and migration source-of-truth rule | Same MySQL decision and migration strategy |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | RBAC roles, permissions, trusted headers, API Gateway/Auth warning | Session visibility reuses exact RBAC model |
| `TaskImplementation/{SERVICE_NAME}/task4_Dependency.md` | Admin metadata forwarding pattern | Session downstream metadata follows the same trusted-context idea |
| `TaskImplementation/{SERVICE_NAME}/task5_Dependency.md` | Downstream HTTP/service-boundary guidance | Same principle: current service should not directly own another service's data |
| `TaskImplementation/{SERVICE_NAME}/Dependency/Go_Modules.md` | Go module setup | Current `go.mod` unchanged |
| `TaskImplementation/{SERVICE_NAME}/Dependency/MySQL.md` | MySQL credentials and local setup | Same database |
| `TaskImplementation/{SERVICE_NAME}/Dependency/Migrations.md` | Complete migration flow | Same manual migration mechanism |
| `TaskImplementation/{SERVICE_NAME}/Dependency/Environment.md` | Env variable list and `.env` loading | Same config loader |
| `TaskImplementation/{SERVICE_NAME}/Dependency/HTTP_API.md` | HTTP server and admin request headers | Same `net/http` transport and actor middleware |
| `TaskImplementation/{SERVICE_NAME}/Dependency/Downstream_HTTP_Services.md` | Downstream service trust model | Useful for gateway-to-Session-Service handoff planning |

## 15. Final Checklist

- [ ] Previous dependency documentation checked before using this file
- [ ] No duplicate Go/MySQL/Docker install steps copied here
- [ ] Existing Go modules downloaded and verified
- [ ] MySQL `8.x` running for functional authorization
- [ ] Complete migrations applied exactly once
- [ ] `sessions:read` and `sessions:risk:read` permissions verified in DB
- [ ] Active local admins provisioned for readonly and risk-allowed roles
- [ ] `SUPERADMIN_DATABASE_DSN` exported for DB-backed checks
- [ ] `SUPERADMIN_SESSION_ANALYTICS_MAX_RANGE` uses Go duration format if overridden
- [ ] `SUPERADMIN_SESSION_ANALYTICS_MAX_PAGE_SIZE` set to a positive integer if overridden
- [ ] Backend service starts on expected `HTTP_ADDR`
- [ ] `/healthz` and `/readyz` checked
- [ ] Dashboard access route verified
- [ ] Authorization route verified for normal session analytics
- [ ] Risk denial verified for readonly admin
- [ ] Risk allow and audit behavior verified for operations or superadmin
- [ ] No Redis, MongoDB, Kafka, RabbitMQ, or gRPC dependency added to current service unnecessarily
- [ ] Gateway/Session Service handoff documented as external integration
- [ ] Logs checked for session analytics access decisions
- [ ] No duplicate setup documentation added
