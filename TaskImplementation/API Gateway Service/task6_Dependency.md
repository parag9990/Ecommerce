# Project Dependency & Setup Guide

Input analyzed:

`TaskImplementation/API Gateway Service/task6.md`

Previous dependency documentation reviewed first:

- `TaskImplementation/API Gateway Service/task1_Dependency.md`
- `TaskImplementation/API Gateway Service/task2_Dependency.md`
- `TaskImplementation/API Gateway Service/task3_Dependency.md`
- `TaskImplementation/API Gateway Service/task4_Dependency.md`
- `TaskImplementation/API Gateway Service/task5_Dependency.md`

Related implementation inspected:

- `backend/services/api-gateway/go.mod`
- `backend/services/api-gateway/cmd/server/main.go`
- `backend/services/api-gateway/internal/config/config.go`
- `backend/services/api-gateway/internal/errors/*`
- `backend/services/api-gateway/internal/transport/http/response.go`
- `backend/services/api-gateway/internal/transport/http/error_mapping_test.go`
- `backend/services/api-gateway/internal/observability/metrics.go`
- `api/master-api.json`

Simple goal: Ye guide beginner developer ko batata hai ki API Gateway Task 6, yani gRPC errors ko safe REST error envelope me map karne ke liye code ke alawa kya setup chahiye. Isme business logic rewrite nahi hai. Sirf dependencies, env impact, local testing, DevOps assumptions, troubleshooting, and security audit covered hai.

---

## 1. Project Overview

Task 6 API Gateway ka error translation layer hai.

Flow:

```text
Client REST request
  -> API Gateway
  -> downstream gRPC service call
  -> gRPC error/status comes back
  -> Gateway maps it to HTTP status + REST error code
  -> Client gets safe JSON envelope
```

Expected public error shape:

```json
{
  "data": null,
  "request_id": "req_123",
  "error": {
    "code": "NOT_FOUND",
    "message": "Product not found",
    "details": null
  }
}
```

Current implementation reality:

| Area | Current Status |
|---|---|
| Runtime language | Go |
| HTTP stack | Go standard `net/http` |
| Error mapping package | `backend/services/api-gateway/internal/errors` |
| HTTP envelope integration | `backend/services/api-gateway/internal/transport/http/response.go` |
| gRPC mapping | `google.golang.org/grpc/status` and `google.golang.org/grpc/codes` |
| Rich error details | `google.golang.org/genproto/googleapis/rpc/errdetails` |
| New database | None |
| New Redis usage | None, Redis is inherited from Task 4 rate limiting |
| New queue/Kafka/RabbitMQ | None |
| New Docker container | None |
| New port | None |
| New `.env` variable | None |

Important beginner note:

```text
Task 6 khud koi DB, Redis, Kafka, Docker service, ya new port introduce nahi karta.
Full Gateway run karne ke liye previous tasks ke dependencies still required hain.
Task 6 ko isolated test karne ke liye sirf Go dependencies enough hain.
```

What Task 6 adds on top of previous tasks:

| New / Task-specific item | Purpose |
|---|---|
| `internal/errors/codes.go` | Public REST error code constants like `VALIDATION_ERROR`, `TIMEOUT`, `INTERNAL_ERROR` |
| `internal/errors/grpc_mapper.go` | gRPC canonical codes ko HTTP status and REST code me convert karta hai |
| `internal/errors/details.go` | `BadRequest`, `RetryInfo`, `ErrorInfo`, `PreconditionFailure` details ko safe JSON me convert karta hai |
| `internal/errors/sanitizer.go` | Raw DB, token, password, stack trace, connection details leak hone se prevent karta hai |
| `writeMappedError(...)` | Mapped error ko common REST envelope me write karta hai |
| Error mapping tests | gRPC status, retry headers, field details, and sanitization verify karte hain |

Already documented elsewhere and not repeated in full:

| Topic | Refer |
|---|---|
| Base Go/Git/Docker install | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `3. Required Software` |
| Go module basics | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `4. Dependency Management` |
| gRPC client setup | `TaskImplementation/API Gateway Service/task2_Dependency.md` -> `6. Redis / Queue / External Services` |
| JWT/JWKS setup | `TaskImplementation/API Gateway Service/task3_Dependency.md` -> `7. Environment Variables` |
| Redis rate limiting | `TaskImplementation/API Gateway Service/task4_Dependency.md` -> `5. Database Setup` |
| Request validation config | `TaskImplementation/API Gateway Service/task5_Dependency.md` -> `7. Environment Variables` |

---

## 2. Tech Stack

### Task 6 Specific Technologies

| Technology | What it is | Why this project uses it | Required? | Beginner Hinglish explanation |
|---|---|---|---:|---|
| Go | Backend programming language | API Gateway service Go me implemented hai | Yes | Go ek fast compiled backend language hai. Gateway isi se build/run/test hota hai. |
| Go Modules | Go dependency system | `go.mod` and `go.sum` dependency versions manage karte hain | Yes | Go modules npm/pip jaisa dependency manager hai, Go projects ke liye. |
| `net/http` | Go standard HTTP package | HTTP status code, headers, and JSON response write karne ke liye | Yes | Is service me external web framework nahi hai; Go ka built-in HTTP package use ho raha hai. |
| `encoding/json` | Go standard JSON encoder | REST response envelope JSON me send karne ke liye | Yes | Error object ko client-readable JSON me convert karta hai. |
| `context` | Go request lifecycle package | Timeout and cancel errors detect karne ke liye | Yes | Agar request timeout ya client disconnect ho jaye, Gateway usko identify karta hai. |
| `errors` | Go error wrapping package | Local errors, context errors, and wrapped errors detect karne ke liye | Yes | `errors.Is` and `errors.As` se wrapped error ka asli type pakda jata hai. |
| gRPC status | gRPC error status helper | `status.FromError(err)` se gRPC code/message nikalne ke liye | Yes | Downstream service ka error structured form me read hota hai. |
| gRPC codes | gRPC canonical error codes | `NotFound`, `InvalidArgument`, `Unavailable` jaise codes map karne ke liye | Yes | Ye service-to-service error categories hain. Gateway inko REST status me translate karta hai. |
| Google RPC errdetails | Rich gRPC error details | Field validation, retry delay, precondition, and error info extract karne ke liye | Yes for rich details | Agar service field-level error bhejti hai, Gateway usko frontend-friendly `details` me convert karta hai. |
| Protobuf runtime | gRPC details serialization runtime | `errdetails` messages ko attach/read karne ke liye | Yes | Protobuf typed messages ka runtime support deta hai. |
| Prometheus metrics | Metrics library | Error response info HTTP metrics me attach hota hai | Inherited/optional | Metrics se pata chalta hai kaunse error codes frequently aa rahe hain. |
| OpenTelemetry tracing | Distributed tracing | Request journey trace karne ke liye | Inherited/optional | Tracing se debugging easy hoti hai jab error Gateway se service tak travel karta hai. |

### Reused Stack

General Gateway, gRPC, JWT, Redis, validation, Docker, and platform dependency stack already explained hai:

- `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `2. Tech Stack`
- `TaskImplementation/API Gateway Service/task2_Dependency.md` -> `2. Tech Stack`
- `TaskImplementation/API Gateway Service/task3_Dependency.md` -> `2. Tech Stack`
- `TaskImplementation/API Gateway Service/task4_Dependency.md` -> `2. Tech Stack`
- `TaskImplementation/API Gateway Service/task5_Dependency.md` -> `2. Tech Stack`

---

## 3. Required Software

### Required for Task 6 Development

| Software | Required? | Why |
|---|---:|---|
| Git | Yes | Repository clone/update ke liye |
| Go `1.26.3` compatible toolchain | Yes | `backend/services/api-gateway/go.mod` declares `go 1.26.3` |
| curl | Optional | Full Gateway HTTP endpoint verify karne ke liye |
| jq | Optional | JSON error response pretty-print karne ke liye |
| grpcurl | Optional | Real downstream gRPC services ke error behavior debug karne ke liye |
| Docker / Docker Compose | Full runtime only | Redis, downstream services, and platform dependencies run karne ke liye |

Installation process already documented hai:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`3. Required Software`

### Setup Modes

| Mode | Redis Needed? | Downstream gRPC Needed? | JWKS Needed? | Use Case |
|---|---:|---:|---:|---|
| Task 6 mapper unit tests | No | No | No | `internal/errors` tests run karne ke liye |
| HTTP error envelope tests | No | No | No | `writeMappedError` envelope and headers verify karne ke liye |
| Full `go run ./cmd/server` | Yes if rate limit enabled | Yes | Yes for protected routes | Real local Gateway runtime |
| Real downstream error smoke test | Depends on route | Yes | Depends on route auth | Actual service error ko REST envelope me dekhne ke liye |

Beginner note:

```text
Agar aap sirf Task 6 verify kar rahe ho, Redis/Auth/JWKS/downstream services start karne ki zarurat nahi.
`go test ./internal/errors ./internal/transport/http` enough hai.
```

---

## 4. Dependency Management

This is a Go project. Base Go modules explanation already covered hai:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`4. Dependency Management`

### Current Go Module

```text
module ecommerce/api-gateway
go 1.26.3
```

### Task 6 Relevant Go Dependencies

Current `backend/services/api-gateway/go.mod` already includes the packages needed by Task 6.

| Dependency | Version in `go.mod` | Task 6 Use |
|---|---:|---|
| `google.golang.org/grpc` | `v1.81.1` | `status.FromError`, `codes.Code`, gRPC status handling |
| `google.golang.org/genproto/googleapis/rpc` | `v0.0.0-20260401024825-9d38bb4040a9` | `errdetails.BadRequest`, `RetryInfo`, `ErrorInfo`, `PreconditionFailure` |
| `google.golang.org/protobuf` | `v1.36.11` | Protobuf runtime for rich error details |
| `github.com/prometheus/client_golang` | `v1.23.2` | Error code/status metrics through HTTP middleware |
| `go.opentelemetry.io/otel` | `v1.43.0` | Inherited tracing around requests |

Standard library packages used by Task 6:

| Package | Why used |
|---|---|
| `context` | `context.DeadlineExceeded` and `context.Canceled` mapping |
| `errors` | Wrapped local errors detect karne ke liye |
| `net/http` | HTTP status constants and response writing |
| `encoding/json` | REST envelope JSON encoding |
| `strconv` | `Retry-After` seconds header format |
| `strings` | Public message sanitization |

### Do You Need To Install a New Package?

No new third-party dependency is required right now.

Task 6 implementation guide mentions packages like `google.golang.org/grpc/status`, `google.golang.org/grpc/codes`, and `google.golang.org/genproto/googleapis/rpc/errdetails`. Current `go.mod` already has them.

Use this only if your local dependency cache is empty:

```bash
cd backend/services/api-gateway
go mod download
```

If someone manually edits imports or dependency versions, clean the module:

```bash
cd backend/services/api-gateway
go mod tidy
```

### Task 6 Focused Commands

Run from API Gateway service directory:

```bash
cd backend/services/api-gateway
```

Download dependencies:

```bash
go mod download
```

Run Task 6 focused tests:

```bash
go test ./internal/errors ./internal/transport/http
```

Run all Gateway tests:

```bash
go test ./...
```

Build Gateway:

```bash
go build ./cmd/server
```

Run Gateway:

```bash
go run ./cmd/server
```

### Common Go Dependency Issues

General Go dependency problems already documented hai:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`4. Dependency Management` -> `Dependency Problems and Fixes`

Task 6 specific notes:

| Error | Likely Cause | Fix |
|---|---|---|
| `cannot find package google.golang.org/genproto/googleapis/rpc/errdetails` | Dependencies download nahi hue ya `go.mod` inconsistent hai | `go mod download`, then `go mod tidy` |
| `missing go.sum entry` | Checksum lock file update required | `go mod tidy` |
| `go: go.mod requires go >= 1.26.3` | Local Go version old hai | Go `1.26.3` compatible version install karo |
| `status.WithDetails` fails in tests | Protobuf/genproto version mismatch ya invalid detail message | Current `go.mod` versions keep karo and `go mod tidy` run karo |
| `undefined: http.StatusClientClosedRequest` | Go standard library me `499` constant nahi hota | Use project constant `gatewayerrors.StatusClientClosedRequest` |
| Test expects raw DB/password message but gets generic message | Sanitizer intentionally hides unsafe internal details | Test expectation update karo; raw secret client ko nahi jana chahiye |

---

## 5. Database Setup

### API Gateway Task 6 Direct Database Requirement

Task 6 does not directly use any SQL/document database.

| Store | Directly used by Task 6? | Purpose |
|---|---:|---|
| MySQL | No | Owned by downstream services |
| MongoDB | No | Owned by downstream services |
| Redis | No new Task 6 usage | Inherited from Task 4 rate limiting |
| Typesense | No | Owned by Search Service |
| Kafka / RabbitMQ | No | Platform async/event services, not error mapping |

### Reuse Previous Database Documentation

Do not duplicate DB installation here. Follow previous guide sections:

| Database / Store | Where setup is already explained |
|---|---|
| MySQL | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `5. Database Setup` -> `MySQL` |
| MongoDB | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `5. Database Setup` -> `MongoDB` |
| Redis | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `5. Database Setup` -> `Redis` |
| Redis rate-limit details | `TaskImplementation/API Gateway Service/task4_Dependency.md` -> `5. Database Setup` -> `Redis for Rate Limiting` |
| Typesense | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `6. Redis / Queue / External Services` -> `Typesense` |
| Kafka / RabbitMQ | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `6. Redis / Queue / External Services` -> `Kafka or RabbitMQ` |

### Migrations

Task 6 does not introduce migrations.

| Migration Type | Needed for Task 6? | Notes |
|---|---:|---|
| API Gateway SQL migrations | No | Gateway has no direct SQL database |
| API Gateway Mongo migrations | No | Gateway has no direct MongoDB collection |
| Redis migration | No | Error mapping does not create Redis schema |
| Downstream service migrations | Full stack only | Run respective service migrations when running full platform |

If full platform setup needs DB migrations, follow:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`10. Running the Project` -> `Migration Commands`

---

## 6. Redis / Queue / External Services

### Direct New External Service for Task 6

None.

Task 6 is in-process Go logic. It maps an `error` value to a REST-safe response. It does not require a new Redis instance, queue, object store, search engine, SMTP provider, payment provider, or external API.

### Important Task 6 Contract Dependency: gRPC Error Shape

Task 6 works best when downstream services return proper gRPC status errors.

#### A. What It Is

gRPC status error ek structured error hota hai jisme canonical code hota hai, jaise:

```text
InvalidArgument
NotFound
FailedPrecondition
Unavailable
DeadlineExceeded
```

Rich details optional typed messages hote hain, jaise:

```text
BadRequest field violations
RetryInfo retry delay
PreconditionFailure business-state violations
ErrorInfo machine-readable reason/metadata
```

#### B. Why This Project Uses It

Gateway ko frontend ko consistent REST error dena hai. Agar downstream service proper gRPC code bhejti hai, Gateway cleanly map kar sakta hai:

```text
gRPC NotFound -> HTTP 404 + NOT_FOUND
gRPC InvalidArgument -> HTTP 400 + VALIDATION_ERROR
gRPC ResourceExhausted + RetryInfo -> HTTP 429 + Retry-After
```

#### C. Required or Optional

| Item | Required? | Notes |
|---|---:|---|
| gRPC canonical status codes | Required for correct mapping | Generic `errors.New(...)` se Gateway `500 INTERNAL_ERROR` return karega |
| Rich error details | Recommended | Field-level validation and retry hints ke liye useful |
| `RetryInfo` for rate limits | Recommended | Gateway `Retry-After` header set kar paata hai |
| Raw DB/stack trace messages | Not allowed | Sanitizer unsafe details hide karega |

#### D. Local Installation / Startup

No separate install. Ye Go gRPC dependency ka part hai.

For actual downstream service testing, downstream gRPC services setup already explained hai:

`TaskImplementation/API Gateway Service/task2_Dependency.md`

Sections:

- `6. Redis / Queue / External Services` -> `Direct New External Dependency for Task 2: Downstream gRPC Services`
- `6. Redis / Queue / External Services` -> `Required gRPC Health Support`

#### E. Docker Setup

Task 6 does not add Docker containers.

If Gateway and downstream services run inside Docker Compose, use the same Docker networking rule from:

`TaskImplementation/API Gateway Service/task2_Dependency.md`

Section:

`8. Docker Setup` -> `Docker Compose Address Rule`

#### F. Start Commands

For Task 6 tests:

```bash
cd backend/services/api-gateway
go test ./internal/errors ./internal/transport/http
```

For full runtime, start inherited dependencies first:

- Redis from Task 4
- Auth JWKS endpoint from Task 3
- downstream gRPC services from Task 2

#### G. Verify Running

Task 6 itself is verified by tests. Expected behavior:

| Test Area | What Should Pass |
|---|---|
| gRPC code mapping | `NotFound` becomes `404 NOT_FOUND`, `Unavailable` becomes `503 SERVICE_UNAVAILABLE` |
| Rich details | `BadRequest` becomes field details |
| Retry hints | `RetryInfo` becomes `Retry-After` header and `retry_after_seconds` detail |
| Sanitization | `sql: password=secret` never appears in public response |
| Request ID | Error envelope includes request id from request context/header |

#### H. Default Port

No new port.

#### I. Connection String / URL Format

No new connection string.

#### J. Where To Place Credentials

Task 6 does not introduce credentials.

| Credential | Needed by Task 6? | Where it belongs |
|---|---:|---|
| DB username/password | No | Downstream service `.env` only |
| Redis password | No new usage | Gateway `.env` only because Task 4 rate limiter uses Redis |
| JWT private key | No | Auth Service secret manager / mounted secret |
| Payment provider secret | No | Payment Service secret manager / mounted secret |
| gRPC TLS certificates | Not new | Deployment secret store if `GRPC_TLS_ENABLED=true` |

### Ports & Networking

Task 6 introduces no new port. Reused ports summary:

| Service | Port | Purpose | New/Reused |
|---|---:|---|---|
| API Gateway HTTP | `8080` default via `HTTP_ADDR=:8080` | Public REST and health endpoints | Reused |
| API Gateway Metrics | `9090` default via `METRICS_PORT=9090` | Prometheus metrics endpoint | Reused |
| Auth Service HTTP/JWKS | `8081` typical local value | JWT public keys for protected routes | Reused from Task 3 |
| Downstream gRPC services | `50051` to `50062` local examples | gRPC service targets and health checks | Reused from Task 2 |
| Redis | `6379` | Rate-limit counters when enabled | Reused from Task 4 |
| MySQL | `3306` | Downstream relational services | Full platform only |
| MongoDB | `27017` | Downstream document services | Full platform only |
| Typesense | `8108` | Search service dependency | Full platform only |
| Kafka / Redpanda | `9092` | Event streaming | Full platform only |
| RabbitMQ AMQP | `5672` | Queue alternative | Full platform only |

Full port table already exists in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`7. Environment Variables` -> `Ports & Networking`

---

## 7. Environment Variables

Task 6 introduces no new and no changed environment variables.

Full Gateway `.env` and export process already documented in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Sections:

- `7. Environment Variables`
- `7. Environment Variables` -> `Complete .env Example`
- `7. Environment Variables` -> `Common .env Mistakes`

### Task 6 `.env` Additions

No additions.

Use this marker in your notes if you want a task-specific snippet:

```env
# Task 6 Error Mapping
# No new env variables are required.
# Keep using the API Gateway env from task1-task5 dependency guides.
```

### Inherited Variables That Affect Error Responses

These variables are not new, but they influence how Task 6 responses are observed/debugged:

| Variable | Owner / Previous Doc | Why it matters for Task 6 |
|---|---|---|
| `REQUEST_ID_HEADER` | Task 1 / observability config | Error envelope returns `request_id`; this controls incoming header name |
| `LOG_LEVEL` | Task 1 | Debugging sanitized internal errors |
| `METRICS_ENABLED` | Task 1 / Task 2 | HTTP metrics include error status/code |
| `METRICS_PORT` / `METRICS_ADDR` | Task 1 / Task 2 | Where to scrape metrics |
| `TRACE_ENABLED` | Task 2 | Trace request path around failures |
| `GRPC_DIAL_TIMEOUT` | Task 2 | Can lead to timeout/unavailable errors in full runtime |
| `RATE_LIMIT_ENABLED` | Task 4 | Rate-limit errors use same envelope |
| `REQUEST_VALIDATION_ENABLED` | Task 5 | Validation errors use same envelope |

### Where To Place Variables

Recommended local file remains:

```text
backend/services/api-gateway/.env
```

Current Go code uses `os.Getenv`. It does not auto-load `.env`. The export process is already documented in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`7. Environment Variables` -> `Where To Create .env`

### Credentials Placement

Task 6 does not need new secrets.

| Data | Put in Gateway `.env`? | Notes |
|---|---:|---|
| Error code mapping rules | No | Hardcoded in code, not runtime secret |
| Public error messages | No | Code-owned fallback messages |
| DB credentials | No for Task 6 | Downstream services own DB credentials |
| Redis password | Inherited only | Needed only if Task 4 Redis rate limiting uses auth |
| JWT private key | No | Auth Service owns private keys |
| Raw downstream error text | No | Never store raw internal errors in `.env` |

---

## 8. Docker Setup

### Current Repo Status

No API Gateway `Dockerfile` or root `docker-compose.yml` was found during this inspection.

Task 6 does not add Docker-specific files.

Base Docker guidance already exists in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`8. Docker Setup`

### Docker Configuration Impact

| Docker Area | Task 6 Change? | Notes |
|---|---:|---|
| API Gateway image | No | Error mapping is compiled into existing Go binary |
| Redis container | No | Still inherited from Task 4 when rate limit enabled |
| DB containers | No | Downstream service responsibility |
| Queue containers | No | Full platform only |
| Volumes | No | Error mapping stores no persistent data |
| Networks | No new network | Use existing Gateway-to-downstream network |
| Restart policies | No new requirement | Existing service restart policy is enough |
| Health checks | No new endpoint | Existing `/health/live` and `/health/ready` remain |

### Docker Networking Reminder

Already documented in:

`TaskImplementation/API Gateway Service/task2_Dependency.md`

Section:

`8. Docker Setup` -> `Docker Compose Address Rule`

Short version:

```text
Host run:
  REDIS_ADDR=localhost:6379
  AUTH_GRPC_ADDR=localhost:50051

Docker Compose run:
  REDIS_ADDR=redis:6379
  AUTH_GRPC_ADDR=auth-service:9090
```

Do not use `localhost` from one container to call another container.

---

## 9. Local Development Setup

### Reused Base Onboarding

Follow the base onboarding first:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`9. Local Development Setup`

This already covers:

- clone repository
- check Go version
- install Go dependencies
- start Redis
- start downstream gRPC services
- start Auth JWKS endpoint
- export `.env`
- run API Gateway
- verify health endpoints

### Task 6 Incremental Setup

After base repo setup, run only Task 6 focused checks:

```bash
cd backend/services/api-gateway
go mod download
go test ./internal/errors ./internal/transport/http
```

Expected result:

```text
ok   ecommerce/api-gateway/internal/errors
ok   ecommerce/api-gateway/internal/transport/http
```

### What To Check While Developing

| Check | Why |
|---|---|
| gRPC code mapping table | `InvalidArgument`, `NotFound`, `Unavailable`, etc. should map predictably |
| Error envelope shape | Client should always see `data`, `request_id`, and `error` |
| Sanitization tests | Internal DB/password/token/connection details must not leak |
| Retry headers | `RetryInfo` should set `Retry-After` |
| Field details | `BadRequest.FieldViolations` should become frontend-friendly details |
| Context errors | `context.DeadlineExceeded` and `context.Canceled` should map correctly |

### Current Manual Testing Limitation

Current `Handler.RouteDefined(...)` still returns:

```text
ROUTE_BRIDGE_NOT_CONFIGURED
```

That means real route-to-downstream business handlers are not fully wired yet. So Task 6 gRPC error mapping is best verified with unit/HTTP tests until actual route bridge code calls downstream services and passes gRPC errors into `writeMappedError(...)`.

Beginner note:

```text
Agar curl se route hit karne par ROUTE_BRIDGE_NOT_CONFIGURED aaye, iska matlab Task 6 fail nahi hai.
Ye current Gateway route bridge placeholder behavior hai.
```

---

## 10. Running the Project

### Minimal Task 6 Verification

Use this when you only want to verify error mapping:

```bash
cd backend/services/api-gateway
go test ./internal/errors
go test ./internal/transport/http -run 'TestWriteMappedError'
```

These tests verify:

- gRPC `NotFound` -> HTTP `404` + `NOT_FOUND`
- gRPC `ResourceExhausted` + `RetryInfo` -> HTTP `429` + `Retry-After`
- gRPC `Internal` unsafe message -> generic `Internal server error`
- request id is included in the response envelope

### Full Gateway Runtime

Full runtime setup is inherited from previous tasks.

Refer:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`10. Running the Project`

Task-specific reminder:

```bash
cd backend/services/api-gateway
set -a
source .env
set +a
go run ./cmd/server
```

Verify basic health:

```bash
curl http://localhost:8080/health/live
curl http://localhost:8080/health/ready
```

If metrics enabled:

```bash
curl http://localhost:9090/metrics
```

### Real Downstream Error Verification

When actual route bridge/downstream handlers are wired, test with a request that should return a known service error.

Examples:

| Scenario | Expected Mapping |
|---|---|
| Missing product id | `404 NOT_FOUND` |
| Invalid quantity | `400 VALIDATION_ERROR` |
| Duplicate email/order conflict | `409 CONFLICT` |
| Order cannot be cancelled after shipment | `422 FAILED_PRECONDITION` |
| Downstream service unavailable | `503 SERVICE_UNAVAILABLE` |
| Downstream deadline exceeded | `504 TIMEOUT` |

Response should never expose:

```text
sql:
password
token
secret
stack trace
connection refused
dial tcp
localhost:
```

### Migrations

No Task 6 migration.

---

## 11. Common Errors & Fixes

### Task 6 Specific Troubleshooting

| Error / Symptom | Likely Cause | Fix |
|---|---|---|
| `cannot find package .../errdetails` | Go dependencies not downloaded or `go.mod` changed | Run `go mod download`, then `go mod tidy` |
| `missing go.sum entry` | Dependency checksum missing | Run `go mod tidy` |
| `undefined: http.StatusClientClosedRequest` | Go does not define HTTP 499 | Use `gatewayerrors.StatusClientClosedRequest` |
| `Retry-After` header missing | Downstream error did not include `RetryInfo` | Service should attach `errdetails.RetryInfo` for retryable/rate-limit errors |
| `details` is null for validation error | Downstream sent only `InvalidArgument` message, no `BadRequest` details | Service should attach `errdetails.BadRequest` field violations |
| Public message becomes generic | Message contained unsafe fragments like `sql`, `token`, `password`, `dial tcp` | This is expected security behavior; send safe public messages only |
| `499` response body missing | Client canceled/disconnected | Expected. Gateway usually cannot write response after client disconnect |
| `ROUTE_BRIDGE_NOT_CONFIGURED` on real route | Route bridge business handlers are not wired yet | This is outside Task 6 dependency setup |
| Error metrics lack downstream service label | `MappedError.DownstreamService` not set by generic mapper | Set downstream service context/field when concrete handlers are integrated |

### Reused Troubleshooting

Do not duplicate previous troubleshooting. Use these:

| Topic | Refer |
|---|---|
| Go module/version issues | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `11. Common Errors & Fixes` |
| gRPC connection/readiness issues | `TaskImplementation/API Gateway Service/task2_Dependency.md` -> `11. Common Errors & Fixes` |
| JWT/JWKS auth errors | `TaskImplementation/API Gateway Service/task3_Dependency.md` -> `11. Common Errors & Fixes` |
| Redis/rate-limit errors | `TaskImplementation/API Gateway Service/task4_Dependency.md` -> `11. Common Errors & Fixes` |
| Validation errors | `TaskImplementation/API Gateway Service/task5_Dependency.md` -> `11. Common Errors & Fixes` |

---

## 12. Security & Best Practices

### Task 6 Security Notes

| Security Area | What to do | Why |
|---|---|---|
| Raw error messages | Never return raw internal error text to clients | DB errors, stack traces, tokens, hosts, and passwords can leak |
| Public messages | Keep messages short, safe, and user-actionable | Frontend users need guidance, not infrastructure details |
| Error details | Only expose safe field/reason/precondition metadata | Rich details are useful but must not include secrets |
| Retry headers | Use `Retry-After` only from trusted `RetryInfo` or local policy | Prevent confusing client retry behavior |
| Request ID | Always include `request_id` in error envelope | Support/debug team can correlate client issue with logs |
| Internal logs | Log internal errors through redacted logging strategy | Operators need details, clients should not see them |
| gRPC codes | Downstream services should use canonical codes correctly | Wrong codes create wrong HTTP behavior |
| 500 errors | Use generic `Internal server error` | Internal implementation details stay private |

### Beginner Best Practices

- Use `InvalidArgument` for request validation mistakes.
- Use `NotFound` when resource does not exist or user cannot access it.
- Use `FailedPrecondition` when request is valid but current state blocks the action.
- Use `AlreadyExists` or `Aborted` for conflicts.
- Use `ResourceExhausted` with `RetryInfo` for rate limits or quota limits.
- Use `Unavailable` for temporary downstream outage.
- Use `DeadlineExceeded` for timeout.
- Do not put secrets in `ErrorInfo.Metadata`.
- Do not include full SQL, Redis, Mongo, or network errors in public messages.
- Add a test whenever a new public error code or detail shape is introduced.

### Already Documented Best Practices

| Topic | Refer |
|---|---|
| General Gateway security | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `12. Security & Best Practices` |
| gRPC client security | `TaskImplementation/API Gateway Service/task2_Dependency.md` -> `12. Security & Best Practices` |
| JWT/JWKS security | `TaskImplementation/API Gateway Service/task3_Dependency.md` -> `12. Security & Best Practices` |
| Redis/rate-limit security | `TaskImplementation/API Gateway Service/task4_Dependency.md` -> `12. Security & Best Practices` |
| Validation security | `TaskImplementation/API Gateway Service/task5_Dependency.md` -> `12. Security & Best Practices` |

---

## 13. Missing or Misconfigured Things

### Audit Findings

| Finding | Severity | Why it matters | Suggested Fix |
|---|---|---|---|
| No new Task 6 `.env` variables | Info | Good. Error mapping is code-owned and deterministic | Keep it this way unless runtime customization is truly needed |
| No API Gateway Dockerfile/compose found | Medium for onboarding | Beginners cannot containerize Gateway directly from repo yet | Add Dockerfile/compose in a separate DevOps task; reuse Task 1 Docker guidance |
| Route bridge still returns `ROUTE_BRIDGE_NOT_CONFIGURED` | Medium | Real downstream gRPC error path is not manually testable through business routes yet | When route bridge is implemented, call `writeMappedError(...)` for downstream errors |
| `MappedError.DownstreamService` exists but generic mapper does not set it | Low/Medium | Downstream error metrics may miss service label unless handlers/interceptors set it | Attach target service when mapping errors from concrete downstream calls |
| `ROUTE_BRIDGE_NOT_CONFIGURED` is a literal code outside core Task 6 code constants | Low | Frontend may see a code not listed in main error catalog | Either document placeholder code or replace it when bridge is implemented |
| Rich details support is selective | Info | Current mapper supports `BadRequest`, `ErrorInfo`, `PreconditionFailure`, `RetryInfo` only | Add tests before supporting new detail types |
| No migrations for Task 6 | Info | Expected because Gateway has no DB | No action |
| No hardcoded credentials found in Task 6 error mapping files | Good | Sanitizer is designed to prevent secret leakage | Keep tests covering token/password/secret cases |

### Hardcoded Credentials Check

Task 6 files do not require credentials and should not contain credentials.

Sensitive values that must never be hardcoded:

- DB usernames/passwords
- Redis password
- JWT private keys
- OAuth client secrets
- payment provider secrets
- SMTP credentials
- S3 access keys
- webhook provider secrets

Where secrets belong:

| Environment | Recommended Location |
|---|---|
| Local development | service-specific `.env`, never committed |
| Docker Compose | `.env` or Docker secrets |
| Kubernetes | Kubernetes Secret / external secret manager |
| Production | managed secret store like AWS Secrets Manager, Vault, GCP Secret Manager, etc. |

---

## 14. References to Previous Dependency Files

Use these instead of duplicating setup:

| Topic | Reference |
|---|---|
| Base project overview and direct/indirect dependencies | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `1. Project Overview` |
| General tech stack | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `2. Tech Stack` |
| Git, Go, Docker install | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `3. Required Software` |
| Go modules explanation | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `4. Dependency Management` |
| MySQL setup | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `5. Database Setup` -> `MySQL` |
| MongoDB setup | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `5. Database Setup` -> `MongoDB` |
| Redis setup | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `5. Database Setup` -> `Redis` |
| Typesense setup | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `6. Redis / Queue / External Services` -> `Typesense` |
| Kafka/RabbitMQ setup | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `6. Redis / Queue / External Services` -> `Kafka or RabbitMQ` |
| Complete Gateway `.env` example | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `7. Environment Variables` -> `Complete .env Example` |
| Full port map | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `7. Environment Variables` -> `Ports & Networking` |
| Docker setup | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `8. Docker Setup` |
| Full local onboarding | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `9. Local Development Setup` |
| Running full project | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `10. Running the Project` |
| Downstream gRPC service setup | `TaskImplementation/API Gateway Service/task2_Dependency.md` -> `6. Redis / Queue / External Services` |
| gRPC env variables | `TaskImplementation/API Gateway Service/task2_Dependency.md` -> `7. Environment Variables` |
| JWKS/Auth setup | `TaskImplementation/API Gateway Service/task3_Dependency.md` -> `6. Redis / Queue / External Services` -> `New/Detailed for Task 3: Auth Service JWKS Endpoint` |
| JWT env variables | `TaskImplementation/API Gateway Service/task3_Dependency.md` -> `7. Environment Variables` |
| Redis rate limiting | `TaskImplementation/API Gateway Service/task4_Dependency.md` -> `5. Database Setup` -> `Redis for Rate Limiting` |
| Rate-limit env variables | `TaskImplementation/API Gateway Service/task4_Dependency.md` -> `7. Environment Variables` |
| Request validation env variables | `TaskImplementation/API Gateway Service/task5_Dependency.md` -> `7. Environment Variables` |
| Validation troubleshooting | `TaskImplementation/API Gateway Service/task5_Dependency.md` -> `11. Common Errors & Fixes` |

---

## 15. Final Checklist

Task 6 setup checklist:

- [ ] Read `TaskImplementation/API Gateway Service/task6.md`
- [ ] Read previous dependency docs before duplicating setup
- [ ] Install Git and Go using `task1_Dependency.md`
- [ ] Confirm Go version is compatible with `go 1.26.3`
- [ ] Run `cd backend/services/api-gateway`
- [ ] Run `go mod download`
- [ ] Run `go test ./internal/errors`
- [ ] Run `go test ./internal/transport/http -run 'TestWriteMappedError'`
- [ ] Run `go test ./...`
- [ ] Confirm no new `.env` variables are needed for Task 6
- [ ] Confirm no new DB migration is needed
- [ ] Confirm no new Redis/Kafka/RabbitMQ/Typesense setup is needed
- [ ] Confirm `RetryInfo` maps to `Retry-After`
- [ ] Confirm `BadRequest` details map to field-level details
- [ ] Confirm unsafe messages like `sql: password=secret` are sanitized
- [ ] Confirm `request_id` appears in error envelopes
- [ ] For full runtime, follow Task 1 to Task 5 dependency guides
- [ ] When route bridge is implemented, verify real downstream gRPC errors through REST endpoints

Final beginner reminder:

```text
Task 6 ka dependency setup mostly Go package verification hai.
System-level setup same hai jo previous API Gateway dependency guides me already documented hai.
Repeat mat karo; previous guides ko refer karo and Task 6 ke tests run karo.
```
