# Project Dependency & Setup Guide

Input analyzed:

`TaskImplementation/API Gateway Service/task8.md`

Previous dependency documentation checked first:

- `TaskImplementation/API Gateway Service/task1_Dependency.md`
- `TaskImplementation/API Gateway Service/task2_Dependency.md`
- `TaskImplementation/API Gateway Service/task3_Dependency.md`
- `TaskImplementation/API Gateway Service/task4_Dependency.md`
- `TaskImplementation/API Gateway Service/task5_Dependency.md`
- `TaskImplementation/API Gateway Service/task6_Dependency.md`
- `TaskImplementation/API Gateway Service/task7_Dependency.md`

Related implementation and configuration inspected:

- `backend/services/api-gateway/go.mod`
- `backend/services/api-gateway/go.sum`
- `backend/services/api-gateway/.env.example`
- `backend/services/api-gateway/cmd/server/main.go`
- `backend/services/api-gateway/internal/config/config.go`
- `backend/services/api-gateway/internal/domain/grpcweb.go`
- `backend/services/api-gateway/internal/repository/json_grpcweb_policy_repository.go`
- `backend/services/api-gateway/internal/usecase/grpcweb_policy_catalog.go`
- `backend/services/api-gateway/internal/transport/grpcweb/*`
- `backend/services/api-gateway/internal/clients/*`
- `backend/services/api-gateway/internal/observability/*`
- `backend/services/api-gateway/config/grpcweb-policies.json`
- `infra/envoy/grpc-web/envoy.yaml`

> Simple goal: Ye guide Task 8 ke implemented gRPC-Web bridge ko run, configure, and debug karne ke liye required non-code setup explain karta hai. Browser gRPC-Web request pehle Envoy ko jayegi, Envoy usko normal gRPC me translate karke API Gateway facade ko bhejega, aur Gateway approved downstream service ko proxy karega.

Important reuse rule:

Base Go installation, Go module basics, databases, Redis, all downstream gRPC addresses, JWT/JWKS, request validation, error mapping, and observability setup previous dependency guides me already documented hain. Is file me sirf Task 8 ke new or changed setup ko detail me explain kiya gaya hai.

---

## 1. Project Overview

Task 8 adds a browser-facing gRPC-Web path without changing the existing REST path.

```text
Browser gRPC-Web client
  -> Envoy :8081
  -> CORS + grpc_web translation
  -> API Gateway gRPC facade :9090
  -> method/service allowlist + JWT/RBAC
  -> approved downstream gRPC service
```

Existing REST flow remains:

```text
Browser REST client
  -> API Gateway HTTP :8080
  -> existing HTTP middleware
  -> downstream gRPC service
```

### Current Implementation Reality

| Area | Current Status |
|---|---|
| Gateway gRPC facade | Implemented in `internal/transport/grpcweb/` |
| Browser protocol translation | Implemented through committed Envoy config |
| Method allowlist and RBAC | Implemented through `config/grpcweb-policies.json` |
| Service-level allowlist | Implemented through `GRPC_WEB_EXPOSED_SERVICES` |
| Transparent downstream proxy | Implemented; raw protobuf bytes are forwarded without generated Gateway handlers |
| Request ID, logs, metrics, traces | Implemented for the gRPC facade |
| JWT verification | Reuses Task 3 JWKS configuration |
| Downstream connections | Reuses Task 2 gRPC client registry |
| Redis rate limiting | Existing HTTP rate limiter is **not** applied to the gRPC-Web facade |
| Envoy Dockerfile / Compose | Not present |
| API Gateway Dockerfile | Not present |
| Kubernetes manifests | Not present |
| Proto source / generated frontend client | Not present |
| Direct Gateway database | None |
| Gateway migrations | None |

### What Is New in Task 8

| New Item | Purpose |
|---|---|
| Envoy Proxy | Browser gRPC-Web ko internal HTTP/2 gRPC me translate karta hai |
| Gateway gRPC facade on `:9090` | Envoy se translated calls receive karta hai |
| `GRPC_WEB_*` environment variables | Facade enablement, policy path, allowlist, and message-size limits configure karte hain |
| `grpcweb-policies.json` | Exact browser-exposed methods, downstream target, auth mode, roles, and timeout define karta hai |
| Envoy CORS policy | Browser origins and gRPC-Web headers control karti hai |
| Inbound gRPC metrics | Facade requests ka method/status/latency record karte hain |

---

## 2. Tech Stack

### Task 8 Specific Technologies

| Technology | What It Is | Why This Project Uses It | Required? | Beginner Hinglish Explanation |
|---|---|---|---:|---|
| gRPC-Web | Browser-compatible gRPC protocol | Browser ko typed protobuf APIs call karne deta hai | Yes for browser gRPC calls | Normal browser native gRPC directly use nahi karta. gRPC-Web browser-friendly format hai. |
| Envoy Proxy | Layer-7 proxy | gRPC-Web ko normal gRPC me translate, CORS enforce, and Gateway ko route karta hai | Yes for browser-to-Gateway flow | Envoy translator aur traffic gate dono ka kaam karta hai. |
| Envoy `grpc_web` filter | Protocol translation filter | Browser payload ko internal gRPC format me convert karta hai | Yes | Filter missing hua to browser request normal gRPC server samajh nahi payega. |
| Envoy CORS filter | Browser-origin policy | Approved frontend origins and headers allow karta hai | Yes for browser calls | CORS decide karta hai kaunsi website browser se API call kar sakti hai. |
| Go gRPC server | Gateway facade runtime | Envoy se normal gRPC request receive karta hai | Yes when bridge enabled | Gateway REST server ke saath ek separate gRPC listener start hota hai. |
| Transparent raw protobuf proxy | Request forwarding layer | Gateway ko every proto service handler generate/register kiye bina approved RPC forward karne deta hai | Yes in current implementation | Gateway message ka business meaning change nahi karta; approved raw protobuf bytes downstream bhejta hai. |
| JSON method policy catalog | Security/routing config | Exact method, downstream, auth mode, roles, and timeout define karta hai | Yes when bridge enabled | Ye browser-exposed RPC methods ki strict allowlist hai. |
| JWT/JWKS and RBAC | Authentication and authorization | Protected gRPC-Web methods secure karta hai | Required by current policies | Token verify hota hai, phir role check hota hai. |
| Prometheus/OpenTelemetry | Metrics and traces | gRPC-Web request visibility provide karte hain | Optional by config, recommended | Errors aur slow methods debug karne me help karte hain. |

### Important Difference From `task8.md`

`task8.md` originally describes a documentation/future implementation plan. Current repository me backend facade, Envoy config, policy config, tests, and `.env.example` ab actually present hain.

Also, `task8.md` mentions these possible items, but they are not implemented in the current repo:

| Mentioned Item | Current Repo Status | Action |
|---|---|---|
| `GRPC_WEB_ALLOWED_ORIGINS` env var | Not read by Gateway or Envoy | Origins directly `infra/envoy/grpc-web/envoy.yaml` me edit karo |
| `VITE_GRPC_WEB_BASE_URL` | No frontend proto-client package or env example found | Frontend implementation task me add karna hoga |
| Buf TypeScript generation | No `buf.yaml`, `buf.gen.yaml`, `.proto`, or generated client found | Platform/frontend proto setup required before browser end-to-end call |
| Compose/Kubernetes examples | No actual Compose/Kubernetes files found | DevOps task me create and validate karna hoga |

### Reused Stack

Do not repeat-install these technologies. Refer:

- General Gateway stack: `task1_Dependency.md` -> `2. Tech Stack`
- Downstream gRPC and metadata: `task2_Dependency.md` -> `2. Tech Stack`
- JWT/JWKS/RBAC: `task3_Dependency.md` -> `2. Tech Stack`
- Redis rate limiting: `task4_Dependency.md` -> `2. Tech Stack`
- Request validation: `task5_Dependency.md` -> `2. Tech Stack`
- Safe error behavior: `task6_Dependency.md` -> `2. Tech Stack`
- Logs, metrics, and traces: `task7_Dependency.md` -> `2. Tech Stack`

---

## 3. Required Software

### Setup Modes

| Mode | Required Software / Services | What You Can Verify |
|---|---|---|
| Task 8 unit tests | Go `1.26.3` compatible toolchain | Policy validation, auth/RBAC, proxying, metadata safety, config |
| Gateway facade runtime | Go, Redis if rate limit enabled, JWKS endpoint, all 12 downstream gRPC services | Gateway HTTP + gRPC listeners start |
| Browser bridge runtime | Everything above + Envoy | CORS, `/grpcweb/*` translation, Envoy-to-Gateway connectivity |
| Browser end-to-end RPC | Everything above + proto contracts/generated browser client + working downstream method | Real typed browser RPC |

### New Required Tool: Envoy

Envoy Task 8 ka only new external runtime service hai.

Recommended beginner approach: Envoy ko Docker container me run karo. Docker installation already documented in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`3. Required Software`

Verify Docker:

```bash
docker version
```

Verify the pinned Envoy image used by this guide:

```bash
docker image inspect envoyproxy/envoy:v1.38.0
```

> Important: `envoyproxy/envoy:latest` currently does not exist in Docker Hub. This guide uses the verified version tag `v1.38.0`. Production deployment me approved version ke saath image digest bhi pin karna better hai.

### Optional Debug Tools

| Tool | Why Useful | Required? |
|---|---|---:|
| `curl` | CORS preflight and HTTP health checks | Recommended |
| `docker logs` | Envoy access/startup errors inspect karne ke liye | Recommended when using Docker |
| `ss` or `lsof` | `8081`, `9090`, `9091` port conflicts check karne ke liye | Recommended |
| `grpcurl` | Direct gRPC debugging | Optional; current repo has no proto source and facade reflection is not enabled |
| Browser DevTools | CORS and gRPC-Web network errors inspect karne ke liye | Required for frontend debugging |

Base Git, Go, curl, Docker, and `grpcurl` installation instructions already exist in `task1_Dependency.md` and `task2_Dependency.md`.

---

## 4. Dependency Management

This is still a Go project. General `go.mod`, `go.sum`, Go modules, `go mod download`, `go mod tidy`, `go build`, and `go run` explanation is already covered in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`4. Dependency Management`

### Task 8 Relevant Go Dependencies

Task 8 did **not** require a new direct Go module. It reuses packages already present in `backend/services/api-gateway/go.mod`.

| Dependency | Current Version | Task 8 Use |
|---|---:|---|
| `google.golang.org/grpc` | `v1.81.1` | Inbound facade server, transparent stream proxy, metadata, status codes, downstream streams |
| `google.golang.org/protobuf` | `v1.36.11` | Protobuf runtime used by gRPC stack |
| `github.com/golang-jwt/jwt/v5` | `v5.3.1` | Protected gRPC-Web method token verification |
| `github.com/prometheus/client_golang` | `v1.23.2` | Inbound gRPC request metrics |
| `go.opentelemetry.io/otel` | `v1.43.0` | Inbound gRPC spans and trace propagation |

Envoy is an external binary/container, not a Go module. Do not run `go get` for Envoy.

### Task 8 Focused Commands

Run from the Gateway service directory:

```bash
cd backend/services/api-gateway
```

Download already-declared modules:

```bash
go mod download
```

Run Task 8 focused tests:

```bash
go test \
  ./internal/domain \
  ./internal/repository \
  ./internal/usecase \
  ./internal/transport/grpcweb \
  ./internal/clients \
  ./internal/config \
  ./internal/observability
```

Run all Gateway tests:

```bash
go test ./...
```

Static analysis:

```bash
go vet ./...
```

Build:

```bash
go build ./cmd/server
```

### Dependency Issues Specific to Task 8

| Error | Cause | Fix |
|---|---|---|
| `cannot find module` | Command wrong directory se run hua | `backend/services/api-gateway` me jaakar run karo |
| Envoy image pull fails | Docker/network/registry issue | Docker daemon and registry access verify karo; approved pinned image use karo |
| Frontend gRPC-Web package/import missing | Frontend proto client is not implemented in repo | Proto generation and frontend transport add karo before browser RPC test |
| `grpcurl` cannot describe facade | Server reflection is not enabled and proto files are missing | Generated client/proto descriptors use karo; reflection ko production me casually enable mat karo |

---

## 5. Database Setup

### Direct Database Requirement

Task 8 adds no database.

The gRPC-Web bridge:

- MySQL directly use nahi karta.
- MongoDB directly use nahi karta.
- Redis ko new purpose ke liye use nahi karta.
- New table, collection, index, or migration create nahi karta.
- Policy config local JSON file se load karta hai, database se nahi.

### Reused Database Documentation

Full platform database and Redis setup already explained in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Sections:

- `5. Database Setup`
- `6. Redis / Queue / External Services`

Redis rate limiting setup already explained in:

`TaskImplementation/API Gateway Service/task4_Dependency.md`

Sections:

- `5. Database Setup`
- `7. Environment Variables`

### Migrations

Task 8 ke liye no migration command required hai.

```text
Gateway gRPC-Web policy file change != database migration
```

Policy file edit ke baad Gateway restart required hai because policies startup par load hote hain.

---

## 6. Redis / Queue / External Services

### New External Service: Envoy Proxy

#### A. What It Is

Envoy ek high-performance proxy hai. Is task me Envoy browser se gRPC-Web request receive karta hai and usko API Gateway facade ke normal gRPC/HTTP2 request me translate karta hai.

#### B. Why This Project Uses It

Normal browser clients native gRPC transport directly use nahi karte. Envoy:

- `grpc_web` protocol translation karta hai.
- Browser CORS preflight handle karta hai.
- `/grpcweb/*` path ko Gateway `:9090` par route karta hai.
- Unsafe identity headers remove karta hai.
- JSON access logs stdout par write karta hai.

#### C. Required or Optional

| Scenario | Envoy Required? |
|---|---:|
| Unit tests | No |
| Gateway direct internal gRPC facade test | No |
| Browser gRPC-Web call | Yes |
| Existing REST API | No |

#### D. Installation / Startup

Recommended local setup is Docker. Exact run commands section 8 me diye gaye hain.

#### E. Current Envoy Configuration

Committed file:

`infra/envoy/grpc-web/envoy.yaml`

| Config | Current Value | Meaning |
|---|---|---|
| Browser listener | `0.0.0.0:8081` | Browser/local host receives gRPC-Web here |
| Admin listener | `127.0.0.1:9901` | Envoy admin endpoint, container/process loopback only |
| Public route | `/grpcweb/*` | Only this prefix proxies to Gateway |
| Upstream host | `api-gateway:9090` | Docker/Kubernetes-style DNS name expected |
| Upstream protocol | HTTP/2 plaintext gRPC | Envoy-to-Gateway TLS is not configured |
| Route timeout | `15s` | Envoy overall route timeout |
| CORS origins | `http://localhost:5173`, `http://localhost:4173` | Exact local frontend origins only |
| CORS methods | `POST,OPTIONS` | gRPC-Web call and preflight |
| Request header limit | `32 KB` | Envoy request header protection |
| Active downstream connection limit | `10000` | Envoy overload protection |

#### F. Verify Running

Check Envoy startup logs:

```bash
docker logs ecommerce-envoy-grpc-web
```

Check listener:

```bash
curl -i -X OPTIONS \
  http://localhost:8081/grpcweb/ecommerce.session.v1.SessionService/IngestEvent \
  -H "Origin: http://localhost:5173" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: authorization,content-type,x-grpc-web,x-request-id"
```

The pinned Envoy image does not include `curl`. Check admin readiness by running a temporary curl container in the Envoy container's network namespace:

```bash
docker run --rm \
  --network container:ecommerce-envoy-grpc-web \
  curlimages/curl:8.12.1 \
  -fsS http://127.0.0.1:9901/ready
```

Expected readiness response:

```text
LIVE
```

#### G. Default Ports

Envoy itself has no universal mandatory port. This repo config uses:

- `8081` for gRPC-Web browser traffic.
- `9901` for Envoy admin on process/container loopback.

#### H. Credentials Placement

Current Envoy config contains no passwords or API keys.

Do not add secrets directly to:

`infra/envoy/grpc-web/envoy.yaml`

Future TLS certificates, private keys, or external auth credentials should come from Docker/Kubernetes secrets or a secret manager.

#### I. Health Check

Current repo has no Docker Compose/Kubernetes healthcheck declaration. Envoy admin `/ready` can be used by future container orchestration health checks.

#### J. Common Envoy Issues

| Issue | Cause | Fix |
|---|---|---|
| Envoy `503` / upstream reset | `api-gateway:9090` DNS/port unreachable | Use correct Docker network/DNS or host mapping; ensure Gateway facade is listening |
| Browser CORS error | Origin/header not present in Envoy exact allowlist | Update and redeploy `envoy.yaml` |
| `404` from Envoy | Path does not start with `/grpcweb/` | Use the configured gRPC-Web base path |
| Envoy starts but RPC fails | Gateway/downstream/policy/auth problem | Check Envoy logs, Gateway logs, and policy file |

### Task 8 Policy Configuration

Committed policy file:

`backend/services/api-gateway/config/grpcweb-policies.json`

Policy fields:

| Field | Required? | Meaning |
|---|---:|---|
| `full_method` | Yes | Exact `/package.Service/Method` allowlisted RPC |
| `downstream` | Yes | Existing Gateway downstream client key |
| `auth_mode` | Yes | `public`, `optional`, or `required` |
| `roles` | Only for required auth | Any accepted role for the method |
| `timeout` | Yes | Positive Go duration such as `800ms`, `1500ms`, or `3s` |

Current allowed methods:

| Full Method | Downstream | Auth Mode | Roles | Timeout |
|---|---|---|---|---:|
| `/ecommerce.session.v1.SessionService/IngestEvent` | `session` | `optional` | None | `800ms` |
| `/ecommerce.session.v1.SessionService/GetLiveSessions` | `session` | `required` | `admin`, `superadmin` | `1500ms` |
| `/ecommerce.superadmin.v1.SuperadminService/ListUsers` | `superadmin` | `required` | `operations_admin`, `superadmin` | `1500ms` |
| `/ecommerce.superadmin.v1.SuperadminService/UpdatePlatformSetting` | `superadmin` | `required` | `superadmin` | `3s` |
| `/ecommerce.cms.v1.CMSService/GetSellerAnalytics` | `cms` | `required` | `seller`, `seller_manager`, `superadmin` | `1500ms` |

Startup validation rejects:

- Empty policy list.
- Unknown JSON fields.
- Duplicate full methods.
- Invalid full-method format.
- Service not included in `GRPC_WEB_EXPOSED_SERVICES`.
- Missing/unavailable downstream connection.
- Invalid auth mode.
- Roles on `public` or `optional` methods.
- Empty/non-positive timeout.

> Security rule: New method expose karne ke liye both service allowlist and JSON method policy review karo. Sirf env allowlist update karna enough nahi hai.

### Reused External Services

| Service | Task 8 Relationship | Refer |
|---|---|---|
| All 12 downstream gRPC services | Gateway startup still requires all; Task 8 policies currently route to Session, CMS, and Superadmin | `task2_Dependency.md` -> `6. Redis / Queue / External Services` |
| Auth Service JWKS endpoint | Required because current Task 8 policies include `optional` and `required` auth | `task3_Dependency.md` -> `6. Redis / Queue / External Services` |
| Redis | Existing HTTP Gateway startup dependency when rate limiting is enabled; not new gRPC-Web rate limiting | `task4_Dependency.md` -> `5. Database Setup` |
| Prometheus/OpenTelemetry stack | Facade emits new inbound gRPC metrics/traces through existing setup | `task7_Dependency.md` -> `6. Redis / Queue / External Services` |
| Kafka/RabbitMQ/Typesense | No new direct Task 8 dependency | `task1_Dependency.md` -> `6. Redis / Queue / External Services` |

### Ports & Networking

| Service / Listener | Port | Purpose | Status | How To Change |
|---|---:|---|---|---|
| API Gateway HTTP | `8080` | Existing REST API and health endpoints | Reused | `HTTP_ADDR` |
| Envoy gRPC-Web listener | `8081` | Browser-facing local gRPC-Web path | **New** | Edit `infra/envoy/grpc-web/envoy.yaml` and published Docker port |
| API Gateway gRPC facade | `9090` | Envoy translated gRPC upstream | **New** | `GRPC_ADDR` and matching Envoy cluster port |
| API Gateway metrics | `9091` in `.env.example` | Avoids collision with facade `9090` | **Changed for Task 8 example** | `METRICS_ADDR` |
| Envoy admin | `9901` | Envoy readiness/config/stats on loopback | **New** | Edit `envoy.yaml`; keep private |
| Auth JWKS HTTP | `8082` in current `.env.example` | Public keys for protected methods | Reused | `JWT_JWKS_URL` |
| Downstream local gRPC services | `50051`-`50062` in `.env.example` | Internal service calls | Reused | Corresponding `*_GRPC_ADDR` |

Critical port rule:

```text
GRPC_ADDR must not equal HTTP_ADDR.
GRPC_ADDR must not equal METRICS_ADDR when metrics are enabled.
```

Code defaults both `GRPC_ADDR` and metrics to `:9090`. Therefore, when enabling gRPC-Web, explicitly set:

```env
GRPC_ADDR=:9090
METRICS_ADDR=:9091
```

---

## 7. Environment Variables

### Where To Create `.env`

Use:

`backend/services/api-gateway/.env`

The root `.gitignore` ignores `.env` files and allows `.env.example`.

Current Go code uses `os.Getenv`; it does **not** automatically load `.env`. Base export instructions already exist in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`7. Environment Variables`

### Task 8 New / Changed `.env` Subset

Add these values on top of the previous Task 1-7 Gateway environment:

```env
# Enable browser gRPC-Web facade
GRPC_WEB_ENABLED=true

# Internal Gateway gRPC listener used by Envoy
GRPC_ADDR=:9090

# Exact method/RBAC policy file
GRPC_WEB_POLICY_PATH=config/grpcweb-policies.json

# Service-level allowlist; every policy service must be present
GRPC_WEB_EXPOSED_SERVICES=ecommerce.session.v1.SessionService,ecommerce.superadmin.v1.SuperadminService,ecommerce.cms.v1.CMSService

# Positive raw protobuf message limits
GRPC_WEB_MAX_RECEIVE_MESSAGE_BYTES=4194304
GRPC_WEB_MAX_SEND_MESSAGE_BYTES=4194304

# Changed from the default :9090 to avoid collision with GRPC_ADDR
METRICS_ADDR=:9091
```

Current safe full example is already committed at:

`backend/services/api-gateway/.env.example`

Create local env:

```bash
cd backend/services/api-gateway
cp .env.example .env
```

Export before running:

```bash
set -a
source .env
set +a
```

### Variable Explanation

| Variable | Required? | Default in Code | Current Example | Purpose | Security / Validation Notes |
|---|---:|---|---|---|---|
| `GRPC_WEB_ENABLED` | Optional | `false` | `true` | Starts/stops the Gateway facade | Keep disabled if bridge is not intentionally deployed |
| `GRPC_ADDR` | Required when enabled | `:9090` | `:9090` | Gateway gRPC facade bind address | Must not equal HTTP or metrics address; protect from direct public access |
| `GRPC_WEB_POLICY_PATH` | Required when enabled | Auto-discovery attempt | `config/grpcweb-policies.json` | Method/RBAC policy JSON path | File must exist and be readable; no hot reload |
| `GRPC_WEB_EXPOSED_SERVICES` | Required when enabled | Empty | Three current services | Service-level allowlist | Values cannot contain slash or whitespace |
| `GRPC_WEB_MAX_RECEIVE_MESSAGE_BYTES` | Required positive integer when enabled | `4194304` | `4194304` | Max inbound protobuf message bytes | Prevent oversized requests; tune carefully |
| `GRPC_WEB_MAX_SEND_MESSAGE_BYTES` | Required positive integer when enabled | `4194304` | `4194304` | Max outbound protobuf message bytes | Prevent oversized responses; tune carefully |
| `METRICS_ADDR` | Required to avoid default collision when metrics enabled | Defaults to `:9090` through metrics config | `:9091` | Existing metrics listener | Keep private; Task 7 documents metrics security |

### Inherited Variables That Task 8 Needs

Do not duplicate their full explanation here:

| Inherited Config | Why Task 8 Needs It | Refer |
|---|---|---|
| All `*_GRPC_ADDR` values | Gateway initializes all downstream connections; current policies route to session, CMS, superadmin | `task2_Dependency.md` -> `7. Environment Variables` |
| `GRPC_TLS_ENABLED`, `GRPC_DIAL_TIMEOUT` | Existing Gateway-to-downstream gRPC connections | `task2_Dependency.md` -> `7. Environment Variables` |
| `JWT_ISSUER`, `JWT_AUDIENCE`, `JWT_ALLOWED_ALGS`, `JWT_JWKS_URL`, JWKS timings | Current Task 8 policy set contains protected/optional-token methods | `task3_Dependency.md` -> `7. Environment Variables` |
| Redis/rate-limit variables | Existing Gateway startup behavior when HTTP rate limiting is enabled | `task4_Dependency.md` -> `7. Environment Variables` |
| Metrics/tracing variables | gRPC facade logs, metrics, and traces | `task7_Dependency.md` -> `7. Environment Variables` |

### Variables That Do Not Exist Yet

Do not add these expecting current code to read them:

```env
GRPC_WEB_ALLOWED_ORIGINS=...
VITE_GRPC_WEB_BASE_URL=...
```

- Allowed origins currently live directly in `infra/envoy/grpc-web/envoy.yaml`.
- No frontend proto client/env implementation was found.

### Common Task 8 `.env` Mistakes

| Mistake | Result | Fix |
|---|---|---|
| `.env` created but not exported | Gateway uses defaults/missing values | `set -a; source .env; set +a` |
| gRPC-Web enabled but metrics left on default `:9090` | Startup fails: `GRPC_ADDR must not equal METRICS_ADDR` | Set `METRICS_ADDR=:9091` |
| Wrong relative policy path | Startup fails: policy file not readable | Run from service directory or use an absolute path |
| Policy service missing from env allowlist | Startup fails with service not in allowlist | Keep JSON policies and `GRPC_WEB_EXPOSED_SERVICES` aligned |
| `JWT_JWKS_URL` blank | Current protected/optional policies cannot create token verifier | Configure Auth JWKS endpoint |
| New policy uses unavailable downstream | Facade initialization fails | Use a supported configured downstream and start it |
| Origins added to `.env` only | Browser still gets CORS error | Edit Envoy YAML because origins are not env-driven |

---

## 8. Docker Setup

### Current Repo Status

| Item | Present? | Notes |
|---|---:|---|
| Envoy config | Yes | `infra/envoy/grpc-web/envoy.yaml` |
| API Gateway `.env.example` | Yes | Includes Task 8 values |
| API Gateway Dockerfile | No | Gateway cannot currently be built as a repo-defined container |
| Envoy custom Dockerfile | No | Standard Envoy image expected |
| Local Docker Compose file | No | No one-command Gateway + Envoy stack |
| Kubernetes Envoy manifests | No | Future DevOps work |
| Docker healthcheck/restart policy | No | Must be added in future Compose/deployment |
| Persistent volume | Not needed | Envoy config is read-only; no Task 8 persistent data |

### Validate Envoy Configuration

Run from repository root:

```bash
docker run --rm \
  -v "$PWD/infra/envoy/grpc-web/envoy.yaml:/etc/envoy/envoy.yaml:ro" \
  envoyproxy/envoy:v1.38.0 \
  --mode validate \
  -c /etc/envoy/envoy.yaml
```

Expected result includes:

```text
configuration ... OK
```

### Run Envoy When Gateway Runs on the Host

Current Envoy upstream is `api-gateway:9090`. When Envoy is in Docker and Gateway runs directly on the host, provide a host mapping:

```bash
docker run -d \
  --name ecommerce-envoy-grpc-web \
  --add-host api-gateway:host-gateway \
  -p 8081:8081 \
  -v "$PWD/infra/envoy/grpc-web/envoy.yaml:/etc/envoy/envoy.yaml:ro" \
  envoyproxy/envoy:v1.38.0
```

Why `--add-host`?

```text
Container ke andar localhost container khud hota hai.
Envoy config `api-gateway:9090` expect karti hai.
Host mapping is name ko host machine tak resolve karwati hai.
```

Gateway `GRPC_ADDR=:9090` uses an all-interface bind, so the container can reach the host listener when local firewall permits it.

Stop the container. Because the run command uses `--rm`, Docker removes it automatically after stop:

```bash
docker stop ecommerce-envoy-grpc-web
```

### Container-to-Container Future Rule

If API Gateway later gets a Dockerfile and both services run in one Compose network:

```text
Envoy upstream `api-gateway:9090`
  -> `api-gateway` must be the Compose service/container DNS name
  -> both containers must share a Docker network
  -> Gateway must listen on `:9090`
```

Do not use `localhost:9090` from the Envoy container to reach another container.

### Docker Security Notes

- Publish `8081` only where browser traffic should reach it.
- Do not publish Gateway `9090` publicly in production.
- Keep Envoy admin `9901` private; current config binds it to loopback.
- Mount Envoy config read-only.
- Keep the Envoy version pinned and use an approved digest for production.
- Add a healthcheck, resource limits, and restart policy when Compose/Kubernetes deployment is created.

Base Docker installation and common commands are already documented in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`8. Docker Setup`

---

## 9. Local Development Setup

### Step 1: Read Previous Dependency Documentation

Follow these first:

| Need | Previous Guide |
|---|---|
| Clone, Go install, base Gateway setup | `task1_Dependency.md` |
| All downstream gRPC services/addresses | `task2_Dependency.md` |
| Auth Service JWKS and JWT/RBAC | `task3_Dependency.md` |
| Redis startup | `task4_Dependency.md` |
| Observability ports/config | `task7_Dependency.md` |

### Step 2: Go To API Gateway Directory

```bash
cd backend/services/api-gateway
```

### Step 3: Install Existing Go Dependencies

```bash
go mod download
```

No new Task 8 Go package needs to be added.

### Step 4: Run Task 8 Tests First

```bash
go test \
  ./internal/domain \
  ./internal/repository \
  ./internal/usecase \
  ./internal/transport/grpcweb \
  ./internal/clients \
  ./internal/config \
  ./internal/observability
```

For full confidence:

```bash
go test ./...
go vet ./...
```

### Step 5: Prepare Environment

```bash
cp .env.example .env
set -a
source .env
set +a
```

Confirm Task 8 values:

```bash
printf '%s\n' \
  "$GRPC_WEB_ENABLED" \
  "$GRPC_ADDR" \
  "$GRPC_WEB_POLICY_PATH" \
  "$GRPC_WEB_EXPOSED_SERVICES" \
  "$METRICS_ADDR"
```

Expected key values:

```text
true
:9090
config/grpcweb-policies.json
...session..., ...superadmin..., ...cms...
:9091
```

### Step 6: Review the Method Policy

Open:

`backend/services/api-gateway/config/grpcweb-policies.json`

Before adding/changing a method, verify:

- Exact full method name matches downstream proto service.
- Downstream key is supported by Gateway client registry.
- Service is present in `GRPC_WEB_EXPOSED_SERVICES`.
- Auth mode is intentional.
- Roles are least-privilege.
- Timeout is positive and lower than Envoy route timeout where appropriate.

### Step 7: Start Reused Runtime Dependencies

Start:

- Redis when `RATE_LIMIT_ENABLED=true`.
- Auth JWKS endpoint.
- All 12 downstream gRPC services because current Gateway startup dials all of them.
- Optional OTel Collector if tracing is enabled.

Use previous dependency guides for exact commands. Task 8 does not change those setup steps.

### Step 8: Start API Gateway

From `backend/services/api-gateway` with env exported:

```bash
go run ./cmd/server
```

Expected listeners with current `.env.example`:

```text
:8080  REST/health
:9090  Gateway gRPC facade
:9091  Prometheus metrics
```

Verify local listening ports:

```bash
ss -ltn | grep -E ':(8080|9090|9091)\b'
```

### Step 9: Validate and Start Envoy

From repository root:

```bash
cd ../../..
```

Validate config using the command from section 8, then start Envoy with the host mapping command.

Verify Envoy listener:

```bash
ss -ltn | grep -E ':8081\b'
```

### Step 10: Verify CORS and Routing

Run the preflight command from section 6.

Expected:

- Allowed origin receives CORS headers.
- `/grpcweb/*` path is accepted by route config.
- A non-`/grpcweb/*` path receives `404`.

Example:

```bash
curl -i http://localhost:8081/not-grpcweb
```

### Step 11: Verify Real Browser RPC

A real end-to-end method requires:

- Downstream method implementation.
- Matching protobuf contract.
- Generated browser client/transport.
- Correct browser base URL, typically `http://localhost:8081/grpcweb`.
- Valid JWT for protected methods.

Current repo has no proto source/generated frontend client, so unit tests plus Envoy CORS/upstream checks are the immediately reproducible Task 8 verification path.

---

## 10. Running the Project

### Minimal Task 8 Verification

No Redis, Envoy, JWKS, or real downstream service is required:

```bash
cd backend/services/api-gateway
go test ./internal/transport/grpcweb ./internal/repository ./internal/usecase ./internal/domain ./internal/config
```

These tests verify:

- Method/service allowlist.
- Policy JSON parsing and strict unknown-field rejection.
- Required/optional auth behavior.
- Role enforcement.
- Unsafe identity metadata stripping.
- Verified identity metadata propagation.
- Request ID propagation.
- Transparent protobuf stream forwarding.
- Unexposed method rejection.
- Missing downstream rejection.

### Full Gateway + Envoy Runtime

```text
1. Follow previous guides and start Redis/JWKS/all downstream services.
2. Export backend/services/api-gateway/.env.
3. Start Gateway with `go run ./cmd/server`.
4. Validate Envoy YAML.
5. Start Envoy.
6. Run CORS preflight.
7. Use generated frontend client when available.
8. Check Gateway logs, Envoy logs, and metrics.
```

### Task 8 Metrics Verification

With metrics enabled and `METRICS_ADDR=:9091`:

```bash
curl -fsS http://localhost:9091/metrics | grep -E 'grpc_server_(requests_total|duration_seconds)'
```

Metrics appear after facade requests have been handled.

### Task 8 Logs

Envoy:

```bash
docker logs -f ecommerce-envoy-grpc-web
```

Gateway logs should include events such as:

```text
grpc_web_facade_starting
grpc_web_request_completed
grpc_web_request_rejected
grpc_web_request_failed
```

### Migrations

No Task 8 migration is required.

---

## 11. Common Errors & Fixes

### Task 8 Specific Troubleshooting

| Error / Symptom | Likely Cause | Fix | Prevention |
|---|---|---|---|
| `GRPC_ADDR must not equal METRICS_ADDR` | Both default/use `:9090` | Set `METRICS_ADDR=:9091` | Keep explicit non-conflicting ports in env |
| `GRPC_ADDR must not equal HTTP_ADDR` | Facade and REST listener use same address | Use separate `:9090` and `:8080` | Maintain port table |
| `GRPC_WEB_POLICY_PATH is required...` | Enabled bridge cannot discover policy file | Set `GRPC_WEB_POLICY_PATH=config/grpcweb-policies.json` | Run from service dir or use absolute path |
| `GRPC_WEB_POLICY_PATH is not readable` | Wrong path/permissions | Fix path and read permission | Mount/copy policy file with deployment |
| `at least one grpc-web method policy is required` | Policy file has empty list | Add reviewed allowlisted method | Do not deploy empty enabled bridge |
| `unexpected trailing content` or decode error | Invalid JSON or extra content | Validate/fix JSON | Keep JSON lint/check in CI |
| `unknown field` decode error | Typo in policy field | Use exact supported schema | Review config changes |
| `policy service ... is not in the exposed service allowlist` | Env service list and JSON policy differ | Add exact service to `GRPC_WEB_EXPOSED_SERVICES` or remove policy | Update both in one change |
| `references unavailable downstream` | Invalid `downstream` value/connection | Use supported key and configure/start service | Review against client registry |
| `JWT_JWKS_URL is required when protected...` | Current policies need token verifier | Configure JWKS URL | Start Auth/JWKS before Gateway |
| Gateway startup hangs then fails on gRPC clients | One of all 12 downstream services is unavailable | Start/fix all configured services | Use shared local stack or future mock mode |
| Envoy `503` | Envoy cannot reach `api-gateway:9090` | Add host mapping or shared Docker network; start facade | Validate DNS and listener before browser test |
| Envoy config validation fails | YAML/API schema/version issue | Read validation output and fix YAML | Validate in CI before deploy |
| Browser CORS error | Origin/header not allowlisted | Edit exact Envoy CORS policy | Keep dev/staging/prod origin lists reviewed |
| Browser gets `404` | Request path lacks `/grpcweb/` prefix | Configure correct gRPC-Web base URL | Centralize frontend transport config |
| Browser gets `Unimplemented: method is not exposed` | Full method missing from policy | Add reviewed method policy | Treat policy as API exposure contract |
| Browser gets `Unauthenticated` | Missing/invalid/expired token | Send valid Bearer token; verify issuer/audience/JWKS | Test token lifecycle |
| Browser gets `PermissionDenied` | Token role not in method policy | Use correct role or fix reviewed policy | Add RBAC tests |
| Browser gets `DeadlineExceeded` | Policy timeout/downstream too slow | Inspect downstream latency; tune carefully | Monitor per-method latency |
| Browser gets `ResourceExhausted` | Message exceeds 4 MiB limit or downstream quota | Reduce payload or intentionally tune limit | Keep bounded message sizes |
| `grpc-status` unreadable in browser | Exposed CORS headers changed/missing | Restore gRPC response exposed headers | Test preflight and error response in CI |
| Metrics missing | No facade call yet, wrong metrics port, or metrics disabled | Call facade; check `:9091`; enable metrics | Keep explicit Task 8 metrics smoke test |
| `.env` values ignored | `.env` was not exported | Source/export before `go run` | Use a documented run script later |

### Reused Troubleshooting

Do not duplicate these:

- General Go/module/Docker errors: `task1_Dependency.md` -> `11. Common Errors & Fixes`
- Downstream gRPC connection/health errors: `task2_Dependency.md` -> `11. Common Errors & Fixes`
- JWT/JWKS/auth errors: `task3_Dependency.md` -> `11. Common Errors & Fixes`
- Redis errors: `task4_Dependency.md` -> `11. Common Errors & Fixes`
- Observability errors: `task7_Dependency.md` -> `11. Common Errors & Fixes`

---

## 12. Security & Best Practices

### Task 8 Security Rules

| Rule | Why It Matters |
|---|---|
| Keep method policy deny-by-default | Unknown methods return `Unimplemented` instead of being silently exposed |
| Require both service and method allowlists | A service env typo alone should not expose every method |
| Review `auth_mode` and roles for every policy change | Browser-exposed admin RPCs are high risk |
| Keep `GRPC_ADDR` private | Browser should reach Envoy, not bypass it and call Gateway facade directly |
| Keep Envoy admin `9901` private | Admin endpoint exposes operational/config data |
| Use exact CORS origins | Wildcards can expose browser APIs to untrusted sites |
| Do not log `Authorization` or raw tokens | Tokens are credentials |
| Strip browser-supplied identity headers | Current Envoy and Gateway prevent spoofed `x-user-id`, `x-session-id`, roles, seller id |
| Use TLS at the public edge | Current local Envoy listener is plaintext |
| Add Envoy-to-Gateway TLS/mTLS for production if required | Current upstream is plaintext HTTP/2 |
| Keep message limits positive and bounded | Protects memory and large-payload abuse |
| Add gRPC-Web-specific rate limiting | Existing Redis HTTP middleware does not protect facade requests |
| Pin Envoy image version/digest | Reproducible and auditable deployments |
| Treat policy JSON as security-sensitive config | It controls public method exposure and RBAC |
| Restart after policy changes | Current implementation loads policy only at startup |

### Request Metadata Behavior

Current implementation safely forwards only approved metadata. Important behavior:

- Envoy removes browser-supplied `x-user-id`, `x-session-id`, `x-roles`, and `x-seller-id`.
- Gateway only forwards `authorization` downstream after token verification.
- Gateway creates trusted identity metadata from verified JWT claims.
- Request/trace metadata is allowlisted.
- Internal/unknown downstream errors are sanitized.

### Task 8 Beginner Best Practices

- First run tests, then start full runtime.
- Keep one source of truth for exposed method policy.
- Change Envoy config, policy JSON, env allowlist, and client contract together when adding a new method.
- Test three auth cases for every protected method: no token, wrong role, correct role.
- Monitor `grpc_server_requests_total` by method and code.
- Keep Envoy route timeout above intended per-method timeout, but avoid very large values.
- Do not expose a whole internal service because one browser method is needed.
- Keep REST and gRPC-Web behavior/error handling intentionally separate.

---

## 13. Missing or Misconfigured Things

### Professional Setup Audit

| Finding | Impact | Suggested Fix |
|---|---|---|
| No API Gateway Dockerfile | Repo cannot build/run Gateway as a defined container | Add a multi-stage Gateway Dockerfile |
| No local Docker Compose file | Gateway + Envoy + reused dependencies cannot start with one command | Add `infra/compose/docker-compose.local.yml` with healthchecks/networks |
| No Kubernetes Envoy manifests | Production/staging bridge deployment is not reproducible | Add ConfigMap, Deployment, Service, Ingress, NetworkPolicy |
| Original `task8.md` Envoy examples use nonexistent `envoyproxy/envoy:latest` | Copying those commands fails before config validation | Use the verified pinned tag from this guide and an approved digest in production |
| Envoy CORS origins are hardcoded | Environment changes require YAML edits; no `GRPC_WEB_ALLOWED_ORIGINS` support | Template config per environment or add controlled config generation |
| `task8.md` mentions `GRPC_WEB_ALLOWED_ORIGINS`, but code does not read it | Developers may set an ineffective env var | Remove stale example or implement templated/env-driven origin config |
| No proto source, Buf config, or generated frontend client | Real browser typed RPC cannot be reproduced from repo | Add reviewed proto contracts and frontend generation pipeline |
| No frontend `VITE_GRPC_WEB_BASE_URL` config | Browser client base URL is not implemented | Add frontend transport and env example |
| Gateway facade binds `:9090` on all interfaces | Direct facade bypass is possible if network exposes the port | Restrict with firewall/container network/Kubernetes NetworkPolicy |
| Envoy-to-Gateway traffic is plaintext | Internal traffic is not encrypted/authenticated | Add TLS/mTLS for non-local environments if required |
| Existing Redis rate limiter only wraps HTTP routes | gRPC-Web calls can bypass current rate limits | Add gRPC stream/method rate-limit interceptor or Envoy rate-limit service |
| No facade/Envoy healthcheck wiring; standard Envoy image has no `curl` utility | Orchestrator cannot reliably gate startup/readiness | Add an external/sidecar Envoy `/ready` check and Gateway facade readiness strategy |
| No Docker restart/resource policy | Local/prod container behavior is undefined | Add restart policy, CPU/memory limits, and connection tuning |
| Policy JSON has no hot reload | Policy updates require process restart | Keep restart explicit or implement audited reload later |
| No automated Envoy config validation in repo CI | Invalid config can reach deployment | Add pinned-image `--mode validate` CI step |
| No browser gRPC-Web integration test | CORS/filter/path regressions may pass Go tests | Add Envoy-backed integration test with generated client |
| No gRPC reflection | Manual `grpcurl` discovery is harder | Keep disabled for security unless a controlled debug environment needs it |
| Envoy accepts `domains: ["*"]` | Host header is not restricted at Envoy virtual host | Use expected hosts in deployed environment |
| Envoy public listener has no TLS config | Direct local pattern is plaintext | Terminate TLS at ingress/load balancer or Envoy in production |
| Current `.env.example` enables gRPC-Web by default | Full local run now needs policy/JWKS/downstream readiness | Keep intentional; document a test-only mode with `GRPC_WEB_ENABLED=false` if needed |

### Hardcoded Credentials Check

No hardcoded password, private key, JWT, database credential, or provider secret was found in Task 8 implementation/config files.

Hardcoded non-secret operational config that needs environment-specific review:

| Item | Current Value | Recommendation |
|---|---|---|
| CORS origins | `http://localhost:5173`, `http://localhost:4173` | Replace with exact environment origins |
| Envoy upstream DNS | `api-gateway` | Ensure deployment DNS/service name matches |
| Envoy/Gateway ports | `8081`, `9090`, `9901` | Keep consistent across firewall/service manifests |
| Exposed method policies | Five committed methods | Security review every change |
| Envoy route timeout | `15s` | Monitor and tune per environment |

---

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/API Gateway Service/task1_Dependency.md` | `3. Required Software` | Git, Go, curl, and Docker installation already documented |
| `TaskImplementation/API Gateway Service/task1_Dependency.md` | `4. Dependency Management` | Go modules and common Go commands are unchanged |
| `TaskImplementation/API Gateway Service/task1_Dependency.md` | `5. Database Setup` | Task 8 adds no database or migration |
| `TaskImplementation/API Gateway Service/task1_Dependency.md` | `7. Environment Variables` | `.env` location/export behavior and baseline env already documented |
| `TaskImplementation/API Gateway Service/task1_Dependency.md` | `8. Docker Setup` | Base Docker commands/install are unchanged |
| `TaskImplementation/API Gateway Service/task1_Dependency.md` | `9. Local Development Setup`, `10. Running the Project` | Base clone/onboarding flow reused |
| `TaskImplementation/API Gateway Service/task2_Dependency.md` | `6. Redis / Queue / External Services` | Existing downstream gRPC services and health requirements reused |
| `TaskImplementation/API Gateway Service/task2_Dependency.md` | `7. Environment Variables` | All `*_GRPC_ADDR`, TLS, and dial timeout config reused |
| `TaskImplementation/API Gateway Service/task2_Dependency.md` | `8. Docker Setup` | Existing gRPC Docker DNS/network rules reused |
| `TaskImplementation/API Gateway Service/task3_Dependency.md` | `6. Redis / Queue / External Services` | Auth Service JWKS endpoint reused for protected gRPC-Web methods |
| `TaskImplementation/API Gateway Service/task3_Dependency.md` | `7. Environment Variables` | JWT issuer/audience/algorithm/JWKS config unchanged |
| `TaskImplementation/API Gateway Service/task4_Dependency.md` | `5. Database Setup`, `7. Environment Variables` | Existing Redis startup/config reused |
| `TaskImplementation/API Gateway Service/task5_Dependency.md` | `6. Redis / Queue / External Services` | API contract baseline remains part of Gateway startup |
| `TaskImplementation/API Gateway Service/task6_Dependency.md` | `12. Security & Best Practices` | Existing safe error principles reused |
| `TaskImplementation/API Gateway Service/task7_Dependency.md` | `6. Redis / Queue / External Services` | Prometheus/OpenTelemetry stack reused |
| `TaskImplementation/API Gateway Service/task7_Dependency.md` | `7. Environment Variables` | Existing metrics/tracing configuration reused; only metrics port changes for Task 8 example |

---

## 15. Final Checklist

```text
[ ] Previous Task 1-7 dependency documentation checked
[ ] Go version is compatible with go 1.26.3
[ ] Ran `go mod download`
[ ] Ran Task 8 focused tests
[ ] Ran `go test ./...`
[ ] Ran `go vet ./...`
[ ] Created `backend/services/api-gateway/.env` from `.env.example`
[ ] Exported `.env` before running Gateway
[ ] `GRPC_WEB_ENABLED=true` intentionally selected
[ ] `GRPC_ADDR=:9090`
[ ] `METRICS_ADDR=:9091` so it does not conflict with the facade
[ ] `GRPC_WEB_POLICY_PATH` points to a readable policy file
[ ] `GRPC_WEB_EXPOSED_SERVICES` matches every service used by policy JSON
[ ] Policy full methods, downstreams, auth modes, roles, and timeouts reviewed
[ ] JWT/JWKS config available for current optional/required auth policies
[ ] Redis running if existing HTTP rate limiting is enabled
[ ] All 12 downstream gRPC services available for full Gateway startup
[ ] Gateway starts REST `:8080`, facade `:9090`, and metrics `:9091`
[ ] Envoy config validates successfully
[ ] Envoy can resolve and reach `api-gateway:9090`
[ ] Envoy listener `:8081` is reachable
[ ] CORS preflight succeeds from an allowed origin
[ ] Non-`/grpcweb/*` Envoy path returns `404`
[ ] Unexposed RPC method is rejected
[ ] Protected method rejects missing/invalid token
[ ] Protected method rejects wrong role
[ ] Envoy and Gateway logs checked
[ ] gRPC server metrics checked
[ ] No Task 8 database migration attempted
[ ] No duplicate previous setup documentation added
```
