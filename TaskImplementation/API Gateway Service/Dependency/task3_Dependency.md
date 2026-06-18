# Project Dependency & Setup Guide

Input analyzed: `TaskImplementation/API Gateway Service/task3.md`

Previous dependency guides reviewed first:

- `TaskImplementation/API Gateway Service/task1_Dependency.md`
- `TaskImplementation/API Gateway Service/task2_Dependency.md`

Related implementation inspected:

- `backend/services/api-gateway/go.mod`
- `backend/services/api-gateway/internal/config/config.go`
- `backend/services/api-gateway/internal/auth/*`
- `backend/services/api-gateway/internal/transport/http/auth_middleware.go`
- `backend/services/api-gateway/internal/transport/http/routes.go`
- `backend/services/api-gateway/internal/clients/metadata.go`
- `api/master-api.json`

Simple goal: Ye guide beginner developer ko batata hai ki API Gateway Task 3 auth middleware run/test karne ke liye code ke alawa kya chahiye: JWT settings, Auth Service JWKS endpoint, RBAC roles, webhook signature header, and safe downstream metadata.

---

## 1. Project Overview

Task 3 ka focus API Gateway ke authentication and authorization layer par hai.

Flow:

```text
Client request
  -> API Gateway route match
  -> Bearer token extraction
  -> JWT signature and claims verification using Auth Service JWKS
  -> RBAC role check
  -> safe user/session metadata sent to downstream gRPC service
```

Current implementation reality:

- API Gateway Go service hai.
- HTTP routing Go standard `net/http` / `http.ServeMux` se ho raha hai.
- Protected routes ke liye JWT required hai.
- JWT verification `github.com/golang-jwt/jwt/v5` se hoti hai.
- Signing public keys Auth Service ke JWKS endpoint se fetch/cache hote hain.
- Route auth level `api/master-api.json` se aata hai: `public`, `buyer`, `seller`, `admin`, `superadmin`, `webhook`.
- Webhook routes user JWT use nahi karte. Gateway currently configured signature header ki presence check karta hai.
- Gateway direct Auth database ya User database query nahi karta.
- Full runtime still needs previous dependencies: Redis, downstream gRPC services, route contract, and env export.

What is new in Task 3:

| Area | New Task 3 Requirement |
|---|---|
| JWT verification | `JWT_ISSUER`, `JWT_AUDIENCE`, `JWT_ALLOWED_ALGS`, `JWT_JWKS_URL`, JWKS cache/fetch settings |
| Auth Service HTTP dependency | JWKS endpoint must return public RSA signing keys |
| RBAC | Token roles must match route auth level |
| Webhook route guard | `WEBHOOK_SIGNATURE_HEADER` must be configured/sent for webhook routes |
| Downstream metadata | Gateway forwards safe claims like `x-user-id`, not raw JWT |
| Security config | Raw tokens/private keys must stay out of Gateway logs/env |

Already documented elsewhere, so not repeated here:

- Base Go install and Go module basics
- Redis local setup
- MySQL/MongoDB/Typesense/Kafka/RabbitMQ setup
- Full Docker dependency compose
- All downstream gRPC service addresses

Refer:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Sections:

- `3. Required Software`
- `4. Dependency Management`
- `5. Database Setup`
- `6. Redis / Queue / External Services`
- `8. Docker Setup`

Refer:

`TaskImplementation/API Gateway Service/task2_Dependency.md`

Sections:

- `6. Redis / Queue / External Services`
- `7. Environment Variables`
- `8. Docker Setup`

---

## 2. Tech Stack

### Task 3 Specific Technologies

| Technology | What it is | Why this project uses it | Required? | Beginner Hinglish explanation |
|---|---|---|---:|---|
| Go | Backend programming language | API Gateway service Go me implemented hai | Yes | Go ek fast compiled backend language hai. Gateway isi me run/test hota hai. |
| Go Modules | Go dependency system | `go.mod` and `go.sum` dependency versions lock karte hain | Yes | Go modules npm/pip jaisa dependency manager hai, bas Go ke liye. |
| `net/http` | Go standard HTTP server/router | Gateway HTTP middleware chain and routes ke liye | Yes | External web framework nahi hai; Go ka built-in HTTP package use ho raha hai. |
| `github.com/golang-jwt/jwt/v5` | JWT parsing and verification library | Access token signature, issuer, audience, expiry, and algorithm verify karne ke liye | Yes | JWT library token ko decode karke verify karti hai ki token trusted hai ya fake. |
| JWT | Signed token format | User identity and roles Gateway tak lane ke liye | Yes for protected routes | Login ke baad client access token bhejta hai: `Authorization: Bearer <token>`. |
| JWKS | JSON Web Key Set endpoint | Auth Service public keys expose karta hai so Gateway JWT signature verify kar sake | Yes for protected routes | JWKS ek public key list hai. Private key Auth Service ke paas rehti hai. |
| RSA algorithms | Asymmetric signing algorithms | Code only supports `RS256`, `RS384`, `RS512` | Yes | Auth Service private key se sign karta hai, Gateway public key se verify karta hai. |
| RBAC | Role-based access control | `buyer`, `seller`, `admin`, `superadmin` route access enforce karne ke liye | Yes | RBAC ka matlab role ke basis par allow/deny. Buyer seller-only route nahi call kar sakta. |
| gRPC metadata | Key/value metadata for gRPC calls | Verified user/session context downstream services ko safely pass karne ke liye | Yes | Ye HTTP headers jaisa hota hai, but internal gRPC call ke saath. |
| Auth Service JWKS endpoint | External HTTP endpoint | JWT public keys fetch karne ke liye | Yes for protected routes | Gateway token issue nahi karta. Auth Service key deta hai, Gateway verify karta hai. |
| Webhook signature header | Provider trust signal | Payment webhook route JWT se alag secure karne ke liye | Required for webhook routes | Webhook browser user nahi hota, isliye Bearer token nahi; provider signature chahiye. |

### Already Explained in Previous Dependency Guides

Do not duplicate these setup sections. Use:

| Topic | Refer |
|---|---|
| General Gateway tech stack | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `2. Tech Stack` |
| gRPC client stack | `TaskImplementation/API Gateway Service/task2_Dependency.md` -> `2. Tech Stack` |
| Redis, DB, queue, Docker stack | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `5. Database Setup`, `6. Redis / Queue / External Services`, `8. Docker Setup` |

---

## 3. Required Software

### Required for Task 3 Code/Test Work

| Software | Required? | Why |
|---|---:|---|
| Git | Yes | Repository clone/update ke liye |
| Go `1.26.3` compatible | Yes | `backend/services/api-gateway/go.mod` me `go 1.26.3` configured hai |
| curl | Yes | JWKS endpoint and Gateway routes verify karne ke liye |
| Docker | Recommended | Auth Service/JWKS, Redis, and other dependencies container me run karne ke liye |
| grpcurl | Recommended for full runtime | Downstream gRPC health debug karne ke liye |
| jq | Optional | JWKS JSON pretty-print/inspect karne ke liye |

Installation process already exists in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`3. Required Software`

Task 3 does not require a new language runtime beyond Go.

### Minimum Test-Only Setup

Agar aap sirf Task 3 auth code tests run karna chahte ho:

```bash
cd backend/services/api-gateway
go test ./internal/auth ./internal/transport/http ./internal/clients ./internal/config
```

Is mode me real Redis, Auth Service, JWKS endpoint, ya downstream gRPC services ki zarurat nahi hoti because tests stubs/httptest use karte hain.

### Full Runtime Setup

Full `go run ./cmd/server` ke liye ye dependencies bhi chahiye:

| Dependency | Why Needed | Already Documented |
|---|---|---|
| Redis | Rate limiter enabled by default | `task1_Dependency.md` -> `5. Database Setup` -> `Redis` |
| Downstream gRPC services | Gateway startup clients and readiness | `task2_Dependency.md` -> `6. Redis / Queue / External Services` |
| Auth Service JWKS HTTP endpoint | JWT verification for protected routes | This file expands Task 3 details; base mention in `task1_Dependency.md` |
| API route contract | Auth levels loaded from route catalog | `task1_Dependency.md` -> `1. Project Overview` |

---

## 4. Dependency Management

This is a Go project. General Go modules explanation is already covered in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`4. Dependency Management`

### Task 3 Specific Go Dependencies

Current `backend/services/api-gateway/go.mod` includes:

| Dependency | Version | Task 3 Use |
|---|---:|---|
| `github.com/golang-jwt/jwt/v5` | `v5.3.1` | JWT claims model, parser, issuer/audience validation, allowed signing algorithms |
| `google.golang.org/grpc` | `v1.81.1` | Safe auth metadata propagation to downstream gRPC calls |
| `github.com/prometheus/client_golang` | `v1.23.2` | Auth failure metrics through Gateway metrics layer |
| `go.opentelemetry.io/otel` | `v1.43.0` | Request tracing context around authenticated requests |

Standard library packages used heavily by Task 3:

| Package | Why |
|---|---|
| `crypto/rsa` | JWKS RSA public key parsing |
| `encoding/base64` / `encoding/json` | JWKS key decode |
| `net/http` | JWKS fetch, middleware, webhook route guard |
| `context` | Store verified auth claims per request |
| `log/slog` | Auth failure logging without raw token leakage |

### Commands

Run from API Gateway service directory:

```bash
cd backend/services/api-gateway
```

Download dependencies:

```bash
go mod download
```

Run Task 3 focused tests:

```bash
go test ./internal/auth ./internal/transport/http ./internal/clients ./internal/config
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

### Common Go/JWT Dependency Issues

General Go module issues are already covered in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`4. Dependency Management` -> `Dependency Problems and Fixes`

Task 3 specific issues:

| Error | Likely Cause | Fix |
|---|---|---|
| `JWT_ALLOWED_ALGS contains unsupported algorithm "HS256"` | Code supports only RSA algorithms | Use `JWT_ALLOWED_ALGS=RS256` locally unless Auth Service uses `RS384` or `RS512` |
| `configure jwt verifier: at least one jwt signing algorithm is required` | Empty `JWT_ALLOWED_ALGS` after parsing | Set `JWT_ALLOWED_ALGS=RS256` |
| `configure jwks key provider: jwks url is required` | Protected routes exist but JWKS URL missing | Set `JWT_JWKS_URL=http://localhost:8081/.well-known/jwks.json` |
| JWT tests fail after dependency edit | `jwt/v5` API changed or version mismatch | Keep `github.com/golang-jwt/jwt/v5 v5.3.1`, then run `go mod tidy` |

---

## 5. Database Setup

### API Gateway Task 3 Direct Database Requirement

Task 3 does not add any direct database dependency.

Important beginner rule:

```text
API Gateway JWT verify karta hai, lekin Auth DB directly query nahi karta.
```

Gateway ko sirf Auth Service ka JWKS public endpoint chahiye. User records, sessions, password hashes, refresh tokens, roles storage, and private signing keys Auth Service ke ownership me hain.

### Detected Platform Databases

These are platform-level dependencies, not new Task 3 Gateway dependencies:

| Database / Store | Directly used by Task 3 Gateway? | Owner |
|---|---:|---|
| MySQL | No | Auth/User/Order/Payment/CMS/Superadmin services |
| MongoDB | No | Product/Cart/Wishlist/Session/Notification services |
| Redis | Not by Task 3 auth itself; Gateway uses it for rate limiting | API Gateway rate limit layer |
| Typesense | No | Search Service |
| Kafka/RabbitMQ | No | Async/event services |

Do not repeat setup here. Full database installation, Docker, ports, volumes, connection strings, and credential placement are already explained in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Sections:

- `5. Database Setup`
- `6. Redis / Queue / External Services`

### Migrations

Task 3 has no API Gateway migration.

| Migration Type | Required for Task 3 Gateway? | Notes |
|---|---:|---|
| API Gateway DB migration | No | Gateway has no direct DB |
| Auth Service user/session/token migrations | Not in Gateway | Follow Auth Service docs when available |
| Redis migration | No | Redis is runtime cache/counter store |

### Credentials Placement

Do not put downstream database credentials in API Gateway `.env`.

| Credential | Correct Place |
|---|---|
| MySQL username/password | Owning service `.env`, not Gateway |
| MongoDB URI | Owning service `.env`, not Gateway |
| JWT private signing key | Auth Service secret manager / `.env` / mounted secret, never Gateway |
| JWT public JWKS URL | API Gateway `.env` as `JWT_JWKS_URL` |
| Redis password | API Gateway `.env`, already documented in `task1_Dependency.md` |

---

## 6. Redis / Queue / External Services

### Reused External Services

These are already documented and not repeated:

| Service | Reuse Documentation |
|---|---|
| Redis | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `5. Database Setup` -> `Redis` |
| Downstream gRPC services | `TaskImplementation/API Gateway Service/task2_Dependency.md` -> `6. Redis / Queue / External Services` |
| Kafka/RabbitMQ | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `6. Redis / Queue / External Services` |
| Typesense | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `6. Redis / Queue / External Services` |

### New/Detailed for Task 3: Auth Service JWKS Endpoint

#### A. What It Is

JWKS means JSON Web Key Set. Ye ek HTTP endpoint hota hai jo public signing keys return karta hai.

Example URL:

```text
http://localhost:8081/.well-known/jwks.json
```

Simple explanation:

```text
Auth Service private key se JWT sign karta hai.
Gateway public key se JWT verify karta hai.
JWKS endpoint public keys provide karta hai.
```

#### B. Why This Project Uses It

Gateway ko har protected request par ye validate karna hota hai:

- Token real Auth Service ne sign kiya hai.
- Token expired nahi hai.
- Token correct issuer/audience ke liye hai.
- Token `access` type ka hai, refresh token nahi.
- Token me required claims present hain.

JWKS se key rotation possible hota hai. Agar Auth Service future me signing key rotate kare, Gateway `kid` header ke basis par correct public key fetch kar sakta hai.

#### C. Required or Optional

| Mode | JWKS Required? | Explanation |
|---|---:|---|
| Unit tests with injected verifier/stubs | No | Tests fake verifier or `httptest` use kar sakte hain |
| Public-route-only contract | No | Agar contract me protected routes nahi hon |
| Current real `api/master-api.json` runtime | Yes | Contract me protected routes hain |
| Production | Yes | Protected APIs ke liye mandatory |

Current router fail closed karta hai:

```text
JWT_JWKS_URL is required when protected routes are configured
```

#### D. Local Installation / Startup

Preferred local approach:

1. Auth Service run karo.
2. Ensure Auth Service HTTP endpoint JWKS expose karta hai.
3. API Gateway `.env` me JWKS URL set karo.

```env
JWT_JWKS_URL=http://localhost:8081/.well-known/jwks.json
```

Verify:

```bash
curl -i http://localhost:8081/.well-known/jwks.json
```

Expected response should be JSON with `keys`:

```json
{
  "keys": [
    {
      "kty": "RSA",
      "use": "sig",
      "kid": "local-key-1",
      "alg": "RS256",
      "n": "...",
      "e": "AQAB"
    }
  ]
}
```

Beginner note: `n` and `e` RSA public key ke encoded parts hain. Inhe manually guess nahi karna. Auth Service generate/expose karega.

#### E. Docker Setup

Base Docker dependency setup is already documented in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`8. Docker Setup`

Task 3 specific Docker rule:

| Where Gateway Runs | Correct JWKS URL |
|---|---|
| Gateway on host machine | `http://localhost:8081/.well-known/jwks.json` |
| Gateway inside Docker Compose | `http://auth-service:8081/.well-known/jwks.json` |
| Gateway inside Kubernetes | Auth Service DNS/Ingress URL, usually HTTPS |

Compose concept only, because current repo does not include a committed API Gateway Dockerfile/compose:

```yaml
services:
  auth-service:
    image: your-auth-service-image
    ports:
      - "8081:8081"
      - "50051:9090"
    environment:
      APP_ENV: local

  api-gateway:
    image: your-api-gateway-image
    environment:
      JWT_ISSUER: ecommerce-auth
      JWT_AUDIENCE: ecommerce-api
      JWT_ALLOWED_ALGS: RS256
      JWT_JWKS_URL: http://auth-service:8081/.well-known/jwks.json
      WEBHOOK_SIGNATURE_HEADER: X-Provider-Signature
```

Warning: Auth Service private signing key should be a secret, not committed in compose YAML.

#### F. Start Commands

Start Auth Service using that service's run guide. Gateway side verification:

```bash
curl -i "$JWT_JWKS_URL"
```

If env is not exported yet:

```bash
curl -i http://localhost:8081/.well-known/jwks.json
```

#### G. Verify Running

Good JWKS endpoint signs:

- HTTP status is `200`.
- `Content-Type` should be JSON.
- Body contains `"keys"`.
- At least one key has `kty=RSA`.
- Token header `kid` should match one JWKS key `kid`.

Optional with `jq`:

```bash
curl -s http://localhost:8081/.well-known/jwks.json | jq '.keys[].kid'
```

#### H. Default Port

| Service | Default Port | Purpose |
|---|---:|---|
| Auth Service HTTP/JWKS | `8081` | Public signing keys for Gateway JWT verification |
| Auth Service gRPC | `50051` local / `9090` container | Downstream Auth gRPC target, inherited from Task 2 |

#### I. Connection String / URL Format

```env
JWT_JWKS_URL=http://localhost:8081/.well-known/jwks.json
```

Production should use HTTPS:

```env
JWT_JWKS_URL=https://auth.example.com/.well-known/jwks.json
```

#### J. Where To Place Credentials

Important distinction:

| Value | Place |
|---|---|
| JWKS public URL | API Gateway `.env` as `JWT_JWKS_URL` |
| JWT issuer/audience | API Gateway `.env`, must match Auth Service |
| JWT private key | Auth Service secret only |
| JWT public key | Auth Service JWKS response |
| Access token | Client request `Authorization` header, never `.env` |

### Task 3 Ports & Networking

Full port map already exists in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`7. Environment Variables` -> `Ports & Networking`

Task 3 specific summary:

| Service | Port | Purpose | New/Reused |
|---|---:|---|---|
| API Gateway HTTP | `8080` | REST API, auth middleware entry point | Reused |
| Auth Service HTTP/JWKS | `8081` | Public JWT signing keys | Task 3 critical |
| Auth Service gRPC | `50051` local / `9090` Docker | Auth downstream service target | Reused from Task 2 |
| Redis | `6379` | Rate limiting, not JWT verification | Reused |
| API Gateway Metrics | `9090` by default | Metrics, including auth failure metrics | Reused |

Networking rule:

- Host to host: use `localhost`.
- Container to container: use service DNS like `auth-service`.
- Never use `localhost` from one container to reach another container.

---

## 7. Environment Variables

Full Gateway `.env` is already documented in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`7. Environment Variables`

This section lists only Task 3 auth-specific variables.

### Where To Create `.env`

Recommended local file:

```text
backend/services/api-gateway/.env
```

Current Go code uses `os.Getenv`. It does not auto-load `.env`. Export before running:

```bash
cd backend/services/api-gateway
set -a
source .env
set +a
```

### Task 3 `.env` Additions

```env
# JWT / Auth Service JWKS
JWT_ISSUER=ecommerce-auth
JWT_AUDIENCE=ecommerce-api
JWT_ALLOWED_ALGS=RS256
JWT_JWKS_URL=http://localhost:8081/.well-known/jwks.json
JWT_JWKS_CACHE_TTL=5m
JWT_JWKS_FETCH_TIMEOUT=3s
JWT_CLOCK_SKEW=30s

# Webhook guard
WEBHOOK_SIGNATURE_HEADER=X-Provider-Signature

# Optional but recommended for auth-related observability
OBSERVABILITY_HASH_SALT=replace-with-local-random-string
```

### Variable Explanation

| Variable | Required? | Default | Purpose | Beginner Notes |
|---|---:|---|---|---|
| `JWT_ISSUER` | Yes | `ecommerce-auth` | Expected `iss` claim | Must match Auth Service token issuer |
| `JWT_AUDIENCE` | Yes | `ecommerce-api` | Expected `aud` claim | Prevents using token meant for another API |
| `JWT_ALLOWED_ALGS` | Yes | `RS256` | Allowed JWT signing algorithms | Code supports only `RS256`, `RS384`, `RS512` |
| `JWT_JWKS_URL` | Yes for protected routes | none | Auth Service public keys URL | Blank value fails when protected routes exist |
| `JWT_JWKS_CACHE_TTL` | Optional | `5m` | How long Gateway caches JWKS keys | Must be positive |
| `JWT_JWKS_FETCH_TIMEOUT` | Optional | `3s` | Max time to fetch JWKS | Must be positive |
| `JWT_CLOCK_SKEW` | Optional | `30s` | Small server clock difference tolerance | Must not be negative or more than `5m` |
| `WEBHOOK_SIGNATURE_HEADER` | Optional | `X-Provider-Signature` | Header required on webhook routes | Header name must not contain whitespace |
| `OBSERVABILITY_HASH_SALT` | Recommended | blank | Hash salt for user id observability | Use a secret-ish random string in shared envs |

### Credential Placement

| Credential / Secret | Put in Gateway `.env`? | Correct Placement |
|---|---:|---|
| JWT private key | No | Auth Service secret manager or mounted secret |
| JWT public key | No manual placement | Auth Service exposes via JWKS |
| JWT access token | No | Client sends in `Authorization: Bearer <token>` |
| Webhook provider secret | Usually no | Payment Service/provider integration config |
| Webhook signature header name | Yes | `WEBHOOK_SIGNATURE_HEADER` |
| Redis password | Yes if Gateway rate limit uses Redis auth | Already documented in `task1_Dependency.md` |

### Required JWT Claims From Auth Service

Gateway expects access tokens to contain:

| Claim | Required? | Purpose |
|---|---:|---|
| `sub` | Yes | User id |
| `sid` | Yes | Session id |
| `roles` | Yes | RBAC role list |
| `seller_id` | Conditional | Seller context, if applicable |
| `token_type` | Yes | Must be `access` |
| `iss` | Yes | Must match `JWT_ISSUER` |
| `aud` | Yes | Must match `JWT_AUDIENCE` |
| `exp` | Yes | Expiry |
| `iat` | Yes | Issued-at time |
| JWT header `kid` | Yes for JWKS lookup | Selects public key from JWKS |

---

## 8. Docker Setup

Docker basics and dependency container setup are already documented in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`8. Docker Setup`

Task 3 does not add a new committed Dockerfile or compose file.

### Current Repo Status for Task 3

No committed API Gateway Dockerfile or docker-compose file was found for this service.

Task 3 adds this container networking requirement:

```text
API Gateway container must reach Auth Service JWKS HTTP endpoint.
```

### Docker Compose Address Rule

If Gateway runs on host:

```env
JWT_JWKS_URL=http://localhost:8081/.well-known/jwks.json
```

If Gateway runs inside Docker Compose:

```env
JWT_JWKS_URL=http://auth-service:8081/.well-known/jwks.json
```

If Auth Service exposes JWKS on a different internal port, update the URL accordingly.

### Docker Healthcheck Recommendation

Auth Service should have an HTTP health/JWKS readiness check. Example concept:

```yaml
healthcheck:
  test: ["CMD", "wget", "-qO-", "http://localhost:8081/.well-known/jwks.json"]
  interval: 10s
  timeout: 3s
  retries: 10
```

This is guidance only. Actual compose files are not present in current repo.

### Secret Handling in Docker/Kubernetes

| Secret | Docker/Kubernetes Best Practice |
|---|---|
| Auth Service private signing key | Docker secret, Kubernetes Secret, or external secret manager |
| Gateway `JWT_JWKS_URL` | Normal env var, not secret |
| `OBSERVABILITY_HASH_SALT` | Secret/env from secret manager in shared environments |
| Webhook provider signing secret | Payment Service secret, not API Gateway unless Gateway later verifies full signature |

---

## 9. Local Development Setup

### Reused Base Onboarding

Clone repo, install Go, run `go mod download`, start Redis, configure full `.env`, and run Gateway are already explained in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Sections:

- `9. Local Development Setup`
- `10. Running the Project`

Downstream gRPC service startup is already explained in:

`TaskImplementation/API Gateway Service/task2_Dependency.md`

Section:

`9. Local Development Setup`

### Task 3 Incremental Setup

1. Move to Gateway service:

```bash
cd backend/services/api-gateway
```

2. Download Go dependencies:

```bash
go mod download
```

3. Run auth-focused tests:

```bash
go test ./internal/auth ./internal/transport/http ./internal/clients ./internal/config
```

4. Start Auth Service JWKS endpoint.

Expected local URL:

```text
http://localhost:8081/.well-known/jwks.json
```

5. Verify JWKS endpoint:

```bash
curl -i http://localhost:8081/.well-known/jwks.json
```

6. Add Task 3 env vars to `backend/services/api-gateway/.env`.

```env
JWT_ISSUER=ecommerce-auth
JWT_AUDIENCE=ecommerce-api
JWT_ALLOWED_ALGS=RS256
JWT_JWKS_URL=http://localhost:8081/.well-known/jwks.json
JWT_JWKS_CACHE_TTL=5m
JWT_JWKS_FETCH_TIMEOUT=3s
JWT_CLOCK_SKEW=30s
WEBHOOK_SIGNATURE_HEADER=X-Provider-Signature
```

7. Export env:

```bash
set -a
source .env
set +a
```

8. Start other inherited runtime dependencies:

- Redis, from `task1_Dependency.md`
- Downstream gRPC services, from `task2_Dependency.md`

9. Run Gateway:

```bash
go run ./cmd/server
```

10. Verify unauthenticated protected route:

```bash
curl -i http://localhost:8080/api/v1/me
```

Expected:

```text
401 Unauthorized
```

11. Verify webhook route requires signature header:

```bash
curl -i -X POST http://localhost:8080/api/v1/webhooks/payments/testpay
```

Expected:

```text
401 Unauthorized
```

With configured header:

```bash
curl -i -X POST \
  -H "X-Provider-Signature: dev-signature" \
  http://localhost:8080/api/v1/webhooks/payments/testpay
```

Current expected after guard passes:

```text
501 ROUTE_BRIDGE_NOT_CONFIGURED
```

This is okay for current Gateway stage because route bridge business calls are later work.

12. Verify protected route with real access token from Auth Service:

```bash
TOKEN="paste-access-token-from-auth-service"

curl -i \
  -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/me
```

Possible expected results:

| Result | Meaning |
|---|---|
| `401` | Token missing/invalid/expired/wrong issuer/wrong audience/JWKS issue |
| `403` | Token valid but role not allowed for route |
| `501 ROUTE_BRIDGE_NOT_CONFIGURED` | Auth passed; downstream route bridge not implemented yet |

### Test-Only Mode

If Auth Service/JWKS is not running:

```bash
cd backend/services/api-gateway
go test ./internal/auth ./internal/transport/http
```

This validates Task 3 auth behavior without full runtime dependencies.

---

## 10. Running the Project

### Minimal Task 3 Verification

Use this when you only want to verify auth code compiles and tests pass:

```bash
cd backend/services/api-gateway
go mod download
go test ./internal/auth ./internal/transport/http ./internal/clients ./internal/config
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

In another terminal:

```bash
curl -i http://localhost:8080/health/live
curl -i http://localhost:8080/health/ready
curl -i http://localhost:8081/.well-known/jwks.json
curl -i http://localhost:8080/api/v1/me
```

Expected auth smoke behavior:

| Command | Expected |
|---|---|
| `GET /health/live` | `200 OK` if Gateway process is alive |
| `GET /health/ready` | `200 OK` only if downstream gRPC services are healthy |
| `GET JWKS URL` | `200 OK` with JSON keys |
| `GET /api/v1/me` without token | `401 Unauthorized` |
| Protected route with valid wrong role | `403 Forbidden` |
| Webhook without signature | `401 Unauthorized` |

### Migrations

No API Gateway migration exists for Task 3.

Database migration setup for full platform is already covered in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Sections:

- `5. Database Setup`
- `10. Running the Project`

---

## 11. Common Errors & Fixes

General Go/Redis/Docker/gRPC troubleshooting is already covered in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`11. Common Errors & Fixes`

Task 2 gRPC-specific troubleshooting is already covered in:

`TaskImplementation/API Gateway Service/task2_Dependency.md`

Section:

`11. Common Errors & Fixes`

Only Task 3 specific troubleshooting is listed below.

| Error / Symptom | Likely Cause | Fix |
|---|---|---|
| `JWT_JWKS_URL is required when protected routes are configured` | Protected route exists and JWKS URL blank | Add `JWT_JWKS_URL=http://localhost:8081/.well-known/jwks.json` |
| `JWT_JWKS_URL is invalid` | URL missing scheme or malformed | Use full URL with `http://` or `https://` |
| `JWT_ALLOWED_ALGS contains unsupported algorithm "HS256"` | Gateway accepts RSA algorithms only | Use `RS256`, `RS384`, or `RS512`; recommended local value is `RS256` |
| `JWT_CLOCK_SKEW must not exceed 5m` | Clock skew env too large | Use `JWT_CLOCK_SKEW=30s` |
| Protected route returns `401` even with token | Expired token, wrong issuer/audience, wrong token type, bad signature, missing `kid`, or JWKS fetch failed | Check token claims, Auth Service config, and `curl $JWT_JWKS_URL` |
| Protected route returns `403` | Token valid but roles do not match route auth level | Use user with correct role or update route auth/role policy intentionally |
| Log contains `jwks_refresh_failed` | Gateway could not fetch JWKS or JWKS JSON invalid | Check Auth Service is running, URL is reachable from Gateway, and JSON contains RSA keys |
| `jwt signing key not found: missing kid` | JWT header has no `kid` | Auth Service must include `kid` header when signing tokens |
| `jwt signing key not found: unknown kid` | Token signed with key not present in JWKS | Refresh/rotate keys correctly; ensure Auth Service exposes current public key |
| JWKS curl returns HTML or 404 | Wrong Auth Service route or port | Fix `JWT_JWKS_URL` to actual JWKS endpoint |
| Webhook route returns `401` even with Bearer token | Webhook auth does not use user JWT | Send configured signature header, for example `X-Provider-Signature` |
| `WEBHOOK_SIGNATURE_HEADER must not contain whitespace` | Header env has spaces | Use `WEBHOOK_SIGNATURE_HEADER=X-Provider-Signature` |
| Token works on host but fails inside Docker | Gateway container uses `localhost` for Auth Service | Use `JWT_JWKS_URL=http://auth-service:8081/.well-known/jwks.json` inside Compose |
| Auth passed but response is `501 ROUTE_BRIDGE_NOT_CONFIGURED` | Current route bridge implementation incomplete | This is expected at current Gateway stage after auth succeeds |

Debug commands:

```bash
curl -i "$JWT_JWKS_URL"
curl -i http://localhost:8080/api/v1/me
curl -i -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/me
go test ./internal/auth -run TestVerifier
go test ./internal/transport/http -run TestProtected
```

Port debug:

```bash
lsof -i :8080
lsof -i :8081
```

Windows:

```powershell
netstat -ano | findstr :8080
netstat -ano | findstr :8081
```

---

## 12. Security & Best Practices

### Task 3 Security Notes

- Never log raw JWT tokens or `Authorization` header values.
- Gateway should store only verified claims in request context, not raw token.
- Gateway should forward only safe metadata: `x-user-id`, `x-session-id`, `x-roles`, `x-seller-id`.
- Do not forward raw JWT to downstream services by default.
- Keep JWT private signing keys only in Auth Service.
- Use HTTPS for `JWT_JWKS_URL` outside local development.
- Keep `JWT_ALLOWED_ALGS` strict. Do not allow `none` or unexpected algorithms.
- Use `RS256` locally unless Auth Service intentionally signs with `RS384` or `RS512`.
- Keep `JWT_CLOCK_SKEW` small, such as `30s`.
- Ensure servers have correct time sync. Expiry/issued-at checks depend on clock.
- Fail closed when JWKS key is unavailable or token is invalid.
- A valid Gateway role check is not enough for business authorization. Downstream services must still enforce ownership rules.
- Webhook header presence is not full cryptographic verification. Payment Service/provider layer must verify actual signature using provider secret and raw payload.
- Use `OBSERVABILITY_HASH_SALT` in shared environments so user id hashes are not easily reversible.

### Beginner Best Practices

| Practice | Why |
|---|---|
| Always test missing-token route first | Confirms middleware is attached |
| Verify JWKS with curl before debugging Go code | Network/config issue clear ho jata hai |
| Keep Auth Service issuer/audience same as Gateway env | Mismatch se every token `401` hoga |
| Use role-specific test users | `401` vs `403` confusion kam hota hai |
| Keep `.env` out of git | Secrets and local config leak nahi hote |
| Use Docker service names inside Compose | Container networking reliable hota hai |
| Keep token TTL short in dev/prod | Stolen token ka blast radius kam hota hai |
| Rotate signing keys with overlapping JWKS window | Old valid tokens and new tokens dono verify ho sakte hain during rollout |

---

## 13. Missing or Misconfigured Things

Professional setup audit for Task 3:

| Finding | Impact | Suggested Fix |
|---|---|---|
| No API Gateway `.env.example` found | Beginners may miss JWT/JWKS variables | Add safe `backend/services/api-gateway/.env.example` with Task 3 variables |
| No committed API Gateway Dockerfile or compose file found | Containerized Gateway auth testing is manual | Add Dockerfile and local compose when runtime packaging is required |
| Full runtime requires Auth Service JWKS endpoint | Gateway cannot start protected routes without it | Document/run Auth Service before Gateway or provide a local JWKS mock for dev |
| Webhook guard currently checks header presence only | Spoofed webhook could pass Gateway guard if left as final security | Payment Service or Gateway future task must verify provider signature cryptographically |
| Auth Service private key management not described in Gateway folder | New developers may confuse public JWKS with private key | Keep private key setup in Auth Service docs and link it from Gateway docs |
| No local token generation helper found | Manual testing protected routes is harder | Add dev-only script/command in Auth Service to issue test access tokens |
| Gateway requires downstream gRPC services for full startup | Auth-only manual testing is difficult without full stack | Add documented mock/stub mode or compose profile later |
| `JWT_JWKS_URL` default is blank | Protected runtime fails unless env is set | Keep it explicit in `.env.example` |
| `OBSERVABILITY_HASH_SALT` default blank | User hash values may be less safe in shared envs | Set salt via secret manager for shared/staging/prod |
| No Gateway DB migration needed | No issue | Keep Gateway DB-free unless a future requirement changes architecture |

No hardcoded JWT private key, database password, or provider secret was found in the Task 3 Gateway implementation inspected here.

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
| Full ports and networking table | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `7. Environment Variables` -> `Ports & Networking` |
| Docker dependency compose examples | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `8. Docker Setup` |
| Full local run flow | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `9. Local Development Setup` and `10. Running the Project` |
| General troubleshooting | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `11. Common Errors & Fixes` |
| Security baseline | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `12. Security & Best Practices` |
| Downstream gRPC addresses and health checks | `TaskImplementation/API Gateway Service/task2_Dependency.md` -> `6. Redis / Queue / External Services` |
| gRPC Docker address rules | `TaskImplementation/API Gateway Service/task2_Dependency.md` -> `8. Docker Setup` |
| gRPC-specific troubleshooting | `TaskImplementation/API Gateway Service/task2_Dependency.md` -> `11. Common Errors & Fixes` |

---

## 15. Final Checklist

Use this before saying "auth middleware is not working":

```text
[ ] Read `task1_Dependency.md` base setup references
[ ] Read `task2_Dependency.md` gRPC dependency references
[ ] Go version compatible with `backend/services/api-gateway/go.mod`
[ ] `go mod download` completed
[ ] `go test ./internal/auth ./internal/transport/http ./internal/clients ./internal/config` passes
[ ] Redis running if full Gateway runtime is used
[ ] Downstream gRPC services running if full Gateway runtime is used
[ ] Auth Service JWKS endpoint running
[ ] `curl -i $JWT_JWKS_URL` returns JSON with `keys`
[ ] `JWT_ISSUER` matches Auth Service token issuer
[ ] `JWT_AUDIENCE` matches Auth Service token audience
[ ] `JWT_ALLOWED_ALGS=RS256` or another supported RSA alg
[ ] `JWT_JWKS_CACHE_TTL`, `JWT_JWKS_FETCH_TIMEOUT`, `JWT_CLOCK_SKEW` are valid durations
[ ] `WEBHOOK_SIGNATURE_HEADER` configured
[ ] `.env` exported before `go run`
[ ] Access token has `sub`, `sid`, `roles`, `token_type=access`, `iss`, `aud`, `exp`, `iat`
[ ] Access token header has `kid` matching JWKS
[ ] Protected route without token returns `401`
[ ] Protected route with wrong role returns `403`
[ ] Webhook route without signature returns `401`
[ ] No raw JWT/private key/Authorization header is logged
[ ] No downstream DB credentials added to Gateway `.env`
```

Final beginner note: Task 3 ka most common confusion ye hota hai ki "Auth Service gRPC address" and "Auth Service JWKS URL" alag cheezein hain. `AUTH_GRPC_ADDR` internal service call/readiness ke liye hai. `JWT_JWKS_URL` public key fetch karke JWT verify karne ke liye hai. Dono ka purpose different hai, and full runtime me dono correctly configured hone chahiye.
