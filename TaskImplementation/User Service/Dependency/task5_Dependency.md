# Project Dependency & Setup Guide

## 1. Project Overview

| Item | Value |
|---|---|
| `SERVICE_NAME` | `{SERVICE_NAME}` |
| `TASK_FILE_NAME` | `{TASK_FILE_NAME}` |
| Input file | `TaskImplementation/{SERVICE_NAME}/{TASK_FILE_NAME}` |
| Output file | `TaskImplementation/{SERVICE_NAME}/{OUTPUT_FILE_NAME}` |
| Task goal | Public REST profile APIs expose karna through API Gateway |
| Main runtime flow | Browser REST -> API Gateway HTTP -> JWT/RBAC -> {SERVICE_NAME} gRPC -> MySQL |

Current task ka setup mostly previous setup ko reuse karta hai. New part API Gateway runtime hai: ek Go HTTP server jo `net/http` routes expose karta hai, JWT token verify karta hai, aur existing {SERVICE_NAME} gRPC methods ko call karta hai.

Important boundary:

- API Gateway public HTTP entrypoint hai.
- API Gateway direct MySQL connect nahi karta.
- {SERVICE_NAME} gRPC process and MySQL setup previous tasks se reused hai.
- Current task koi new database table, migration, Redis, Kafka, RabbitMQ, or Docker container introduce nahi karta.

## 2. Tech Stack

| Technology | Required? | Status | Beginner-friendly explanation |
|---|---:|---|---|
| Go `1.24` | Yes | Reused | Go backend language hai. Gateway aur service dono Go modules hain. Full Go setup previous docs me hai. |
| Go `net/http` | Yes | New for current task | Standard library HTTP server/router hai. Is task me REST APIs ke liye external router framework add nahi kiya gaya. |
| API Gateway module | Yes | New | Browser-facing process hai jo `/api/v1/me` type REST endpoints serve karta hai. |
| gRPC client | Yes | Reused + newly consumed by Gateway | Gateway internal typed calls ke liye {SERVICE_NAME} gRPC client use karta hai. gRPC explanation previous docs me available hai. |
| Protobuf generated Go code | Yes | Reused | `backend/shared/gen/go/ecommerce/user/v1` se typed request/response structs milte hain. |
| Shared validation module | Yes | New for current task | Common phone, GSTIN, email, URL, postal code validation helpers ko Gateway handlers reuse karte hain. |
| `github.com/nyaruka/phonenumbers` | Indirect | New via shared validation | Phone number normalize/validate karne ke liye use hota hai. Direct Gateway code me import nahi hai. |
| JWT HS256 verifier | Yes | New Gateway config | Gateway bearer token verify karta hai using shared secret, issuer, audience, expiry, token type, and roles. |
| `log/slog` | Yes | Reused | Structured JSON logs ke liye Go standard logging package. |
| MySQL | Yes for {SERVICE_NAME} runtime | Reused | Gateway direct DB use nahi karta, but upstream {SERVICE_NAME} ko real runtime me MySQL chahiye. |
| Docker | Optional | Reused | Local MySQL container ke liye useful. Current task me app Dockerfile/compose add nahi hua. |

Already explained setup:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`1. Project Tech Stack Analysis`
`2. Language-Specific Dependency System: Go`
`5. External Services Analysis -> gRPC / Protobuf and Buf`
```

## 3. Required Software

Use the previous dependency files first:

| Software | Needed for current task? | Setup status |
|---|---:|---|
| Go `1.24` | Yes | Reuse `task1_Dependency.md` |
| MySQL 8.x | Yes when running real {SERVICE_NAME} | Reuse `task1_Dependency.md` and `task2_Dependency.md` |
| MySQL CLI | Recommended | Reuse `task2_Dependency.md` |
| Docker | Optional | Reuse MySQL Docker setup from `task1_Dependency.md` |
| Buf CLI | Only if proto changed | Reuse `task1_Dependency.md` |
| grpcurl | Optional {SERVICE_NAME} debug tool | Reuse `task1_Dependency.md` and `task4_Dependency.md` |
| curl/Postman | Yes for REST verification | New for current task |

New beginner note: `curl` simple HTTP testing tool hai. Isse browser ke bina terminal se Gateway REST endpoints test kar sakte ho.

## 4. Dependency Management

Go modules, `go.mod`, `go.sum`, `go.work`, `go mod download`, `go mod tidy`, proxy/cache issues, and version mismatch troubleshooting already explained hain:

```md
This setup is already explained in:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`2. Language-Specific Dependency System: Go`
```

Current task specific Go module changes:

| File | Current task relevance |
|---|---|
| `backend/services/api-gateway/go.mod` | New Gateway module dependencies define karta hai. |
| `backend/go.work` | Includes `./services/api-gateway`, `./services/user-service`, `./shared/gen/go`, and `./shared/validation`. |
| `backend/shared/validation/go.mod` | Phone validation dependency `github.com/nyaruka/phonenumbers` rakhta hai. |

Current task Gateway dependencies:

| Dependency | Why needed |
|---|---|
| `github.com/parag/ecommerce/backend/shared/gen/go` | Generated {SERVICE_NAME} gRPC client and proto messages. |
| `github.com/parag/ecommerce/backend/shared/validation` | REST request normalization and validation helpers. |
| `google.golang.org/grpc` | Gateway -> {SERVICE_NAME} internal gRPC call. |
| `google.golang.org/protobuf` | Field masks and generated protobuf runtime. |

Fresh clone command for Gateway module:

```bash
cd backend/services/api-gateway
go mod download
```

Run Gateway tests:

```bash
cd backend/services/api-gateway
go test ./...
```

Only run `go mod tidy` if dependencies are missing or `go test` reports module errors:

```bash
cd backend/services/api-gateway
go mod tidy
```

## 5. Database Setup

No new database setup is introduced by current task.

API Gateway does not open a MySQL connection. Gateway receives REST request, validates auth, then calls {SERVICE_NAME} through gRPC. MySQL is still mandatory only because {SERVICE_NAME} stores profile, address, and seller profile data there.

Reuse previous database documentation:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`3. Database Analysis`
`4. Environment Variables`
`7. Docker and DevOps Setup`
```

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task2_Dependency.md`

Sections:
`5. Database Setup`
`Apply Migration`
`Schema Verification Queries`
```

Current task database impact:

| Database item | Status |
|---|---|
| New database | No |
| New table | No |
| New migration | No |
| New index | No |
| Gateway DB credentials | Not needed |
| {SERVICE_NAME} MySQL setup | Reused |

Beginner note: Agar `/api/v1/me` Gateway se call kar rahe ho and response me upstream error aa raha hai, issue Gateway DB ka nahi hoga. Usually {SERVICE_NAME} gRPC process down hota hai, MySQL down hota hai, migration missing hoti hai, ya JWT/gRPC metadata issue hota hai.

## 6. Redis / Queue / External Services

Current task API Gateway does not require Redis, Kafka, RabbitMQ, NATS, MinIO, Elasticsearch, SMTP, Stripe, Firebase, or OAuth provider.

| External service | Required for current task? | Notes |
|---|---:|---|
| {SERVICE_NAME} gRPC | Yes | Gateway must dial {SERVICE_NAME} at `localhost:50052` by default. |
| Auth/JWT issuer | Yes for real tokens | Gateway validates HS256 JWT. Token secret/issuer/audience must match Auth Service. |
| MySQL | Indirect | Required by {SERVICE_NAME}, not by Gateway. |
| Redis | No | No Gateway Redis client/env found. |
| Kafka | No | Not needed for current task REST profile APIs. |
| RabbitMQ | No | Keep outbox worker disabled unless working on event/outbox tasks. |
| Docker | Optional | Only for reused local MySQL setup. |

{SERVICE_NAME} gRPC dependency:

- Default Gateway upstream address: `localhost:50052`
- Default {SERVICE_NAME} listen address: `:50052`
- Health/debug command is already documented with `grpcurl` in `task4_Dependency.md`

JWT dependency:

- Gateway expects `Authorization: Bearer <access_token>`.
- Token must be HS256 signed with the same secret configured in `API_GATEWAY_JWT_HS256_SECRET`.
- Token must contain `token_type=access`.
- Buyer/self APIs need roles like `buyer`, `seller`, `admin`, or `superadmin`.
- Seller APIs need roles like `seller`, `seller_manager`, or `superadmin`, plus `seller_id`.

## 7. Environment Variables

Previous {SERVICE_NAME} `.env` is already documented:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`4. Environment Variables`
```

Current task introduces API Gateway environment variables. Create this file locally:

```text
backend/services/api-gateway/.env
```

Example:

```bash
# HTTP server config.
API_GATEWAY_HTTP_ADDRESS=:8080
API_GATEWAY_SHUTDOWN_TIMEOUT=10s

# Gateway -> {SERVICE_NAME} gRPC config.
# Use localhost when both processes run directly on the host.
USER_SERVICE_GRPC_ADDRESS=localhost:50052
API_GATEWAY_USER_SERVICE_TIMEOUT=2s

# JWT verification config.
# Must match Auth Service token signing config.
API_GATEWAY_JWT_HS256_SECRET=change-me-local-only
API_GATEWAY_JWT_ISSUER=ecommerce-auth
API_GATEWAY_JWT_AUDIENCE=ecommerce-api
API_GATEWAY_JWT_CLOCK_SKEW=30s

# Validation and logs.
API_GATEWAY_VALIDATION_PHONE_REGION=IN
API_GATEWAY_LOG_LEVEL=info
```

Variable details:

| Variable | Required? | Example | Purpose | Security note |
|---|---:|---|---|---|
| `API_GATEWAY_HTTP_ADDRESS` | No | `:8080` | Gateway HTTP listen address | Public in local dev, restrict properly in prod. |
| `API_GATEWAY_SHUTDOWN_TIMEOUT` | No | `10s` | Graceful shutdown wait time | Must be positive duration. |
| `USER_SERVICE_GRPC_ADDRESS` | No | `localhost:50052` | Gateway upstream target for {SERVICE_NAME} | Do not expose gRPC publicly in prod. |
| `API_GATEWAY_USER_SERVICE_TIMEOUT` | No | `2s` | Per-request gRPC timeout | Too low causes false timeout. Too high makes frontend wait. |
| `API_GATEWAY_JWT_HS256_SECRET` | Yes | `change-me-local-only` | JWT signature verification secret | Real value is secret. Never commit or print it. |
| `API_GATEWAY_JWT_ISSUER` | No | `ecommerce-auth` | Expected JWT issuer | Must match token `iss`. |
| `API_GATEWAY_JWT_AUDIENCE` | No | `ecommerce-api` | Expected JWT audience | Must match token `aud`. |
| `API_GATEWAY_JWT_CLOCK_SKEW` | No | `30s` | Small time tolerance for JWT clocks | Keep small. Do not hide clock problems with huge value. |
| `API_GATEWAY_VALIDATION_PHONE_REGION` | No | `IN` | Default phone parse region | Required by config; default is `IN`. |
| `API_GATEWAY_LOG_LEVEL` | No | `info` | JSON log level | Avoid debug logs with sensitive headers. |

Supported fallback aliases:

| Preferred variable | Fallback alias |
|---|---|
| `API_GATEWAY_HTTP_ADDRESS` | `HTTP_ADDR` |
| `API_GATEWAY_SHUTDOWN_TIMEOUT` | `HTTP_SHUTDOWN_TIMEOUT` |
| `USER_SERVICE_GRPC_ADDRESS` | `USER_GRPC_ADDR` |
| `API_GATEWAY_USER_SERVICE_TIMEOUT` | `GRPC_DIAL_TIMEOUT` |
| `API_GATEWAY_JWT_HS256_SECRET` | `JWT_HS256_SECRET` |
| `API_GATEWAY_JWT_ISSUER` | `JWT_ISSUER` |
| `API_GATEWAY_JWT_AUDIENCE` | `JWT_AUDIENCE` |
| `API_GATEWAY_JWT_CLOCK_SKEW` | `JWT_CLOCK_SKEW` |
| `API_GATEWAY_VALIDATION_PHONE_REGION` | `VALIDATION_PHONE_REGION` |
| `API_GATEWAY_LOG_LEVEL` | `LOG_LEVEL` |

How to load `.env`:

```bash
cd backend/services/api-gateway
set -a
. ./.env
set +a
go run ./cmd/server
```

Important caveat:

`USER_SERVICE_GRPC_ADDRESS` is used by both services with different meanings:

- In {SERVICE_NAME}, it is the server listen address, usually `:50052`.
- In API Gateway, it is the upstream dial target, usually `localhost:50052`.

Beginner-safe rule: run {SERVICE_NAME} and API Gateway in separate terminals and source each service's `.env` only inside that terminal. Agar same shell me dono `.env` source kar doge, wrong address accidentally use ho sakta hai.

## 8. Docker Setup

Current task does not add a Dockerfile, `docker-compose.yml`, new container, volume, network, restart policy, or health check.

Reuse Docker/MySQL setup:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Section:
`7. Docker and DevOps Setup`
```

Current Docker status:

| Docker item | Status |
|---|---|
| API Gateway Dockerfile | Not found |
| {SERVICE_NAME} Dockerfile | Not found |
| Project docker-compose | Not found |
| MySQL container | Optional, reused from previous setup |
| Gateway container network | Not configured |
| Gateway health check | HTTP `/healthz` exists in code, but no container health check is configured |

If containerized later:

- Gateway should call {SERVICE_NAME} by service DNS like `user-service:50052`, not `localhost:50052`.
- MySQL should stay private to {SERVICE_NAME}.
- Gateway HTTP port `8080` can be exposed to host/ingress.
- {SERVICE_NAME} gRPC port `50052` should stay internal.

## 9. Local Development Setup

### Step 1: Read previous dependency documentation

Follow these first:

| Order | File | Why |
|---:|---|---|
| 1 | `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | Base Go, MySQL, Docker, `.env`, ports, gRPC tools |
| 2 | `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | MySQL schema/migration setup |
| 3 | `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | Repository runtime assumptions |
| 4 | `TaskImplementation/{SERVICE_NAME}/task4_Dependency.md` | {SERVICE_NAME} gRPC startup and verification |

### Step 2: Go to project directory

```bash
cd /home/parag/Ecommerce
```

### Step 3: Install only current task Gateway dependencies

```bash
cd backend/services/api-gateway
go mod download
```

### Step 4: Setup databases/services

No new database or external service setup for current task.

Do this from previous docs:

- Start MySQL.
- Apply {SERVICE_NAME} migrations.
- Start {SERVICE_NAME} gRPC on `:50052`.

### Step 5: Add only new Gateway environment variables

Create:

```text
backend/services/api-gateway/.env
```

Use the example from section `7. Environment Variables`.

### Step 6: Run migrations if needed

Current task introduces no migration.

If this is a fresh machine, apply previous {SERVICE_NAME} migrations from `task2_Dependency.md`. If database is already migrated through Task 4, nothing new is required.

### Step 7: Start backend services

Terminal 1: start {SERVICE_NAME} after following previous `.env` and DB setup.

```bash
cd backend/services/user-service
set -a
. ./.env
set +a
go run ./cmd/server
```

Expected log should include:

```text
user_service_grpc_listening
```

Terminal 2: start API Gateway.

```bash
cd backend/services/api-gateway
set -a
. ./.env
set +a
go run ./cmd/server
```

Expected log should include:

```text
api_gateway_http_listening
```

### Step 8: Verify APIs related to `{TASK_FILE_NAME}`

Health check, no token needed:

```bash
curl -i http://localhost:8080/healthz
```

Profile endpoint, valid access token needed:

```bash
export ACCESS_TOKEN='<valid-access-token-from-auth-service>'
curl -i \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  http://localhost:8080/api/v1/me
```

Address list:

```bash
curl -i \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  "http://localhost:8080/api/v1/me/addresses?page=1&page_size=20"
```

Seller profile:

```bash
curl -i \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  http://localhost:8080/api/v1/sellers/me
```

Token note: seller endpoint needs a token that contains seller role and `seller_id`. Buyer-only token will correctly return `403`.

## 10. Running the Project

Minimum local runtime for current task:

| Runtime piece | Command/source | Required? |
|---|---|---:|
| MySQL | Previous docs | Yes for real data |
| {SERVICE_NAME} gRPC | `backend/services/user-service` -> `go run ./cmd/server` | Yes |
| API Gateway HTTP | `backend/services/api-gateway` -> `go run ./cmd/server` | Yes |
| Auth Service | Needed to issue real tokens, or use a valid local test token | Yes for real auth flow |

REST endpoints added by current task:

| Method | Path | Role requirement | Upstream gRPC method |
|---|---|---|---|
| `GET` | `/api/v1/me` | `buyer`, `seller`, `admin`, `superadmin` | `GetUser` |
| `PATCH` | `/api/v1/me` | `buyer`, `seller`, `admin`, `superadmin` | `UpdateUserProfile` |
| `GET` | `/api/v1/me/addresses` | `buyer`, `seller`, `admin`, `superadmin` | `ListUserAddresses` |
| `POST` | `/api/v1/me/addresses` | `buyer`, `seller`, `admin`, `superadmin` | `CreateAddress` |
| `PATCH` | `/api/v1/me/addresses/{address_id}` | `buyer`, `seller`, `admin`, `superadmin` | `UpdateAddress` |
| `DELETE` | `/api/v1/me/addresses/{address_id}` | `buyer`, `seller`, `admin`, `superadmin` | `DeleteAddress` |
| `GET` | `/api/v1/sellers/me` | `seller`, `seller_manager`, `superadmin` | `GetSellerProfile` |
| `PATCH` | `/api/v1/sellers/me` | `seller`, `seller_manager`, `superadmin` | `UpdateSellerProfile` |

Ports and networking:

| Service | Port | Purpose | Status |
|---|---:|---|---|
| API Gateway HTTP | `8080` | Public REST entrypoint and `/healthz` | New for current task |
| {SERVICE_NAME} gRPC | `50052` | Internal profile/address/seller gRPC APIs | Reused |
| MySQL | `3306` | {SERVICE_NAME} database | Reused |
| MySQL alternate | `3307` | Local Docker fallback if `3306` busy | Reused |
| Redis | `6379` | Not used by current task | Not required |
| Kafka | `9092` | Not used by current task | Not required |
| RabbitMQ | `5672` | Not used by current task | Not required |

Port conflict fix:

```bash
# Gateway HTTP alternate port
export API_GATEWAY_HTTP_ADDRESS=':8081'

# Gateway upstream target if {SERVICE_NAME} runs on alternate gRPC port
export USER_SERVICE_GRPC_ADDRESS='localhost:50053'
```

## 11. Common Errors & Fixes

Generic Go/MySQL/Docker/gRPC errors already exist in previous docs. Current task specific additions:

| Error | Cause | Fix | Prevention |
|---|---|---|---|
| `API_GATEWAY_JWT_HS256_SECRET is required` | Gateway `.env` missing or not sourced | Add `API_GATEWAY_JWT_HS256_SECRET` and source `.env` before `go run` | Keep sanitized `.env.example` with required names |
| `dial user service at localhost:50052: connection refused` | {SERVICE_NAME} gRPC not running or wrong target | Start {SERVICE_NAME} and confirm port `50052` | Start upstream first, then Gateway |
| `UPSTREAM_UNAVAILABLE` | Gateway reached no healthy {SERVICE_NAME} | Check {SERVICE_NAME} logs, address, firewall, Docker networking | Use clear local addresses and health checks |
| `UPSTREAM_TIMEOUT` | {SERVICE_NAME} call exceeded Gateway timeout | Check DB slowness or increase `API_GATEWAY_USER_SERVICE_TIMEOUT` locally | Keep DB indexed and use realistic timeouts |
| `401 AUTHENTICATION_REQUIRED` | Missing bearer token, invalid signature, expired token, wrong issuer/audience, or wrong token type | Use valid access token signed with matching HS256 secret | Keep Auth and Gateway JWT config aligned |
| `403 PERMISSION_DENIED` | Token roles do not match route requirement | Use token with correct role; seller route also needs `seller_id` | Test buyer and seller tokens separately |
| `400 INVALID_REQUEST` | Bad JSON, unknown field, empty body, or multiple JSON objects | Send one valid JSON object with expected snake_case fields | Use DTO examples from `{TASK_FILE_NAME}` |
| `VALIDATION_ERROR` with phone field | Phone cannot be parsed for configured region | Use E.164 value or set `API_GATEWAY_VALIDATION_PHONE_REGION` correctly | Keep frontend and backend phone rules aligned |
| `listen tcp :8080: bind: address already in use` | Another process uses Gateway port | Change `API_GATEWAY_HTTP_ADDRESS=:8081` | Reserve local ports per service |
| Health check works but `/api/v1/me` fails | Gateway is up, but auth/upstream dependency is failing | Check token and {SERVICE_NAME} gRPC | Verify `/healthz`, then gRPC, then REST |

## 12. Security & Best Practices

Reuse previous security notes for DB credentials, `.env`, Docker, gRPC reflection, and plaintext gRPC:

```md
Refer:
`TaskImplementation/{SERVICE_NAME}/task1_Dependency.md`

Sections:
`9. Common Errors and Fixes`
`10. Security and Configuration Audit`
`11. Best Practices`
```

Current task specific best practices:

- Keep `API_GATEWAY_JWT_HS256_SECRET` same as Auth Service signing secret in local dev, but never commit the real value.
- Do not accept `user_id` from request body for self APIs. Gateway must use JWT claims.
- Keep {SERVICE_NAME} gRPC private. Browser clients should call only Gateway REST.
- Use short but realistic gRPC deadlines through `API_GATEWAY_USER_SERVICE_TIMEOUT`.
- Keep Gateway and {SERVICE_NAME} `.env` files separate to avoid `USER_SERVICE_GRPC_ADDRESS` confusion.
- Log request IDs and upstream errors, but never log full JWTs or secrets.
- Use HTTPS at ingress/proxy level in real environments.
- Add CORS policy before browser frontend integration if frontend runs on a different origin.
- Keep seller APIs strict: token should include seller role and `seller_id`.
- Prefer service-specific env names in future, for example `API_GATEWAY_USER_SERVICE_GRPC_ADDRESS`, to avoid naming collision.

## 13. Missing or Misconfigured Things

| Finding | Impact | Suggested fix |
|---|---|---|
| No committed `.env.example` found for API Gateway | Beginners may miss required JWT secret and upstream address | Add sanitized `backend/services/api-gateway/.env.example` |
| Gateway and {SERVICE_NAME} both use `USER_SERVICE_GRPC_ADDRESS` for different meanings | Same shell can accidentally mix listen address and dial address | Prefer a Gateway-specific variable name or document separate terminals clearly |
| No API Gateway Dockerfile or compose service found | Cannot run full local stack with one command | Add Dockerfile and compose in DevOps task |
| `/healthz` checks only Gateway process | Health can pass while {SERVICE_NAME} is down | Add optional upstream readiness check or separate `/readyz` |
| gRPC dial uses insecure credentials | Fine for local/private network, unsafe over public network | Use TLS/mTLS or private service network in production |
| JWT verifier supports HS256 only | Works if Auth Service signs HS256, but secret sharing is operationally sensitive | Consider JWKS/RS256 later for cleaner key separation |
| No CORS middleware visible | Browser app on different origin may fail preflight | Add explicit CORS config before frontend integration |
| No rate limiting visible | Public endpoints can be abused | Add Gateway-level rate limiting after core route setup |
| No Docker health check configured | Container orchestration cannot use `/healthz` automatically | Add health check in future compose/Kubernetes manifests |

## 14. References to Previous Dependency Files

| Previous Dependency File | Section / Topic Reused | Why Reused |
|---|---|---|
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `1. Project Tech Stack Analysis` | Base Go/MySQL/gRPC stack already explained |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `2. Language-Specific Dependency System: Go` | Go modules, `go.mod`, `go.sum`, `go.work`, and common module commands already documented |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `3. Database Analysis` | MySQL install, Docker setup, DSN, default port, and credentials placement already documented |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `4. Environment Variables` | {SERVICE_NAME} `.env` setup already documented |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `5. External Services Analysis -> gRPC / Protobuf and Buf` | gRPC, grpcurl, protobuf, and Buf setup already explained |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `6. Ports and Networking` | MySQL `3306`, {SERVICE_NAME} gRPC `50052`, and port conflict notes already documented |
| `TaskImplementation/{SERVICE_NAME}/task1_Dependency.md` | `7. Docker and DevOps Setup` | Local Docker MySQL setup and missing Dockerfile caveat already covered |
| `TaskImplementation/{SERVICE_NAME}/task2_Dependency.md` | `5. Database Setup` | Schema/migration setup for {SERVICE_NAME} data already documented |
| `TaskImplementation/{SERVICE_NAME}/task3_Dependency.md` | `7. Database Setup Required For Task 3` | Repository runtime depends on migrated MySQL schema |
| `TaskImplementation/{SERVICE_NAME}/task4_Dependency.md` | `4. Dependency Management` | Proto generation and gRPC transport dependency flow already documented |
| `TaskImplementation/{SERVICE_NAME}/task4_Dependency.md` | `9. Local Development Setup` | Starting and verifying {SERVICE_NAME} gRPC already documented |
| `TaskImplementation/{SERVICE_NAME}/task4_Dependency.md` | `11. Common Errors & Fixes` | Existing gRPC status/reflection/deadline issues already documented |

## 15. Final Checklist

- [ ] Previous dependency documentation checked before current task setup.
- [ ] No duplicate Go/MySQL/Docker setup copied into this file.
- [ ] `backend/services/api-gateway/go.mod` dependencies downloaded.
- [ ] `backend/go.work` includes API Gateway, {SERVICE_NAME}, generated proto, and shared validation modules.
- [ ] MySQL setup reused from previous docs.
- [ ] {SERVICE_NAME} migrations applied if this is a fresh machine.
- [ ] {SERVICE_NAME} gRPC started before API Gateway.
- [ ] `backend/services/api-gateway/.env` created locally.
- [ ] `API_GATEWAY_JWT_HS256_SECRET` configured and not committed.
- [ ] Gateway upstream address points to running {SERVICE_NAME} gRPC.
- [ ] Gateway HTTP port `8080` is free or changed intentionally.
- [ ] `/healthz` verified with `curl`.
- [ ] Profile REST endpoint verified with valid access token.
- [ ] Seller endpoint verified with seller token containing `seller_id`.
- [ ] Gateway tests pass with `go test ./...`.
- [ ] Logs checked for `api_gateway_http_listening`.
- [ ] No Redis/Kafka/RabbitMQ setup added for current task.
- [ ] Missing `.env.example`, Dockerfile, CORS, and readiness gaps noted for future DevOps tasks.
