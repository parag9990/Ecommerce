# Project Dependency & Setup Guide

> Incremental beginner-friendly onboarding handbook for the user/seller control task. Shared setup is referenced from previous dependency guides so duplicate installation documentation is avoided.

## Document Variables

| Variable | Meaning |
|---|---|
| `SERVICE_NAME` | Service folder value supplied by the task request |
| `TASK_FILE_NAME` | Current implementation task filename supplied by the task request |
| `INPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | Replace `.md` in `TASK_FILE_NAME` with `_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}` |

## 1. Project Overview

`INPUT_FILE_PATH` defines admin controls for:

- User list/search
- User block/unblock
- Seller list/search and KYC metadata view
- Seller approve/reject/suspend/reinstate
- User Service coordination
- RBAC, reason, request context, review-task closure, and audit recording

### Current implementation reality

The task document describes a future gRPC-oriented implementation guide, but the current repository now contains a runnable Go HTTP implementation:

```text
backend/services/superadmin-service/
├── cmd/server/main.go
├── internal/clients/user_service_http_client.go
├── internal/domain/user_seller_controls.go
├── internal/transport/http/control_handler.go
├── internal/usecase/user_seller_controls.go
├── internal/repository/mysql_review_task_repository.go
└── migrations/
```

Current runtime flow:

```text
Trusted admin caller / API Gateway
        |
        | HTTP + trusted admin headers
        v
Current service :8088
        |
        +--> MySQL :3306 for RBAC, review tasks, and audit logs
        |
        +--> User Service internal admin HTTP API
```

> **Important limitation:** `backend/services/user-service/` currently has no runnable source code, module, routes, or documented start command. Therefore current service unit tests can pass, but a real user/seller end-to-end flow requires an external User Service implementation or a contract-compatible test double.

### Task-specific dependency result

| Dependency area | Result |
|---|---|
| New Go package | None |
| New runtime/framework | None |
| New environment variable | None; existing User Service variables become mandatory for this flow |
| New database product | None |
| Task-relevant migrations | Existing migrations `001`, `002`, `003`, `004`, and `006` |
| New port | None |
| New Docker container | None |
| Redis / Kafka / RabbitMQ | Not used by current implementation |
| Mandatory downstream | User Service internal admin HTTP API |

## 2. Tech Stack

### Task-specific technologies

| Technology / component | Required? | Why used | Simple Hinglish explanation |
|---|---|---|---|
| Go `net/http` server | Required | Exposes current admin user/seller routes | Current code external framework ke bina Go standard HTTP server use karti hai. |
| Go HTTP client | Required | Calls User Service internal admin endpoints | Current service user/seller data directly own nahi karti; HTTP se User Service ko call karti hai. |
| RBAC | Required | Checks read and status-update permissions before downstream calls | RBAC decide karta hai kaunsa admin read ya mutation action kar sakta hai. |
| MySQL | Required for functional integrated setup | Stores admin permissions, review-task state, and audit records | User/seller main records yahan store nahi hote; current service ke control-plane records store hote hain. |
| JSON | Required, built into Go | Request, response, downstream payload, and audit snapshot format | Services structured data JSON me exchange karti hain. |
| API Gateway / trusted internal caller | Required for secure production use | Must authenticate admin and inject trusted headers | Public caller ke raw admin headers par directly trust karna unsafe hai. |
| User Service internal admin API | Required for every user/seller endpoint | Owns user, seller, and KYC source-of-truth data | User Service actual user/seller status read aur update karti hai. |

### Already documented shared technologies

Go installation, Go modules, MySQL driver, MySQL, Docker, HTTP runtime, migrations, environment loading, and RBAC are already explained.

Refer:

`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:

- `2. Tech Stack`
- `3. Required Software`
- `4. Dependency Management`
- `5. Database Setup`
- `6. External Services`

Also refer:

`TaskImplementation/{SERVICE_NAME}/task3_Dependency.md`

Sections:

- `2. Tech Stack`
- `5. Database Setup`
- `6. Redis / Queue / External Services`

### Planned items that are not current dependencies

| Item mentioned by `INPUT_FILE_PATH` | Current status | Setup action now |
|---|---|---|
| `google.golang.org/grpc` | Not present in `go.mod`; current downstream client uses HTTP | Do not install |
| Protobuf/generated code | Not present | Do not generate |
| `golang-migrate/migrate` | Optional local/CI tool, not configured as a project dependency | Use only if the team deliberately adopts it |
| Redis rate limiter | Not implemented | No Redis setup required |
| Superadmin frontend | Not part of the current backend folder | No frontend setup available for this task |

Running future-oriented `go get` commands from `INPUT_FILE_PATH` would add unused dependencies to the current implementation.

## 3. Required Software

No new software is introduced by this task.

| Software | Required for this task? | Status |
|---|---|---|
| Go `1.26.3` | Yes, for build/test/run | Reused |
| MySQL `8.x` and MySQL client | Yes, for functional RBAC, audit, and review-task behavior | Reused |
| curl | Recommended for route verification | Reused |
| Docker | Optional for local MySQL | Reused |
| Runnable User Service | Yes, for end-to-end user/seller APIs | Required but missing from this repository |

For installation and version checks, refer to:

`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:

`3. Required Software`

## 4. Dependency Management

### New dependency result

No new Go module is required.

Current direct dependency remains:

```text
github.com/go-sql-driver/mysql v1.10.0
```

Do not add gRPC, Gin, Redis, or migration CLI packages merely because they appear in the future design examples.

For full Go module setup and troubleshooting, refer to:

`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:

`4. Dependency Management`

Also refer:

`TaskImplementation/{SERVICE_NAME}/Dependency/Go_Modules.md`

### Task-specific verification

Run from the current backend service directory:

```bash
cd backend/services/superadmin-service

go mod verify
go test ./internal/domain ./internal/rbac ./internal/usecase ./internal/transport/http
go build -o /tmp/superadmin-service ./cmd/server
```

These commands verify current domain transitions, permission guards, control use cases, HTTP handlers, modules, and compilation. They do not verify a real User Service or real MySQL integration.

## 5. Database Setup

### Database ownership

MySQL setup is reused, but ownership must be clear:

| Data | Source of truth |
|---|---|
| Admin identity and exact permissions | Current service MySQL |
| Seller KYC review-task status | Current service MySQL `admin_review_tasks` |
| Admin mutation audit records | Current service MySQL `admin_audit_logs` |
| User profile and user status | User Service |
| Seller profile, seller status, and KYC document metadata | User Service |

Current service must not directly connect to or update the User Service database.

### Reused MySQL installation and credentials

Do not repeat MySQL installation, Docker, database creation, DSN, or local admin seed steps here.

Refer:

`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:

- `5. Database Setup`
- `Recommended Docker setup`
- `Run migrations`
- `Seed a local admin`
- `Verify database`

Also refer:

`TaskImplementation/{SERVICE_NAME}/task2_Dependency.md`

Section:

`5. Database Setup`

### Task-relevant migrations

| Migration | Required for this task | Why |
|---|---|---|
| `001_create_superadmin_rbac.up.sql` | Yes | Creates active admin and permission tables |
| `002_seed_admin_rbac.up.sql` | Yes | Seeds `users:*` and `sellers:*` permissions and role mappings |
| `003_create_admin_review_tasks.up.sql` | Required for complete seller KYC workflow | Creates review-task storage |
| `004_add_open_review_task_uniqueness.up.sql` | Required for complete seller KYC workflow | Prevents duplicate open review tasks for the same resource |
| `006_create_admin_audit_logs.up.sql` | Yes for DB-backed mutations | Stores required user/seller status audit records |
| `005_create_platform_settings.up.sql` | Not task-specific | Still apply it when following the complete ordered migration flow |

> Apply the complete repository migration sequence exactly once and in filename order. Migration `004` is not safely repeatable.

### Current task permission mappings

Exact authorization comes from the active admin's MySQL role mappings.

| Role | Users read | Users status update | Sellers read | Sellers status update |
|---|---|---|---|---|
| `superadmin` | Yes | Yes | Yes | Yes |
| `operations_admin` | Yes | Yes | Yes | Yes |
| `catalog_admin` | No | No | Yes | Yes |
| `readonly_admin` | Yes | No | Yes | No |
| `finance_admin` | No | No | No | No |
| `admin` | No | No | No | No |

> **Security review required:** `INPUT_FILE_PATH` describes `catalog_admin` as seller/KYC read-only, but current Go matrix and migration `002` grant `sellers:status:update`. Confirm whether that broader permission is intentional before production use.

### Task-specific database verification

```bash
mysql -h 127.0.0.1 -P 3306 -u superadmin -p -D superadmin_db -e \
  "SHOW TABLES LIKE 'admin_review_tasks'; SHOW TABLES LIKE 'admin_audit_logs';"

mysql -h 127.0.0.1 -P 3306 -u superadmin -p -D superadmin_db -e \
  "SELECT role, permission_key FROM admin_role_permissions WHERE permission_key IN ('users:read','users:status:update','sellers:read','sellers:status:update') ORDER BY role, permission_key;"

mysql -h 127.0.0.1 -P 3306 -u superadmin -p -D superadmin_db -e \
  "SELECT admin_id, role, status FROM admin_users ORDER BY admin_id;"

mysql -h 127.0.0.1 -P 3306 -u superadmin -p -D superadmin_db -e \
  "SHOW CREATE TABLE admin_review_tasks\G"
```

Look for:

- An active local admin
- Required user/seller permission mappings
- `admin_review_tasks`
- `open_resource_key`
- `uk_admin_review_tasks_open_resource`
- `admin_audit_logs`

### Review-task behavior

Current seller status workflow only **closes** an existing open `seller_kyc` review task when a `pending_review` seller is approved or rejected.

- It does not create the KYC review task.
- Missing review-task rows do not block seller status success.
- Review-task database errors are logged as warnings and do not fail the response.

Another workflow must create the open task before approval/rejection if review-task history is required.

## 6. Redis / Queue / External Services

### User Service: mandatory task-specific external service

User Service ek mandatory downstream dependency hai for every route in this task. It owns user/seller/KYC data and must implement the internal admin HTTP contract below.

| Current service operation | Required downstream request |
|---|---|
| List users | `GET {USER_SERVICE_ADMIN_BASE_URL}/users` |
| Read user before mutation | `GET {USER_SERVICE_ADMIN_BASE_URL}/users/{user_id}` |
| Update user status | `PATCH {USER_SERVICE_ADMIN_BASE_URL}/users/{user_id}/status` |
| List sellers/KYC metadata | `GET {USER_SERVICE_ADMIN_BASE_URL}/sellers` |
| Read seller before mutation | `GET {USER_SERVICE_ADMIN_BASE_URL}/sellers/{seller_id}` |
| Update seller status | `PATCH {USER_SERVICE_ADMIN_BASE_URL}/sellers/{seller_id}/status` |

If the configured URL has no path, the client automatically uses `/internal/admin`. If the URL already contains a path, that path is used as the base path.

Example configured value:

```text
http://127.0.0.1:8081/internal/admin
```

### Downstream contract requirements

The User Service must:

- Accept `http` or `https`.
- Return JSON matching current user/seller response structures for successful `GET` requests.
- Accept status mutation JSON containing `status` and `reason`.
- Return any `2xx` for a successful mutation; a response body is not required.
- Accept trusted admin context headers forwarded by the current service.
- Return stable `400`, `401`, `403`, `404`, `429`, and `5xx` statuses so errors map correctly.

Forwarded headers:

| Header | Forwarded? | Notes |
|---|---|---|
| `X-Admin-Id` | Yes | Acting admin ID |
| `X-User-Id` | Yes | Acting admin's user ID |
| `X-Admin-Roles` | Yes | Comes from incoming trusted actor context |
| `X-Session-Id` | Yes | Admin session context |
| `X-Request-Id` | Yes | Required for status mutations |
| `X-IP-Hash` | When present | PII-safe source context |
| `X-Admin-Reason` | Mutation only | Normalized reason |
| `Authorization` | No | Current client does not forward a bearer token |

> **Repository gap:** The required User Service internal admin endpoints are not implemented in the checked workspace. Exact User Service installation, database setup, and start commands therefore cannot be supplied from current repository evidence.

For shared downstream setup, refer to:

`TaskImplementation/{SERVICE_NAME}/Dependency/Downstream_HTTP_Services.md`

### API Gateway / trusted caller

Incoming routes trust admin identity headers. Secure production use requires a gateway or authenticated internal caller that verifies identity/session and strips forged client headers.

For the existing gateway and RBAC security explanation, refer to:

`TaskImplementation/{SERVICE_NAME}/task3_Dependency.md`

Sections:

- `6. Redis / Queue / External Services`
- `13. Security & Best Practices`

### Redis, queues, and gRPC

| Service / technology | Current task status |
|---|---|
| Redis | Not used |
| Kafka / RabbitMQ / NATS | Not used |
| gRPC | Planned in docs but not implemented |
| Rate limiter | Not implemented in current service |
| Notification service | Not called by current task implementation |

No container, credentials, or installation steps are required for these items.

### Ports and networking

No port is newly introduced or changed.

| Service / component | Port / address | Purpose | Status |
|---|---:|---|---|
| Current backend HTTP server | `8088` | Admin routes and health/readiness | Reused |
| MySQL | `3306` | RBAC, review tasks, and audit logs | Reused |
| User Service | Example `8081` | User/seller internal admin API | Reused, task-critical, implementation missing here |
| API Gateway | Not defined | Secure external entry point | Required for production, not configured here |

`8081` is an example from the local ignored `.env`, not a hardcoded User Service port in the Go client. The configured URL must match the real downstream service.

For port conflicts, firewall guidance, and Docker networking, refer to:

`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:

`8. Ports & Networking`

## 7. Environment Variables

### New or changed variables

There are no new or changed environment-variable names introduced by this task. Do not create another complete `.env` example.

Use the shared environment guide:

`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:

`7. Environment Variables`

Also refer:

`TaskImplementation/{SERVICE_NAME}/Dependency/Environment.md`

### Existing variables required for this task

| Variable | Current default | Task requirement | Purpose / security note |
|---|---|---|---|
| `SUPERADMIN_DATABASE_DSN` | Empty | Required for functional RBAC and durable audit | Contains credentials; keep secret |
| `SUPERADMIN_REQUIRE_DATABASE` | `false` | Recommended `true` for functional local testing | Fails startup if DB is missing |
| `USER_SERVICE_ADMIN_BASE_URL` | Empty | Required for all user/seller routes | Use trusted private network or HTTPS |
| `USER_SERVICE_TIMEOUT` | `5s` | Optional | Total downstream HTTP client timeout |
| `SUPERADMIN_REQUIRE_USER_SERVICE` | `false` | Recommended `true` for task-specific local testing | Fails startup if User Service URL is missing |
| `APP_ENV` | `local` | Use `local` for task-only development | Non-local values also force unrelated Order and Payment URLs |
| `HTTP_ADDR` | `:8088` | Required | Prefer `127.0.0.1:8088` for direct local testing |

Task-focused local delta:

```env
APP_ENV=local
SUPERADMIN_REQUIRE_DATABASE=true
USER_SERVICE_ADMIN_BASE_URL=http://127.0.0.1:8081/internal/admin
USER_SERVICE_TIMEOUT=5s
SUPERADMIN_REQUIRE_USER_SERVICE=true
```

These are existing variables, not new variables. Keep the existing DSN and shared HTTP values from the previous guide.

### Important loading behavior

The application does not auto-load `.env`. Source it before running:

```bash
cd backend/services/superadmin-service
set -a
source .env
set +a
```

### Incoming request headers

Headers are not environment variables, but they are mandatory runtime setup for direct local API verification.

| Header | Read routes | Status mutation routes |
|---|---|---|
| `X-Admin-Id` | Required | Required |
| `X-User-Id` or `X-Subject-Id` | Required | Required |
| `X-Admin-Roles` or `X-Roles` | Required | Required |
| `X-Session-Id` | Required | Required |
| `X-Request-Id` | Recommended | Required |
| `X-IP-Hash` | Optional | Optional |
| `X-MFA-Verified` | Not required by current task policy | Not required by current task policy |

`INPUT_FILE_PATH` recommends MFA for some user/seller controls, but current `users:status:update` and `sellers:status:update` policies require reason and request ID, not MFA.

## 8. Docker Setup

No new container, image, volume, network, health check, or restart policy is introduced.

Reuse the existing MySQL Docker setup:

`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:

- `Recommended Docker setup`
- `Docker Compose example for MySQL`
- `9. Docker Setup`

Current repository still has:

- No service Dockerfile
- No complete local Compose stack
- No User Service Dockerfile or Compose service
- No API Gateway Docker/route setup for this flow

If User Service runs in another container, `USER_SERVICE_ADMIN_BASE_URL` must use its container/network DNS name instead of `127.0.0.1`.

## 9. Local Development Setup

### Step 1: Read reused setup first

Follow:

`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:

- `5. Database Setup`
- `10. Local Development Setup`
- `11. Running the Project`

Also follow:

`TaskImplementation/{SERVICE_NAME}/task3_Dependency.md`

Sections:

- `5. Database Setup`
- `10. Local Development Setup`

### Step 2: Enter the current backend service directory

```bash
cd backend/services/superadmin-service
```

### Step 3: Install dependencies

No new package is required. Use the previous guide's `go mod download` flow.

### Step 4: Start MySQL and apply migrations

Use the reused MySQL setup and apply the complete ordered migration sequence. Confirm task-relevant migrations `001`, `002`, `003`, `004`, and `006`.

### Step 5: Provision an active local admin

Use the previous guide's local-only admin insert.

Recommended local role:

```text
superadmin
```

`operations_admin` can also read and mutate users/sellers. `readonly_admin` can read but cannot mutate.

### Step 6: Configure task-required existing environment values

Load the existing `.env`, then verify:

```bash
printenv SUPERADMIN_DATABASE_DSN
printenv USER_SERVICE_ADMIN_BASE_URL
printenv USER_SERVICE_TIMEOUT
printenv SUPERADMIN_REQUIRE_USER_SERVICE
```

Do not print a production DSN in shared logs or screenshots.

### Step 7: Provide a compatible User Service

Start a real external User Service or contract-compatible test double that implements the six required internal admin routes.

Current repository does not provide a User Service start command. Do not claim end-to-end readiness until this dependency exists and is reachable.

### Step 8: Verify User Service reachability

If the external service exposes health:

```bash
curl -i http://127.0.0.1:8081/healthz
```

Health endpoint availability is downstream-specific. Also verify at least one internal admin route using the trusted headers expected by that service.

### Step 9: Run tests and build

```bash
go mod verify
go test ./internal/domain ./internal/rbac ./internal/usecase ./internal/transport/http
go build -o /tmp/superadmin-service ./cmd/server
```

### Step 10: Start current service

```bash
go run ./cmd/server
```

### Step 11: Verify health, readiness, and RBAC

Use the shared `/healthz`, `/readyz`, and RBAC verification steps from the previous guides.

### Step 12: Verify task-specific database state

Run the verification queries from section 5 of this file.

## 10. Running the Project

No new start command is introduced.

Refer:

`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:

- `10. Local Development Setup`
- `11. Running the Project`
- `Quick Command Reference`

### Verify user list

```bash
curl -i \
  -H 'X-Admin-Id: admin_local' \
  -H 'X-User-Id: user_local' \
  -H 'X-Admin-Roles: superadmin' \
  -H 'X-Session-Id: session_local_001' \
  -H 'X-Request-Id: request_users_001' \
  'http://127.0.0.1:8088/api/v1/admin/users?status=active&page=1&page_size=20'
```

### Verify seller list / KYC metadata

```bash
curl -i \
  -H 'X-Admin-Id: admin_local' \
  -H 'X-User-Id: user_local' \
  -H 'X-Admin-Roles: superadmin' \
  -H 'X-Session-Id: session_local_001' \
  -H 'X-Request-Id: request_sellers_001' \
  'http://127.0.0.1:8088/api/v1/admin/sellers?status=pending_review&page=1&page_size=20'
```

### Verify a status mutation

Use only disposable local data whose current downstream status permits the requested transition:

```bash
curl -i -X PATCH \
  -H 'Content-Type: application/json' \
  -H 'X-Admin-Id: admin_local' \
  -H 'X-User-Id: user_local' \
  -H 'X-Admin-Roles: superadmin' \
  -H 'X-Session-Id: session_local_001' \
  -H 'X-Request-Id: request_user_block_001' \
  --data '{"status":"blocked","reason":"confirmed local integration test action"}' \
  http://127.0.0.1:8088/api/v1/admin/users/user_test_001/status
```

After success, verify both systems:

```bash
mysql -h 127.0.0.1 -P 3306 -u superadmin -p -D superadmin_db -e \
  "SELECT action, resource_type, resource_id, request_id, reason, created_at FROM admin_audit_logs WHERE request_id='request_user_block_001';"
```

Also verify `user_test_001` in the User Service through its supported admin/read interface.

### Task-specific success criteria

```text
MySQL is reachable and task-relevant migrations are applied.
An active admin has the required DB-backed permission.
User Service internal admin API is reachable and contract-compatible.
Read routes return valid User Service JSON.
Valid status mutation returns HTTP 200.
User Service source-of-truth status changes.
An admin_audit_logs row exists for the mutation.
Seller approval/rejection closes an existing matching open review task when present.
```

## 11. Common Errors & Fixes

Generic Go, MySQL, Docker, port, `.env`, and RBAC errors are already documented in:

`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:

`12. Common Errors & Fixes`

Only task-specific issues are listed here.

| Error / symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `503 DOWNSTREAM_UNAVAILABLE` with "user service is not configured" | `USER_SERVICE_ADMIN_BASE_URL` is empty | Set the URL, source `.env`, and restart | Set `SUPERADMIN_REQUIRE_USER_SERVICE=true` in task-focused local profiles |
| Startup fails because User Service URL is required | Strict flag or non-local environment is active | Configure a valid `http`/`https` URL | Validate env before deployment |
| Startup also requires Order/Payment URLs during task-only testing | Any non-local `APP_ENV` forces all downstream URLs | Use `APP_ENV=local` for isolated task development | Remove unrelated startup coupling in a future config change |
| User Service connection refused or timeout | Service missing, wrong host/port/path, firewall issue, or timeout too short | Start/fix downstream service and verify URL/network | Add dependency monitoring and contract tests |
| User Service returns invalid-response error | Successful list/detail response is empty or does not match expected JSON | Align downstream JSON schema | Add consumer-driven contract tests |
| Downstream `401` or `403` becomes Superadmin `503` | User Service denied forwarded admin context | Verify trusted headers and downstream auth policy | Add authenticated service-to-service contract and clearer error mapping |
| Read route returns `403 FORBIDDEN` before downstream call | Admin DB role lacks `users:read` or `sellers:read` | Check active admin and seeded mappings | Reuse audited RBAC provisioning |
| Mutation returns `400 REQUEST_CONTEXT_MISSING` | `X-Request-Id` missing | Send a non-empty request ID | Gateway should always inject request IDs |
| Mutation returns `400 REASON_REQUIRED` | Reason shorter than 10 characters or blank | Send a meaningful 10-512 character reason | Validate in admin UI/gateway |
| Mutation returns `400 INVALID_STATUS_TRANSITION` | Downstream current status does not allow requested transition | Fetch current resource and choose a valid local test transition | Keep state-transition contract synchronized |
| Audit insert foreign-key failure | `X-Admin-Id` does not exist in `admin_users` | Use/provision an active DB admin row | Test with a real provisioned admin |
| Mutation changed downstream state but response is `500` | Downstream mutation succeeded, then audit insert failed | Reconcile downstream state and audit record before retrying | Add idempotency and durable workflow/outbox |
| Seller status succeeds but review task remains open | Review-task close failed or no matching open `seller_kyc` task exists | Inspect logs/table and repair workflow state | Add integration tests and reliable task creation/closure |
| User Service implementation/start command cannot be found | It is absent from the checked repository | Obtain or implement the external service contract | Track it as a blocking dependency |
| Expected rate limit does not trigger | Current code has no rate limiter | Enforce at gateway or implement reviewed service limiter | Add explicit rate-limit tests and monitoring |

## 12. Security & Best Practices

### Task-specific security and configuration audit

| Severity | Finding | Recommended fix |
|---|---|---|
| Critical for public deployment | Incoming admin headers are trusted without JWT or signed-context verification | Keep port private behind a trusted gateway; strip client headers; authenticate gateway-to-service traffic |
| High | Current downstream User Service client sends trusted admin headers but no bearer token, service credential, or mTLS configuration | Add authenticated service-to-service transport and private networking/TLS |
| High | User/seller mutation policies require reason/request ID but do not require MFA, despite stronger direction in `INPUT_FILE_PATH` | Decide the required risk policy, then enforce and test it consistently |
| High | Current RBAC grants `catalog_admin` seller status-update permission, while `INPUT_FILE_PATH` describes seller mutation as limited to operations/superadmin roles | Confirm intent; remove the mapping through coordinated code and migration changes if it is too broad |
| High | Downstream mutation occurs before audit insert; audit failure returns an error after state already changed | Add idempotency, durable audit/outbox, reconciliation, and retry-safe semantics |
| High | Seller review-task close is best-effort and can silently leave workflow state open | Make failure observable and add reconciliation or transactional workflow coordination |
| High | Required User Service internal admin implementation is absent from this repository | Implement or supply the contract before production readiness |
| Medium | No consumer-driven contract tests verify User Service request/response schemas | Add contract tests for all six internal routes and error mappings |
| Medium | `/readyz` checks MySQL only, not User Service | Define integrated readiness or expose separate dependency health status |
| Medium | No retry, circuit breaker, or idempotency key exists for downstream mutations | Add carefully reviewed resilience behavior; never blindly retry non-idempotent mutations |
| Medium | Downstream `401`/`403` maps to `503`, reducing security/debug clarity | Define safe, explicit internal-auth error mapping |
| Medium | No task-specific rate limit exists | Add gateway rate limiting and alerting for admin mutations |
| Medium | Review-task creation owner is not implemented or documented in current runtime | Assign ownership and test task lifecycle end to end |
| Low | Task docs describe gRPC while current runtime uses HTTP | Update architecture contract or implement the intended transport |

### Task-specific best practices

- User/seller source-of-truth data ko current service DB me duplicate ya directly modify mat karo.
- Status mutation se pehle meaningful reason and unique request ID require karo.
- Admin headers sirf trusted gateway/internal caller se accept karo.
- Downstream internal admin APIs ko public internet par expose mat karo.
- User Service calls ke liye service authentication, TLS/mTLS, timeouts, and observability add karo.
- Request ID ko current service, User Service, logs, and audit records me preserve karo.
- Mutation retry se pehle downstream state check karo because response failure state-change ke baad ho sakta hai.
- Audit insert failures and review-task close warnings par alerts banao.
- User/seller/KYC raw PII, document URLs, tokens, or secrets logs me mat likho.
- Read and denied cases ke saath valid/invalid status transitions bhi test karo.

For shared security guidance, refer to:

`TaskImplementation/{SERVICE_NAME}/task3_Dependency.md`

Section:

`13. Security & Best Practices`

## 13. Missing or Misconfigured Things

Only task-specific gaps are listed here:

- [ ] No runnable User Service implementation in the checked repository
- [ ] No User Service internal admin route implementation or start command
- [ ] No User Service Docker/Compose setup
- [ ] No authenticated service-to-service credential, token, or mTLS configuration
- [ ] No consumer-driven User Service contract tests
- [ ] No real-MySQL plus real-User-Service integration test
- [ ] No User Service dependency in readiness checks
- [ ] No mutation idempotency key or retry-safe workflow
- [ ] Audit insert can fail after downstream state already changed
- [ ] Seller review-task closure is best-effort only
- [ ] Review-task creation ownership is not implemented in this flow
- [ ] No current task-specific rate limiter
- [ ] User/seller status mutations do not require MFA in current policy
- [ ] `catalog_admin` seller status-update permission differs from `INPUT_FILE_PATH`
- [ ] Incoming trusted roles are forwarded downstream without reconciliation against the authoritative DB role
- [ ] Downstream `401`/`403` is mapped to service unavailable
- [ ] No service Dockerfile or full local Compose stack
- [ ] Current HTTP implementation differs from the gRPC-oriented design in `INPUT_FILE_PATH`

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `task1_Dependency.md` | Tech stack and required software | Same Go, HTTP, MySQL, Docker, and curl setup |
| `task1_Dependency.md` | Dependency management | No new Go package |
| `task1_Dependency.md` | MySQL installation, Docker, DSN, migrations, and admin seed | Same database setup |
| `task1_Dependency.md` | Environment variables and `.env` loading | No new variable names |
| `task1_Dependency.md` | Ports, Docker, local run, and generic troubleshooting | No shared setup change |
| `task2_Dependency.md` | MySQL ownership and migration source of truth | Same database decision |
| `task3_Dependency.md` | RBAC permissions, admin provisioning, trusted headers, and gateway security | Same authorization setup |
| `Dependency/main_dependency.md` | Existing runtime architecture and missing-service audit | Same service-wide dependency map |
| `Dependency/Go_Modules.md` | Go module setup | No module change |
| `Dependency/MySQL.md` | MySQL setup and credentials | Same database |
| `Dependency/Migrations.md` | Complete migration flow | Same migration mechanism |
| `Dependency/Environment.md` | Existing User Service variables and environment loading | Same variables |
| `Dependency/HTTP_API.md` | HTTP server and incoming admin headers | Same transport |
| `Dependency/Downstream_HTTP_Services.md` | User Service base URL and shared downstream behavior | Same downstream client setup |

Full reference paths:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
TaskImplementation/{SERVICE_NAME}/Dependency/main_dependency.md
TaskImplementation/{SERVICE_NAME}/Dependency/Go_Modules.md
TaskImplementation/{SERVICE_NAME}/Dependency/MySQL.md
TaskImplementation/{SERVICE_NAME}/Dependency/Migrations.md
TaskImplementation/{SERVICE_NAME}/Dependency/Environment.md
TaskImplementation/{SERVICE_NAME}/Dependency/HTTP_API.md
TaskImplementation/{SERVICE_NAME}/Dependency/Downstream_HTTP_Services.md
```

## 15. Final Checklist

- [ ] Previous dependency documentation read first
- [ ] No unused gRPC, Gin, Redis, or migration-tool package installed
- [ ] Existing Go modules downloaded and verified
- [ ] Task-specific unit tests pass
- [ ] Current service build passes
- [ ] MySQL `8.x` running
- [ ] Complete repository migrations applied exactly once
- [ ] Migrations `001`, `002`, `003`, `004`, and `006` verified
- [ ] Active local admin provisioned
- [ ] Required user/seller permissions verified
- [ ] `catalog_admin` seller mutation access reviewed and approved or removed
- [ ] Existing `.env` loaded into the shell
- [ ] `SUPERADMIN_REQUIRE_DATABASE=true` used for functional local testing
- [ ] `USER_SERVICE_ADMIN_BASE_URL` points to a real compatible service
- [ ] `SUPERADMIN_REQUIRE_USER_SERVICE=true` used for task-focused local testing
- [ ] User Service internal admin routes are reachable
- [ ] `/healthz`, `/readyz`, and RBAC checks pass
- [ ] User list and seller list routes verified
- [ ] Valid user or seller status mutation verified with disposable data
- [ ] Downstream source-of-truth state verified after mutation
- [ ] Audit row verified after mutation
- [ ] Seller review-task closure verified when applicable
- [ ] Invalid transition, missing reason, missing request ID, and permission denial tested
- [ ] Production admin headers can only come from a trusted gateway
- [ ] Service-to-service authentication and transport security planned
- [ ] Task-specific security gaps reviewed
- [ ] No duplicate shared setup documentation added
