# Project Dependency & Setup Guide

Input analyzed: `TaskImplementation/API Gateway Service/task2.md`

Previous dependency documentation checked first:

- `TaskImplementation/API Gateway Service/task1_Dependency.md`

Related implementation inspected:

- `backend/services/api-gateway/go.mod`
- `backend/services/api-gateway/go.sum`
- `backend/go.work`
- `backend/services/api-gateway/cmd/server/main.go`
- `backend/services/api-gateway/internal/config/config.go`
- `backend/services/api-gateway/internal/clients/*`
- `backend/services/api-gateway/internal/transport/http/handler.go`
- `TaskImplementation/Platform Foundation/task2.md`
- `docs/03-folder-structure.md`
- `docs/13-developer-guide.md`

> Simple goal: Ye guide API Gateway ke Task 2, yani gRPC clients setup, ko run/debug karne ke liye required dependencies, env vars, service addresses, health checks, ports, and DevOps assumptions explain karta hai. Business logic ya API implementation yahan repeat nahi ki gayi.

---

## 1. Project Overview

API Gateway Service public REST entry point hai. Task 2 ka focus Gateway ke andar outbound gRPC client layer hai.

Flow:

```text
Browser / Frontend
  -> REST API Gateway
  -> reusable gRPC client registry
  -> downstream microservice
```

Current code me gRPC client setup ye files handle karti hain:

| File | Purpose |
|---|---|
| `internal/config/config.go` | `*_GRPC_ADDR`, `GRPC_TLS_ENABLED`, `GRPC_DIAL_TIMEOUT` env vars load/validate karta hai |
| `internal/clients/service.go` | Downstream service list, health service names, default timeouts define karta hai |
| `internal/clients/dialer.go` | gRPC connection create karta hai with TLS/insecure, backoff, keepalive, and timeout |
| `internal/clients/clients.go` | 12 downstream services ka central client registry initialize/close karta hai |
| `internal/clients/health.go` | `/health/ready` ke liye downstream gRPC health checks chalata hai |
| `internal/clients/metadata.go` | Request id, trace headers, and safe auth metadata downstream forward karta hai |
| `internal/clients/timeouts.go` | Service-wise RPC deadline strategy define karta hai |

Important current reality:

- Gateway direct database use nahi karta.
- Gateway startup par all 12 downstream gRPC addresses required hain.
- Startup client initialization fail-fast hai: agar koi required service address missing hai ya service reachable nahi hai, Gateway start nahi hoga.
- Current repo me proto source folder and generated proto client packages abhi present nahi hain. Current client layer typed wrapper interfaces expose karta hai, but real generated RPC methods future proto generation ke baad wire honge.

---

## 2. Tech Stack

### Task 2 Specific Technologies

| Technology | What it is | Why this task uses it | Required? | Beginner Hinglish explanation |
|---|---|---|---:|---|
| Go | Backend programming language | API Gateway Go service me implemented hai | Yes | Go ek compiled backend language hai jo fast services banane ke liye use hoti hai. |
| Go Modules | Dependency management system | `go.mod` and `go.sum` se versions lock/manage hote hain | Yes | Go modules npm/pip jaisa dependency system hai, bas Go ke liye. |
| gRPC | RPC framework | Gateway internal services ko typed network calls bhejne ke liye | Yes | gRPC service-to-service communication ke liye fast protocol hai. REST public side par hai, gRPC internal side par. |
| Protobuf | Contract/data format | Future generated clients and request/response types ke liye | Required for full bridge | Protobuf schema file se typed Go clients generate hote hain. Abhi repo me proto generated clients missing hain. |
| Buf CLI | Proto lint/generate tool | Future `buf generate` workflow ke liye | Required when proto files are added | Buf proto files ko lint karta hai aur Go/TS clients generate karta hai. |
| gRPC Health Checking | Standard health service | Gateway `/health/ready` downstream services verify karta hai | Yes for readiness | Har downstream service ko batana hoga ki wo `SERVING` hai ya nahi. |
| gRPC Metadata | Request metadata carrier | Request id, trace context, and user context downstream bhejne ke liye | Yes | Metadata headers jaisa hota hai, bas gRPC calls ke saath. |
| gRPC Keepalive/Backoff | Connection reliability settings | Long-lived connections healthy and reconnect-friendly rakhne ke liye | Yes | Connection tootne par controlled reconnect hota hai. |
| TLS / insecure gRPC credentials | Transport security mode | Local me insecure, prod me TLS/mTLS expected | Local optional, prod required | Local dev me TLS off ho sakta hai, production me encrypted service traffic hona chahiye. |
| Prometheus client | Metrics library | gRPC client request/error metrics expose karne ke liye | Optional by config | Metrics se pata chalta hai kaunsi downstream service slow/failing hai. |
| OpenTelemetry | Tracing library | Outbound gRPC client spans create karne ke liye | Optional by config | Tracing se ek request ka full journey Gateway se service tak track hota hai. |

### Already Explained in Previous Dependency Guide

Do not repeat these setup sections here. Refer:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Sections:

- `2. Tech Stack`
- `3. Required Software`
- `4. Dependency Management`
- `5. Database Setup`
- `6. Redis / Queue / External Services`
- `8. Docker Setup`

---

## 3. Required Software

### Required for Task 2 Code/Test Work

| Software | Required? | Why |
|---|---:|---|
| Git | Yes | Repository clone and source control |
| Go `1.26.3` compatible toolchain | Yes | `backend/go.work` and `backend/services/api-gateway/go.mod` declare `go 1.26.3` |
| curl | Recommended | HTTP health endpoint check karne ke liye |
| grpcurl | Recommended | Downstream gRPC services manually verify karne ke liye |
| Buf CLI | Future required | Proto files add hone ke baad generated clients banane ke liye |

Base installation steps for Git, Go, Docker, and OS-specific commands are already explained in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`3. Required Software`

### grpcurl Install

`grpcurl` ek CLI tool hai jisse aap gRPC endpoint ko terminal se test kar sakte ho.

macOS:

```bash
brew install grpcurl
```

Go install option:

```bash
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

Verify:

```bash
grpcurl --version
```

### Buf CLI Install

Proto strategy already documented in:

`TaskImplementation/Platform Foundation/task2.md`

Relevant sections:

- `Step 9: Buf as Proto Tooling Define Kiya`
- `Step 10: gRPC and Protobuf Libraries Define Kiye`

If proto files are later added, install Buf:

```bash
brew install bufbuild/buf/buf
```

Verify:

```bash
buf --version
```

---

## 4. Dependency Management

This is a Go project. General Go modules explanation is already covered in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`4. Dependency Management`

### Task 2 Specific Go Dependencies

Current `backend/services/api-gateway/go.mod` includes these Task 2 relevant dependencies:

| Dependency | Version | Why needed |
|---|---:|---|
| `google.golang.org/grpc` | `v1.81.1` | gRPC client connections, dial options, health client, metadata, status codes |
| `google.golang.org/protobuf` | `v1.36.11` | Protobuf runtime for generated messages once proto clients are added |
| `google.golang.org/genproto/googleapis/rpc` | `v0.0.0-20260401024825-9d38bb4040a9` | Google RPC status/details support |
| `go.opentelemetry.io/otel` | `v1.43.0` | Client-side tracing spans and propagation |
| `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc` | `v1.43.0` | Sends traces to OTLP collector over gRPC |
| `go.opentelemetry.io/otel/sdk` | `v1.43.0` | OpenTelemetry tracer provider |
| `github.com/prometheus/client_golang` | `v1.23.2` | gRPC client metrics |

### Commands

Run from API Gateway service directory:

```bash
cd backend/services/api-gateway
go mod download
go test ./internal/clients ./internal/config
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

### Proto Generation Status

Task 2 implementation guide expects generated proto clients in the future, but current repo does not contain:

```text
proto/
proto/buf.yaml
proto/buf.gen.yaml
backend/shared/gen/go/...
```

So this command is future workflow only:

```bash
buf generate
```

Warning: Abhi `buf generate` run karne ke liye required proto config/files repo me nahi mile. Jab Platform Foundation proto files add ho jayenge, tab Gateway generated clients ko import kar payega.

### Common Go/gRPC Dependency Issues

| Error | Cause | Fix |
|---|---|---|
| `go: go.mod requires go >= 1.26.3` | Local Go version old hai | Go `1.26.3` compatible version install karo |
| `missing go.sum entry` | Dependency checksum missing | `go mod tidy` run karo |
| `cannot find package .../gen/go/...` | Generated proto clients missing hain | Proto files and Buf config add karke `buf generate` run karo |
| `module lookup disabled` | Go proxy/network issue | `go env GOPROXY` check karo |
| `grpc: no transport security set` | gRPC credentials missing in code | Current dialer already credentials set karta hai; custom code me `WithTransportCredentials` use karo |

---

## 5. Database Setup

### API Gateway Task 2 Direct Database Requirement

No direct database is required.

Gateway gRPC clients:

- MySQL connect nahi karte.
- MongoDB connect nahi karte.
- Migrations run nahi karte.
- DB credentials store nahi karte.

Downstream services apni database ownership maintain karenge. Example: Auth Service MySQL/Redis use karega, Product Service MongoDB use karega, Search Service Typesense use karega.

### Reuse Previous Database Documentation

Full platform database setup already explained in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Sections:

- `5. Database Setup`
- `MySQL`
- `MongoDB`
- `Redis`
- `Typesense`
- `Kafka or RabbitMQ`

Task 2 me koi new DB, new DB port, new DB credentials, or new migration command introduce nahi hua.

---

## 6. Redis / Queue / External Services

### Direct New External Dependency for Task 2: Downstream gRPC Services

Task 2 ka main external dependency downstream gRPC services hain. Gateway startup par in sab services ke addresses validate/dial karta hai.

| Env Var | Service | Health Service Name | Local Example | Docker/K8s Example | Default RPC Timeout |
|---|---|---|---|---|---:|
| `AUTH_GRPC_ADDR` | Auth Service | `ecommerce.auth.v1.AuthService` | `localhost:50051` | `auth-service:9090` | `300ms` |
| `USER_GRPC_ADDR` | User Service | `ecommerce.user.v1.UserService` | `localhost:50052` | `user-service:9090` | `700ms` |
| `PRODUCT_GRPC_ADDR` | Product Service | `ecommerce.product.v1.ProductService` | `localhost:50053` | `product-service:9090` | `500ms` |
| `CART_GRPC_ADDR` | Cart Service | `ecommerce.cart.v1.CartService` | `localhost:50054` | `cart-service:9090` | `700ms` |
| `WISHLIST_GRPC_ADDR` | Wishlist Service | `ecommerce.wishlist.v1.WishlistService` | `localhost:50055` | `wishlist-service:9090` | `700ms` |
| `ORDER_GRPC_ADDR` | Order Service | `ecommerce.order.v1.OrderService` | `localhost:50056` | `order-service:9090` | `1500ms` |
| `PAYMENT_GRPC_ADDR` | Payment Service | `ecommerce.payment.v1.PaymentService` | `localhost:50057` | `payment-service:9090` | `1500ms` |
| `SEARCH_GRPC_ADDR` | Search Service | `ecommerce.search.v1.SearchService` | `localhost:50058` | `search-service:9090` | `300ms` |
| `CMS_GRPC_ADDR` | CMS Service | `ecommerce.cms.v1.CMSService` | `localhost:50059` | `cms-service:9090` | `1000ms` |
| `SESSION_GRPC_ADDR` | Session Service | `ecommerce.session.v1.SessionService` | `localhost:50060` | `session-service:9090` | `200ms` |
| `NOTIFICATION_GRPC_ADDR` | Notification Service | `ecommerce.notification.v1.NotificationService` | `localhost:50061` | `notification-service:9090` | `700ms` |
| `SUPERADMIN_GRPC_ADDR` | Superadmin Service | `ecommerce.superadmin.v1.SuperadminService` | `localhost:50062` | `superadmin-service:9090` | `1000ms` |

Beginner explanation:

`AUTH_GRPC_ADDR=localhost:50051` ka matlab Gateway same machine par port `50051` par Auth Service gRPC server se connect karega. Docker Compose ke andar `localhost` use nahi hota; wahan service name use hota hai, jaise `auth-service:9090`.

### Required gRPC Health Support

Gateway `/health/ready` downstream health check karta hai using standard gRPC health service:

```text
grpc.health.v1.Health/Check
```

Each downstream service should register health status for its exact service name. Example:

```text
ecommerce.product.v1.ProductService -> SERVING
```

Verify with grpcurl:

```bash
grpcurl -plaintext \
  -d '{"service":"ecommerce.product.v1.ProductService"}' \
  localhost:50053 \
  grpc.health.v1.Health/Check
```

If server reflection is disabled, use the health request manually with proto support in the downstream repo. At minimum, service logs should confirm health service registration.

### Ports & Networking

Full platform port mapping is already available in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`7. Environment Variables` -> `Ports & Networking`

Task 2 specific gRPC client ports:

| Service | Port | Purpose | Status |
|---|---:|---|---|
| API Gateway HTTP | `8080` | Public REST and health endpoints | Reused from Task 1 |
| API Gateway Metrics | `9090` | Prometheus `/metrics` endpoint | Reused; may conflict with Docker gRPC examples if same host |
| Auth Service gRPC | `50051` local / `9090` container | Auth downstream client and health | Task 2 required |
| User Service gRPC | `50052` local / `9090` container | User downstream client and health | Task 2 required |
| Product Service gRPC | `50053` local / `9090` container | Product downstream client and health | Task 2 required |
| Cart Service gRPC | `50054` local / `9090` container | Cart downstream client and health | Task 2 required |
| Wishlist Service gRPC | `50055` local / `9090` container | Wishlist downstream client and health | Task 2 required |
| Order Service gRPC | `50056` local / `9090` container | Order downstream client and health | Task 2 required |
| Payment Service gRPC | `50057` local / `9090` container | Payment downstream client and health | Task 2 required |
| Search Service gRPC | `50058` local / `9090` container | Search downstream client and health | Task 2 required |
| CMS Service gRPC | `50059` local / `9090` container | CMS downstream client and health | Task 2 required |
| Session Service gRPC | `50060` local / `9090` container | Session downstream client and health | Task 2 required |
| Notification Service gRPC | `50061` local / `9090` container | Notification downstream client and health | Task 2 required |
| Superadmin Service gRPC | `50062` local / `9090` container | Superadmin downstream client and health | Task 2 required |

Networking rule:

- Local host run: use unique host ports like `localhost:50051`, `localhost:50052`.
- Docker Compose run: each service can listen on `:9090` inside its own container, and Gateway uses DNS names like `auth-service:9090`.
- Kubernetes run: use service DNS like `auth-service.core.svc.cluster.local:9090`.

### Redis, JWT/JWKS, Queues

Redis, JWT/JWKS, Kafka/RabbitMQ, Typesense, MySQL, and MongoDB are already explained in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Sections:

- `5. Database Setup`
- `6. Redis / Queue / External Services`
- `7. Environment Variables`

Task 2 does not add a new Redis or queue requirement. Redis is still used by Gateway rate limiting, not by the gRPC client registry itself.

### Optional Observability Services

Current code initializes tracing and gRPC client interceptors. If tracing is enabled:

```env
TRACE_ENABLED=true
TRACE_EXPORTER_OTLP_ENDPOINT=otel-collector:4317
TRACE_EXPORTER_OTLP_INSECURE=true
```

For beginner local development, if no OpenTelemetry Collector is running, set:

```env
TRACE_ENABLED=false
```

Metrics endpoint is controlled by:

```env
METRICS_ENABLED=true
METRICS_PORT=9090
```

These observability variables are also included in the broader Task 1 dependency guide.

---

## 7. Environment Variables

Full `.env` example is already documented in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`7. Environment Variables`

This section only lists Task 2 gRPC-client-specific variables.

### Task 2 `.env` Additions

Create local env file:

```text
backend/services/api-gateway/.env
```

Current code uses `os.Getenv`; it does not auto-load `.env`. Export before running:

```bash
cd backend/services/api-gateway
set -a
source .env
set +a
```

Task 2 gRPC variables:

```env
# gRPC transport config
GRPC_TLS_ENABLED=false
GRPC_DIAL_TIMEOUT=3s

# Downstream gRPC targets
AUTH_GRPC_ADDR=localhost:50051
USER_GRPC_ADDR=localhost:50052
PRODUCT_GRPC_ADDR=localhost:50053
CART_GRPC_ADDR=localhost:50054
WISHLIST_GRPC_ADDR=localhost:50055
ORDER_GRPC_ADDR=localhost:50056
PAYMENT_GRPC_ADDR=localhost:50057
SEARCH_GRPC_ADDR=localhost:50058
CMS_GRPC_ADDR=localhost:50059
SESSION_GRPC_ADDR=localhost:50060
NOTIFICATION_GRPC_ADDR=localhost:50061
SUPERADMIN_GRPC_ADDR=localhost:50062
```

### Variable Explanation

| Variable | Required? | Default | Purpose | Notes |
|---|---:|---|---|---|
| `GRPC_TLS_ENABLED` | Optional | `false` | Downstream gRPC TLS enable/disable | Local dev usually `false`; prod should use TLS/mTLS |
| `GRPC_DIAL_TIMEOUT` | Optional | `3s` | Startup dial max wait per service | Must be positive duration |
| `AUTH_GRPC_ADDR` | Yes | none | Auth Service gRPC target | No whitespace |
| `USER_GRPC_ADDR` | Yes | none | User Service gRPC target | No whitespace |
| `PRODUCT_GRPC_ADDR` | Yes | none | Product Service gRPC target | No whitespace |
| `CART_GRPC_ADDR` | Yes | none | Cart Service gRPC target | No whitespace |
| `WISHLIST_GRPC_ADDR` | Yes | none | Wishlist Service gRPC target | No whitespace |
| `ORDER_GRPC_ADDR` | Yes | none | Order Service gRPC target | No whitespace |
| `PAYMENT_GRPC_ADDR` | Yes | none | Payment Service gRPC target | No whitespace |
| `SEARCH_GRPC_ADDR` | Yes | none | Search Service gRPC target | No whitespace |
| `CMS_GRPC_ADDR` | Yes | none | CMS Service gRPC target | No whitespace |
| `SESSION_GRPC_ADDR` | Yes | none | Session Service gRPC target | No whitespace |
| `NOTIFICATION_GRPC_ADDR` | Yes | none | Notification Service gRPC target | No whitespace |
| `SUPERADMIN_GRPC_ADDR` | Yes | none | Superadmin Service gRPC target | No whitespace |

### Credentials Placement

Task 2 gRPC address variables do not include passwords.

Where to place config:

| Config | File |
|---|---|
| API Gateway gRPC targets | `backend/services/api-gateway/.env` |
| Downstream DB credentials | Downstream service `.env`, not Gateway `.env` |
| Redis password | Gateway `.env`, already covered by Task 1 guide |
| TLS certs/mTLS secrets | Secret manager, Kubernetes Secret, or mounted secret files in future infra |

Warning: Do not commit `.env` files.

---

## 8. Docker Setup

Docker basics and dependency containers are already documented in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`8. Docker Setup`

### Current Repo Status for Task 2

No committed API Gateway Dockerfile or docker-compose file was found for this service.

Task 2 does not introduce a new Docker container by itself. It introduces a networking requirement:

```text
API Gateway container must be able to resolve/connect to all downstream gRPC service containers.
```

### Docker Compose Address Rule

If API Gateway and downstream services run inside the same Compose network, use service DNS names:

```env
AUTH_GRPC_ADDR=auth-service:9090
USER_GRPC_ADDR=user-service:9090
PRODUCT_GRPC_ADDR=product-service:9090
CART_GRPC_ADDR=cart-service:9090
WISHLIST_GRPC_ADDR=wishlist-service:9090
ORDER_GRPC_ADDR=order-service:9090
PAYMENT_GRPC_ADDR=payment-service:9090
SEARCH_GRPC_ADDR=search-service:9090
CMS_GRPC_ADDR=cms-service:9090
SESSION_GRPC_ADDR=session-service:9090
NOTIFICATION_GRPC_ADDR=notification-service:9090
SUPERADMIN_GRPC_ADDR=superadmin-service:9090
```

Beginner rule:

- Gateway on host machine -> use `localhost:500xx` when ports are published.
- Gateway inside Docker Compose -> use `service-name:9090`.
- Gateway inside Kubernetes -> use service DNS, for example `product-service.core.svc.cluster.local:9090`.

### Docker Healthcheck Requirement for Downstream Services

Each downstream container should expose a gRPC health endpoint. Docker healthcheck can call an internal health command or use `grpc-health-probe` if included in the image.

Example concept:

```yaml
healthcheck:
  test: ["CMD", "grpc-health-probe", "-addr=:9090"]
  interval: 10s
  timeout: 3s
  retries: 10
```

This compose snippet is guidance only. Actual compose files are not present in current repo.

---

## 9. Local Development Setup

### Reused Base Onboarding

Clone repo, install Go, run `go mod download`, start Redis, configure full `.env`, and run the Gateway are already explained in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Sections:

- `9. Local Development Setup`
- `10. Running the Project`

### Task 2 Incremental Setup

1. Move to Gateway service:

```bash
cd backend/services/api-gateway
```

2. Download Go dependencies:

```bash
go mod download
```

3. Run gRPC client/config tests:

```bash
go test ./internal/clients ./internal/config
```

4. Add/export Task 2 gRPC env vars:

```bash
set -a
source .env
set +a
```

5. Start all downstream gRPC services.

Required services:

```text
auth-service
user-service
product-service
cart-service
wishlist-service
order-service
payment-service
search-service
cms-service
session-service
notification-service
superadmin-service
```

6. Verify at least one downstream endpoint manually:

```bash
grpcurl -plaintext \
  -d '{"service":"ecommerce.product.v1.ProductService"}' \
  localhost:50053 \
  grpc.health.v1.Health/Check
```

7. Run Gateway:

```bash
go run ./cmd/server
```

8. Check readiness:

```bash
curl -i http://localhost:8080/health/ready
```

If any downstream service is not `SERVING`, readiness will return unavailable.

### Test-Only Mode

If downstream services are not implemented/running yet, use tests only:

```bash
cd backend/services/api-gateway
go test ./...
```

This is enough to validate code dependencies. Full runtime requires real or mock downstream gRPC services.

---

## 10. Running the Project

### Minimal Task 2 Verification

Use this when you only want to verify gRPC client layer compiles and unit tests pass:

```bash
cd backend/services/api-gateway
go mod download
go test ./internal/clients ./internal/config
```

### Full Runtime Verification

Use this when all dependencies are available:

```bash
cd backend/services/api-gateway

set -a
source .env
set +a

go run ./cmd/server
```

Expected useful startup logs:

```text
grpc_client_connected service=auth target=localhost:50051
grpc_client_connected service=product target=localhost:50053
api_gateway_starting addr=:8080
```

In another terminal:

```bash
curl -i http://localhost:8080/health/live
curl -i http://localhost:8080/health/ready
```

Expected behavior:

- `/health/live` checks process is alive.
- `/health/ready` checks route catalog plus all downstream gRPC health statuses.

### Migrations

No API Gateway Task 2 migration exists.

Database migration setup for full platform is already covered in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Sections:

- `5. Database Setup`
- `10. Running the Project`

---

## 11. Common Errors & Fixes

Only Task 2 specific troubleshooting is listed here. General Go/Redis/Docker/JWT errors are already covered in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`11. Common Errors & Fixes`

| Error | Likely Cause | Fix |
|---|---|---|
| `AUTH_GRPC_ADDR is required` | Env var missing | Add all 12 `*_GRPC_ADDR` variables to Gateway `.env` and export them |
| `PRODUCT_GRPC_ADDR must not contain whitespace` | Value has spaces/newlines | Use `PRODUCT_GRPC_ADDR=localhost:50053`, no quotes/spaces |
| `initialize grpc clients: dial product service: context deadline exceeded` | Product Service not running, wrong port, firewall, or Docker DNS issue | Start service, correct `PRODUCT_GRPC_ADDR`, verify with `grpcurl` |
| `/health/ready` returns `DOWNSTREAM_SERVICES_UNAVAILABLE` | One or more health checks not `SERVING` | Check service logs and gRPC health registration |
| gRPC call works on host but fails in Docker | Using `localhost` inside container | Use Compose service name like `product-service:9090` |
| `tls: first record does not look like a TLS handshake` | `GRPC_TLS_ENABLED=true` but downstream is plaintext | Set `GRPC_TLS_ENABLED=false` locally or enable TLS on service |
| `transport: authentication handshake failed` | TLS cert/server name mismatch | Configure proper certs/mTLS in future infra; use plaintext only for local dev |
| `grpc: the client connection is closing` | Gateway shutdown or failed partial initialization | Check earlier dial error in logs |
| `unknown service grpc.health.v1.Health` | Downstream did not register gRPC health service | Add standard gRPC health server to downstream service |
| `unknown service ecommerce.product.v1.ProductService` in health response | Health registered wrong service name | Register exact health service name from section 6 |
| `cannot find package .../backend/shared/gen/go/...` | Generated proto clients missing | Add proto files/Buf config and run `buf generate` |
| Gateway starts slowly | Dial timeout/retry waiting for unreachable services | Lower `GRPC_DIAL_TIMEOUT` for dev or start dependencies first |
| Tracing startup/export issue | OTLP endpoint unavailable/misconfigured | Set `TRACE_ENABLED=false` for local or start an OTEL Collector |
| Metrics port conflict on `:9090` | Another service uses port 9090 | Set `METRICS_PORT=19090` or `METRICS_ENABLED=false` |

Debug commands:

```bash
grpcurl -plaintext localhost:50051 list
grpcurl -plaintext \
  -d '{"service":"ecommerce.auth.v1.AuthService"}' \
  localhost:50051 \
  grpc.health.v1.Health/Check
lsof -i :50051
lsof -i :8080
lsof -i :9090
```

Windows port debug:

```powershell
netstat -ano | findstr :50051
netstat -ano | findstr :8080
```

---

## 12. Security & Best Practices

### Task 2 Security Notes

- Use `GRPC_TLS_ENABLED=false` only for local development.
- Production should use TLS/mTLS or service mesh encryption.
- Do not log JWT tokens, Authorization headers, Redis passwords, or DB DSNs.
- Current metadata propagation allowlists safe keys. `Authorization` is intentionally not forwarded by `WithMetadata`.
- Forward stable request identifiers like `x-request-id` and tracing headers for debugging.
- Keep gRPC deadlines on every downstream call. No unbounded RPC calls.
- Do not dial gRPC per request. Reuse startup-created connections from registry.
- Do not put downstream DB credentials in API Gateway env. Gateway should talk to services, not databases.
- Keep health check names stable and documented.
- Use `GRPC_DIAL_TIMEOUT` small enough for fast feedback, but not so small that normal local startup fails.

### Beginner Best Practices

| Practice | Why |
|---|---|
| Start downstream services before Gateway | Gateway dials all services during startup |
| Keep a service address table in `.env.example` | New developers do not guess ports |
| Use Docker service names inside Compose | `localhost` inside container points to same container, not another service |
| Check `/health/live` first, `/health/ready` second | Live means process up; ready means dependencies healthy |
| Use `grpcurl` before debugging Gateway code | Direct connectivity issue quickly clear hota hai |
| Keep proto contract and generated code in sync | Generated client mismatch compile/runtime issues create karta hai |
| Disable tracing locally if collector absent | Local startup/debug simpler hota hai |

---

## 13. Missing or Misconfigured Things

Professional setup audit for Task 2:

| Finding | Impact | Suggested Fix |
|---|---|---|
| No `proto/` folder or Buf config found | Generated service clients cannot be produced yet | Add Platform Foundation proto structure, `buf.yaml`, `buf.gen.yaml`, and service proto files |
| No generated Go proto clients found | Current client wrappers do not expose actual RPC methods like `GetProduct` | Generate Go clients and wire registry fields to generated clients |
| Current service client files are typed wrappers around connection/descriptor | Useful for health/readiness, but not enough for real route bridging | Replace or extend wrappers with generated `AuthServiceClient`, `ProductServiceClient`, etc. |
| No API Gateway `.env.example` found | Beginners may miss all 12 `*_GRPC_ADDR` variables | Add safe `.env.example` with local and Docker examples |
| No committed API Gateway Dockerfile found | Containerized run/deploy not ready | Add service Dockerfile once runtime packaging is required |
| No committed local compose file found | New devs must manually start dependencies/services | Add `infra/compose/docker-compose.local.yml` |
| Gateway requires all 12 downstream services at startup | Partial local development is hard | Add explicit mock/stub mode or optional dependency mode for route-contract-only dev |
| `GRPC_TLS_ENABLED=true` uses default system CA and no custom server name/mTLS config | Production TLS may be incomplete | Add cert path, server name, and mTLS env/config when infra is ready |
| Health checks require exact service names | Wrong downstream health registration makes readiness fail | Share the table from section 6 with every service team |
| Tracing defaults to `otel-collector:4317` | Local run can be confusing if collector is absent | Document `TRACE_ENABLED=false` for simple local runs |
| No per-service gRPC port standard file found | Port/address guessing risk | Add a central local port map or compose service DNS standard |

No hardcoded database credentials were introduced by Task 2.

---

## 14. References to Previous Dependency Files

To avoid duplicate documentation, use these existing sections:

| Topic | Refer |
|---|---|
| Git/Go/Docker installation | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `3. Required Software` |
| Go modules, `go.mod`, `go.sum`, `go mod download`, `go test` basics | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `4. Dependency Management` |
| MySQL setup | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `5. Database Setup` -> `MySQL` |
| MongoDB setup | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `5. Database Setup` -> `MongoDB` |
| Redis setup | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `5. Database Setup` -> `Redis` |
| Typesense setup | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `6. Redis / Queue / External Services` -> `Typesense` |
| Kafka/RabbitMQ setup | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `6. Redis / Queue / External Services` -> `Kafka or RabbitMQ` |
| Full Gateway `.env` example | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `7. Environment Variables` |
| Docker dependency compose examples | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `8. Docker Setup` |
| Full local run flow | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `9. Local Development Setup` and `10. Running the Project` |
| General troubleshooting | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `11. Common Errors & Fixes` |
| Security baseline | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `12. Security & Best Practices` |
| Proto strategy and Buf workflow | `TaskImplementation/Platform Foundation/task2.md` -> `Step 9` and `Step 10` |

---

## 15. Final Checklist

Task 2 specific checklist:

```text
[ ] Read task1_Dependency.md for base Go/Docker/Redis setup
[ ] Go version is compatible with backend/go.work and go.mod
[ ] Ran: cd backend/services/api-gateway
[ ] Ran: go mod download
[ ] Ran: go test ./internal/clients ./internal/config
[ ] Created local backend/services/api-gateway/.env
[ ] Exported .env into shell before go run
[ ] GRPC_TLS_ENABLED set correctly for local/prod
[ ] GRPC_DIAL_TIMEOUT is positive
[ ] AUTH_GRPC_ADDR configured
[ ] USER_GRPC_ADDR configured
[ ] PRODUCT_GRPC_ADDR configured
[ ] CART_GRPC_ADDR configured
[ ] WISHLIST_GRPC_ADDR configured
[ ] ORDER_GRPC_ADDR configured
[ ] PAYMENT_GRPC_ADDR configured
[ ] SEARCH_GRPC_ADDR configured
[ ] CMS_GRPC_ADDR configured
[ ] SESSION_GRPC_ADDR configured
[ ] NOTIFICATION_GRPC_ADDR configured
[ ] SUPERADMIN_GRPC_ADDR configured
[ ] Downstream services are running or test-only mode is being used
[ ] Downstream gRPC ports match the env file
[ ] Downstream services register standard gRPC health service
[ ] Health service names match section 6 exactly
[ ] grpcurl can reach at least one downstream service directly
[ ] Gateway starts without grpc dial timeout errors
[ ] /health/live returns ok
[ ] /health/ready returns ready when all downstream services are serving
[ ] TRACE_ENABLED=false locally if no OTEL Collector is running
[ ] METRICS_PORT does not conflict with other local services
[ ] .env and secrets are not staged in git
[ ] Proto/generation gap is understood before implementing real REST-to-gRPC route bridge
```

Final beginner note: Task 2 ka sabse important point ye hai ki Gateway DB se direct baat nahi karta. Gateway ko correct `*_GRPC_ADDR` values, reachable downstream services, and healthy gRPC connections chahiye. Agar aap sirf code check kar rahe ho to `go test ./...` enough hai; agar actual Gateway run karna hai to downstream services first, Gateway second.
