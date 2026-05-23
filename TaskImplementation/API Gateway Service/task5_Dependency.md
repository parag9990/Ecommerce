# Project Dependency & Setup Guide

Input analyzed:

`TaskImplementation/API Gateway Service/task5.md`

Previous dependency documentation reviewed first:

- `TaskImplementation/API Gateway Service/task1_Dependency.md`
- `TaskImplementation/API Gateway Service/task2_Dependency.md`
- `TaskImplementation/API Gateway Service/task3_Dependency.md`
- `TaskImplementation/API Gateway Service/task4_Dependency.md`

Related implementation inspected:

- `backend/services/api-gateway/go.mod`
- `backend/services/api-gateway/cmd/server/main.go`
- `backend/services/api-gateway/internal/config/config.go`
- `backend/services/api-gateway/internal/validation/*`
- `backend/services/api-gateway/internal/transport/http/validation_middleware.go`
- `backend/services/api-gateway/internal/transport/http/routes.go`
- `backend/services/api-gateway/internal/transport/http/response.go`
- `api/master-api.json`

Simple goal: Ye guide API Gateway Task 5, yani request validation, ko run/debug karne ke liye required setup explain karta hai. Isme business logic rewrite nahi hai. Sirf dependencies, env, schema contract, local run, troubleshooting, and DevOps notes hain.

---

## 1. Project Overview

Task 5 API Gateway me incoming REST request ko downstream gRPC service tak bhejne se pehle validate karta hai.

Flow:

```text
Client request
  -> API Gateway route match
  -> rate limit from Task 4
  -> auth/RBAC from Task 3, if protected route
  -> request validation from Task 5
  -> route bridge / downstream gRPC service
```

Current implementation reality:

| Area | Current Status |
|---|---|
| Runtime language | Go |
| HTTP stack | Go standard `net/http` / `http.ServeMux` |
| Validation engine | Custom Gateway validation package using `api/master-api.json` schemas |
| JSON parsing | Go standard `encoding/json` with `UseNumber` |
| Content-Type parsing | Go standard `mime.ParseMediaType` |
| Body limit | `http.MaxBytesReader` |
| Path params | Go `Request.PathValue`, so no chi router is required |
| External validator library | None currently |
| Direct DB requirement | None |
| New Docker service | None |
| Migrations | None |
| Default validation mode | `REQUEST_VALIDATION_ENABLED=true` |

Important beginner note:

```text
Task 5 does not create a new database, Redis instance, Kafka topic, or Docker container.
It uses the existing API route contract file and validates request shape/format/size.
```

What Task 5 adds on top of previous tasks:

| New / Task-specific item | Purpose |
|---|---|
| `internal/validation/*` | Header, query, path, body, schema, and idempotency validation |
| `RequestValidationMiddleware` | Runs validation before the handler |
| `REQUEST_VALIDATION_*` env settings | Turn validation on/off and tune limits |
| `api/master-api.json` schema dependency | Request schemas are loaded from this contract |
| `VALIDATION_ERROR` response | Consistent public error shape for invalid input |

Already documented elsewhere and not repeated in full:

| Topic | Refer |
|---|---|
| Base Go/Git/Docker install | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `3. Required Software` |
| Go modules basics | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `4. Dependency Management` |
| Full platform DB setup | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `5. Database Setup` |
| Redis setup | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `5. Database Setup` -> `Redis` |
| Downstream gRPC addresses | `TaskImplementation/API Gateway Service/task2_Dependency.md` -> `7. Environment Variables` |
| JWT/JWKS setup | `TaskImplementation/API Gateway Service/task3_Dependency.md` -> `7. Environment Variables` |
| Rate-limit Redis setup | `TaskImplementation/API Gateway Service/task4_Dependency.md` -> `5. Database Setup` |

---

## 2. Tech Stack

### Task 5 Specific Technologies

| Technology | What it is | Why this project uses it | Required? | Beginner Hinglish explanation |
|---|---|---|---:|---|
| Go | Backend programming language | API Gateway service Go me implemented hai | Yes | Go ek fast compiled backend language hai. Gateway isi se build/run/test hota hai. |
| Go Modules | Go dependency system | `go.mod` and `go.sum` dependency versions manage karte hain | Yes | Go modules npm/pip jaisa dependency manager hai, Go projects ke liye. |
| `net/http` | Go standard HTTP server and middleware package | Request body limit, handlers, headers, path values ke liye | Yes | Is service me external web framework nahi hai; Go ka built-in HTTP package use ho raha hai. |
| `encoding/json` | Go standard JSON parser | Request body ko `map[string]any` me decode karne ke liye | Yes | JSON body ko Go data structure me convert karta hai. |
| `mime` | Go standard media-type parser | `Content-Type: application/json; charset=utf-8` ko safely parse karne ke liye | Yes | Header ko exact string compare karne ke bajay proper parse karta hai. |
| `regexp` | Go standard regex package | Public id, provider, coupon code, SKU, status patterns validate karne ke liye | Yes | Regex se format rules check hote hain. |
| `net/mail` | Go standard email parser | Email format validation ke liye | Yes | Email string valid address jaisi hai ya nahi check karta hai. |
| `api/master-api.json` | Route and schema contract file | Request schema names, fields, required fields, enums, min/max limits yahin se load hote hain | Yes | Ye Gateway ka API map hai. Validation ko pata chalta hai kaunsa route kaunsa schema expect karta hai. |
| `VALIDATION_ERROR` envelope | Consistent REST error response | Frontend ko predictable error format dene ke liye | Yes | Har invalid request same shape me error return karti hai. |

### Important Difference From `task5.md` Recommendation

`task5.md` implementation guide ne `github.com/go-playground/validator/v10` and `github.com/go-chi/chi/v5` ko recommended options ke roop me mention kiya tha.

Current repo me ye libraries add nahi ki gayi hain:

| Library | Current status | What to do |
|---|---|---|
| `github.com/go-playground/validator/v10` | Not present in `go.mod` | Install mat karo unless implementation intentionally struct-tag validator par shift hoti hai |
| `github.com/go-chi/chi/v5` | Not present in `go.mod` | Install mat karo; current code Go `net/http` path params use karta hai |

Beginner rule: `go get` tabhi run karo jab code actually new package import karta ho. Sirf documentation me mentioned package ke liye dependency add karna unnecessary dependency bloat create karta hai.

### Reused Stack

General Gateway, gRPC, JWT, Redis, metrics, and tracing stack already explained hai:

- `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `2. Tech Stack`
- `TaskImplementation/API Gateway Service/task2_Dependency.md` -> `2. Tech Stack`
- `TaskImplementation/API Gateway Service/task3_Dependency.md` -> `2. Tech Stack`
- `TaskImplementation/API Gateway Service/task4_Dependency.md` -> `2. Tech Stack`

---

## 3. Required Software

### Required for Task 5 Development

| Software | Required? | Why |
|---|---:|---|
| Git | Yes | Repository clone/update ke liye |
| Go `1.26.3` compatible toolchain | Yes | `backend/services/api-gateway/go.mod` declares `go 1.26.3` |
| curl | Recommended | Validation errors manually test karne ke liye |
| Docker | Recommended for full runtime | Redis and downstream dependencies run karne ke liye |
| jq | Optional | JSON error responses pretty-print karne ke liye |

Installation process already documented hai:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`3. Required Software`

### Setup Modes

| Mode | Redis Needed? | Downstream gRPC Needed? | JWKS Needed? | Use Case |
|---|---:|---:|---:|---|
| Task 5 unit tests only | No | No | No | `internal/validation` tests run karne ke liye |
| Router validation tests | No | No | No, stub verifier is used | HTTP middleware behavior verify karne ke liye |
| Full `go run ./cmd/server` | Yes if rate limit enabled | Yes | Yes for protected routes | Real local Gateway runtime |

Task 5 khud Redis/JWKS/gRPC introduce nahi karta. Full runtime me ye dependencies previous tasks ki wajah se required hain.

---

## 4. Dependency Management

This is a Go project. Base Go module explanation already covered hai:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`4. Dependency Management`

### Current Go Module

```text
module ecommerce/api-gateway
go 1.26.3
```

### Task 5 New External Libraries

No new third-party Go dependency is required for current Task 5 implementation.

Task 5 uses mostly Go standard library packages:

| Package | Why used |
|---|---|
| `encoding/json` | JSON body decode and schema value parsing |
| `net/http` | Middleware, `http.MaxBytesReader`, status codes |
| `mime` | Content-Type parsing |
| `regexp` | Format rules |
| `net/mail` | Email validation |
| `time` | RFC3339 date-time and date range validation |
| `strings` | Trim/normalize field values |

### Commands

Run from API Gateway service directory:

```bash
cd backend/services/api-gateway
```

Download existing dependencies:

```bash
go mod download
```

Run Task 5 focused tests:

```bash
go test ./internal/validation ./internal/transport/http ./internal/config
```

Run all Gateway tests:

```bash
go test ./...
```

Build:

```bash
go build ./cmd/server
```

Run:

```bash
go run ./cmd/server
```

### Common Go Dependency Issues

General fixes already documented:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`4. Dependency Management` -> `Dependency Problems and Fixes`

Task 5 specific notes:

| Error | Likely Cause | Fix |
|---|---|---|
| `go: go.mod requires go >= 1.26.3` | Local Go version old hai | Install compatible Go version |
| `cannot find module github.com/go-playground/validator/v10` | Developer copied code from task guide but dependency not added | Current implementation does not need it; either remove that copied import or intentionally add dependency |
| `cannot find module github.com/go-chi/chi/v5` | Developer copied chi example from task guide | Current router uses `net/http`; use `r.PathValue(...)` instead |
| `missing go.sum entry` | Dependency checksum missing after module edits | Run `go mod tidy` |

---

## 5. Database Setup

### API Gateway Task 5 Direct Database Requirement

Task 5 does not directly use any SQL/document database.

| Store | Directly used by Task 5? | Purpose |
|---|---:|---|
| MySQL | No | Owned by downstream services |
| MongoDB | No | Owned by downstream services |
| Redis | No direct validation store | Inherited from Task 4 rate limiting |
| Typesense | No | Owned by Search Service |
| Kafka/RabbitMQ | No | Platform async/event services, not validation |

### Reuse Previous Database Documentation

Do not duplicate DB installation here. Follow previous guide sections:

| Database / Store | Where setup is already explained |
|---|---|
| MySQL | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `5. Database Setup` -> `MySQL` |
| MongoDB | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `5. Database Setup` -> `MongoDB` |
| Redis | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `5. Database Setup` -> `Redis` |
| Typesense | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `6. Redis / Queue / External Services` -> `Typesense` |
| Kafka / RabbitMQ | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `6. Redis / Queue / External Services` -> `Kafka or RabbitMQ` |

### Migrations

No migration is required for Task 5.

Why:

- API Gateway validation is in-memory code.
- It reads `api/master-api.json`.
- It does not create database tables.
- It does not create Redis keys for validation.

If a downstream service needs DB schema changes because a field became required, that migration belongs to that downstream service, not API Gateway Task 5.

---

## 6. Redis / Queue / External Services

### Direct New External Service for Task 5

None.

Task 5 does not add Redis, Kafka, RabbitMQ, S3, SMTP, Stripe, Twilio, Firebase, MinIO, or Elasticsearch.

### Important Task 5 Configuration Dependency: API Contract File

Task 5 depends heavily on:

```text
api/master-api.json
```

What this file provides:

| Contract Data | Used By Validation For |
|---|---|
| `rest_endpoints[].id` | Route-specific policy, for example `order.checkout` needs idempotency |
| `rest_endpoints[].method` | GET/DELETE body rejection, POST/PATCH JSON body validation |
| `rest_endpoints[].path` | Path param names like `{product_id}` |
| `rest_endpoints[].auth` | Webhook route special handling |
| `rest_endpoints[].request_schema` | Schema name used for body/query validation |
| `schemas` | Required fields, object properties, min/max, enum, format |

Where to configure it:

```env
API_CONTRACT_PATH=../../../api/master-api.json
```

This variable is already documented in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`7. Environment Variables`

Beginner explanation: API contract file validation ka map hai. Agar schema missing ya stale hai, Gateway ko sahi field rules pata nahi chalenge.

### Reused External Services

These are inherited from earlier tasks:

| Service | Why Gateway still needs it in full runtime | Setup reference |
|---|---|---|
| Redis | Rate limiting from Task 4 | `task4_Dependency.md` -> `5. Database Setup` |
| Auth Service JWKS endpoint | JWT verification from Task 3 | `task3_Dependency.md` -> `6. Redis / Queue / External Services` |
| Downstream gRPC services | Route clients and readiness from Task 2 | `task2_Dependency.md` -> `6. Redis / Queue / External Services` |
| Metrics server | Validation failure metrics if metrics enabled | `task1_Dependency.md` -> `7. Environment Variables` and Task 7 docs later |
| OTLP collector | Tracing if trace enabled | `task1_Dependency.md` -> `7. Environment Variables`; optional for local |

### Ports & Networking

Task 5 introduces no new port.

| Service | Port | Purpose | Task 5 Status |
|---|---:|---|---|
| API Gateway HTTP | `8080` by default | REST routes and validation responses | Reused |
| API Gateway metrics | `9090` by default | Metrics including validation failures | Reused |
| Redis | `6379` | Rate-limit counters from Task 4 | Inherited |
| Auth Service HTTP/JWKS | `8081` example | JWT public keys from Task 3 | Inherited |
| Downstream gRPC services | `50051` to `50062` local examples | Internal service targets | Inherited |
| MySQL | `3306` | Downstream service DB | Not direct |
| MongoDB | `27017` | Downstream service DB | Not direct |
| Typesense | `8108` | Search Service index | Not direct |
| Kafka / Redpanda | `9092` | Event streaming | Not direct |
| RabbitMQ AMQP | `5672` | Queue alternative | Not direct |

Networking rules are unchanged from previous tasks. Refer:

- `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `7. Environment Variables` -> `Ports & Networking`
- `TaskImplementation/API Gateway Service/task4_Dependency.md` -> `6. Redis / Queue / External Services` -> `Ports & Networking`

---

## 7. Environment Variables

Full Gateway `.env` and export process already documented in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Sections:

- `7. Environment Variables`
- `7. Environment Variables` -> `Complete .env Example`
- `7. Environment Variables` -> `Common .env Mistakes`

Task 5 does not add variables that were not already listed in Task 1. This section explains how to tune the validation-specific variables.

### Task 5 Validation Variables

| Variable | Required? | Default in code | Purpose | Beginner Notes |
|---|---:|---|---|---|
| `REQUEST_VALIDATION_ENABLED` | Optional | `true` | Enables request validation middleware | Local debugging me false kar sakte ho, but normal dev/staging/prod me true rakho |
| `REQUEST_VALIDATION_DEFAULT_MAX_BODY_BYTES` | Optional | `131072` | Fallback body limit, 128 KB | Route-specific limits can override this |
| `REQUEST_VALIDATION_MAX_HEADER_BYTES` | Optional | `32768` | Total request header limit, 32 KB | Too many/large headers reject honge |
| `REQUEST_VALIDATION_MAX_QUERY_BYTES` | Optional | `8192` | Raw query string limit, 8 KB | Huge query string abuse prevent karta hai |

Recommended local values are the same as Task 1:

```env
REQUEST_VALIDATION_ENABLED=true
REQUEST_VALIDATION_DEFAULT_MAX_BODY_BYTES=131072
REQUEST_VALIDATION_MAX_HEADER_BYTES=32768
REQUEST_VALIDATION_MAX_QUERY_BYTES=8192
```

This snippet is repeated here only because Task 5 owns these settings. Full `.env` should still be maintained using Task 1 guide.

### Where To Place These Variables

Recommended local file:

```text
backend/services/api-gateway/.env
```

Current Go code uses `os.Getenv`. It does not auto-load `.env`. Before running:

```bash
cd backend/services/api-gateway
set -a
source .env
set +a
```

Windows PowerShell export process is already shown in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`7. Environment Variables`

### Validation Limit Rules

| Setting | Must Be | If Invalid |
|---|---|---|
| `REQUEST_VALIDATION_DEFAULT_MAX_BODY_BYTES` | Positive integer | Startup config validation fails |
| `REQUEST_VALIDATION_MAX_HEADER_BYTES` | Positive integer | Startup config validation fails |
| `REQUEST_VALIDATION_MAX_QUERY_BYTES` | Positive integer | Startup config validation fails |
| `REQUEST_VALIDATION_ENABLED` | Valid boolean like `true`, `false`, `1`, `0`, `yes`, `no` | Startup config load fails |

### Credentials Placement

Task 5 does not introduce a new secret.

| Data | Put in Gateway `.env`? | Notes |
|---|---:|---|
| Validation enable/limit values | Yes | Safe config values |
| API contract path | Yes | Safe path value |
| DB username/password | No for Task 5 | Downstream service `.env` owns DB credentials |
| Redis password | Inherited from Task 4 | Use Gateway `.env` only if rate limiter uses Redis auth |
| JWT private key | No | Auth Service secret manager / mounted secret |
| Webhook provider secret | Usually no | Payment Service should own provider secret |

---

## 8. Docker Setup

Docker basics, dependency compose examples, and Redis container setup are already documented in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Sections:

- `5. Database Setup` -> `Redis` -> `E. Docker Setup`
- `8. Docker Setup`

Task 5 does not require a new Docker image, new container, new volume, or new network.

### Current Repo Status

No committed API Gateway Dockerfile or docker-compose file was found for this service.

Recommended beginner flow remains:

```text
Run dependencies with Docker where needed.
Run API Gateway locally with go run.
Use previous task docs for Redis/JWKS/downstream gRPC service setup.
```

### Docker Configuration Impact

If a Dockerfile or compose service is added later, include these validation env vars in the Gateway container environment:

```yaml
environment:
  REQUEST_VALIDATION_ENABLED: "true"
  REQUEST_VALIDATION_DEFAULT_MAX_BODY_BYTES: "131072"
  REQUEST_VALIDATION_MAX_HEADER_BYTES: "32768"
  REQUEST_VALIDATION_MAX_QUERY_BYTES: "8192"
  API_CONTRACT_PATH: "/app/api/master-api.json"
```

Also make sure `api/master-api.json` is copied or mounted into the container. Without the contract file, startup fails because route catalog cannot load.

Example future volume idea:

```yaml
volumes:
  - ../../../api/master-api.json:/app/api/master-api.json:ro
```

This is guidance only. Do not create compose changes unless infra ownership requires it.

### Volumes

Task 5 does not create persistent data. No new volume is needed.

Existing volumes for Redis/MySQL/MongoDB/Typesense are already covered in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`8. Docker Setup` -> `Docker Volumes`

---

## 9. Local Development Setup

### Reused Base Onboarding

Clone repo, install Go, download modules, and export the base `.env` using:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Sections:

- `9. Local Development Setup`
- `10. Running the Project`

### Task 5 Incremental Setup

1. Go to API Gateway service:

```bash
cd backend/services/api-gateway
```

2. Confirm the API contract file exists:

```bash
ls ../../../api/master-api.json
```

3. Confirm validation env values are present in your `.env` or shell:

```bash
printenv REQUEST_VALIDATION_ENABLED
printenv REQUEST_VALIDATION_DEFAULT_MAX_BODY_BYTES
printenv REQUEST_VALIDATION_MAX_HEADER_BYTES
printenv REQUEST_VALIDATION_MAX_QUERY_BYTES
```

Expected local values:

```text
true
131072
32768
8192
```

4. Run validation-focused tests:

```bash
go test ./internal/validation
```

5. Run router/config tests that include validation middleware:

```bash
go test ./internal/transport/http ./internal/config
```

6. Run all Gateway tests:

```bash
go test ./...
```

7. For full runtime, start inherited dependencies:

| Dependency | Setup reference |
|---|---|
| Redis | `task4_Dependency.md` -> `9. Local Development Setup` |
| Auth JWKS endpoint | `task3_Dependency.md` -> `9. Local Development Setup` |
| Downstream gRPC services | `task2_Dependency.md` -> `9. Local Development Setup` |

8. Start Gateway:

```bash
go run ./cmd/server
```

9. Verify health:

```bash
curl http://localhost:8080/health/live
curl http://localhost:8080/health/ready
```

Beginner note: `/health/live` only proves process is alive. `/health/ready` can fail if downstream gRPC services are not ready.

---

## 10. Running the Project

### Minimal Task 5 Verification

Use this when you only want to verify validation code:

```bash
cd backend/services/api-gateway
go test ./internal/validation ./internal/transport/http ./internal/config
```

This does not need Redis, JWKS, databases, or downstream services.

### Full Runtime Verification

Full runtime still follows previous task dependencies:

1. Export Gateway `.env`.
2. Start Redis if `RATE_LIMIT_ENABLED=true`.
3. Start or mock Auth JWKS endpoint.
4. Start downstream gRPC services or use dev stubs if available.
5. Run:

```bash
cd backend/services/api-gateway
go run ./cmd/server
```

6. Send a validation test request.

Public route example:

```bash
curl -i -X POST http://localhost:8080/api/v1/auth/signup \
  -H "Content-Type: application/json" \
  -d '{"email":"not-an-email","password":"short","full_name":"Buyer One"}'
```

Expected result:

```text
HTTP 400
error.code = VALIDATION_ERROR
```

Unsupported content type example:

```bash
curl -i -X POST http://localhost:8080/api/v1/auth/signup \
  -H "Content-Type: text/plain" \
  -d '{"email":"buyer@example.com"}'
```

Expected result:

```text
HTTP 415
error.code = VALIDATION_ERROR
```

Large body example:

```bash
python3 -c 'print("{\"email\":\"buyer@example.com\",\"password\":\"" + "x"*200000 + "\",\"full_name\":\"Buyer One\"}")' \
  | curl -i -X POST http://localhost:8080/api/v1/auth/signup \
      -H "Content-Type: application/json" \
      --data-binary @-
```

Expected result:

```text
HTTP 413
error.code = VALIDATION_ERROR
```

Note: If the route passes validation, current handler may still return `501 ROUTE_BRIDGE_NOT_CONFIGURED` because downstream route bridge is a later Gateway task.

### Migrations

No migration command is required for Task 5.

---

## 11. Common Errors & Fixes

### Task 5 Specific Troubleshooting

| Error / Symptom | Likely Cause | Fix |
|---|---|---|
| `API_CONTRACT_PATH is not readable` | Wrong relative path or file missing | From `backend/services/api-gateway`, use `API_CONTRACT_PATH=../../../api/master-api.json` |
| Validation seems disabled | `REQUEST_VALIDATION_ENABLED=false` or env not exported | Export `.env` and set `REQUEST_VALIDATION_ENABLED=true` |
| Startup fails with `REQUEST_VALIDATION_DEFAULT_MAX_BODY_BYTES must be positive` | Env value is `0`, negative, or invalid | Use a positive integer like `131072` |
| `415 VALIDATION_ERROR` on POST/PATCH | Missing or wrong `Content-Type` | Send `Content-Type: application/json` for JSON routes |
| `413 VALIDATION_ERROR` | Request body larger than route policy | Reduce payload size or adjust limit carefully |
| `400 VALIDATION_ERROR` with `query unknown_field` | Query param is not present in route request schema | Add correct schema field in `api/master-api.json` or remove query param |
| `400 VALIDATION_ERROR` with `body not_allowed` | GET/DELETE route received a body, or schema says `Empty` | Do not send body for GET/DELETE/Empty routes |
| `400 VALIDATION_ERROR` with `Idempotency-Key required` | Route `order.checkout`, `payment.retry`, or `payment.refund` needs idempotency | Send `Idempotency-Key` header or matching `idempotency_key` body field |
| `400 VALIDATION_ERROR` with `Idempotency-Key mismatch` | Header and body idempotency keys differ | Use one canonical key and keep both same if both are sent |
| Valid request returns `501 ROUTE_BRIDGE_NOT_CONFIGURED` | Validation passed, but route bridge not implemented yet | This is expected in current Gateway state |

### Reused Troubleshooting

Do not duplicate previous troubleshooting. Refer:

| Problem Area | Reference |
|---|---|
| Go module download/version issues | `task1_Dependency.md` -> `4. Dependency Management` |
| Redis ping/rate-limit startup failure | `task4_Dependency.md` -> `11. Common Errors & Fixes` |
| Missing JWKS/Auth config | `task3_Dependency.md` -> `11. Common Errors & Fixes` |
| Missing downstream gRPC targets | `task2_Dependency.md` -> `11. Common Errors & Fixes` |
| Docker port/container issues | `task1_Dependency.md` -> `8. Docker Setup` and `11. Common Errors & Fixes` |

---

## 12. Security & Best Practices

### Task 5 Security Notes

| Security Area | Current Behavior | Recommendation |
|---|---|---|
| Body size | Uses `http.MaxBytesReader` | Keep route-specific limits strict |
| Header size | Counts total header bytes and rejects large headers | Keep `REQUEST_VALIDATION_MAX_HEADER_BYTES` sane |
| Query size | Rejects huge query strings | Keep `REQUEST_VALIDATION_MAX_QUERY_BYTES` sane |
| Unknown JSON fields | Schema validator rejects unknown fields | Keep schemas updated so client typos are caught |
| Sensitive fields | Validation errors do not echo password values in tests | Continue avoiding raw token/password/OTP in logs |
| Webhooks | Raw body preserved in context for webhook routes | Signature verification should remain provider-aware in Payment Service or webhook auth layer |
| Idempotency | Checkout/payment routes require idempotency key | Prefer `Idempotency-Key` header for command safety |
| Request schemas | Loaded from `api/master-api.json` | Treat contract changes like API changes and review carefully |

### Beginner Best Practices

- Keep `REQUEST_VALIDATION_ENABLED=true` in local/staging/prod.
- Do not increase body limits just to make a failing request pass. First ask why payload is large.
- Add schema fields in `api/master-api.json` before sending new fields from frontend.
- Do not place DB passwords, JWT private keys, or payment provider secrets in validation config.
- Keep Gateway validation for shape/format/size only. Business rules should stay in downstream services.
- When adding a new route, update route contract, request schema, tests, and validation expectations together.
- Prefer allowlists for enums like `payment_provider`, `sort`, and `status`.

---

## 13. Missing or Misconfigured Things

### Audit Findings

| Finding | Impact | Suggested Fix |
|---|---|---|
| No committed API Gateway Dockerfile/compose file | Beginner cannot run full stack with one command | Add maintained Dockerfile and local compose when infra ownership is ready |
| `api/master-api.json` schema is runtime-critical | Missing/stale schema can weaken or break validation | Add contract validation CI and schema coverage tests |
| Missing schema currently may skip body schema validation | A route with unknown `request_schema` may not get full schema checks | Fail closed when route schema name is missing from contract |
| No third-party validator library is used | Current custom validator is fine, but schema behavior must be tested thoroughly | Keep table-driven tests for each route group |
| Webhook route allows JSON/form content types but signature algorithm is not verified in Task 5 | Basic header/body validation only | Keep real signature verification in Task 3/Payment Service flow |
| Full runtime still fails if downstream gRPC services are unavailable | Validation-only developers may get blocked by unrelated deps | Use focused tests for Task 5, and provide local gRPC stubs later |
| No automatic `.env` loading | Beginners may create `.env` but forget to export | Document `set -a; source .env; set +a` and consider dev-only env loader |
| No migrations | Correct for Gateway, but downstream field changes may need migrations | Coordinate schema changes with owning service migrations |

### Hardcoded Credentials Check

Task 5 did not introduce hardcoded credentials.

Values such as body limits, content types, route-id rules, and regex patterns are configuration/validation policy, not secrets.

---

## 14. References to Previous Dependency Files

Use these instead of duplicating setup:

| Topic | Reference |
|---|---|
| Full beginner setup and `.env` | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `7. Environment Variables` |
| Required software install | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `3. Required Software` |
| Go module explanation | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `4. Dependency Management` |
| MySQL/MongoDB/Redis install and Docker | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `5. Database Setup` |
| Docker compose examples | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `8. Docker Setup` |
| Downstream gRPC service addresses | `TaskImplementation/API Gateway Service/task2_Dependency.md` -> `7. Environment Variables` |
| Auth/JWT/JWKS setup | `TaskImplementation/API Gateway Service/task3_Dependency.md` -> `6. Redis / Queue / External Services` and `7. Environment Variables` |
| Redis rate-limit setup | `TaskImplementation/API Gateway Service/task4_Dependency.md` -> `5. Database Setup` and `7. Environment Variables` |
| Rate-limit troubleshooting | `TaskImplementation/API Gateway Service/task4_Dependency.md` -> `11. Common Errors & Fixes` |

---

## 15. Final Checklist

Before saying Task 5 setup is ready, verify:

| Check | Done |
|---|---|
| Go version is compatible with `go 1.26.3` |  |
| `backend/services/api-gateway/go mod download` completed |  |
| `api/master-api.json` exists and is readable |  |
| `API_CONTRACT_PATH` points to the correct contract file |  |
| `REQUEST_VALIDATION_ENABLED=true` is exported |  |
| Validation limits are positive integers |  |
| `go test ./internal/validation` passes |  |
| `go test ./internal/transport/http ./internal/config` passes |  |
| `go test ./...` passes |  |
| Full runtime dependencies from Tasks 2, 3, and 4 are started if using `go run` |  |
| Invalid JSON request returns `VALIDATION_ERROR` |  |
| Unsupported content type returns HTTP `415` |  |
| Oversized body returns HTTP `413` |  |
| GET/DELETE body is rejected where body is not allowed |  |
| Idempotency key requirement works for checkout/payment routes |  |
| No secrets are stored in Task 5 validation config |  |

Final mental model:

```text
Task 5 setup = Go service + API contract schemas + validation env limits.
Databases, Redis, JWKS, and gRPC services are inherited from previous Gateway tasks.
```
