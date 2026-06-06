# Project Dependency & Setup Guide

Input analyzed:

`TaskImplementation/API Gateway Service/task4.md`

Previous dependency documentation reviewed first:

- `TaskImplementation/API Gateway Service/task1_Dependency.md`
- `TaskImplementation/API Gateway Service/task2_Dependency.md`
- `TaskImplementation/API Gateway Service/task3_Dependency.md`

Related implementation inspected:

- `backend/services/api-gateway/go.mod`
- `backend/services/api-gateway/cmd/server/main.go`
- `backend/services/api-gateway/internal/config/config.go`
- `backend/services/api-gateway/internal/ratelimit/*`
- `backend/services/api-gateway/internal/transport/http/rate_limit_middleware.go`
- `backend/services/api-gateway/internal/transport/http/routes.go`

Simple goal: Ye guide beginner developer ko batata hai ki API Gateway Task 4, yani Redis-backed rate limiting, ko run/debug karne ke liye code ke alawa kya setup chahiye. Isme business logic rewrite nahi hai. Sirf dependencies, env, Redis, Docker, local run, troubleshooting, and DevOps notes hain.

---

## 1. Project Overview

Task 4 API Gateway me request traffic control karta hai.

Flow simple words me:

```text
Client request
  -> API Gateway
  -> pre-auth rate limit: IP / target / route
  -> auth middleware if protected route
  -> post-auth rate limit: user
  -> handler / downstream gRPC service
```

Current implementation reality:

| Area | Current Status |
|---|---|
| Runtime language | Go |
| HTTP stack | Standard `net/http` / `http.ServeMux` |
| Rate-limit backend | Redis |
| Rate-limit algorithm | Redis Lua token bucket |
| Go Redis client | `github.com/redis/go-redis/v9` |
| Default mode | `RATE_LIMIT_ENABLED=true` |
| Redis startup behavior | Gateway pings Redis during startup and fails if Redis is unavailable |
| Pre-auth limits | IP, target, and route dimensions |
| Post-auth limits | User dimension from Task 3 auth claims |
| API Gateway SQL DB | None |
| API Gateway migrations | None |

Important beginner note:

```text
API Gateway directly MySQL/MongoDB query nahi karta.
Rate limit counters Redis me store hote hain.
Full runtime still needs Task 2 gRPC services and Task 3 JWKS config.
```

Already documented elsewhere, so not repeated here:

- Base Go install and Go module basics
- Full Redis installation and Docker run examples
- MySQL/MongoDB/Typesense/Kafka/RabbitMQ setup
- Downstream gRPC address table
- JWT/JWKS setup
- Full Gateway `.env` example

Use the references in section 14 for those topics.

---

## 2. Tech Stack

### Task 4 Specific Technologies

| Technology | What it is | Why this task uses it | Required? | Beginner Hinglish explanation |
|---|---|---|---:|---|
| Go | Backend programming language | API Gateway service Go me implemented hai | Yes | Go fast compiled backend language hai. Gateway isi se build/run hota hai. |
| Go Modules | Go dependency system | `go.mod` and `go.sum` dependency versions manage karte hain | Yes | Go modules npm/pip jaisa dependency manager hai, Go projects ke liye. |
| `net/http` | Go standard HTTP server | Middleware chain and routes run karne ke liye | Yes | External web framework nahi hai; Go ka built-in HTTP package use ho raha hai. |
| Redis | In-memory data store | Shared rate-limit counters store karne ke liye | Yes when rate limit enabled | Redis fast memory database hai. Multiple Gateway instances same counters share kar sakte hain. |
| `go-redis/v9` | Go Redis client library | Go code se Redis `PING`, Lua script, and commands chalane ke liye | Yes | Ye Gateway aur Redis ke beech driver/connector hai. |
| Redis Lua script | Atomic Redis script | Token read/refill/consume ek atomic operation me karne ke liye | Yes | Lua script race condition avoid karta hai jab many requests ek saath aati hain. |
| Token bucket | Rate-limit algorithm | Short burst allow karta hai but sustained abuse block karta hai | Yes | Bucket me tokens hote hain. Request token consume karti hai. Time ke saath token refill hota hai. |
| SHA-256 hashing | Standard hashing | Email, phone, IP, user id raw Redis key me expose na ho | Yes | Sensitive identity ko readable key ke badle hashed value me convert karta hai. |
| `net/netip` | Go IP parser | Client IP normalize and trusted proxy CIDR validate karne ke liye | Yes | IP string ko safe/consistent format me handle karta hai. |
| JWT auth context | Verified user context | Post-auth user limit ke liye user id read karna | Required for user limits | Auth ke baad Gateway ko `sub` user id milta hai, usi se user quota lagta hai. |
| Prometheus metrics | Metrics collection | Rate-limited request count expose karne ke liye | Optional by config | Metrics se pata chalta hai kaunse route par throttling ho rahi hai. |
| Docker | Local dependency runner | Redis ko easy local container me run karne ke liye | Recommended | Docker se Redis setup repeatable and clean hota hai. |

### Already Explained

General Gateway stack is already explained in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`2. Tech Stack`

gRPC stack is already explained in:

`TaskImplementation/API Gateway Service/task2_Dependency.md`

Section:

`2. Tech Stack`

JWT/JWKS stack is already explained in:

`TaskImplementation/API Gateway Service/task3_Dependency.md`

Section:

`2. Tech Stack`

---

## 3. Required Software

### Required for Task 4 Development

| Software | Required? | Why |
|---|---:|---|
| Git | Yes | Repository clone/update ke liye |
| Go `1.26.3` compatible toolchain | Yes | `backend/services/api-gateway/go.mod` declares `go 1.26.3` |
| Redis Server | Yes for full runtime | Rate-limit counters store karne ke liye |
| Redis CLI | Recommended | `PING`, key scan, TTL debug ke liye |
| Docker / Docker Compose | Recommended | Local Redis container run karne ke liye |
| curl | Recommended | Gateway HTTP endpoints and rate-limit smoke tests ke liye |
| lsof / netstat | Optional | Port conflict debug karne ke liye |

Base installation steps for Git, Go, Docker, Redis CLI, and OS-specific commands are already documented in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`3. Required Software`

### Setup Modes

| Mode | Redis Needed? | Downstream gRPC Needed? | JWKS Needed? | Use Case |
|---|---:|---:|---:|---|
| Rate-limit unit tests | No | No | No | Middleware/key/policy tests run karne ke liye |
| API Gateway full runtime with rate limit on | Yes | Yes | Yes for protected routes | Real local Gateway run |
| Temporary route/debug mode with `RATE_LIMIT_ENABLED=false` | No | Yes | Yes for protected routes | Sirf Redis issue isolate karne ke liye |

Warning:

`RATE_LIMIT_ENABLED=false` local debugging ke liye useful ho sakta hai, but realistic dev/staging/prod me rate limiting enabled rehni chahiye.

---

## 4. Dependency Management

This is a Go project. General Go module explanation is already covered in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`4. Dependency Management`

### Task 4 Specific Go Dependencies

Current `backend/services/api-gateway/go.mod` includes:

| Dependency | Version | Task 4 Use |
|---|---:|---|
| `github.com/redis/go-redis/v9` | `v9.19.0` | Redis client, Lua script execution, startup `PING` |
| `github.com/prometheus/client_golang` | `v1.23.2` | Rate-limited request metric |
| `github.com/golang-jwt/jwt/v5` | `v5.3.1` | User id is available after Task 3 JWT auth for post-auth limits |
| `go.opentelemetry.io/otel` | `v1.43.0` | Request tracing around Gateway middleware chain |

Standard library packages used by Task 4:

| Package | Why |
|---|---|
| `crypto/sha256` | Redis key identity hashing |
| `encoding/hex` | Hash formatting |
| `net/netip` | Trusted proxy CIDR and IP parsing |
| `encoding/json` | OTP target/body extraction |
| `time` | Limit windows, retry-after, TTL |

### Commands

Run from API Gateway service directory:

```bash
cd backend/services/api-gateway
```

Download dependencies:

```bash
go mod download
```

Run Task 4 focused tests:

```bash
go test ./internal/ratelimit ./internal/transport/http
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

### Task 4 Dependency Issues

| Error | Likely Cause | Fix |
|---|---|---|
| `missing go.sum entry` | Dependency checksum missing | Run `go mod tidy` |
| `go: go.mod requires go >= 1.26.3` | Local Go version old | Install compatible Go version |
| `redis limiter is not configured` | Code path created limiter with nil Redis client | In runtime, keep `RATE_LIMIT_ENABLED=true` and let `main.go` create Redis client |
| `redis ping failed` | Redis dependency unavailable or wrong env | See sections 5, 7, and 11 |

General Go module troubleshooting is already in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`4. Dependency Management` -> `Dependency Problems and Fixes`

---

## 5. Database Setup

### API Gateway Direct Database Requirement

API Gateway Task 4 does not introduce a SQL or document database.

| Store | Directly used by Gateway Task 4? | Purpose |
|---|---:|---|
| MySQL | No | Owned by downstream services |
| MongoDB | No | Owned by downstream services |
| Redis | Yes | Rate-limit counters and token bucket state |
| Typesense | No | Owned by Search Service |
| Kafka/RabbitMQ | No | Platform async events, not Gateway rate limiting |

### Redis for Rate Limiting

#### A. What It Is

Redis ek fast in-memory store hai. Is project me Redis request counters/token buckets store karta hai.

Example:

```text
IP 203.0.113.10 ne login route 10 minutes me kitni baar hit kiya?
User user_123 ne checkout route 10 minutes me kitni baar hit kiya?
```

#### B. Why This Project Uses It

Gateway ke multiple instances ho sakte hain. Agar limiter memory me hota, har instance ka counter alag hota. Redis shared hai, isliye total platform-wide limit consistent rehta hai.

#### C. Required or Optional

| Context | Required? |
|---|---:|
| `RATE_LIMIT_ENABLED=true` | Yes |
| `RATE_LIMIT_ENABLED=false` | No, but only for temporary local debug |
| Production/Staging | Yes |

#### D. Local Installation

Redis installation already exists in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`5. Database Setup` -> `Redis` -> `D. Local Installation`

Do not duplicate the install steps here. Follow that section exactly.

#### E. Docker Setup

Redis Docker `docker run` and Docker Compose examples already exist in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`5. Database Setup` -> `Redis` -> `E. Docker Setup`

Use the same setup. Task 4 does not require a different Redis image or port.

#### F. Start Commands

Redis start commands are already documented in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`5. Database Setup` -> `Redis` -> `F. Start Commands`

#### G. Verify Running

After following the Redis setup guide, verify:

```bash
redis-cli -h localhost -p 6379 PING
```

If Redis has password:

```bash
redis-cli -h localhost -p 6379 -a localredispass PING
```

Expected:

```text
PONG
```

#### H. Default Port

```text
6379
```

#### I. Connection Env Format

Gateway uses separate Redis env vars:

```env
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_TLS_ENABLED=false
REDIS_DIAL_TIMEOUT=3s
```

Inside Docker Compose network, host value usually changes:

```env
REDIS_ADDR=redis:6379
```

#### J. Where To Place Credentials

Local Gateway env file:

```text
backend/services/api-gateway/.env
```

Important:

- `.env` is ignored by `.gitignore`.
- Do not commit Redis password.
- Do not put MySQL/MongoDB credentials in Gateway `.env`; those belong to downstream services.
- Production Redis password/TLS should come from secret manager, Kubernetes Secret, Docker secret, or managed platform secret storage.

### Redis Keys Created by Task 4

Redis keys use this shape:

```text
<RATE_LIMIT_KEY_PREFIX>:<dimension>:<policy_name>:<hashed_identity>
```

Examples:

```text
rl:v1:ip:auth.login.ip:<hash>
rl:v1:target:auth.otp_send.target:<hash>
rl:v1:user:checkout.user:<hash>
```

Raw email, phone, IP, and user id should not appear in keys. Current implementation hashes identity with SHA-256 and keeps the first 32 hex characters.

Local inspect command:

```bash
redis-cli --scan --pattern 'rl:v1:*'
```

Warning:

Do not run broad delete commands in production. For local reset, prefer deleting only your local Redis container/volume or scanning the exact `RATE_LIMIT_KEY_PREFIX`.

### Migrations

No API Gateway database migration is required for Task 4.

Redis keys are runtime cache/counter data. They expire automatically using TTL based on rate-limit window.

---

## 6. Redis / Queue / External Services

### Direct Task 4 External Service

| Service | Directly used? | Why |
|---|---:|---|
| Redis | Yes | Shared token bucket counters |
| Auth Service JWKS | Inherited from Task 3 | Needed when protected routes exist |
| Downstream gRPC services | Inherited from Task 2 | Gateway startup/readiness and route bridge |
| Kafka/RabbitMQ | No new Task 4 dependency | Platform-level async services only |
| Typesense | No new Task 4 dependency | Search Service owns it |

### Rate-Limit Policy Matrix

Current default policy behavior from `internal/ratelimit/policy.go`:

| Policy | Dimension | Route IDs / Purpose | Limit |
|---|---|---|---|
| `global.ip` | IP | Fallback for all routes | `RATE_LIMIT_DEFAULT_IP_LIMIT` / `RATE_LIMIT_DEFAULT_IP_WINDOW` |
| `auth.login.ip` | IP | Login brute-force control | 5 / 10 min |
| `auth.otp_send.target` | Target | Same email/phone OTP spam control | 3 / 15 min |
| `auth.otp_send.ip` | IP | OTP abuse by IP | 10 / 15 min |
| `auth.password_reset.ip` | IP | Password reset abuse | 5 / 15 min |
| `search.public.ip` | IP | Search/autocomplete | 120 / min |
| `catalog.public.ip` | IP | Product/category browsing | 240 / min |
| `cart.user` | User | Cart mutations | 60 / min |
| `wishlist.user` | User | Wishlist mutations | 60 / min |
| `checkout.user` | User | Checkout | 10 / 10 min |
| `seller.mutation.user` | User | Seller mutations | 60 / min |
| `admin.mutation.user` | User | Admin/superadmin mutations | 30 / min |
| `webhook.provider.ip` | IP | Payment webhook route | 120 / min |

Beginner explanation:

Ek request par multiple policies apply ho sakti hain. Example login route par `global.ip` plus `auth.login.ip` dono check ho sakte hain. Request tabhi continue karegi jab applicable policies pass hongi.

### Ports & Networking

| Service | Port | Purpose | Task 4 Status |
|---|---:|---|---|
| API Gateway HTTP | `8080` | REST routes and health endpoints | Reused |
| API Gateway metrics | `9090` | Prometheus metrics, including rate-limited count | Reused |
| Redis | `6379` | Rate-limit counters | Direct Task 4 requirement |
| Auth Service HTTP/JWKS | `8081` | JWT public keys | Inherited from Task 3 |
| Downstream gRPC services | `50051` to `50062` local | Internal service targets | Inherited from Task 2 |

Networking rules:

- Gateway running on host and Redis Docker port published: `REDIS_ADDR=localhost:6379`.
- Gateway running inside Docker Compose: `REDIS_ADDR=redis:6379`.
- Gateway running in Kubernetes: use service DNS, for example `redis.data.svc.cluster.local:6379`.
- Do not use `localhost` from one container to reach another container.
- If Gateway is behind Nginx/Ingress/LB, configure `RATE_LIMIT_TRUSTED_PROXY_CIDRS` so real client IP can be read safely.

---

## 7. Environment Variables

Full Gateway `.env` is already documented in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`7. Environment Variables`

This section lists only Task 4 rate-limit variables. Some of these already appear in the full Task 1 `.env` example; they are repeated here only as the Task 4 subset to verify.

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

### Task 4 `.env` Subset

```env
# Redis for API Gateway rate limiting
RATE_LIMIT_ENABLED=true
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_TLS_ENABLED=false
REDIS_DIAL_TIMEOUT=3s

# Rate-limit behavior
RATE_LIMIT_KEY_PREFIX=rl:v1
RATE_LIMIT_FAIL_OPEN=false
RATE_LIMIT_DEFAULT_IP_LIMIT=600
RATE_LIMIT_DEFAULT_IP_WINDOW=1m
RATE_LIMIT_TRUSTED_PROXY_CIDRS=
RATE_LIMIT_TARGET_BODY_LIMIT_BYTES=65536
```

### Variable Explanation

| Variable | Required? | Default | Purpose | Beginner Notes |
|---|---:|---|---|---|
| `RATE_LIMIT_ENABLED` | Optional | `true` | Enables Redis rate limiter | Keep true for realistic local testing |
| `REDIS_ADDR` | Required when enabled | `localhost:6379` | Redis host/port | Use `redis:6379` inside Compose |
| `REDIS_PASSWORD` | Optional | blank | Redis auth password | Secret; do not commit |
| `REDIS_DB` | Optional | `0` | Redis logical DB number | Must be non-negative |
| `REDIS_TLS_ENABLED` | Optional | `false` | Enable TLS for Redis | Local usually false; managed prod Redis may require true |
| `REDIS_DIAL_TIMEOUT` | Optional | `3s` | Max time for Redis connect/ping | Must be positive |
| `RATE_LIMIT_KEY_PREFIX` | Required when enabled | `rl:v1` | Namespace for Redis keys | Change for test isolation, for example `rl:test` |
| `RATE_LIMIT_FAIL_OPEN` | Optional | `false` | Redis error pe allow or block | `false` means fail closed, safer for auth/checkout/admin |
| `RATE_LIMIT_DEFAULT_IP_LIMIT` | Optional | `600` | Global fallback IP limit count | Positive integer |
| `RATE_LIMIT_DEFAULT_IP_WINDOW` | Optional | `1m` | Global fallback IP limit window | Valid Go duration like `30s`, `1m`, `10m` |
| `RATE_LIMIT_TRUSTED_PROXY_CIDRS` | Optional | blank | Trusted proxy ranges for forwarded IP headers | Example: `10.0.0.0/8,127.0.0.1/32` |
| `RATE_LIMIT_TARGET_BODY_LIMIT_BYTES` | Optional | `65536` | Max body bytes read to extract OTP target | Must be positive |

### Credentials Placement

| Value | Put in Gateway `.env`? | Correct Place |
|---|---:|---|
| Redis host/port | Yes | `REDIS_ADDR` |
| Redis password | Yes for local, secret manager for shared/prod | `REDIS_PASSWORD` |
| Redis TLS toggle | Yes | `REDIS_TLS_ENABLED` |
| MySQL password | No | Owning downstream service `.env` |
| MongoDB URI | No | Owning downstream service `.env` |
| JWT private key | No | Auth Service secret only |
| JWT public key URL | Yes, inherited Task 3 | `JWT_JWKS_URL` |

---

## 8. Docker Setup

Docker basics and Redis container setup are already documented in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Sections:

- `5. Database Setup` -> `Redis` -> `E. Docker Setup`
- `8. Docker Setup`

Task 4 does not require a new Docker image, new port, or different Redis container.

### Current Repo Status

No committed API Gateway Dockerfile or docker-compose file was found for this service.

So recommended beginner flow remains:

```text
Run Redis using Docker.
Run API Gateway locally with go run.
Run downstream services using their own setup when full runtime is needed.
```

### Docker Networking Rule for Redis

| Gateway Location | Redis Env |
|---|---|
| Gateway on host machine, Redis container exposes `6379:6379` | `REDIS_ADDR=localhost:6379` |
| Gateway inside same Compose network as Redis service named `redis` | `REDIS_ADDR=redis:6379` |
| Gateway in Kubernetes | `REDIS_ADDR=<redis-service-dns>:6379` |

### Redis Healthcheck Expectation

Use the healthcheck pattern already shown in `task1_Dependency.md`. For password-protected Redis, healthcheck must include the same password that Gateway uses.

Debug:

```bash
docker ps
docker logs ecommerce-redis
redis-cli -h localhost -p 6379 PING
```

With password:

```bash
redis-cli -h localhost -p 6379 -a localredispass PING
```

---

## 9. Local Development Setup

### Reused Base Onboarding

Clone repo, install Go, install Docker, download Go dependencies, create/export `.env`, and full Gateway run flow are already explained in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Sections:

- `9. Local Development Setup`
- `10. Running the Project`

Downstream gRPC service setup is in:

`TaskImplementation/API Gateway Service/task2_Dependency.md`

Section:

`9. Local Development Setup`

JWT/JWKS setup is in:

`TaskImplementation/API Gateway Service/task3_Dependency.md`

Section:

`9. Local Development Setup`

### Task 4 Incremental Setup

1. Move to Gateway service:

```bash
cd backend/services/api-gateway
```

2. Download dependencies:

```bash
go mod download
```

3. Run rate-limit focused tests:

```bash
go test ./internal/ratelimit ./internal/transport/http
```

4. Start Redis by following:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`5. Database Setup` -> `Redis`

5. Verify Redis:

```bash
redis-cli -h localhost -p 6379 PING
```

6. Add/export Task 4 env vars:

```bash
set -a
source .env
set +a
```

7. For full Gateway runtime, also start inherited dependencies:

- Downstream gRPC services from `task2_Dependency.md`
- Auth Service JWKS endpoint from `task3_Dependency.md`

8. Run Gateway:

```bash
go run ./cmd/server
```

9. Verify health:

```bash
curl -i http://localhost:8080/health/live
curl -i http://localhost:8080/health/ready
```

10. Verify Redis keys after making requests:

```bash
redis-cli --scan --pattern 'rl:v1:*'
```

---

## 10. Running the Project

### Minimal Task 4 Verification

Use this when Redis/downstream services are not running and you only want to validate Task 4 code dependencies:

```bash
cd backend/services/api-gateway
go mod download
go test ./internal/ratelimit ./internal/transport/http
```

### Full Runtime Verification

Use this when Redis, downstream gRPC services, and JWKS are available:

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
```

### Manual Rate-Limit Smoke Test

Login route policy is tight: `5 / 10 min`.

```bash
for i in $(seq 1 6); do
  curl -i -X POST http://localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"email":"demo@example.com","password":"wrong-password"}'
done
```

Expected behavior:

| Request | Expected |
|---|---|
| First few requests | Normal Gateway route/auth/validation response |
| After configured limit | `429 Too Many Requests` with `RATE_LIMITED` |

Current Gateway may return `501 ROUTE_BRIDGE_NOT_CONFIGURED` after middleware passes because route bridge work is separate. That is okay for this task. For Task 4, important check is that repeated requests eventually return `429`.

### Expected Rate-Limit Headers

Blocked response should include useful headers:

```text
RateLimit-Limit: <limit>
RateLimit-Remaining: 0
RateLimit-Reset: <seconds>
Retry-After: <seconds>
```

### Migrations

No API Gateway migration exists for Task 4.

Full platform database setup/migration references are in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`5. Database Setup`

---

## 11. Common Errors & Fixes

General Go/Docker/Redis errors are already covered in:

`TaskImplementation/API Gateway Service/task1_Dependency.md`

Section:

`11. Common Errors & Fixes`

Task 4 specific troubleshooting:

| Error / Symptom | Likely Cause | Fix |
|---|---|---|
| `redis ping failed` on startup | Redis not running, wrong host/port, password mismatch, or TLS mismatch | Verify `redis-cli PING`, then check `REDIS_ADDR`, `REDIS_PASSWORD`, `REDIS_TLS_ENABLED` |
| `NOAUTH Authentication required` | Redis requires password but `REDIS_PASSWORD` is blank | Set `REDIS_PASSWORD` to match Redis server |
| `ERR invalid password` | Wrong password in Gateway env | Fix `REDIS_PASSWORD` or restart local Redis with expected password |
| `REDIS_ADDR is required when rate limiting is enabled` | `RATE_LIMIT_ENABLED=true` and blank Redis address | Set `REDIS_ADDR=localhost:6379` or Compose DNS |
| `REDIS_DB must not be negative` | Invalid env value | Use `REDIS_DB=0` for local |
| `RATE_LIMIT_KEY_PREFIX must not contain whitespace` | Prefix has spaces/newline | Use simple value like `rl:v1` |
| `RATE_LIMIT_DEFAULT_IP_WINDOW must be positive` | Bad duration or zero/negative value | Use `1m`, `30s`, or another positive Go duration |
| `RATE_LIMIT_TRUSTED_PROXY_CIDRS contains invalid CIDR` | CIDR typo | Use valid CIDR like `10.0.0.0/8` |
| Every request returns `503 Service temporarily unavailable` | Redis command failing and fail-closed mode active | Fix Redis connectivity or temporarily set `RATE_LIMIT_FAIL_OPEN=true` for local debugging only |
| Every request returns `429` | Limit too low, same IP shared, old local Redis keys, or unknown target bucket | Increase local limit, clear local Redis data, or inspect policy/key behavior |
| Real client IP seems wrong behind proxy | `RATE_LIMIT_TRUSTED_PROXY_CIDRS` not configured | Add trusted proxy/LB CIDRs so `X-Forwarded-For` is trusted safely |
| Redis keys show only one shared bucket | Gateway sees all users as same proxy IP or target extraction failed | Configure trusted proxies and check request body field names |
| OTP target limit not behaving as expected | Body missing `target`, invalid JSON, or body bigger than `RATE_LIMIT_TARGET_BODY_LIMIT_BYTES` | Send valid JSON with `target` and keep body under configured limit |
| Local test passes but full run fails | Unit tests use stubs, full runtime needs real Redis and inherited services | Start Redis, JWKS, and gRPC services before `go run` |
| Gateway in Docker cannot reach Redis | Using `localhost` inside container | Use Compose service DNS like `redis:6379` |

Useful debug commands:

```bash
redis-cli -h localhost -p 6379 PING
redis-cli --scan --pattern 'rl:v1:*'
curl -i http://localhost:8080/health/live
curl -i http://localhost:8080/health/ready
lsof -i :6379
lsof -i :8080
```

Windows port debug:

```powershell
netstat -ano | findstr :6379
netstat -ano | findstr :8080
```

---

## 12. Security & Best Practices

### Task 4 Security Notes

- Keep `RATE_LIMIT_ENABLED=true` in realistic environments.
- Use Redis password and network isolation outside local development.
- Use `REDIS_TLS_ENABLED=true` when managed/production Redis requires TLS.
- Keep `RATE_LIMIT_FAIL_OPEN=false` for auth, OTP, checkout, admin, and webhook routes.
- Configure `RATE_LIMIT_TRUSTED_PROXY_CIDRS` carefully. Blindly trusting `X-Forwarded-For` allows IP spoofing.
- Do not store raw email, phone, JWT, refresh token, or payment data in Redis keys.
- Do not log raw request bodies while debugging target extraction.
- Do not expose Redis publicly on the internet.
- Do not run broad Redis deletes in production.
- Do not add downstream DB credentials to Gateway `.env`.

### Beginner Best Practices

| Practice | Why |
|---|---|
| Test Redis with `redis-cli PING` before debugging Go | Confirms dependency is alive |
| Use separate `RATE_LIMIT_KEY_PREFIX` for tests | Test keys do not mix with local dev keys |
| Keep local limits realistic | Too high hides bugs, too low blocks normal manual testing |
| Use user limits for logged-in workflows | IP-only limits can hurt offices/shared networks |
| Keep route IDs normalized | Path params should not create unlimited unique buckets |
| Check `Retry-After` on `429` | Frontend can back off cleanly |
| Start Redis before Gateway | Gateway pings Redis during startup |
| Use Docker service names inside Compose | Container networking works predictably |

---

## 13. Missing or Misconfigured Things

Professional setup audit for Task 4:

| Finding | Impact | Suggested Fix |
|---|---|---|
| No API Gateway `.env.example` found | Beginners may miss rate-limit env vars | Add `backend/services/api-gateway/.env.example` with safe local defaults |
| No committed API Gateway Dockerfile or compose file found | Containerized Gateway run is manual | Add service Dockerfile and local compose when packaging is needed |
| Redis is required at startup when rate limiting is enabled | Full Gateway run fails if Redis is not started | Document Redis-first startup or provide compose profile |
| Rate-limit policies are currently code-defined defaults | Runtime tuning requires code/config change | Consider external policy config later if product limits change often |
| `RATE_LIMIT_FAIL_OPEN` is global | Cannot fail-open only for catalog while fail-closed for auth | Add per-policy fail mode later if availability/security needs differ |
| `RATE_LIMIT_TRUSTED_PROXY_CIDRS` default is blank | Behind a proxy, Gateway may rate-limit proxy IP instead of real client IP | Set trusted proxy CIDRs per environment |
| No Redis cluster/sentinel config found | Single Redis endpoint assumption may not fit production HA | Add managed Redis/cluster config when infra design is finalized |
| Redis TLS config only toggles TLS minimum version | Advanced prod TLS options like CA/server name are not exposed | Add CA cert/server name/mTLS envs if required |
| Current tests do not run a real Redis integration token-bucket test | Lua/runtime Redis behavior has limited integration coverage | Add Docker-backed or testcontainer Redis integration test |
| OTP target extraction depends on request body field names | Missing/malformed target may use shared `unknown` bucket | Keep Task 5 validation strict and align request schema |
| Full runtime still requires Task 2 and Task 3 dependencies | Rate-limit-only manual testing can be blocked by unrelated services | Add documented stub/mock mode if developers need isolated Gateway runtime |

No hardcoded Redis password, MySQL password, MongoDB URI, JWT private key, or payment provider secret was found in the Task 4 Gateway files inspected.

---

## 14. References to Previous Dependency Files

To avoid duplicate documentation, use these existing sections:

| Topic | Refer |
|---|---|
| Git/Go/Docker installation | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `3. Required Software` |
| Go modules, `go.mod`, `go.sum`, `go mod download`, `go test` basics | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `4. Dependency Management` |
| MySQL setup | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `5. Database Setup` -> `MySQL` |
| MongoDB setup | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `5. Database Setup` -> `MongoDB` |
| Redis install/start/Docker setup | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `5. Database Setup` -> `Redis` |
| Typesense setup | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `6. Redis / Queue / External Services` -> `Typesense` |
| Kafka/RabbitMQ setup | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `6. Redis / Queue / External Services` -> `Kafka or RabbitMQ` |
| Full Gateway `.env` example | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `7. Environment Variables` |
| Full ports and networking table | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `7. Environment Variables` -> `Ports & Networking` |
| Docker dependency compose examples | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `8. Docker Setup` |
| Full local run flow | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `9. Local Development Setup` and `10. Running the Project` |
| General troubleshooting | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `11. Common Errors & Fixes` |
| Security baseline | `TaskImplementation/API Gateway Service/task1_Dependency.md` -> `12. Security & Best Practices` |
| Downstream gRPC addresses and health checks | `TaskImplementation/API Gateway Service/task2_Dependency.md` -> `6. Redis / Queue / External Services` |
| gRPC Docker networking rules | `TaskImplementation/API Gateway Service/task2_Dependency.md` -> `8. Docker Setup` |
| gRPC-specific troubleshooting | `TaskImplementation/API Gateway Service/task2_Dependency.md` -> `11. Common Errors & Fixes` |
| JWT/JWKS setup | `TaskImplementation/API Gateway Service/task3_Dependency.md` -> `6. Redis / Queue / External Services` -> `Auth Service JWKS Endpoint` |
| JWT/JWKS env vars | `TaskImplementation/API Gateway Service/task3_Dependency.md` -> `7. Environment Variables` |
| Auth-specific troubleshooting | `TaskImplementation/API Gateway Service/task3_Dependency.md` -> `11. Common Errors & Fixes` |

---

## 15. Final Checklist

Task 4 specific checklist:

```text
[ ] Read task1_Dependency.md Redis setup section
[ ] Read task2_Dependency.md if running full Gateway with downstream gRPC
[ ] Read task3_Dependency.md if protected routes/JWKS are involved
[ ] Go version is compatible with backend/services/api-gateway/go.mod
[ ] Ran: cd backend/services/api-gateway
[ ] Ran: go mod download
[ ] Ran: go test ./internal/ratelimit ./internal/transport/http
[ ] Redis server/container is running
[ ] redis-cli PING returns PONG
[ ] backend/services/api-gateway/.env exists locally
[ ] .env is exported before go run
[ ] RATE_LIMIT_ENABLED=true for realistic testing
[ ] REDIS_ADDR points to correct host/port
[ ] REDIS_PASSWORD matches Redis server config
[ ] REDIS_DB is non-negative
[ ] REDIS_TLS_ENABLED matches Redis server TLS mode
[ ] REDIS_DIAL_TIMEOUT is a positive duration
[ ] RATE_LIMIT_KEY_PREFIX has no whitespace
[ ] RATE_LIMIT_FAIL_OPEN is intentionally chosen
[ ] RATE_LIMIT_DEFAULT_IP_LIMIT is positive
[ ] RATE_LIMIT_DEFAULT_IP_WINDOW is positive
[ ] RATE_LIMIT_TRUSTED_PROXY_CIDRS is configured if behind proxy/LB
[ ] RATE_LIMIT_TARGET_BODY_LIMIT_BYTES is positive
[ ] Downstream gRPC services are running for full Gateway runtime
[ ] Auth JWKS endpoint is running for protected routes
[ ] Gateway starts without redis ping errors
[ ] Repeated login/search test eventually returns 429
[ ] 429 response includes Retry-After and RateLimit headers
[ ] redis-cli --scan --pattern 'rl:v1:*' shows hashed rate-limit keys
[ ] Redis keys do not expose raw email, phone, JWT, or user id
[ ] .env and Redis password are not staged in git
```

Final beginner note:

Task 4 ka core dependency Redis hai. Agar `go test` pass ho raha hai but `go run ./cmd/server` fail ho raha hai, sabse pehle Redis `PING`, `REDIS_ADDR`, `REDIS_PASSWORD`, and Docker networking check karo. Uske baad inherited gRPC/JWKS dependencies check karo.
