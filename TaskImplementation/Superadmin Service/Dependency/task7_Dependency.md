# Project Dependency & Setup Guide

> Incremental beginner-friendly dependency handbook for platform settings. Common Go, MySQL, Docker, `.env`, migration, and HTTP setup is reused from previous dependency guides. Is file ka focus sirf current task ke new setup points par hai.

## Document Variables

| Variable | Meaning |
|---|---|
| `SERVICE_NAME` | Service folder value supplied by the task request |
| `TASK_FILE_NAME` | Current implementation task filename supplied by the task request |
| `INPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | Replace `.md` in `TASK_FILE_NAME` with `_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}` |

## 1. Project Overview

`INPUT_FILE_PATH` defines platform settings: maintenance mode, commission rules, feature flags, and search synonyms. Simple meaning: admin ek central control plane se important platform behavior change karega, but har change permission, reason, version, validation, audit, and event logging ke saath hoga.

### Current implementation reality

The current repository has a runnable Go HTTP implementation for platform settings:

```text
backend/services/superadmin-service/
|-- internal/domain/platform_setting.go
|-- internal/usecase/platform_settings.go
|-- internal/usecase/platform_settings_validation.go
|-- internal/repository/mysql_platform_settings_repository.go
|-- internal/transport/http/settings_handler.go
|-- internal/clients/settings_event_publisher.go
|-- internal/config/config.go
`-- migrations/005_create_platform_settings.up.sql
```

Current flow:

```text
Admin client / API Gateway
        |
        | trusted admin headers
        v
GET   /api/v1/admin/settings
PATCH /api/v1/admin/settings/{key}
GET   /api/v1/admin/search/synonyms
POST  /api/v1/admin/search/synonyms
        |
        v
MySQL RBAC permission check
        |
        v
MySQL platform_settings update with version check
        |
        v
Audit record + in-memory cache invalidation + log-based update event
```

Important boundary:

- Current service stores platform setting values in MySQL.
- Current service does not directly call CMS Service, Search Service, Typesense, API Gateway, Kafka, RabbitMQ, or Redis.
- Current event publisher logs the event topic only. It is not a real message broker publisher yet.
- Current cache is in-memory TTL cache inside the process, not Redis.
- Actual maintenance enforcement must happen in API Gateway/frontend consumers.
- Actual search synonym application must happen in Search Service/Typesense consumers.
- Actual commission usage must happen in CMS, Order, Payment, or settlement consumers.

Beginner note: Is task me "settings update" ka matlab DB me safe config save karna hai. "Setting apply karna" alag service ka kaam ho sakta hai. Example: `search_synonyms` row yahan save hoti hai, but Typesense synonym config Search Service apply karega.

## 2. Tech Stack

### Task-specific technologies

| Technology / component | Required? | Why used | Simple Hinglish explanation |
|---|---|---|---|
| Go `net/http` settings routes | Required | Exposes platform settings and synonym APIs | External web framework nahi hai; Go standard HTTP server use ho raha hai. |
| MySQL `platform_settings` table | Required for functional settings | Stores versioned settings as JSON with risk and update metadata | MySQL ek relational DB hai; yahan settings durable rows ke form me store hoti hain. |
| MySQL JSON column | Required | Stores different schemas for maintenance, commission, flags, and synonyms | JSON column flexible value store karta hai, but Go validation schema ko strict rakhta hai. |
| RBAC permission service | Required | Checks `platform:settings:*` and `search:synonyms:*` permissions | Role ke through exact permission verify hoti hai. Sirf header role par blind trust nahi. |
| High-risk guard | Required | Requires reason/request ID, and for platform settings write also MFA | Sensitive config change ke liye extra safety check hai. |
| Audit recorder | Required for DB-backed mutation trace | Records `platform_setting.update` after successful update | Audit se pata chalta hai kis admin ne kya setting change ki. |
| In-memory platform settings cache | Required internally | Avoids repeated DB reads during cache TTL | Ye local process cache hai. Redis setup current code me nahi hai. |
| Log-based settings event publisher | Required internally | Logs `PlatformSettingUpdated` event with topic and version | Abhi ye real broker nahi, sirf structured log event hai. |
| API Gateway / Auth Service | Required for secure production | Validates real admin identity before setting trusted headers | Public client ko direct admin headers bhejne dena unsafe hai. |
| Search Service / Typesense | Required only for end-to-end synonym apply | Consumes saved synonym changes outside this service | Current service Typesense API key ya URL use nahi karta. |
| CMS / Order / Payment consumers | Required only for end-to-end commission behavior | Consumers read/apply commission rules outside this service | Current service commission calculation engine nahi chalata. |

### Already documented shared technologies

Go installation, Go modules, MySQL driver, MySQL setup, Docker-for-MySQL, HTTP server basics, migrations, request headers, admin provisioning, and generic troubleshooting are already explained.

Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`2. Tech Stack`, `3. Required Software`, `4. Dependency Management`, `5. Database Setup`, `7. Environment Variables`, `8. Docker Setup`, and `12. Common Errors & Fixes`

Also refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/main_dependency.md`

Sections:
`2. All Detected Dependencies`, `8. Database Setup`, `9. Environment Variables`, and `12. Missing / Incomplete Implementation Audit`

### Technologies mentioned in broader design but not current direct dependencies

| Technology | Current status in this service | Setup action for current task |
|---|---|---|
| Redis | Docs mention platform settings Redis cache, but current code uses in-memory TTL cache | No Redis container required for this service task |
| Kafka / RabbitMQ / NATS | No client package, broker URL, or producer code exists | No broker setup required |
| Typesense | Search Service owns Typesense; current service has no Typesense client/env | Do not configure Typesense here |
| CMS Service | Conceptual consumer of commission rules, no current CMS URL/env in this service | No CMS setup required for local settings API verification |
| gRPC / protobuf | Not present in current `go.mod` | Do not install unless transport changes |

## 3. Required Software

No new local software is required beyond previous setup guides.

| Software | Required? | Why | Status |
|---|---|---|---|
| Go `1.26.3` compatible toolchain | Yes | Build, test, run current Go service | Reused |
| MySQL `8.x` and MySQL client | Yes for DB-backed settings | Stores RBAC, `platform_settings`, and audit logs | Reused |
| curl or similar HTTP client | Recommended | Verify settings APIs | Reused |
| Docker Engine / Docker Desktop | Optional | Run local MySQL container if local MySQL is not installed | Reused |
| Migration CLI | Optional but recommended | Apply migration `005` and the complete sequence | Reused |
| Redis / broker / Typesense | No for this service task | Not wired in current code | Not needed |

For install commands, refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/Go_Modules.md`

Section:
`5. Installation Steps`

And:
`TaskImplementation/{SERVICE_NAME}/Dependency/MySQL.md`

Sections:
`5. Installation Steps` and `6. Docker Setup, If Possible`

## 4. Dependency Management

### New dependency result

No new Go module is required for current platform settings.

Current direct dependency remains:

```text
github.com/go-sql-driver/mysql v1.10.0
```

Current implementation uses:

- Go standard library `net/http`
- Go standard library `database/sql`
- Go standard library `encoding/json`
- Existing internal RBAC, audit, config, repository, and HTTP packages
- Existing MySQL migrations

Do not add Redis, Kafka, RabbitMQ, NATS, Typesense, gRPC, or CMS client packages for current local verification. Agar code import nahi karta, package install karna unnecessary churn hai.

### Task-specific verification

From the service directory:

```bash
go test ./internal/domain ./internal/rbac ./internal/usecase ./internal/transport/http
go mod verify
```

These tests verify setting validation, role permission behavior, version conflict handling, event publish call behavior, and HTTP request mapping. They do not verify real Redis, real broker, real Typesense, real CMS, or real API Gateway integration.

For Go module setup and common Go module errors, refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/Go_Modules.md`

Sections:
`7. Local Setup Without Docker` and `11. Common Errors And Fixes`

## 5. Database Setup

### Reused MySQL setup

MySQL installation, Docker MySQL run command, database creation, application user, DSN format, migration CLI setup, and generic DB troubleshooting are already documented.

Refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/MySQL.md`

Sections:
`5. Installation Steps`, `6. Docker Setup, If Possible`, `8. Required Environment Variables`, and `11. Common Errors And Fixes`

Also refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/Migrations.md`

Sections:
`6. Docker Setup, If Possible`, `7. Local Setup Without Docker`, and `10. Verify Running Commands`

### Task-specific database change

Current task introduces the `platform_settings` table through migration:

```text
backend/services/superadmin-service/migrations/005_create_platform_settings.up.sql
```

Rollback file:

```text
backend/services/superadmin-service/migrations/005_create_platform_settings.down.sql
```

### What the table stores

| Column / concept | Purpose |
|---|---|
| `setting_key` | Stable key like `maintenance_mode`, `commission_rules`, `feature_flags`, `search_synonyms` |
| `setting_type` | Category enum: `maintenance`, `commission`, `feature_flags`, `search` |
| `value_json` | JSON payload validated by Go before update |
| `risk_level` | Risk marker used for safe admin behavior |
| `version` | Optimistic locking value to prevent overwriting another admin's change |
| `updated_by_admin_id` | Admin who last changed the setting |
| `update_reason` | Required reason for audit and review |
| `created_at`, `updated_at` | Operational timestamps |

### Seeded default settings

Migration `005` inserts four default rows:

| Setting key | Default behavior | Required consumer |
|---|---|---|
| `maintenance_mode` | Disabled, admin and health checks allowed | API Gateway and frontend apps |
| `commission_rules` | `1000` bps, `INR`, effective from `2026-01-01T00:00:00Z` | CMS, Order, Payment, settlement flows |
| `feature_flags` | `new_checkout` disabled with `0` percent rollout | Gateway, frontend apps, backend consumers |
| `search_synonyms` | Empty synonym list | Search Service / Typesense |

### Task-specific migration requirement

Apply the complete migration sequence exactly once. For current settings APIs, these migrations matter most:

| Migration | Why it matters |
|---|---|
| `001_create_superadmin_rbac.up.sql` | Creates admin identity and permission tables |
| `002_seed_admin_rbac.up.sql` | Seeds `platform:settings:*` and `search:synonyms:*` permissions |
| `005_create_platform_settings.up.sql` | Creates and seeds platform settings |
| `006_create_admin_audit_logs.up.sql` | Stores setting mutation audit records when DB-backed audit is enabled |

Task-specific verification:

```bash
mysql -h 127.0.0.1 -P 3306 -u superadmin -p -D superadmin_db -e \
"SELECT setting_key, setting_type, risk_level, version FROM platform_settings ORDER BY setting_key;"
```

Expected keys:

```text
commission_rules
feature_flags
maintenance_mode
search_synonyms
```

### Credentials placement

No new database credential variable was introduced. Use the existing DSN:

```text
SUPERADMIN_DATABASE_DSN
```

Where to place it:

```text
backend/services/superadmin-service/.env
```

Important: Code reads OS environment variables with `os.Getenv`; it does not auto-load `.env`. Source the file before running.

## 6. Redis / Queue / External Services

### Current direct dependencies

| Service | Required for current settings APIs? | Current status | Setup needed now |
|---|---:|---|---|
| Redis | No | Not imported or configured. In-memory cache is used. | None |
| Kafka / RabbitMQ / NATS | No | Event publisher only writes structured logs. | None |
| Typesense | No | No direct client or env variable in current service. | None |
| Search Service | Not for local API verification | Future consumer of search synonym changes. | External integration only |
| CMS Service | Not for local API verification | Future consumer of commission settings. | External integration only |
| API Gateway | Required for secure production | Should validate auth and pass trusted admin headers. | Configure outside this service |

### New event topic variable

The current service has one task-specific event topic variable:

```text
SUPERADMIN_PLATFORM_SETTINGS_EVENT_TOPIC=platform.settings.updated
```

This value is used by `LoggingSettingsEventPublisher`. Simple meaning: update event ka topic log me dikhega, but message broker me publish nahi hoga.

### Redis note

Broader architecture docs mention Redis cache for platform settings with 5 minute TTL. Current code does not use Redis. So Redis setup ko duplicate ya force mat karo for this task.

Refer:
`TaskImplementation/{SERVICE_NAME}/task6_Dependency.md`

Section:
`2. Tech Stack`

Same rule applies here: architecture-level services are not local dependencies unless current code imports/configures them.

## 7. Environment Variables

### Reused environment setup

Shared `.env` location, loading behavior, common variables, duration format, and generic env errors are already documented.

Refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/Environment.md`

Sections:
`5. Installation Steps`, `7. Local Setup Without Docker`, `8. Required Environment Variables`, and `11. Common Errors And Fixes`

### New or task-specific variables

Only these variables are task-specific for platform settings:

| Variable | Required? | Default / example | Purpose | Security note |
|---|---|---|---|---|
| `SUPERADMIN_PLATFORM_SETTINGS_CACHE_TTL` | Optional | `5m` | In-memory settings cache TTL | Set low enough that stale config does not live too long |
| `SUPERADMIN_PLATFORM_SETTINGS_EVENT_TOPIC` | Optional but must not be blank after config load | `platform.settings.updated` | Topic label used by log-based settings event publisher | Not a secret |

Task-specific `.env` snippet:

```bash
SUPERADMIN_PLATFORM_SETTINGS_CACHE_TTL=5m
SUPERADMIN_PLATFORM_SETTINGS_EVENT_TOPIC=platform.settings.updated
```

Current code behavior:

- If cache TTL is omitted, default is `5m`.
- If cache TTL is `0`, service code also falls back to `5m`.
- If cache TTL is negative, config validation fails.
- If event topic is blank after trimming, config validation fails.
- Event topic is only a log topic right now, not a broker credential.

### Request headers required for settings APIs

Header behavior is reused from the HTTP API dependency guide, but settings writes have stricter high-risk requirements.

| Header | Needed for read? | Needed for write? | Why |
|---|---:|---:|---|
| `X-Admin-Id` | Yes | Yes | DB RBAC lookup uses admin id |
| `X-User-Id` or `X-Subject-Id` | Yes | Yes | Actor context validation |
| `X-Admin-Roles` or `X-Roles` | Yes | Yes | Actor context, but DB permission remains final |
| `X-Session-Id` | Yes | Yes | Actor context validation |
| `X-Request-Id` | Recommended | Required for writes | Audit context for high-risk mutation |
| `X-MFA-Verified` | No | Required for `platform:settings:write` | Critical platform settings require MFA |
| `X-IP-Hash` | Optional | Recommended | PII-safe audit metadata |

For full header explanation, refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/HTTP_API.md`

Section:
`8. Required Environment Variables`

## 8. Docker Setup

No new Docker container is required for this task.

Reuse existing MySQL Docker setup:

Refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/MySQL.md`

Section:
`6. Docker Setup, If Possible`

Current Docker/DevOps status:

| Item | Status | Task-specific note |
|---|---|---|
| MySQL container | Optional and reused | Must contain migration `005` seed rows |
| Service Dockerfile | Missing | No new Dockerfile added by current task |
| docker-compose service | Missing | No compose update exists for settings |
| Redis container | Not required | Current cache is in-memory |
| Message broker container | Not required | Current event publisher is log-only |
| Typesense container | Not required here | Search Service owns Typesense setup |

If a Compose stack is added later, include these task-specific env values in the service environment:

```yaml
environment:
  SUPERADMIN_PLATFORM_SETTINGS_CACHE_TTL: 5m
  SUPERADMIN_PLATFORM_SETTINGS_EVENT_TOPIC: platform.settings.updated
```

## 9. Local Development Setup

### Step 1: Read reused setup first

Read these first to avoid duplicate setup confusion:

```text
TaskImplementation/{SERVICE_NAME}/Dependency/Go_Modules.md
TaskImplementation/{SERVICE_NAME}/Dependency/MySQL.md
TaskImplementation/{SERVICE_NAME}/Dependency/Migrations.md
TaskImplementation/{SERVICE_NAME}/Dependency/Environment.md
TaskImplementation/{SERVICE_NAME}/Dependency/HTTP_API.md
```

### Step 2: Enter the backend service directory

```bash
cd backend/services/superadmin-service
```

### Step 3: Install dependencies

No new dependency install is required. Use the existing Go module flow:

```bash
go mod download
```

### Step 4: Start MySQL and apply migrations

Use the reused MySQL setup and apply all migrations. Confirm migration `005` created settings rows:

```bash
mysql -h 127.0.0.1 -P 3306 -u superadmin -p -D superadmin_db -e \
"SELECT setting_key, version FROM platform_settings ORDER BY setting_key;"
```

### Step 5: Provision an active local admin

Use the existing admin provisioning guidance. For full platform setting updates, use an active admin with `superadmin` role because `platform:settings:write` is critical.

For search synonym write testing, `catalog_admin` can be used if that admin exists in `admin_users` and migration `002` role permissions are applied.

### Step 6: Configure only task-specific env values

Add or confirm:

```bash
SUPERADMIN_PLATFORM_SETTINGS_CACHE_TTL=5m
SUPERADMIN_PLATFORM_SETTINGS_EVENT_TOPIC=platform.settings.updated
```

Do not add Redis, broker, Typesense, CMS, or Search URLs unless their real integration code is implemented.

### Step 7: Run tests

```bash
go test ./internal/domain ./internal/rbac ./internal/usecase ./internal/transport/http
go mod verify
```

### Step 8: Start backend service

```bash
set -a
source .env
set +a
go run ./cmd/server
```

### Step 9: Verify health and readiness

```bash
curl http://127.0.0.1:8088/healthz
curl http://127.0.0.1:8088/readyz
```

### Step 10: Verify current task APIs

List settings:

```bash
curl http://127.0.0.1:8088/api/v1/admin/settings \
  -H "X-Admin-Id: superadmin_1" \
  -H "X-User-Id: user_superadmin_1" \
  -H "X-Admin-Roles: superadmin" \
  -H "X-Session-Id: sess_1" \
  -H "X-Request-Id: req_settings_list_1"
```

List search synonyms:

```bash
curl http://127.0.0.1:8088/api/v1/admin/search/synonyms \
  -H "X-Admin-Id: superadmin_1" \
  -H "X-User-Id: user_superadmin_1" \
  -H "X-Admin-Roles: superadmin" \
  -H "X-Session-Id: sess_1" \
  -H "X-Request-Id: req_synonyms_list_1"
```

Update maintenance mode requires reason, request ID, and MFA:

```bash
curl -X PATCH http://127.0.0.1:8088/api/v1/admin/settings/maintenance_mode \
  -H "Content-Type: application/json" \
  -H "X-Admin-Id: superadmin_1" \
  -H "X-User-Id: user_superadmin_1" \
  -H "X-Admin-Roles: superadmin" \
  -H "X-Session-Id: sess_1" \
  -H "X-Request-Id: req_maintenance_update_1" \
  -H "X-MFA-Verified: true" \
  -d '{
    "version": 1,
    "reason": "scheduled maintenance window",
    "value": {
      "enabled": true,
      "message": "Scheduled maintenance is active",
      "starts_at": "2026-06-01T01:00:00Z",
      "ends_at": "2026-06-01T02:00:00Z",
      "allow_admins": true,
      "allow_health_checks": true
    }
  }'
```

Upsert search synonym requires reason and request ID, but not MFA:

```bash
curl -X POST http://127.0.0.1:8088/api/v1/admin/search/synonyms \
  -H "Content-Type: application/json" \
  -H "X-Admin-Id: catalog_1" \
  -H "X-User-Id: user_catalog_1" \
  -H "X-Admin-Roles: catalog_admin" \
  -H "X-Session-Id: sess_2" \
  -H "X-Request-Id: req_synonym_update_1" \
  -d '{
    "version": 1,
    "root": "mobile",
    "synonyms": ["phone", "smartphone"],
    "reason": "catalog search synonym improvement"
  }'
```

## 10. Running the Project

Use the same project run flow from previous dependency docs. Current task adds these success criteria:

| Check | Expected result |
|---|---|
| `/healthz` | `{"status":"ok"}` |
| `/readyz` | `{"status":"ready"}` when DB is reachable |
| `GET /api/v1/admin/settings` | Returns four settings after migration `005` |
| `PATCH /api/v1/admin/settings/{key}` | Returns updated setting with incremented `version` |
| `GET /api/v1/admin/search/synonyms` | Returns synonym snapshot and version |
| `POST /api/v1/admin/search/synonyms` | Upserts normalized root/synonym values |
| Service logs | Contains `platform setting updated event` with configured topic |
| MySQL `admin_audit_logs` | Contains `platform_setting.update` if DB-backed audit is configured |

Ports and networking:

| Service | Port | Purpose | Status |
|---|---:|---|---|
| Backend HTTP API | `8088` | Main admin HTTP routes | Reused |
| MySQL | `3306` | RBAC, settings, audit storage | Reused |
| Redis | `6379` | Not used by current service code | Not required |
| Kafka / RabbitMQ / NATS | N/A | No current broker publisher | Not required |
| Typesense | N/A | Owned by Search Service, not current service | External |

For port conflicts and Docker networking basics, refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`10. Ports & Networking`

## 11. Common Errors & Fixes

Generic Go, MySQL, Docker, `.env`, readiness, and admin header errors are already documented.

Refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/Go_Modules.md`

Section:
`11. Common Errors And Fixes`

Refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/MySQL.md`

Section:
`11. Common Errors And Fixes`

Refer:
`TaskImplementation/{SERVICE_NAME}/Dependency/HTTP_API.md`

Section:
`11. Common Errors And Fixes`

### Task-specific errors

| Error / symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `PLATFORM_SETTING_NOT_FOUND` | Migration `005` not applied or seed rows missing | Apply migration `005` and verify `platform_settings` rows | Run complete migrations in order |
| `SETTING_VERSION_CONFLICT` | Request `version` is old compared to DB version | Re-fetch setting, use latest version, retry intentionally | UI should send latest version from GET response |
| `MFA_REQUIRED` on settings update | `platform:settings:write` requires `X-MFA-Verified: true` | Add MFA verified header through trusted gateway | Enforce MFA in admin login before critical settings |
| `REQUEST_CONTEXT_MISSING` | Write request missing `X-Request-Id` | Send unique request ID | Gateway should generate request ID for every mutation |
| `REASON_REQUIRED` | Reason missing or shorter than 10 chars | Send meaningful reason, max 512 chars | UI should require reason textarea before submit |
| `VALIDATION_FAILED` for maintenance | Message/timestamps/allow flags invalid | Use RFC3339 times, end after start, keep admin and health checks allowed | Validate form client-side too |
| `VALIDATION_FAILED` for commission | BPS outside `0..5000`, non-INR currency, duplicate override IDs, bad timestamp | Fix payload schema | Keep commission editor schema-driven |
| `VALIDATION_FAILED` for feature flags | Unknown flag, rollout outside `0..100`, unknown role | Use known flag names and roles from code | Add reviewed registry before adding new flags |
| `VALIDATION_FAILED` for search synonyms | Empty root, duplicate terms, root equals synonym, too many entries | Normalize values and remove duplicates | UI should trim/lowercase preview before submit |
| Update succeeds but downstream behavior does not change | Current publisher logs only; Search/CMS/Gateway consumers are not wired | Implement real consumer/publisher integration in relevant services | Do not assume DB save means downstream applied |
| Logs show event topic but no broker message exists | `LoggingSettingsEventPublisher` is log-only | Add real broker publisher when architecture requires it | Document broker URL/topic/credentials in a future dependency guide |

## 12. Security & Best Practices

### Task-specific security audit

| Severity | Finding | Suggested fix |
|---|---|---|
| Critical for public deployment | Current service trusts admin headers after `ActorMiddleware`; it does not validate JWT/session itself | Keep service private behind trusted API Gateway, strip client-supplied admin headers, use service-to-service trust |
| Critical | `platform:settings:write` can affect revenue, availability, checkout, and search behavior | Keep MFA, reason, request ID, RBAC, and audit mandatory |
| High | Settings event publisher is log-only, so downstream invalidation is not guaranteed | Add real broker or consumer integration before depending on instant propagation |
| High | Current cache is per-process memory; multiple replicas can have stale settings until TTL expires | Use real invalidation via broker/Redis or keep TTL conservative |
| High | `.env` contains real-looking DB credentials | Use `.env.example` with placeholders and keep real secrets in ignored files or secret manager |
| Medium | No service Dockerfile or Compose stack exists | Add container setup when deployment/local stack requires it |
| Medium | No Typesense/CMS/Gateway consumer wiring in this service | Track those as separate integration tasks |
| Medium | Feature flag names are hardcoded in Go registry | Add reviewed registry process for new flags |

### Task-specific best practices

- Always update settings through API, not manual SQL, because API performs validation, versioning, audit, and event logging.
- Always send `version` from latest GET response. Ye accidental overwrite prevent karta hai.
- Keep `reason` human-readable: "scheduled maintenance window" is better than "update".
- Do not reduce `allow_admins` or `allow_health_checks` for maintenance mode; current validation blocks this for safety.
- Treat commission changes as finance-impacting. Review rate BPS, effective date, category override, and seller override carefully.
- Treat feature flags like production config. Unknown flag names should require code review before rollout.
- Treat search synonyms as catalog quality config. Duplicate, vague, or overly broad synonyms search quality ko hurt kar sakte hain.
- Watch logs for `platform setting updated event` after every successful mutation.

## 13. Missing or Misconfigured Things

Task-specific current gaps:

- [ ] No real Redis-backed platform settings cache.
- [ ] No real Kafka/RabbitMQ/NATS publisher for `platform.settings.updated`.
- [ ] No Search Service consumer that applies `search_synonyms` to Typesense from this service's event.
- [ ] No CMS/Order/Payment consumer documented here for `commission_rules`.
- [ ] No API Gateway enforcement implementation for `maintenance_mode`.
- [ ] No frontend platform settings UI in this backend task.
- [ ] No service Dockerfile or full local Compose stack.
- [ ] No `.env.example` with safe placeholders.
- [ ] No direct JWT/JWKS/Auth Service verification in this service.
- [ ] No cross-replica cache invalidation mechanism.

These are not blockers for local settings API verification, but they matter before production rollout.

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `task1_Dependency.md` | Tech stack, required software, ports, base run flow, generic troubleshooting | Same Go HTTP service, MySQL, Docker, and curl setup |
| `task2_Dependency.md` | MySQL decision and platform data ownership | Same database choice for structured admin/control-plane data |
| `task3_Dependency.md` | RBAC permissions, admin headers, exact permission behavior | Settings APIs depend on existing RBAC model |
| `task4_Dependency.md` | High-risk mutation context, audit direction, downstream boundary style | Settings writes reuse reason/request/audit pattern |
| `task5_Dependency.md` | Order/payment downstream boundary and high-risk control notes | Commission settings are consumed by financial/order workflows later |
| `task6_Dependency.md` | Architecture-mentioned services vs current direct dependencies | Redis/broker/analytics style boundary reused |
| `Dependency/Go_Modules.md` | Go module setup and errors | No new Go package added |
| `Dependency/MySQL.md` | MySQL install, Docker, DSN, credentials, common DB errors | Same MySQL database and credentials |
| `Dependency/Migrations.md` | Migration tooling and ordered migration flow | Current task only adds migration `005` details |
| `Dependency/Environment.md` | `.env` loading and common environment variables | Only two task-specific variables are new |
| `Dependency/HTTP_API.md` | HTTP server setup and admin request headers | Settings APIs use same transport/header contract |

Full setup references:

```text
TaskImplementation/{SERVICE_NAME}/Dependency/Go_Modules.md
TaskImplementation/{SERVICE_NAME}/Dependency/MySQL.md
TaskImplementation/{SERVICE_NAME}/Dependency/Migrations.md
TaskImplementation/{SERVICE_NAME}/Dependency/Environment.md
TaskImplementation/{SERVICE_NAME}/Dependency/HTTP_API.md
```

## 15. Final Checklist

- [ ] Previous dependency documentation checked before starting.
- [ ] No duplicate Go/MySQL/Docker setup copied into this task guide.
- [ ] No new Go module installed for platform settings.
- [ ] MySQL is running.
- [ ] Complete migrations applied, including `005_create_platform_settings.up.sql`.
- [ ] `platform_settings` has four default rows.
- [ ] Active local admin exists with required role/permissions.
- [ ] `SUPERADMIN_PLATFORM_SETTINGS_CACHE_TTL` confirmed or default accepted.
- [ ] `SUPERADMIN_PLATFORM_SETTINGS_EVENT_TOPIC` confirmed or default accepted.
- [ ] No Redis/broker/Typesense setup forced for current local verification.
- [ ] Backend service started on `HTTP_ADDR`.
- [ ] `/healthz` and `/readyz` verified.
- [ ] `GET /api/v1/admin/settings` verified.
- [ ] `PATCH /api/v1/admin/settings/{key}` verified with version, reason, request ID, and MFA.
- [ ] `GET /api/v1/admin/search/synonyms` verified.
- [ ] `POST /api/v1/admin/search/synonyms` verified with version, reason, and request ID.
- [ ] Service logs checked for `platform setting updated event`.
- [ ] Audit log behavior checked when DB-backed audit is enabled.
- [ ] Production gaps tracked separately: real gateway auth, cross-replica cache invalidation, real event broker, Search/CMS consumers, and safe secret handling.
