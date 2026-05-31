# Superadmin Service - Main Dependency Documentation

## 1. Service Overview

Superadmin Service platform owner/admin ke liye control plane hai. Ye service users, sellers, orders, payments, session analytics access, platform settings, search synonyms, RBAC permissions, review tasks, and admin audit logs ko handle karti hai.

Current codebase me Superadmin Service ka backend available hai:

| Item | Value |
|------|-------|
| Service folder | `backend/services/superadmin-service` |
| Go module | `ecommerce/superadmin-service` |
| Runtime protocol | HTTP server using Go `net/http` |
| Default port | `8088` from `HTTP_ADDR=:8088` |
| Primary database | MySQL `superadmin_db` |
| Main entrypoint | `backend/services/superadmin-service/cmd/server/main.go` |
| Workspace registration | `backend/go.work` includes `./services/superadmin-service` |

Important: project architecture docs and `api/master-api.json` mention gRPC methods for Superadmin Service, but actual backend code currently exposes HTTP routes and uses HTTP clients for downstream User, Order, and Payment services. Proto files/generated gRPC code are not clearly found in project files.

## 2. All Detected Dependencies

| Dependency | Status | Required? | Where Found | Notes |
|------------|--------|-----------|-------------|-------|
| Go toolchain | Completed | Required | `backend/services/superadmin-service/go.mod` | `go 1.26.3` declared |
| Go modules | Completed | Required | `go.mod`, `go.sum` | MySQL driver present |
| MySQL | Completed | Required in real environment | migrations + repository code | Stores RBAC, review tasks, settings, audit logs |
| MySQL driver | Completed | Required | `github.com/go-sql-driver/mysql v1.10.0` | Used through `database/sql` |
| Environment variables | Partial | Required | `.env`, `internal/config/config.go` | `.env` exists, but code reads OS env only |
| SQL migrations | Completed | Required before DB-backed run | `migrations/*.sql` | Up/down migrations exist |
| HTTP server | Completed | Required | `cmd/server/main.go`, `internal/transport/http` | Uses Go standard library |
| Admin HTTP API | Completed | Required | `internal/transport/http/*` | Routes, admin headers, health/readiness |
| Downstream HTTP services | Partial | Required for user/order/payment flows | `internal/clients` | User, Order, Payment base URLs configurable |
| Platform settings in-memory cache | Completed | Required internally | `internal/usecase/platform_settings.go` | Local TTL cache, not Redis |
| Health/readiness checks | Completed | Required | `/healthz`, `/readyz` | Readiness pings DB when configured |
| Logging | Completed | Required | `internal/logging/logger.go` | Structured service logger |
| Validation | Completed | Required | domain/usecase/http validation files | JSON and business validations exist |
| Docker | Missing | Optional for local setup | Not clearly found in project files | No Dockerfile or docker-compose found |
| Frontend | Not Applicable | Optional | Not clearly found in project files | No `frontend/` folder found |
| Redis | Not Applicable | Optional future | Not used in backend code | Docs mention future cache/rate-limit only |
| gRPC | Missing | Planned by docs | Not clearly found in code | No `google.golang.org/grpc` in `go.mod` |
| Protobuf | Missing | Planned by docs | Not clearly found in code | No `.proto` or generated `.pb.go` found |

## 3. Setup Order

1. Install Go version compatible with `go.mod`.
2. Start MySQL and create `superadmin_db`.
3. Create a least-privilege MySQL user for the service.
4. Apply migrations from `backend/services/superadmin-service/migrations`.
5. Export environment variables from `.env` or your shell.
6. Start downstream services if you need user/seller/order/payment flows.
7. Start Superadmin Service.
8. Verify `/healthz` and `/readyz`.
9. Call protected admin APIs with admin headers.

## 4. Backend Start Process

Exact start command is not clearly found in project files.

Suggested command based on project structure:

```bash
cd backend/services/superadmin-service
set -a
source .env
set +a
go run ./cmd/server
```

Verify:

```bash
curl http://127.0.0.1:8088/healthz
curl http://127.0.0.1:8088/readyz
```

Expected health response:

```json
{"status":"ok"}
```

## 5. Frontend Start Process

Not Applicable.

Frontend code related to this service is not clearly found in project files. Docs mention a future Superadmin Panel, but no actual `frontend/superadmin-panel` implementation exists in the checked project tree.

## 6. gRPC / Proto Setup

Current status: Missing / not implemented in actual backend code.

Docs and `api/master-api.json` list gRPC-style methods like:

| Method |
|--------|
| `SuperadminService.ListUsersForAdmin` |
| `SuperadminService.UpdateUserStatus` |
| `SuperadminService.ListSellersForAdmin` |
| `SuperadminService.UpdateSellerStatus` |
| `SuperadminService.ReviewRefund` |
| `SuperadminService.GetPlatformSettings` |
| `SuperadminService.UpdatePlatformSetting` |
| `SuperadminService.ListAuditLogs` |

But these are not wired in actual code:

| Item | Status |
|------|--------|
| `.proto` file | Missing |
| generated protobuf Go code | Missing |
| `transport/grpc` package | Missing |
| gRPC server registration | Missing |
| gRPC client configuration | Missing |
| `google.golang.org/grpc` in `go.mod` | Missing |

## 7. Docker Setup

Docker setup for this service is missing.

Not clearly found in project files:

| Expected Item | Status |
|---------------|--------|
| Service Dockerfile | Missing |
| docker-compose service for `superadmin-service` | Missing |
| compose MySQL service tied to Superadmin Service | Missing |
| compose env file | Missing |

Docker can still be used manually for MySQL. See `MySQL.md`.

## 8. Database Setup

MySQL is required for production-like behavior. Without `SUPERADMIN_DATABASE_DSN`, the service starts in local mode but RBAC checks deny admin routes and DB-backed features become unavailable.

Suggested command based on project structure:

```bash
mysql -uroot -p -e "CREATE DATABASE IF NOT EXISTS superadmin_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
```

Suggested migration command based on project structure:

```bash
migrate \
  -path backend/services/superadmin-service/migrations \
  -database "mysql://superadmin:Ecom8880Admin@tcp(127.0.0.1:3306)/superadmin_db?parseTime=true&charset=utf8mb4&loc=UTC" \
  up
```

Tables created by migrations:

| Table | Purpose |
|-------|---------|
| `admin_users` | Admin identity, role, active/disabled status |
| `admin_permissions` | Permission registry |
| `admin_role_permissions` | Role-to-permission mapping |
| `admin_review_tasks` | Manual review queues |
| `platform_settings` | Maintenance, commission, feature flags, search synonyms |
| `admin_audit_logs` | Immutable-style admin action audit trail |

## 9. Environment Variables

| Variable | Example / Default | Required? | Notes |
|----------|-------------------|-----------|-------|
| `SERVICE_NAME` | `superadmin-service` | Yes | Service identity |
| `APP_ENV` | `local` | Yes | Non-local requires DB/downstream unless flags disable |
| `HTTP_ADDR` | `:8088` | Yes | HTTP bind address |
| `LOG_LEVEL` | `debug` | Optional | Fallback is `info` |
| `SUPERADMIN_DATABASE_DSN` | MySQL DSN | Required for DB-backed run | Fallback key: `MYSQL_DSN` |
| `SUPERADMIN_REQUIRE_DATABASE` | `false` | Optional | Set `true` in strict env |
| `USER_SERVICE_ADMIN_BASE_URL` | `http://127.0.0.1:8081/internal/admin` | Required for user/seller flows | HTTP downstream |
| `ORDER_SERVICE_ADMIN_BASE_URL` | `http://127.0.0.1:8084/internal/admin` | Required for order flows | HTTP downstream |
| `PAYMENT_SERVICE_ADMIN_BASE_URL` | `http://127.0.0.1:8085/internal/admin` | Required for payment/refund flows | HTTP downstream |
| `SUPERADMIN_PLATFORM_SETTINGS_CACHE_TTL` | `5m` | Optional | In-memory cache TTL |
| `SUPERADMIN_PLATFORM_SETTINGS_EVENT_TOPIC` | `platform.settings.updated` | Optional | Logging publisher topic |
| `SUPERADMIN_SESSION_ANALYTICS_MAX_RANGE` | `720h` | Optional | Session authorization max range |
| `SUPERADMIN_SESSION_ANALYTICS_MAX_PAGE_SIZE` | `100` | Optional | Session authorization page size |
| `HTTP_READ_HEADER_TIMEOUT` | `5s` | Optional | HTTP hardening |
| `HTTP_SHUTDOWN_TIMEOUT` | `10s` | Optional | Graceful shutdown |

Security note: `.env` contains a real-looking DB password. Production secrets should not be committed. Add `.env.example` with safe placeholders and load real secrets from secret manager or local ignored files.

## 10. Ports Table

| Component | Port / Address | Source | Status |
|-----------|----------------|--------|--------|
| Superadmin HTTP | `:8088` | `.env`, config default | Completed |
| MySQL | `127.0.0.1:3306` | `SUPERADMIN_DATABASE_DSN` | Completed |
| User Service admin API | `127.0.0.1:8081/internal/admin` | `.env` | Partial, downstream implementation not verified here |
| Order Service admin API | `127.0.0.1:8084/internal/admin` | `.env` | Partial, downstream implementation not verified here |
| Payment Service admin API | `127.0.0.1:8085/internal/admin` | `.env` | Partial, downstream implementation not verified here |
| gRPC | Not clearly found | docs only | Missing |
| Frontend | Not clearly found | docs only | Not Applicable |

## 11. Mermaid Diagrams

### Dependency Flow

```mermaid
flowchart TD
    AdminClient[Admin Client or Gateway] -->|HTTP + admin headers| SA[Superadmin Service]
    SA -->|database/sql + MySQL driver| DB[(MySQL superadmin_db)]
    SA -->|HTTP admin API| UserSvc[User Service]
    SA -->|HTTP admin API| OrderSvc[Order Service]
    SA -->|HTTP admin API| PaymentSvc[Payment Service]
    SA --> Cache[In-memory settings cache]
    SA --> Logs[Structured logs]
```

### Service Startup Flow

```mermaid
flowchart TD
    A[Start process] --> B[Load env with config.Load]
    B --> C[Validate config]
    C --> D{SUPERADMIN_DATABASE_DSN set?}
    D -->|Yes| E[Open MySQL and ping]
    D -->|No| F[Use static deny repository]
    E --> G[Build repositories]
    F --> G
    G --> H[Build usecases]
    H --> I[Build HTTP handlers]
    I --> J[Register routes]
    J --> K[Listen on HTTP_ADDR]
```

### Backend to DB / Downstream Flow

```mermaid
flowchart LR
    Req[Admin HTTP request] --> Actor[ActorMiddleware extracts headers]
    Actor --> Authz[AuthorizationService]
    Authz --> DB[(MySQL RBAC tables)]
    Authz --> Usecase[Control/usecase logic]
    Usecase --> DB2[(Review tasks/settings/audit tables)]
    Usecase --> U[User Service HTTP]
    Usecase --> O[Order Service HTTP]
    Usecase --> P[Payment Service HTTP]
    Usecase --> Audit[Audit recorder]
    Audit --> DB3[(admin_audit_logs)]
```

## 12. Missing / Incomplete Implementation Audit

| Area | Status | Finding | Suggested Fix |
|------|--------|---------|---------------|
| Task implementation incomplete | Partial | Backend code exists for tasks 3-8 areas, but docs still mention future gRPC/proto/frontend. | Update task docs or add follow-up tasks for gRPC/frontend if required. |
| Dependency not installed | Partial | Runtime Go dependency is installed in `go.mod`; migration CLI/Docker are not project dependencies. | Document local tools in README or add Makefile targets. |
| go.mod dependency missing | Completed | MySQL driver exists; no unused gRPC/protobuf dependency in current HTTP implementation. | Add gRPC/protobuf only when real gRPC code is added. |
| go.sum not updated | Completed | `go.sum` contains MySQL driver checksum. | Run `go mod tidy` after future dependency changes. |
| Service not registered in go.work | Completed | `backend/go.work` includes `./services/superadmin-service`. | No action needed. |
| Dockerfile missing | Missing | No Dockerfile found for service. | Add `backend/services/superadmin-service/Dockerfile` if container deployment is required. |
| docker-compose service missing | Missing | No compose file found. | Add local compose with MySQL and service if Docker workflow is required. |
| `.env.example` missing | Missing | `.env` exists, but safe example file not found. | Add `.env.example` with placeholders only. |
| Environment variables missing | Partial | Required vars are documented in config and `.env`; `.env` contains secrets. | Use placeholders and secret manager/ignored local env files. |
| Config not loaded | Partial | Code reads OS env via `os.Getenv`; it does not automatically parse `.env`. | Source `.env` before start or add a safe dotenv loader if desired. |
| Migration missing | Completed | Migrations `001` to `006` exist. | No action for current schema. |
| Migration rollback missing | Completed | Each migration has a `.down.sql`. | No action needed. |
| MySQL table/index/foreign key missing | Partial | Required Superadmin tables/indexes exist; `admin_audit_logs` FK to `admin_users` exists. Admin seed user not found. | Add secure admin provisioning/seed flow outside docs. |
| Repository not wired | Completed | MySQL permission, review task, settings, audit repositories are wired in `main.go`. | No action needed. |
| Usecase/service not wired | Completed | Authorization, controls, order/payment, settings, audit log, session visibility usecases are wired. | No action needed. |
| Handler/controller not wired | Completed | HTTP handlers are created and passed to `NewServeMux`. | No action needed. |
| Routes not registered | Completed | Admin HTTP routes plus `/healthz` and `/readyz` are registered. | No action needed. |
| gRPC proto missing | Missing | No Superadmin `.proto` file found. | Add proto only if project decides to implement gRPC contract. |
| Generated protobuf code missing | Missing | No generated Go protobuf files found. | Generate after proto exists. |
| gRPC server not registered | Missing | No gRPC server startup found. | Add gRPC transport if needed. |
| gRPC client not configured | Missing | Current code uses HTTP clients for downstream services. | Either document HTTP as final choice or implement gRPC clients. |
| Frontend API integration missing | Not Applicable | No frontend code found. | Add Superadmin Panel when frontend scope starts. |
| Frontend env missing | Not Applicable | No frontend implementation found. | Add when frontend exists. |
| Health check missing | Completed | `/healthz` and `/readyz` exist. | No action needed. |
| Logging missing | Completed | Internal structured logger exists. | Consider external log sink later. |
| Validation missing | Completed | Request, domain, permission, and high-risk validation exist. | No action needed. |

## 13. Final Checklist

- [x] `TaskImplementation/Dependency/` folder created.
- [x] `main_dependency.md` created.
- [x] Service-specific dependencies documented.
- [x] MySQL dependency documented.
- [x] Go modules documented.
- [x] Environment variables documented.
- [x] Migrations documented.
- [x] Admin HTTP API documented.
- [x] Downstream HTTP service dependencies documented.
- [x] Frontend status checked.
- [x] Docker status checked.
- [x] gRPC/proto status checked.
- [x] Missing/incomplete audit added.
- [x] No business logic modified.
