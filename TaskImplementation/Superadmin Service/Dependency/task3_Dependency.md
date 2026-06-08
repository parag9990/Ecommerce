# Project Dependency & Setup Guide

> Incremental beginner-friendly handbook for the Admin RBAC task. Shared project setup is referenced from earlier dependency guides so it is not duplicated here.

## Document Variables

| Variable | Meaning |
|---|---|
| `SERVICE_NAME` | Service folder value supplied by the task request |
| `TASK_FILE_NAME` | Current implementation task filename supplied by the task request |
| `INPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | Replace `.md` in `TASK_FILE_NAME` with `_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}` |

## 1. Project Overview

`INPUT_FILE_PATH` defines Admin RBAC: kaunsa admin role kaunsi exact permission use kar sakta hai, aur sensitive actions ke liye kaunsa extra security context required hai.

### Current implementation reality

The task document describes a documentation-only future design, but the current repository now contains a runnable Go RBAC implementation:

```text
backend/services/superadmin-service/
├── internal/domain/permission.go
├── internal/rbac/
├── internal/repository/mysql_permission_repository.go
├── internal/usecase/authorization.go
├── internal/transport/http/
└── migrations/
```

Current authorization flow:

```text
Trusted API Gateway / internal caller
        |
        | HTTP admin-context headers
        v
Actor middleware
        |
        v
Authorization service
        |
        v
MySQL admin_users + admin_role_permissions + admin_permissions
```

> **Important:** Current service JWT validate nahi karti. Auth Service and API Gateway production security boundary hain, but their runnable configuration is not present in this service folder.

### Scope of this guide

- Existing clone, Go, MySQL, Docker, `.env`, and normal run steps ko reference karta hai.
- RBAC-specific database seed, admin provisioning, request metadata, and verification explain karta hai.
- Suggested-but-unused JWT, Gin, gRPC, Redis, and migration tooling ko clearly separate karta hai.
- Business workflows or original task document ko modify nahi karta.

## 2. Tech Stack

### Task-specific technologies

| Technology / component | Required? | Why used | Simple Hinglish explanation |
|---|---|---|---|
| RBAC | Required | Exact admin permission enforcement | RBAC me role ek permission group hota hai. Service final access decision admin ke DB role aur requested permission se leti hai. |
| Go domain permission registry | Required | Valid roles, permission keys, descriptions, and risk levels define karta hai | Code ko pata rehta hai kaunsi permissions valid hain. |
| Go `net/http` actor middleware | Required | Trusted HTTP headers ko `AdminActor` context me convert karta hai | Gateway se aayi admin identity request ke context me attach hoti hai. |
| MySQL RBAC tables | Required for functional RBAC | Active admin, role mapping, and permission catalog ka runtime source of truth | MySQL decide karta hai ki given `admin_id` active hai aur uske DB role ko requested permission mili hai ya nahi. |
| Auth Service + API Gateway | Required for secure production use; direct local testing me optional | JWT/session/MFA verify karke trusted admin context inject karna chahiye | Public client ke raw headers par trust nahi karna chahiye. Gateway pehle identity verify karega. |
| High-risk permission policy | Required for sensitive actions | Selected permissions ke liye reason, request ID, and sometimes MFA require karta hai | Sirf role enough nahi hai; sensitive action ke liye extra proof/context chahiye. |

### Already documented shared technologies

Go `1.26.3`, Go modules, MySQL driver, MySQL installation, Docker, and standard HTTP runtime already explained hain.

Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`2. Tech Stack`, `3. Required Software`, and `4. Dependency Management`

### Suggested packages that are not current dependencies

`INPUT_FILE_PATH` recommends future packages, but current `go.mod` contains only the MySQL driver and its indirect dependency.

| Suggested item | Current status | Setup action now |
|---|---|---|
| `github.com/golang-jwt/jwt/v5` | Not installed; service does not parse JWTs | Do not install for the current implementation |
| `google.golang.org/grpc` / metadata | Not installed; current transport is HTTP | Do not install unless gRPC is actually implemented |
| `github.com/gin-gonic/gin` | Not installed; service uses `net/http` | Do not install |
| `github.com/stretchr/testify` | Not installed; tests use standard `testing` | Do not install |
| `golang-migrate/migrate` | Not configured in repository | Optional future tooling, not a current run requirement |
| Redis | Not implemented for permission caching | No Redis setup required |

Running the example `go get` commands from `INPUT_FILE_PATH` would add unused dependencies and create unnecessary `go.mod`/`go.sum` changes.

## 3. Required Software

No new software is introduced by this task.

| Software | Task requirement | Status |
|---|---|---|
| Go `1.26.3` | Build and run RBAC code/tests | Reused |
| MySQL `8.x` | Store active admins and permission mappings | Reused and mandatory for functional RBAC |
| MySQL client | Apply and verify RBAC migrations | Reused |
| curl | Verify protected RBAC HTTP endpoints | Reused |
| Docker | Optional local MySQL setup | Reused |

For installation and version checks, refer to:

`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`3. Required Software`

## 4. Dependency Management

### New dependency result

No new Go module is required for this task.

Current direct dependency:

```text
github.com/go-sql-driver/mysql v1.10.0
```

The MySQL driver and common Go commands are already documented.

Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`4. Dependency Management`

### Task-specific verification

From the service directory:

```bash
go test ./internal/domain ./internal/rbac ./internal/usecase ./internal/transport/http
go mod verify
```

These tests verify permission definitions, role matrix behavior, high-risk guards, authorization behavior, and HTTP actor handling. They do not test the real MySQL seed or an actual API Gateway.

## 5. Database Setup

### Reused MySQL setup

MySQL installation, Docker container setup, DSN, database creation, complete migration flow, local admin seed, and rollback warnings are already documented.

Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`5. Database Setup`, `Run migrations`, `Seed a local admin`, and `Verify database`

Also refer:
`TaskImplementation/{SERVICE_NAME}/task2_Dependency.md`

Sections:
`5. Database Setup` and `Source-of-truth rule`

### RBAC-specific tables

| Table | Required? | RBAC purpose |
|---|---|---|
| `admin_users` | Required | Maps `admin_id` to one DB role and active/disabled status |
| `admin_permissions` | Required | Stores the valid permission catalog |
| `admin_role_permissions` | Required | Maps each specialized role to exact permission keys |

### RBAC-specific migrations

| Migration | Purpose |
|---|---|
| `001_create_superadmin_rbac.up.sql` | Creates the three RBAC tables |
| `002_seed_admin_rbac.up.sql` | Seeds 20 permissions and specialized role mappings |

Use repository migration files as the executable source of truth. The conceptual SQL in `INPUT_FILE_PATH` should not be copied into the database.

### Admin role provisioning rule

Migration `002` seeds permission mappings, but it does **not** create an `admin_users` row. At least one active admin must be securely provisioned before protected RBAC endpoints can succeed.

Supported exact-permission roles:

```text
superadmin
operations_admin
finance_admin
catalog_admin
readonly_admin
```

The code also recognizes broad role `admin`, but that role intentionally has no exact permissions in the static matrix and no mappings in migration `002`. An `admin_users` row with role `admin` will therefore be denied by current MySQL RBAC.

For the local-only admin insert example, refer to:

`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`Seed a local admin`

### Verify RBAC seed

After migrations:

```bash
mysql -h 127.0.0.1 -P 3306 -u superadmin -p -D superadmin_db -e \
  "SELECT COUNT(*) AS permission_count FROM admin_permissions;"

mysql -h 127.0.0.1 -P 3306 -u superadmin -p -D superadmin_db -e \
  "SELECT role, COUNT(*) AS permission_count FROM admin_role_permissions GROUP BY role ORDER BY role;"

mysql -h 127.0.0.1 -P 3306 -u superadmin -p -D superadmin_db -e \
  "SELECT admin_id, user_id, role, status, mfa_required FROM admin_users ORDER BY admin_id;"
```

Expected initial seed direction:

| Role | Seeded exact permissions |
|---|---:|
| `superadmin` | 20 |
| `operations_admin` | 9 |
| `finance_admin` | 2 |
| `catalog_admin` | 6 |
| `readonly_admin` | 7 |
| `admin` | 0 |

### Permission-change deployment requirement

Adding or renaming a permission is a coordinated deployment, not only a DB edit. Update all applicable items:

1. Go permission constant and definition.
2. Static role-permission matrix.
3. A new versioned SQL migration for permission and role mappings.
4. Unit tests and real-MySQL integration tests.
5. Deployment notes and security review.

If DB seed and Go registry drift, permission-list endpoints can fail with an internal error for an unknown database permission.

## 6. Redis / Queue / External Services

### Auth Service and API Gateway

This task conceptually depends on Auth Service. Current service does not call Auth Service directly and has no Auth Service URL, JWT issuer, JWT audience, or JWKS environment variable.

Production integration must ensure:

- Auth Service issues valid admin identity/session claims.
- API Gateway validates token signature, expiry, issuer, audience, session, and MFA state.
- Gateway removes any client-supplied admin context headers.
- Gateway injects verified admin context before forwarding the request.
- Service port remains private and reachable only by trusted internal callers.

For the shared API Gateway warning, refer to:

`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`6. External Services` -> `API Gateway`

### Redis and queues

| Service | Task-specific status |
|---|---|
| Redis | Not used. `INPUT_FILE_PATH` mentions only a future permission-cache idea. |
| Kafka / RabbitMQ / NATS | Not used by RBAC. |
| User / Order / Payment services | Not needed to verify RBAC catalog/role endpoints. |
| gRPC | Not implemented in current service. |

No new external-service container or credential is required for this task.

## 7. Environment Variables

### New or changed variables

There are no new or changed environment variables for this task.

Use the complete existing `.env` guidance from:

`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`7. Environment Variables`

For functional RBAC, the important existing variable is:

```env
SUPERADMIN_DATABASE_DSN='superadmin:devpassword@tcp(127.0.0.1:3306)/superadmin_db?parseTime=true&charset=utf8mb4&loc=UTC'
```

No `.env` variable currently configures JWT verification, Auth Service, gateway trust, allowed proxy identities, issuer, audience, JWKS, or header-signing keys. This is a production security gap, not a missing local RBAC setup step.

### Request headers are not environment variables

Current HTTP middleware reads trusted request metadata:

| Header | Required by current service? | Purpose |
|---|---|---|
| `X-Admin-Id` | Yes | DB permission lookup identity |
| `X-User-Id` or `X-Subject-Id` | Yes | Authenticated user identity |
| `X-Admin-Roles` or `X-Roles` | Yes, at least one value | Broad route context; exact permission still comes from DB |
| `X-Session-Id` | Yes | Admin session context |
| `X-Request-Id` | Required for selected high-risk actions; recommended always | Trace and security context |
| `X-MFA-Verified` | Required as `true` for refund review and platform-setting writes | Trusted MFA result |
| `X-IP-Hash` | Optional | PII-safe source context |

> Incoming high-risk action reasons are generally read from request bodies. Downstream HTTP clients may forward a normalized reason as `X-Admin-Reason`. The `x-action-reason` gRPC metadata example in `INPUT_FILE_PATH` is not the current incoming HTTP contract.

## 8. Docker Setup

No new container, image, port mapping, volume, network, health check, or restart policy is introduced by this task.

Use the existing MySQL Docker setup from:

`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`Recommended Docker setup`, `Docker Compose example for MySQL`, and `9. Docker Setup`

Current repository still has no service Dockerfile or complete Compose stack. RBAC-specific Docker requirement is only that the application can reach the MySQL database containing migrations `001` and `002`.

## 9. Ports & Networking

No port is new or changed.

| Service / component | Port | Purpose | Status |
|---|---:|---|---|
| Current backend HTTP server | `8088` | RBAC and other admin HTTP APIs | Reused |
| MySQL | `3306` | Admin identity and permission lookup | Reused |
| Auth Service | Not configured here | Issues/verifies authentication state through the platform auth architecture | External dependency, no current service config |
| API Gateway | Not configured here | Validates authentication and injects trusted headers | External dependency, no current service config |

For conflicts, firewall guidance, and Docker networking, refer to:

`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`8. Ports & Networking`

> **Production warning:** Port `8088` ko public internet par expose mat karo. Otherwise a caller trusted admin and MFA headers forge kar sakta hai.

## 10. Local Development Setup

### Step 1: Read reused setup first

Follow:

`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`5. Database Setup`, `10. Local Development Setup`, and `11. Running the Project`

Also read:

`TaskImplementation/{SERVICE_NAME}/task2_Dependency.md`

Section:
`5. Database Setup`

### Step 2: Enter the service directory

```bash
cd backend/services/superadmin-service
```

### Step 3: Install dependencies

No new package install is required. Use the existing `go mod download` flow from the previous guide.

### Step 4: Start MySQL and apply migrations

Use the previous guide's MySQL and complete migration flow. Confirm that migrations `001` and `002` were applied.

### Step 5: Provision an active local admin

Use the local-only insert from the previous guide and select one of the seeded specialized roles. For the RBAC catalog endpoints below, use `superadmin` because those endpoints require `admin:users:read`.

### Step 6: Configure environment

Load the existing DSN-based `.env`. There are no task-specific variables to add.

### Step 7: Run tests and start

Use the previous guide's test, build, and start commands.

### Step 8: Verify RBAC database state

Run the RBAC seed verification queries from section 5 of this file.

### Step 9: Verify permission catalog

```bash
curl -i \
  -H 'X-Admin-Id: admin_local' \
  -H 'X-User-Id: user_local' \
  -H 'X-Admin-Roles: superadmin' \
  -H 'X-Session-Id: session_local_001' \
  -H 'X-Request-Id: request_local_001' \
  http://127.0.0.1:8088/api/v1/admin/rbac/permissions
```

Expected result: HTTP `200` and the permission catalog, provided `admin_local` exists, is active, and its DB role has `admin:users:read`.

### Step 10: Verify one role mapping

```bash
curl -i \
  -H 'X-Admin-Id: admin_local' \
  -H 'X-User-Id: user_local' \
  -H 'X-Admin-Roles: superadmin' \
  -H 'X-Session-Id: session_local_001' \
  -H 'X-Request-Id: request_local_002' \
  http://127.0.0.1:8088/api/v1/admin/rbac/roles/finance_admin/permissions
```

Expected role permissions:

```text
payments:read
payments:refund:review
```

### Step 11: Verify denial behavior

Change the DB role for a disposable local admin to a role without `admin:users:read`, or disable that admin, then repeat the catalog request. Expected result is HTTP `403`.

Do not use header changes alone as an exact permission test. Runtime MySQL authorization uses `X-Admin-Id` to load the admin's DB role; the role header is not the exact permission source of truth.

## 11. Running the Project

No new start command is introduced.

Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`10. Local Development Setup`, `11. Running the Project`, and `Quick Command Reference`

### RBAC-specific success criteria

```text
MySQL is reachable.
Migrations 001 and 002 are applied.
An active admin_users row uses a seeded specialized role.
The process has SUPERADMIN_DATABASE_DSN loaded.
The permission catalog and role-permission endpoints return expected results.
A disabled or unauthorized admin receives HTTP 403.
```

## 12. Common Errors & Fixes

Generic Go, MySQL, Docker, port, `.env`, and migration errors are already documented in:

`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`12. Common Errors & Fixes`

Only RBAC-specific issues are listed here.

| Error / symptom | Cause | Fix | Prevention |
|---|---|---|---|
| Protected request returns `401 ADMIN_CONTEXT_MISSING` | One of admin ID, user ID, roles, or session ID headers is missing | Send all required trusted headers for local testing | Define and test the gateway-to-service identity contract |
| Invalid/missing bearer token is not independently rejected by this service | Current service does not parse or validate JWTs | Put service behind a gateway that validates JWT/session state | Keep direct service port private and add authenticated service-to-service trust |
| Request header says `superadmin`, but response is `403 FORBIDDEN` | Exact permission comes from the active `admin_users` DB role, not the role header | Check `admin_users` and `admin_role_permissions` for the supplied `admin_id` | Provision admins through a controlled workflow |
| Role `admin` receives `403` | Broad `admin` role has no exact permission mappings | Use a seeded specialized role or deliberately add reviewed mappings | Document broad versus exact roles clearly |
| Every protected RBAC request returns `403` in local mode | DSN is empty, so startup uses an empty static deny repository | Configure MySQL, load DSN, apply migrations, and provision an admin | Set `SUPERADMIN_REQUIRE_DATABASE=true` for functional local profiles |
| Permission catalog returns internal error for an unknown permission | DB contains a permission not recognized by the Go registry | Reconcile code and migration, then redeploy safely | Add code/seed parity tests |
| Expected role permission is absent | Migration `002` missing, modified, or not aligned with code matrix | Inspect DB seed and apply a reviewed follow-up migration | Never manually patch production RBAC without versioned change control |
| High-risk action returns `MFA_REQUIRED` | Trusted actor context does not have verified MFA | Complete MFA at Auth/Gateway layer; local test may use trusted header only on private localhost | Never accept public client MFA headers |
| High-risk action returns `REASON_REQUIRED` or `REQUEST_CONTEXT_MISSING` | Required reason body or request ID missing | Send a valid reason and `X-Request-Id` | Enforce request schema and request-ID injection at gateway |
| `go get` adds JWT/Gin/gRPC packages but code still uses none of them | Future examples from `INPUT_FILE_PATH` were treated as current setup | Revert unintended dependency additions if approved and use current `go.mod` | Install packages only when implementation imports them |

## 13. Security & Best Practices

### Task-specific security audit

| Severity | Finding | Recommended fix |
|---|---|---|
| Critical for public deployment | Service trusts admin ID, roles, session, request ID, and MFA HTTP headers without validating JWTs | Keep it private behind a trusted gateway; strip client headers; authenticate gateway-to-service traffic |
| High | No JWT issuer, audience, JWKS, Auth Service URL, or gateway trust configuration exists in this service | Implement and document the intended trust model before production |
| High | `X-MFA-Verified` is a plain trusted header | Only a verified gateway should set it; never expose direct service access |
| High | Incoming role headers are not reconciled with the admin's DB role before being forwarded downstream | Gateway must inject verified roles; preferably derive/validate forwarded roles against authoritative admin state |
| High | No secure admin provisioning or role-change workflow is present | Add audited provisioning, role review, MFA enrollment, and break-glass controls |
| High | Permission seed and Go role matrix are separate definitions and can drift | Add a parity test against migration seed or generate both from one reviewed source |
| Medium | `admin_users.mfa_required` exists, but current authorization checks only the trusted `X-MFA-Verified` actor value | Define how the DB flag and verified session MFA state interact, then enforce and test that policy |
| Medium | `admin_users.role` has no database constraint limiting known roles | Add a safe role registry/constraint or validate every provisioning/update path |
| Medium | Real-MySQL RBAC integration tests are missing | Test migrations, active/disabled admins, every role, and permission denial in CI |
| Medium | Gateway/Auth integration tests are missing | Test invalid, expired, wrong-audience, revoked-session, and forged-header cases |
| Medium | Broad role `admin` is valid in code but has no exact permissions | Keep this deliberate behavior documented and tested |
| Medium | No permission cache exists | This is secure but may increase DB load; add only a short-TTL, revocation-aware cache if needed |

### Task-specific best practices

- Least privilege follow karo: specialized role ko only required permissions do.
- Permission changes ko code review, security review, and versioned migration ke through deploy karo.
- Production DB me direct manual role edits avoid karo.
- Disabled admins ko immediately deny hona chahiye.
- High-risk actions ke liye reason, request ID, audit record, and required MFA preserve karo.
- Admin and MFA headers ko public clients se accept mat karo.
- Permission denial logs monitor karo, but tokens, secrets, and raw sensitive data log mat karo.
- Break-glass access time-limited, approved, MFA-protected, and audited hona chahiye.
- Role/permission mapping changes ke baad allowed and denied cases dono test karo.

## 14. Missing or Misconfigured Things

Only task-specific gaps are listed here:

- [ ] No direct JWT validation or signed identity-context validation
- [ ] No Auth Service/JWKS/issuer/audience configuration
- [ ] No API Gateway implementation or trusted-header sanitization config in this service folder
- [ ] No authenticated gateway-to-service channel configuration
- [ ] No secure admin provisioning, role-change, or disable workflow
- [ ] Incoming role headers are not reconciled with the authoritative DB role before downstream forwarding
- [ ] `admin_users.mfa_required` is not used by current authorization checks
- [ ] No database constraint for known admin roles
- [ ] No automated parity check between Go permission matrix and migration seed
- [ ] No real-MySQL RBAC integration tests
- [ ] No gateway/Auth integration tests
- [ ] No documented emergency/break-glass admin process
- [ ] No permission change runbook or approval workflow
- [ ] Future JWT/Gin/gRPC/migration-tool examples in `INPUT_FILE_PATH` do not match current runtime dependencies
- [ ] Current HTTP implementation differs from the gRPC-oriented future structure in `INPUT_FILE_PATH`

## 15. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `task1_Dependency.md` | `2. Tech Stack` | Same Go, HTTP, MySQL driver, and runtime stack |
| `task1_Dependency.md` | `3. Required Software` | No new software introduced |
| `task1_Dependency.md` | `4. Dependency Management` | No new Go module introduced |
| `task1_Dependency.md` | `5. Database Setup` | MySQL installation, Docker, DSN, migrations, admin seed, and verification already complete |
| `task1_Dependency.md` | `6. External Services` | Existing API Gateway security boundary already explained |
| `task1_Dependency.md` | `7. Environment Variables` | No new or changed variables |
| `task1_Dependency.md` | `8. Ports & Networking` | No port changes |
| `task1_Dependency.md` | `9. Docker Setup` | No Docker changes |
| `task1_Dependency.md` | `10. Local Development Setup` and `11. Running the Project` | Same clone, test, build, and start flow |
| `task1_Dependency.md` | `12. Common Errors & Fixes` | Generic setup troubleshooting already covered |
| `task2_Dependency.md` | `5. Database Setup` | Same MySQL decision and migration source-of-truth rule |
| `Dependency/HTTP_API.md` | `8. Required Environment Variables` and admin headers | Same current HTTP transport contract |
| `Dependency/Migrations.md` | Migration setup and rollback | Same repository migration mechanism |
| `Dependency/MySQL.md` | MySQL dependency details | Same database installation and credentials |

Full reference paths:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
TaskImplementation/{SERVICE_NAME}/Dependency/HTTP_API.md
TaskImplementation/{SERVICE_NAME}/Dependency/Migrations.md
TaskImplementation/{SERVICE_NAME}/Dependency/MySQL.md
```

## 16. Final Checklist

- [ ] Previous dependency documentation read first
- [ ] No unused JWT, Gin, gRPC, Redis, or migration-tool dependency installed
- [ ] Existing Go modules downloaded and verified
- [ ] MySQL `8.x` running
- [ ] Complete repository migrations applied
- [ ] RBAC migrations `001` and `002` confirmed
- [ ] Permission and role seed counts verified
- [ ] Active local admin provisioned with a seeded specialized role
- [ ] Existing DSN environment loaded
- [ ] Service tests pass
- [ ] Service starts and health/readiness checks pass
- [ ] Permission catalog endpoint returns expected permissions
- [ ] Role-permission endpoint returns expected mapping
- [ ] Unauthorized or disabled admin receives `403`
- [ ] Required admin headers are supplied only by a trusted caller
- [ ] Production service port is private
- [ ] Gateway/Auth Service JWT, session, and MFA verification is planned or implemented
- [ ] Permission change and admin provisioning processes are audited
- [ ] Task-specific security gaps reviewed
- [ ] No duplicate shared setup documentation added
