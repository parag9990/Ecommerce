# Project Dependency & Setup Guide

> Incremental beginner-friendly onboarding handbook for the order/payment controls task. Shared setup is referenced from previous dependency guides so duplicate installation documentation is avoided.

## Document Variables

| Variable | Meaning |
|---|---|
| `SERVICE_NAME` | Service folder value supplied by the task request |
| `TASK_FILE_NAME` | Current implementation task filename supplied by the task request |
| `INPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}` |
| `OUTPUT_FILE_NAME` | Replace `.md` in `TASK_FILE_NAME` with `_Dependency.md` |
| `OUTPUT_FILE_PATH` | `TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}` |

## 1. Project Overview

`INPUT_FILE_PATH` defines admin workflows for order and payment control:

- Admin order list/search
- Admin payment list/search
- Refund approve/reject review
- Manual order review queue creation
- Order dispute view aggregation
- RBAC, reason, request context, downstream admin calls, review-task storage, and audit recording

### Current implementation reality

The task document describes a future gRPC-oriented guide, but the current repository contains a runnable Go HTTP implementation:

```text
backend/services/superadmin-service/
|-- cmd/server/main.go
|-- internal/clients/order_payment_http_clients.go
|-- internal/domain/order_payment_controls.go
|-- internal/transport/http/order_payment_handler.go
|-- internal/usecase/order_payment_controls.go
|-- internal/repository/mysql_review_task_repository.go
|-- internal/repository/mysql_audit_log_repository.go
`-- migrations/
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
        +--> Order Service internal admin HTTP API
        |
        +--> Payment Service internal admin HTTP API
```

### Task-specific dependency result

| Dependency area | Result |
|---|---|
| New Go package | None |
| New runtime/framework | None |
| New environment variable names | None; existing Order/Payment variables become task-critical |
| New database product | None |
| Task-relevant migrations | Existing migrations `001`, `002`, `003`, `004`, and `006` |
| New port | None |
| New Docker container | None |
| Redis / Kafka / RabbitMQ | Not used by current implementation |
| Mandatory downstream | Order Service and Payment Service internal admin HTTP APIs |

> Beginner note: Is task me current service order/payment tables ka owner nahi hai. Order data Order Service se aata hai, refund/payment data Payment Service se aata hai, aur current service only admin-control state like review tasks and audit logs apni MySQL DB me store karti hai.

## 2. Tech Stack

### Task-specific technologies

| Technology / component | Required? | Why used | Simple Hinglish explanation |
|---|---|---|---|
| Go `net/http` server | Required | Exposes admin order/payment/refund routes | Current code external web framework ke bina Go standard HTTP server use karti hai. |
| Go HTTP clients | Required | Calls Order Service and Payment Service internal admin APIs | Current service doosre service ke DB ko direct touch nahi karti; HTTP se trusted admin request bhejti hai. |
| RBAC | Required | Checks `orders:*` and `payments:*` permissions before work | RBAC decide karta hai kaunsa admin order dekh sakta hai, manual review bana sakta hai, ya refund approve/reject kar sakta hai. |
| High-risk policy | Required | Requires reason/request ID, and MFA for refund review | Sensitive finance actions me sirf role enough nahi hota; extra context chahiye. |
| MySQL | Required for functional task run | Stores active admin permissions, manual review tasks, and audit records | MySQL structured and audit-sensitive admin data ke liye reliable storage hai. |
| JSON | Required, built into Go | HTTP request/response payloads and audit snapshots | Services structured data JSON format me exchange karti hain. |
| API Gateway / trusted internal caller | Required for secure production use | Authenticates admin and injects trusted headers | Public client ke raw admin headers par direct trust unsafe hai. Gateway verify karke headers set karega. |
| Order Service internal admin API | Required for order list, order detail, status history, manual review, dispute view | Order source of truth wahi service hai | Order ka actual status current service me store nahi hota. |
| Payment Service internal admin API | Required for payment list, refund read/review, refund lists, dispute view | Payment/refund source of truth wahi service hai | Refund provider/payment state current service direct update nahi karti. |

### Already documented shared technologies

Go installation, Go modules, MySQL driver, MySQL setup, Docker-for-MySQL, HTTP server basics, migrations, environment loading, and RBAC are already explained.

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
- `13. Security & Best Practices`

### Planned items that are not current dependencies

| Item mentioned in design docs or `INPUT_FILE_PATH` | Current status | Setup action now |
|---|---|---|
| gRPC / protobuf | Not present in `go.mod`; current runtime uses HTTP | Do not install unless implementation changes |
| Redis | Not used for this task | No Redis setup required |
| Kafka / RabbitMQ / NATS | Not used for this task | No queue setup required |
| Payment provider SDK | Not used by current service | Provider integration belongs inside Payment Service |
| Direct Order DB / Payment DB credentials | Not used and should not be added | Use downstream services only |
| Service Dockerfile / Compose service | Missing from repository | No task-specific Docker setup available |

## 3. Required Software

No new local software is introduced by this task.

| Software | Required for this task? | Status |
|---|---|---|
| Go `1.26.3` compatible toolchain | Yes, for build/test/run | Reused |
| MySQL `8.x` and MySQL client | Yes, for functional RBAC, manual review tasks, and audit logs | Reused |
| curl or similar HTTP client | Recommended for verification | Reused |
| Docker | Optional for local MySQL | Reused |
| Runnable Order Service | Yes for order list, manual review, and dispute view | Required but not implemented in this repository tree |
| Runnable Payment Service | Yes for payment list, refund review, and dispute view | Required but not implemented in this repository tree |

For installation and version checks, refer to:

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

Do not add gRPC, Redis, payment-provider SDKs, or migration CLI packages just because they appear in future design examples. Current code imports none of them.

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

These commands verify domain validation, permission guards, order/payment usecases, HTTP handlers, modules, and compilation. They do not verify real Order Service, real Payment Service, or real MySQL integration.

## 5. Database Setup

### Database ownership

MySQL setup is reused, but ownership must be clear:

| Data | Source of truth |
|---|---|
| Admin identity and exact permissions | Current service MySQL |
| Manual order review queue | Current service MySQL `admin_review_tasks` |
| Refund review task closure when a matching open task exists | Current service MySQL `admin_review_tasks` |
| Admin mutation audit records | Current service MySQL `admin_audit_logs` |
| Order status, amount, items, and status history | Order Service |
| Payment status and payment list | Payment Service |
| Refund record and provider refund workflow | Payment Service |

Current service must not directly connect to or update the Order Service database or Payment Service database.

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
| `002_seed_admin_rbac.up.sql` | Yes | Seeds `orders:*` and `payments:*` permissions and role mappings |
| `003_create_admin_review_tasks.up.sql` | Yes for manual order review | Creates review-task storage |
| `004_add_open_review_task_uniqueness.up.sql` | Yes for manual order review | Prevents duplicate open task for same task/resource |
| `006_create_admin_audit_logs.up.sql` | Yes for DB-backed mutations | Stores audit records for refund review and manual review |
| `005_create_platform_settings.up.sql` | Not task-specific | Still apply it when following complete ordered migration flow |

> Apply the complete repository migration sequence exactly once and in filename order. Migration `004` is not safely repeatable if manually re-run after success.

### Task permission mappings

Exact authorization comes from the active admin's MySQL role mappings, not from headers alone.

| Action | Permission | Seeded roles that should pass |
|---|---|---|
| List/search orders | `orders:read` | `superadmin`, `operations_admin`, `readonly_admin` |
| View order dispute context | `orders:read` | `superadmin`, `operations_admin`, `readonly_admin` |
| Mark order for manual review | `orders:manual_review:write` | `superadmin`, `operations_admin` |
| List/search payments | `payments:read` | `superadmin`, `finance_admin` |
| Approve/reject refund | `payments:refund:review` | `superadmin`, `finance_admin` |

### Task-specific database verification

```bash
mysql -h 127.0.0.1 -P 3306 -u superadmin -p -D superadmin_db -e \
  "SHOW TABLES LIKE 'admin_review_tasks'; SHOW TABLES LIKE 'admin_audit_logs';"

mysql -h 127.0.0.1 -P 3306 -u superadmin -p -D superadmin_db -e \
  "SELECT role, permission_key FROM admin_role_permissions WHERE permission_key IN ('orders:read','orders:manual_review:write','payments:read','payments:refund:review') ORDER BY role, permission_key;"

mysql -h 127.0.0.1 -P 3306 -u superadmin -p -D superadmin_db -e \
  "SELECT admin_id, role, status, mfa_required FROM admin_users ORDER BY admin_id;"

mysql -h 127.0.0.1 -P 3306 -u superadmin -p -D superadmin_db -e \
  "SHOW CREATE TABLE admin_review_tasks\G"
```

Look for:

- An active local admin
- Required `orders:*` and `payments:*` permission mappings
- `admin_review_tasks`
- `open_resource_key`
- `uk_admin_review_tasks_open_resource`
- `admin_audit_logs`

### Review-task behavior

Manual order review creates an open task with:

```text
task_type=order_manual_review
resource_type=order
resource_id={order_id}
status=open
```

The unique open-task constraint means duplicate open manual review tasks for the same order are rejected with `REVIEW_TASK_ALREADY_EXISTS`.

Refund review tries to close an existing open refund review task with:

```text
task_type=refund_review
resource_type=refund
resource_id={refund_id}
```

If the close fails, current code logs a warning and continues. The Payment Service refund decision and audit behavior are still handled separately.

## 6. Redis / Queue / External Services

### Order Service: task-critical downstream

Order Service is required for order list, order details, status history, manual order review validation, and dispute view.

| Current service operation | Required downstream request |
|---|---|
| List orders | `GET {ORDER_SERVICE_ADMIN_BASE_URL}/orders` |
| Read order before manual review | `GET {ORDER_SERVICE_ADMIN_BASE_URL}/orders/{order_id}` |
| Read order for dispute view | `GET {ORDER_SERVICE_ADMIN_BASE_URL}/orders/{order_id}` |
| Read order status history | `GET {ORDER_SERVICE_ADMIN_BASE_URL}/orders/{order_id}/status-history` |

List query parameters supported by current client:

```text
status
user_id
seller_id
page
page_size
cursor
```

Expected status values from current domain code:

```text
created
pending_payment
paid
packed
shipped
delivered
cancelled
refunded
payment_failed
```

### Payment Service: task-critical downstream

Payment Service is required for payment list, refund read/review, payment lookups by order, refund lists, and dispute view.

| Current service operation | Required downstream request |
|---|---|
| List payments | `GET {PAYMENT_SERVICE_ADMIN_BASE_URL}/payments` |
| List payments for an order | `GET {PAYMENT_SERVICE_ADMIN_BASE_URL}/payments?order_id={order_id}` |
| List refunds for a payment | `GET {PAYMENT_SERVICE_ADMIN_BASE_URL}/payments/{payment_id}/refunds` |
| Read refund before review | `GET {PAYMENT_SERVICE_ADMIN_BASE_URL}/refunds/{refund_id}` |
| Apply refund review | `POST {PAYMENT_SERVICE_ADMIN_BASE_URL}/refunds/{refund_id}/review` |

List query parameters supported by current client:

```text
status
provider
order_id
page
page_size
cursor
```

Expected payment status values:

```text
initiated
requires_action
authorized
captured
failed
refunded
partially_refunded
```

Expected refund status values:

```text
requested
approved
rejected
processing
succeeded
failed
```

Only refunds currently in `requested` state can be approved or rejected by this service. If a refund is already in the same final review state, the code treats the request as idempotent and records an idempotent audit action.

### Downstream base-path behavior

The current HTTP clients accept `http` or `https` base URLs.

If a configured base URL has no path, client code automatically uses:

```text
/internal/admin
```

Example valid values:

```text
ORDER_SERVICE_ADMIN_BASE_URL=http://127.0.0.1:8084/internal/admin
PAYMENT_SERVICE_ADMIN_BASE_URL=http://127.0.0.1:8085/internal/admin
```

Use the real local ports for your actual services. The current code does not hardcode Order/Payment ports.

### Downstream contract requirements

Order Service and Payment Service must:

- Accept trusted admin context headers forwarded by this service.
- Return JSON matching current Go domain structs.
- Return valid known status strings only.
- Return stable `400`, `401`, `403`, `404`, `409`, `412`, `429`, and `5xx` statuses so current error mapping behaves predictably.
- Keep internal admin APIs private.

Forwarded headers:

| Header | Forwarded? | Notes |
|---|---|---|
| `X-Admin-Id` | Yes | Acting admin ID |
| `X-User-Id` | Yes | Acting admin's user ID |
| `X-Admin-Roles` | Yes | Comes from incoming trusted actor context |
| `X-Session-Id` | Yes | Admin session context |
| `X-Request-Id` | Yes | Required for high-risk actions |
| `X-IP-Hash` | When present | PII-safe source context |
| `X-Admin-Reason` | Mutation only | Normalized reason for refund review |
| `Authorization` | No | Current client does not forward a bearer token |

> Repository gap: no runnable Order Service or Payment Service implementation/start command is present in the checked workspace. End-to-end testing needs external services or contract-compatible test doubles.

### Redis, queues, and gRPC

| Service / technology | Current task status |
|---|---|
| Redis | Not used |
| Kafka / RabbitMQ / NATS | Not used |
| gRPC | Planned by docs, not implemented in current code |
| Payment provider SDK | Not called by current service |
| Rate limiter | Not implemented in current service |

No container, credentials, broker topic, queue, or installation steps are required for these items.

## 7. Environment Variables

### New or changed variables

There are no new environment-variable names introduced by this task. Do not create another complete `.env` example.

Use the shared environment guide:

`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:

`7. Environment Variables`

Also refer:

`TaskImplementation/{SERVICE_NAME}/Dependency/Environment.md`

### Existing variables required for this task

| Variable | Current default | Task requirement | Purpose / security note |
|---|---|---|---|
| `SUPERADMIN_DATABASE_DSN` | Empty | Required for functional RBAC, review tasks, and audit | Contains credentials; keep secret |
| `SUPERADMIN_REQUIRE_DATABASE` | `false` | Recommended `true` for task-focused local testing | Fails startup if DB is missing |
| `ORDER_SERVICE_ADMIN_BASE_URL` | Empty | Required for order list, manual review, and dispute view | Use trusted private network or HTTPS |
| `ORDER_SERVICE_TIMEOUT` | `5s` | Optional | Total Order Service HTTP timeout |
| `SUPERADMIN_REQUIRE_ORDER_SERVICE` | `false` | Recommended `true` for task-focused local testing | Fails startup if Order URL is missing |
| `PAYMENT_SERVICE_ADMIN_BASE_URL` | Empty | Required for payment list, refund review, and dispute view | Use trusted private network or HTTPS |
| `PAYMENT_SERVICE_TIMEOUT` | `5s` | Optional | Total Payment Service HTTP timeout |
| `SUPERADMIN_REQUIRE_PAYMENT_SERVICE` | `false` | Recommended `true` for task-focused local testing | Fails startup if Payment URL is missing |
| `APP_ENV` | `local` | Use `local` for isolated task development | Non-local values force DB, User, Order, and Payment URLs |
| `HTTP_ADDR` | `:8088` | Required | Prefer `127.0.0.1:8088` for direct local testing |

Task-focused local delta:

```env
APP_ENV=local
SUPERADMIN_REQUIRE_DATABASE=true

ORDER_SERVICE_ADMIN_BASE_URL=http://127.0.0.1:8084/internal/admin
ORDER_SERVICE_TIMEOUT=5s
SUPERADMIN_REQUIRE_ORDER_SERVICE=true

PAYMENT_SERVICE_ADMIN_BASE_URL=http://127.0.0.1:8085/internal/admin
PAYMENT_SERVICE_TIMEOUT=5s
SUPERADMIN_REQUIRE_PAYMENT_SERVICE=true
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

Headers are not environment variables, but they are mandatory runtime setup for local API verification.

| Header | Read routes | Manual review | Refund review |
|---|---|---|---|
| `X-Admin-Id` | Required | Required | Required |
| `X-User-Id` or `X-Subject-Id` | Required | Required | Required |
| `X-Admin-Roles` or `X-Roles` | Required | Required | Required |
| `X-Session-Id` | Required | Required | Required |
| `X-Request-Id` | Recommended | Required | Required |
| `X-IP-Hash` | Optional | Optional | Optional |
| `X-MFA-Verified` | Not required | Not required by current policy | Required as `true` |

Reason rules:

```text
Minimum length: 10 characters
Maximum length: 512 characters
```

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
- No Order Service Dockerfile or Compose service
- No Payment Service Dockerfile or Compose service
- No API Gateway Docker/route setup for this flow

If Order Service or Payment Service runs in another container, use container/network DNS names instead of `127.0.0.1`:

```env
ORDER_SERVICE_ADMIN_BASE_URL=http://order-service:8084/internal/admin
PAYMENT_SERVICE_ADMIN_BASE_URL=http://payment-service:8085/internal/admin
```

Use the actual service names and ports from your Compose/Kubernetes setup.

## 9. Ports & Networking

No port is newly introduced or changed by this task.

| Service / component | Port / address | Purpose | Status |
|---|---:|---|---|
| Current backend HTTP server | `8088` via `HTTP_ADDR` | Admin routes and health/readiness | Reused |
| MySQL | `3306` | RBAC, review tasks, and audit logs | Reused |
| Order Service admin API | From `ORDER_SERVICE_ADMIN_BASE_URL` | Order list/detail/history | Existing variable, task-critical |
| Payment Service admin API | From `PAYMENT_SERVICE_ADMIN_BASE_URL` | Payment list/refund review/refund lists | Existing variable, task-critical |
| API Gateway | Not defined here | Secure external entry point | Required for production, not configured here |

For port conflicts, firewall guidance, and Docker networking, refer to:

`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:

`8. Ports & Networking`

Production reminders:

- Current backend port should stay private behind a trusted gateway.
- Order/Payment internal admin APIs should not be public.
- MySQL should not be exposed to the public internet.
- In Docker networks, use service DNS names, not `localhost`, when one container calls another.

## 10. Local Development Setup

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

And:

`TaskImplementation/{SERVICE_NAME}/task4_Dependency.md`

Sections:

- `6. Redis / Queue / External Services`
- `7. Environment Variables`

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

Recommended local roles:

| Testing goal | Recommended role |
|---|---|
| Verify every current-task route quickly | `superadmin` |
| Verify order list/manual review/dispute view only | `operations_admin` |
| Verify payment list/refund review only | `finance_admin` |
| Verify read-only order dispute access | `readonly_admin` |

Remember: exact permission is loaded from MySQL using `X-Admin-Id`. Changing only `X-Admin-Roles` does not grant permission.

### Step 6: Provide compatible Order and Payment services

Start real external services or contract-compatible test doubles that implement the internal admin routes from section 6.

Current repository does not provide start commands for those services. Do not claim end-to-end readiness until both downstream contracts are reachable.

### Step 7: Configure task-required existing environment values

Load the existing `.env`, then verify required keys are present:

```bash
printenv SUPERADMIN_DATABASE_DSN
printenv ORDER_SERVICE_ADMIN_BASE_URL
printenv ORDER_SERVICE_TIMEOUT
printenv SUPERADMIN_REQUIRE_ORDER_SERVICE
printenv PAYMENT_SERVICE_ADMIN_BASE_URL
printenv PAYMENT_SERVICE_TIMEOUT
printenv SUPERADMIN_REQUIRE_PAYMENT_SERVICE
```

Do not print a production DSN in shared logs or screenshots.

### Step 8: Verify downstream reachability

If the external services expose health endpoints:

```bash
curl -i http://127.0.0.1:8084/healthz
curl -i http://127.0.0.1:8085/healthz
```

Health endpoints are downstream-specific. Also verify at least one internal admin route using trusted headers expected by that service.

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

## 11. Running the Project

No new start command is introduced.

Refer:

`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:

- `10. Local Development Setup`
- `11. Running the Project`
- `Quick Command Reference`

### Verify order list

```bash
curl -i \
  -H 'X-Admin-Id: admin_local' \
  -H 'X-User-Id: user_local' \
  -H 'X-Admin-Roles: superadmin' \
  -H 'X-Session-Id: session_local_001' \
  -H 'X-Request-Id: request_orders_001' \
  'http://127.0.0.1:8088/api/v1/admin/orders?status=paid&page=1&page_size=20'
```

Expected: HTTP `200` with an `orders` JSON array, assuming the Order Service is reachable and `admin_local` has `orders:read`.

### Verify manual order review

Use disposable local order data:

```bash
curl -i -X POST \
  -H 'Content-Type: application/json' \
  -H 'X-Admin-Id: admin_local' \
  -H 'X-User-Id: user_local' \
  -H 'X-Admin-Roles: superadmin' \
  -H 'X-Session-Id: session_local_001' \
  -H 'X-Request-Id: request_manual_review_001' \
  --data '{"category":"payment_mismatch","reason":"payment captured but order is still pending"}' \
  http://127.0.0.1:8088/api/v1/admin/orders/order_test_001/manual-review
```

Expected: HTTP `201` with a review task whose `resource_type` is `order`, `resource_id` is `order_test_001`, and `status` is `open`.

After success:

```bash
mysql -h 127.0.0.1 -P 3306 -u superadmin -p -D superadmin_db -e \
  "SELECT task_id, task_type, resource_type, resource_id, status, reason FROM admin_review_tasks WHERE resource_type='order' AND resource_id='order_test_001';"

mysql -h 127.0.0.1 -P 3306 -u superadmin -p -D superadmin_db -e \
  "SELECT action, resource_type, resource_id, request_id, reason FROM admin_audit_logs WHERE request_id='request_manual_review_001';"
```

### Verify order dispute view

```bash
curl -i \
  -H 'X-Admin-Id: admin_local' \
  -H 'X-User-Id: user_local' \
  -H 'X-Admin-Roles: superadmin' \
  -H 'X-Session-Id: session_local_001' \
  -H 'X-Request-Id: request_dispute_view_001' \
  http://127.0.0.1:8088/api/v1/admin/orders/order_test_001/dispute-view
```

Expected: HTTP `200` with:

```text
order
status_history
payments
refunds
review_tasks
risk_flags
```

Risk flags can include `payment_mismatch`, `amount_mismatch`, `payment_failed`, `refund_requested`, `refund_processing`, `refund_failed`, and `manual_review_open`.

### Verify payment list

```bash
curl -i \
  -H 'X-Admin-Id: admin_local' \
  -H 'X-User-Id: user_local' \
  -H 'X-Admin-Roles: superadmin' \
  -H 'X-Session-Id: session_local_001' \
  -H 'X-Request-Id: request_payments_001' \
  'http://127.0.0.1:8088/api/v1/admin/payments?status=captured&page=1&page_size=20'
```

Expected: HTTP `200` with a `payments` JSON array, assuming the Payment Service is reachable and `admin_local` has `payments:read`.

### Verify refund review

Use only disposable local refund data in `requested` state:

```bash
curl -i -X POST \
  -H 'Content-Type: application/json' \
  -H 'X-Admin-Id: admin_local' \
  -H 'X-User-Id: user_local' \
  -H 'X-Admin-Roles: superadmin' \
  -H 'X-Session-Id: session_local_001' \
  -H 'X-Request-Id: request_refund_review_001' \
  -H 'X-MFA-Verified: true' \
  --data '{"decision":"approved","reason":"duplicate charge verified by finance team"}' \
  http://127.0.0.1:8088/api/v1/admin/refunds/refund_test_001/review
```

Expected: HTTP `200` with a refund response showing the reviewed status.

After success:

```bash
mysql -h 127.0.0.1 -P 3306 -u superadmin -p -D superadmin_db -e \
  "SELECT action, resource_type, resource_id, request_id, reason FROM admin_audit_logs WHERE request_id='request_refund_review_001';"
```

Also verify the refund state through Payment Service because Payment Service is the source of truth.

### Task-specific success criteria

```text
MySQL is reachable and task-relevant migrations are applied.
An active admin has the required DB-backed permission.
Order Service internal admin API is reachable and contract-compatible.
Payment Service internal admin API is reachable and contract-compatible.
Order and payment list routes return valid downstream JSON.
Manual order review creates an admin_review_tasks row.
Dispute view aggregates order, status history, payments, refunds, review tasks, and risk flags.
Refund review updates Payment Service source-of-truth state.
Audit rows exist for manual order review and refund review.
Missing reason, missing request ID, missing MFA, invalid role, and downstream unavailable cases are tested.
```

## 12. Common Errors & Fixes

Generic Go, MySQL, Docker, port, `.env`, and RBAC errors are already documented in:

`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:

`12. Common Errors & Fixes`

Only task-specific issues are listed here.

| Error / symptom | Cause | Fix | Prevention |
|---|---|---|---|
| `503 DOWNSTREAM_UNAVAILABLE` with "order service is not configured" | `ORDER_SERVICE_ADMIN_BASE_URL` is empty | Set the URL, source `.env`, and restart | Set `SUPERADMIN_REQUIRE_ORDER_SERVICE=true` for task-focused local profiles |
| `503 DOWNSTREAM_UNAVAILABLE` with "payment service is not configured" | `PAYMENT_SERVICE_ADMIN_BASE_URL` is empty | Set the URL, source `.env`, and restart | Set `SUPERADMIN_REQUIRE_PAYMENT_SERVICE=true` for task-focused local profiles |
| Startup fails because Order/Payment URL is required | Strict flag or non-local environment is active | Configure valid `http`/`https` URLs | Validate env before deployment |
| Startup unexpectedly requires User Service too | Non-local `APP_ENV` forces all downstream URLs | Use `APP_ENV=local` for isolated current-task testing, or configure every downstream | Reduce unrelated startup coupling in a future config change |
| Base URL validation fails | Missing scheme or host | Use values like `http://127.0.0.1:8084/internal/admin` | Keep an `.env.example` with correct format |
| Order/Payment connection refused or timeout | Service missing, wrong host/port/path, firewall issue, or timeout too short | Start/fix downstream service and verify URL/network | Add dependency monitoring and contract tests |
| Downstream successful response becomes invalid-response error | JSON shape does not match current Go structs | Align downstream response schema | Add consumer-driven contract tests |
| Downstream `401` or `403` becomes Superadmin `503` | Downstream denied forwarded admin context | Verify trusted headers and internal auth policy | Add authenticated service-to-service contract and clearer error mapping |
| List route returns `403 FORBIDDEN` before downstream call | Admin DB role lacks required permission | Check `admin_users` and `admin_role_permissions` | Provision admins through audited RBAC flow |
| Manual review returns `REVIEW_TASK_ALREADY_EXISTS` | Open task already exists for same order | Close/cancel existing task or use another test order | Use UI/flow guard before creating another open task |
| Manual review returns "admin review task storage is not configured" | DB DSN is empty, so review-task repository is nil | Configure MySQL DSN and migrations | Set `SUPERADMIN_REQUIRE_DATABASE=true` |
| Refund review returns `MFA_REQUIRED` | `X-MFA-Verified` is missing or false | Send trusted MFA context for local private testing | Gateway must inject verified MFA state |
| Refund/manual review returns `REQUEST_CONTEXT_MISSING` | `X-Request-Id` missing | Send a non-empty request ID | Gateway should always inject request IDs |
| Refund/manual review returns `REASON_REQUIRED` | Reason is blank or shorter than 10 characters | Send a meaningful 10-512 character reason | Validate in admin UI/gateway |
| Refund review returns `REFUND_NOT_REVIEWABLE` | Refund is not in `requested` state, or downstream reported conflict/precondition failure | Fetch refund current state and choose valid test data | Keep refund workflow states synchronized |
| Payment summary missing from manual review metadata | Payment lookup for the order failed; current code logs and continues | Check Payment Service URL/logs if summary is required | Add observability and contract tests |
| Audit insert foreign-key failure | `X-Admin-Id` does not exist in `admin_users` | Use/provision an active DB admin row | Test with real provisioned admins |
| Refund state changed but response is `500` | Payment Service mutation succeeded, then audit insert failed | Reconcile Payment Service state and audit record before retrying | Add idempotency, durable audit/outbox, and reconciliation |
| Order/Payment service implementation/start command cannot be found | It is absent from the checked repository | Obtain or implement the external service contract | Track as a blocking dependency |

## 13. Security & Best Practices

### Task-specific security and configuration audit

| Severity | Finding | Recommended fix |
|---|---|---|
| Critical for public deployment | Incoming admin headers are trusted without JWT or signed-context verification | Keep port private behind a trusted gateway; strip client headers; authenticate gateway-to-service traffic |
| Critical | Refund review is finance-critical and can call Payment Service before audit write completes | Add idempotency, durable audit/outbox, reconciliation, and alerts |
| High | Downstream Order/Payment clients send trusted admin headers but no bearer token, service credential, or mTLS config | Add authenticated service-to-service transport and private networking/TLS |
| High | Required Order Service and Payment Service implementations are absent from this workspace | Implement or supply contract-compatible services before production readiness |
| High | No consumer-driven contract tests verify Order/Payment request/response schemas | Add contract tests for all internal routes and error mappings |
| High | `/readyz` checks MySQL only, not Order/Payment dependency state | Define integrated readiness or expose separate dependency health status |
| High | Manual review task creation and downstream state are not in one transaction | Use workflow/reconciliation patterns for cross-service consistency |
| Medium | Downstream `401`/`403` maps to `503`, reducing security/debug clarity | Define safe, explicit internal-auth error mapping |
| Medium | No retry, circuit breaker, or idempotency key exists for downstream mutations | Add carefully reviewed resilience behavior; never blindly retry non-idempotent mutations |
| Medium | No task-specific rate limit exists | Add gateway rate limiting and alerts for finance/admin mutations |
| Medium | Current HTTP implementation differs from gRPC-oriented design docs | Update architecture docs or implement the intended transport |
| Medium | No service Dockerfile or full local Compose stack exists | Add reproducible container/development stack |

### Task-specific best practices

- Order/payment source-of-truth data ko current service DB me duplicate ya directly modify mat karo.
- Refund review ke liye `finance_admin` or `superadmin` permission, MFA, request ID, and meaningful reason require karo.
- Manual order review ke liye duplicate open task create karne se pehle existing open task check karo.
- Request ID ko current service, downstream service, logs, and audit records me preserve karo.
- Payment Service mutation retry se pehle current refund state check karo.
- Downstream admin APIs ko public internet par expose mat karo.
- Service-to-service calls ke liye TLS/mTLS, credentials, timeouts, and observability add karo.
- Audit insert failures, downstream timeouts, and review-task duplicate errors par alerts banao.
- Payment/refund raw provider secrets, card data, tokens, or PII logs me mat likho.
- Valid and denied cases dono test karo: permissions, reason, request ID, MFA, invalid statuses, duplicate review tasks, and downstream outage.

For shared security guidance, refer to:

`TaskImplementation/{SERVICE_NAME}/task3_Dependency.md`

Section:

`13. Security & Best Practices`

## 14. Missing or Misconfigured Things

Only task-specific gaps are listed here:

- [ ] No runnable Order Service implementation in the checked repository
- [ ] No runnable Payment Service implementation in the checked repository
- [ ] No Order/Payment internal admin route implementations or start commands
- [ ] No Order/Payment Docker/Compose setup
- [ ] No authenticated service-to-service credential, token, or mTLS configuration
- [ ] No consumer-driven Order/Payment contract tests
- [ ] No real-MySQL plus real-Order/Payment integration tests
- [ ] No Order/Payment dependency status in readiness checks
- [ ] No mutation idempotency key or durable workflow/outbox
- [ ] Audit insert can fail after Payment Service refund state already changed
- [ ] Payment summary lookup during manual review is best-effort only
- [ ] Refund review task closure is best-effort only
- [ ] No current task-specific rate limiter
- [ ] Incoming trusted roles are forwarded downstream without reconciliation against authoritative DB role
- [ ] Downstream `401`/`403` is mapped to service unavailable
- [ ] No service Dockerfile or full local Compose stack
- [ ] Current HTTP implementation differs from the gRPC-oriented design in `INPUT_FILE_PATH`

## 15. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `task1_Dependency.md` | Tech stack and required software | Same Go, HTTP, MySQL, Docker, and curl setup |
| `task1_Dependency.md` | Dependency management | No new Go package |
| `task1_Dependency.md` | MySQL installation, Docker, DSN, migrations, and admin seed | Same database setup |
| `task1_Dependency.md` | Environment variables and `.env` loading | No new variable names |
| `task1_Dependency.md` | Ports, Docker, local run, and generic troubleshooting | No shared setup change |
| `task2_Dependency.md` | MySQL ownership and migration source of truth | Same database decision |
| `task3_Dependency.md` | RBAC permissions, admin provisioning, trusted headers, high-risk context, and gateway security | Same authorization setup |
| `task4_Dependency.md` | Downstream HTTP service pattern and environment loading | Same current HTTP client approach |
| `Dependency/main_dependency.md` | Existing runtime architecture and missing-service audit | Same service-wide dependency map |
| `Dependency/Go_Modules.md` | Go module setup | No module change |
| `Dependency/MySQL.md` | MySQL setup and credentials | Same database |
| `Dependency/Migrations.md` | Complete migration flow | Same migration mechanism |
| `Dependency/Environment.md` | Existing Order/Payment variables and environment loading | Same variable names |
| `Dependency/HTTP_API.md` | HTTP server and incoming admin headers | Same transport |
| `Dependency/Downstream_HTTP_Services.md` | Order/Payment base URLs and shared downstream behavior | Same downstream client setup |

Full reference paths:

```text
TaskImplementation/{SERVICE_NAME}/task1_Dependency.md
TaskImplementation/{SERVICE_NAME}/task2_Dependency.md
TaskImplementation/{SERVICE_NAME}/task3_Dependency.md
TaskImplementation/{SERVICE_NAME}/task4_Dependency.md
TaskImplementation/{SERVICE_NAME}/Dependency/main_dependency.md
TaskImplementation/{SERVICE_NAME}/Dependency/Go_Modules.md
TaskImplementation/{SERVICE_NAME}/Dependency/MySQL.md
TaskImplementation/{SERVICE_NAME}/Dependency/Migrations.md
TaskImplementation/{SERVICE_NAME}/Dependency/Environment.md
TaskImplementation/{SERVICE_NAME}/Dependency/HTTP_API.md
TaskImplementation/{SERVICE_NAME}/Dependency/Downstream_HTTP_Services.md
```

## 16. Final Checklist

- [ ] Previous dependency documentation read first
- [ ] No unused gRPC, Redis, Kafka, RabbitMQ, payment SDK, or migration-tool package installed
- [ ] Existing Go modules downloaded and verified
- [ ] Task-specific unit tests pass
- [ ] Current service build passes
- [ ] MySQL `8.x` running
- [ ] Complete repository migrations applied exactly once
- [ ] Migrations `001`, `002`, `003`, `004`, and `006` verified
- [ ] Active local admin provisioned
- [ ] Required order/payment permissions verified
- [ ] Existing `.env` loaded into the shell
- [ ] `SUPERADMIN_REQUIRE_DATABASE=true` used for functional local testing
- [ ] `ORDER_SERVICE_ADMIN_BASE_URL` points to a real compatible service
- [ ] `PAYMENT_SERVICE_ADMIN_BASE_URL` points to a real compatible service
- [ ] `SUPERADMIN_REQUIRE_ORDER_SERVICE=true` used for task-focused local testing
- [ ] `SUPERADMIN_REQUIRE_PAYMENT_SERVICE=true` used for task-focused local testing
- [ ] Order Service internal admin routes are reachable
- [ ] Payment Service internal admin routes are reachable
- [ ] `/healthz`, `/readyz`, and RBAC checks pass
- [ ] Order list route verified
- [ ] Payment list route verified
- [ ] Manual order review verified with disposable data
- [ ] Dispute view verified with disposable data
- [ ] Refund review verified with disposable requested refund
- [ ] Downstream source-of-truth states verified after mutation
- [ ] Review task row verified after manual order review
- [ ] Audit rows verified after manual order review and refund review
- [ ] Duplicate manual review, missing reason, missing request ID, missing MFA, invalid status, and permission denial tested
- [ ] Production admin headers can only come from a trusted gateway
- [ ] Service-to-service authentication and transport security planned
- [ ] Task-specific security gaps reviewed
- [ ] No duplicate shared setup documentation added
