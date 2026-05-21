# Project Dependency & Setup Guide

Input implementation file:

```text
TaskImplementation/Auth Service/task5.md
```

Generated dependency file:

```text
TaskImplementation/Auth Service/task5_Dependency.md
```

This guide is for **Auth Service - Task 5: OTP Verification**.

Simple goal: beginner developer ko clearly samajh aaye ki OTP send/verify flow chalane ke liye kaunsi dependency, env variable, database table, Redis setup, Notification Service endpoint, and debugging steps chahiye.

Important reuse rule followed:

- Previous Auth Service dependency files were inspected first.
- Common Go, MySQL, Redis, Docker, JWT, password, and full local run setup is already documented in earlier files.
- This file does not repeat those full installation steps. It references them and explains only Task 5 specific OTP setup and differences.

Previous dependency files inspected:

```text
TaskImplementation/Auth Service/task1_Dependency.md
TaskImplementation/Auth Service/task2_Dependency.md
TaskImplementation/Auth Service/task3_Dependency.md
TaskImplementation/Auth Service/task4_Dependency.md
```

Current repository note:

- `task5.md` is a detailed implementation guide.
- Current repo already contains OTP implementation under `backend/services/auth-service/internal/security/otp/`, `internal/usecase/otp.go`, `internal/repository/mysql_otp_repository.go`, `internal/repository/redis_otp_rate_repository.go`, and HTTP handlers.
- `task5.md` mentions some future/recommended packages like gRPC, `redis/go-redis/v9`, and validator. Current code does not import those packages.

---

## 1. Project Overview

Task 5 adds the setup required for OTP verification:

- Generate email/phone OTP codes.
- Store only OTP hash in MySQL.
- Use Redis for resend cooldown and rate limits.
- Send the plain OTP to Notification Service for delivery.
- Verify OTP with expiry, attempts, replay protection, and source/IP throttling.

Hinglish explanation:

OTP ka full form One-Time Password hai. User ke email ya phone par short code send hota hai. Auth Service OTP create and verify karta hai. Notification Service sirf delivery karta hai. Database me raw OTP kabhi store nahi hota, only HMAC hash store hota hai.

Task 5 runtime dependencies:

| Dependency | Required? | Why |
|---|---:|---|
| Go | Yes | Auth Service OTP code build/test/run karne ke liye |
| MySQL 8+ | Yes for full OTP flow | `otp_challenges` table stores OTP challenge hash, attempts, expiry |
| Redis | Yes for full OTP flow | Cooldown, send quota, daily limit, verify source throttling |
| Notification Service endpoint | Yes for real OTP send | Auth Service HTTP POST karke OTP delivery request bhejta hai |
| JWT keys and refresh token config | Required for full server startup | Current `main.go` initializes token components even when testing OTP endpoints |
| Docker | Optional | MySQL/Redis dependencies run karne ke liye easiest local option |

Baseline full service setup is already explained in:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Sections:
3. Required Software
5. Database Setup
6. Redis / Queue / External Services
7. Environment Variables
8. Docker Setup
9. Local Development Setup
10. Running the Project
```

---

## 2. Tech Stack

### Task 5 specific technologies

| Technology | What it is | Why Task 5 uses it | Required? | Beginner explanation |
|---|---|---|---:|---|
| Go | Backend programming language | OTP generator, hasher, usecase, repositories, HTTP handlers | Yes | Go ek compiled backend language hai. Is service ka OTP code Go me likha gaya hai. |
| Go standard `net/http` | Built-in HTTP server/client library | OTP APIs expose karne and Notification Service ko HTTP call karne ke liye | Yes | Current service Gin/Fiber/gRPC server nahi use karta. Plain Go HTTP use hota hai. |
| Go `crypto/rand` | Secure random source | OTP and challenge ID predictable na ho | Yes | OTP random hona chahiye. `math/rand` insecure hota hai; `crypto/rand` safer hai. |
| Go `crypto/hmac` + `sha256` | Keyed hashing | OTP hash and Redis lookup keys protect karne ke liye | Yes | HMAC me secret pepper use hota hai, isliye DB/Redis leak se raw OTP/target easily recover nahi hota. |
| MySQL 8+ | Relational database | OTP challenge durable state store karne ke liye | Yes | OTP expiry, attempts, verified state reliable table me store hote hain. |
| Redis | In-memory key-value store | Cooldown and rate limit counters ke liye | Yes for running OTP APIs | Redis fast hai and short-lived counters ke liye perfect hai. |
| HTTP Notification Client | Outbound HTTP integration | Notification Service ko OTP send request dene ke liye | Yes for real send | Auth Service OTP generate karta hai, but email/SMS send Notification Service karega. |
| MySQL row lock `FOR UPDATE` | DB locking feature | Same OTP do baar verify na ho | Yes | Verification transaction me row lock hota hai taaki race condition avoid ho. |

### Current code vs `task5.md` recommended packages

`task5.md` includes future-oriented examples. Current repository implementation differs in these important ways:

| Item mentioned in `task5.md` | Current repository status | What beginner should do |
|---|---|---|
| `google.golang.org/grpc` | Not imported by current Auth Service code | Do not install unless code later adds gRPC handlers |
| `github.com/redis/go-redis/v9` | Not imported | Current repo has custom RESP Redis client in `internal/repository/redis_client.go` |
| `github.com/go-playground/validator/v10` | Not imported | Current handlers do manual validation through usecases |
| gRPC Auth handler | Not present | OTP endpoints are HTTP routes in `internal/transport/http/handler.go` |
| `github.com/go-sql-driver/mysql` | Already present | Needed for MySQL through `database/sql` |

Do not run `go get` for packages that current code does not import. Extra dependencies create confusion and unnecessary `go.mod` churn.

Baseline stack already documented:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 2. Tech Stack

TaskImplementation/Auth Service/task4_Dependency.md
Section: 2. Tech Stack
```

---

## 3. Required Software

Task 5 does not introduce new software beyond the existing Auth Service baseline. Use previous setup docs for installation.

| Software/service | Required for OTP unit tests? | Required for full OTP HTTP flow? | Setup reference |
|---|---:|---:|---|
| Go | Yes | Yes | `task1_Dependency.md`, section `3. Required Software` |
| MySQL 8+ | No for pure unit tests | Yes | `task1_Dependency.md`, section `5. Database Setup` |
| Redis | No for fake-repo unit tests | Yes | `task1_Dependency.md`, section `6. Redis / Queue / External Services` |
| Notification Service or mock endpoint | No | Yes for `/otp/send` success | `task1_Dependency.md`, section `6. Redis / Queue / External Services`, subsection `Notification Service` |
| OpenSSL / JWT keys | No for OTP package tests | Yes for full server startup | `task1_Dependency.md`, section `6. Redis / Queue / External Services`, subsection `OpenSSL / JWT Keys` |
| Docker | Optional | Optional but recommended | `task1_Dependency.md`, section `8. Docker Setup` |
| curl | Optional | Recommended | `task1_Dependency.md`, section `3. Required Software` |

Quick verify commands:

```bash
go version
mysql --version
redis-cli --version
curl --version
```

Note: Full Auth Service startup validates MySQL, Redis, JWT config, OTP config, Notification URL, and Session Link config. OTP-only package tests do not need all of these services.

---

## 4. Dependency Management

### Language-specific dependency system: Go modules

Full beginner explanation of:

- `go.mod`
- `go.sum`
- Go modules
- `go mod download`
- `go mod tidy`
- `go build`
- `go run`
- common Go module issues

already exists in:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 4. Dependency Management
```

### Task 5 current dependency impact

Current `backend/services/auth-service/go.mod` direct dependencies are:

```text
github.com/go-sql-driver/mysql v1.9.3
golang.org/x/crypto v0.32.0
```

Task 5 OTP implementation mostly uses Go standard library packages:

| Package area | Why used |
|---|---|
| `crypto/rand` | Secure OTP and challenge ID generation |
| `crypto/hmac` | OTP and lookup-key HMAC hashing |
| `crypto/sha256` | HMAC SHA-256 digest |
| `encoding/hex` | Store hash as hex string |
| `net/mail` | Email validation |
| `regexp` | E.164 phone validation |
| `database/sql` | MySQL repositories and transactions |
| `net/http` | OTP HTTP endpoints and Notification Service HTTP client |

### Commands for Task 5 verification

Run OTP focused tests:

```bash
cd backend/services/auth-service
go test ./internal/security/otp ./internal/usecase
```

Run all Auth Service tests:

```bash
cd backend/services/auth-service
go test ./...
```

If Go cache is not writable in your environment:

```bash
cd backend/services/auth-service
GOCACHE=/tmp/auth-go-cache go test ./...
```

### Common dependency confusion

| Symptom | Likely reason | Fix |
|---|---|---|
| Developer expects `redis/go-redis/v9` | `task5.md` recommended it, but current code uses custom Redis client | Do not add it unless imports are changed |
| Developer expects gRPC server | `task5.md` talks about gRPC methods, current code exposes HTTP routes | Test HTTP endpoints under `/api/v1/auth/otp/*` |
| `go: go.mod requires go >= 1.26.3` | Local Go version old hai | Install Go version compatible with `go.mod` |
| `package not in std` | Command wrong folder se run hua | Run from `backend/services/auth-service` |

---

## 5. Database Setup

### A. Database detected

Task 5 uses the existing Auth Service MySQL database:

```text
auth_db
```

Main OTP table:

```text
otp_challenges
```

Migration files involved:

```text
backend/services/auth-service/migrations/001_create_auth_tables.up.sql
backend/services/auth-service/migrations/003_add_otp_account_fk.up.sql
```

### B. What MySQL is

MySQL ek relational database hai. Data tables, rows, and columns me store hota hai. OTP flow me challenge state reliable hona chahiye: code hash, expiry, attempts, and verified status. Isliye MySQL use hota hai.

Full MySQL install, Docker setup, start commands, verify commands, default port, DSN format, and credential placement already exist in:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 5. Database Setup

TaskImplementation/Auth Service/task2_Dependency.md
Section: 5. Database Setup
```

### C. Why Task 5 uses MySQL

Task 5 MySQL ko OTP challenge lifecycle ke liye use karta hai:

| Column/feature | Why needed |
|---|---|
| `challenge_id` | Client ko public-safe ID return hota hai; verify request me ye ID aata hai |
| `target` | Normalized email or phone |
| `channel` | `email` or `phone` |
| `purpose` | `signup`, `login`, `password_reset`, `phone_verify`, `email_verify` |
| `otp_hash` | Plain OTP ke badle HMAC hash store hota hai |
| `attempts` and `max_attempts` | Wrong OTP retries limit karne ke liye |
| `verified_at` | OTP replay block karne ke liye |
| `expires_at` | Old OTP reject karne ke liye |
| `FOR UPDATE` row lock | Concurrent verify requests se double-use prevent karne ke liye |

### D. Required or optional

| Scenario | MySQL needed? | Notes |
|---|---:|---|
| `go test ./internal/security/otp` | No | Pure generator/hasher/validation tests |
| `go test ./internal/usecase` | No | OTP usecase tests use fake repositories |
| Full Auth Service startup | Yes | `main.go` pings MySQL |
| Real `/api/v1/auth/otp/send` | Yes | Challenge hash insert hota hai |
| Real `/api/v1/auth/otp/verify` | Yes | Challenge row lock + update hota hai |

### E. Local installation

Do not duplicate MySQL install steps here.

Refer:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 5. Database Setup
Subsections: D. Local installation, E. Docker setup, F. Start commands
```

### F. Docker setup

Task 5 does not change MySQL Docker settings.

Use the same MySQL container setup from:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 8. Docker Setup
```

Same important settings:

| Setting | Value |
|---|---|
| MySQL port | `3306` |
| Database | `auth_db` |
| Suggested user | `auth_user` |
| Persistent volume | `ecommerce-auth-mysql-data` |

### G. Apply Task 5 related migrations

For current full Auth Service, easiest is to apply all migrations in order as described in `task1_Dependency.md`.

Task 5 specifically needs:

1. Migration `001` for base `otp_challenges` table.
2. Migration `003` for optional account foreign key hardening.

Run from repo root:

```bash
mysql -u root -p < backend/services/auth-service/migrations/001_create_auth_tables.up.sql
mysql -u root -p < backend/services/auth-service/migrations/003_add_otp_account_fk.up.sql
```

If migration `003` says foreign key already exists, it may already have been applied. Verify before re-running.

### H. Verify Task 5 schema

```bash
mysql -u root -p auth_db
```

Inside MySQL:

```sql
SHOW TABLES LIKE 'otp_challenges';
SHOW COLUMNS FROM otp_challenges;
SHOW INDEX FROM otp_challenges;

SELECT CONSTRAINT_NAME, TABLE_NAME, REFERENCED_TABLE_NAME
FROM information_schema.KEY_COLUMN_USAGE
WHERE TABLE_SCHEMA = 'auth_db'
  AND TABLE_NAME = 'otp_challenges'
  AND REFERENCED_TABLE_NAME IS NOT NULL;
```

Expected important indexes:

| Index | Purpose |
|---|---|
| `uk_otp_challenges_challenge_id` | Fast and unique challenge lookup |
| `idx_otp_target_purpose_expiry` | Expire old active challenges for same target/purpose |
| `idx_otp_account` | Account-specific audit/debug lookup |

Expected foreign key after migration `003`:

```text
fk_otp_challenges_account
```

### I. Rollback Task 5 migration

Only rollback if you intentionally want to undo Task 5 FK hardening:

```bash
mysql -u root -p < backend/services/auth-service/migrations/003_add_otp_account_fk.down.sql
```

Warning: rollback of migration `001` drops base auth tables. Do not run `001_create_auth_tables.down.sql` on a DB with useful local data unless you intentionally want data loss.

### J. Connection string and credentials

Connection string format is unchanged:

```text
AUTH_MYSQL_DSN='auth_user:auth_password@tcp(127.0.0.1:3306)/auth_db?parseTime=true&charset=utf8mb4&loc=UTC'
```

Credentials placement is already explained in:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 5. Database Setup
Subsection: J. Where to place credentials

TaskImplementation/Auth Service/task1_Dependency.md
Section: 7. Environment Variables
```

Important for OTP:

- Keep `parseTime=true`; OTP code scans `TIMESTAMP` fields into Go `time.Time`.
- Use least privilege app user in production.
- Do not hardcode DB password in Go files.

---

## 6. Redis / Queue / External Services

### Redis

Redis setup and installation are already documented in:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 6. Redis / Queue / External Services
Subsection: Redis
```

Task 5 specific Redis usage:

| Redis key pattern | Purpose |
|---|---|
| `otp:cooldown:{purpose}:{target_hash}` | Blocks resend before cooldown ends |
| `otp:send:window:{purpose}:{target_hash}` | Limits sends per rolling window |
| `otp:send:daily:{purpose}:{target_hash}:{yyyyMMdd}` | Limits sends per UTC day |
| `otp:verify:ip:{source_hash}:{challenge_id}` | Limits verify attempts per source/IP and challenge |

Hinglish explanation:

Redis yahan OTP code store karne ke liye nahi hai. Redis sirf short-lived counters and cooldown keys store karta hai. Raw email, phone, and IP Redis key me direct store nahi hote; code HMAC lookup hash banata hai using `OTP_RATE_LIMIT_PEPPER`.

Task 5 Redis commands:

```bash
redis-cli PING
redis-cli --scan --pattern 'otp:*'
redis-cli TTL 'otp:cooldown:signup:PASTE_HASH_HERE'
```

Note: Actual key me hash value hoti hai, raw email/phone nahi. Isliye direct `otp:cooldown:signup:user@example.com` search nahi milega.

### Notification Service

Current Auth Service OTP delivery uses HTTP, not gRPC.

Config:

```text
NOTIFICATION_OTP_ENDPOINT=http://localhost:8084/internal/v1/notifications/otp
NOTIFICATION_TIMEOUT=3s
```

What Auth Service sends to Notification Service:

```json
{
  "target": "user@example.com",
  "channel": "email",
  "purpose": "signup",
  "otp": "123456",
  "challenge_id": "otp_chal_...",
  "expires_in_seconds": 300
}
```

Important:

- Notification Service must return HTTP `2xx`.
- If endpoint is down or returns non-2xx, Auth Service expires the created challenge and returns `OTP_DELIVERY_UNAVAILABLE`.
- Notification Service must not log raw OTP.
- Production should authenticate internal Notification requests. Current client does not add auth headers.

Setup reference:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 6. Redis / Queue / External Services
Subsection: Notification Service
```

### Kafka/RabbitMQ/NATS status

Task 5 OTP implementation does not introduce Kafka, RabbitMQ, or NATS.

Session outbox/event setup is separate and already covered in:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 6. Redis / Queue / External Services
Subsection: Session Link / Outbox

TaskImplementation/Auth Service/task4_Dependency.md
Section: 6. Redis / Queue / External Services
Subsection: Session link / outbox
```

---

## 7. Environment Variables

Task 5 introduces no new environment variables beyond the OTP and Notification variables already documented in `task1_Dependency.md`.

Refer:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 7. Environment Variables
Subsection: Complete local `.env` example
Subsection: Environment variable explanation
```

### Task 5 env keys to verify

| Variable | Required? | Default/current behavior | Why Task 5 needs it |
|---|---:|---|---|
| `OTP_HASH_PEPPER` | Yes | No default | HMAC secret for OTP hashes stored in MySQL |
| `OTP_RATE_LIMIT_PEPPER` | Optional if hash pepper set, recommended separate | Falls back to `OTP_HASH_PEPPER` | HMAC secret for Redis target/source lookup keys |
| `OTP_LENGTH` | Optional | `6` | OTP digit length, allowed 6 to 10 |
| `OTP_TTL` or `OTP_TTL_SECONDS` | Optional | `5m` | Challenge expiry |
| `OTP_MAX_ATTEMPTS` | Optional | `5` | Wrong OTP attempts per challenge |
| `OTP_RESEND_COOLDOWN` or `OTP_RESEND_COOLDOWN_SECONDS` | Optional | `1m` | Minimum resend gap |
| `OTP_SEND_LIMIT_WINDOW` or `OTP_SEND_LIMIT_WINDOW_SECONDS` | Optional | `15m` | Send quota window |
| `OTP_SEND_LIMIT_PER_WINDOW` | Optional | `3` | Sends allowed per window |
| `OTP_DAILY_LIMIT` | Optional | `10` | Sends allowed per UTC day |
| `OTP_VERIFY_IP_WINDOW` or `OTP_VERIFY_IP_WINDOW_SECONDS` | Optional | `5m` | Verify source/IP throttle window |
| `OTP_VERIFY_IP_LIMIT` | Optional | `20` | Verify attempts per source/IP window |
| `NOTIFICATION_OTP_ENDPOINT` | Optional default, but must be absolute URL | `http://localhost:8084/internal/v1/notifications/otp` | OTP delivery HTTP endpoint |
| `NOTIFICATION_TIMEOUT` | Optional | `3s` | Outbound notification call timeout |

No separate `.env` block is repeated here because the complete local `.env` already exists in `task1_Dependency.md`.

### Credentials placement

Use `.env` locally or your shell environment:

```bash
cd backend/services/auth-service
set -a
source .env
set +a
```

Production placement:

| Secret | Recommended placement |
|---|---|
| `OTP_HASH_PEPPER` | Secret manager or Kubernetes Secret |
| `OTP_RATE_LIMIT_PEPPER` | Secret manager or Kubernetes Secret |
| `AUTH_MYSQL_DSN` | Secret manager because it includes password |
| `AUTH_REDIS_PASSWORD` | Secret manager if Redis auth enabled |
| `JWT_PRIVATE_KEY_PEM_PATH` content | Mounted secret file |
| `REFRESH_TOKEN_PEPPER` | Secret manager |

Warning: Current repo contains `backend/services/auth-service/.env` and `backend/services/auth-service/secrets/` as untracked files. Keep them out of commits.

---

## 8. Docker Setup

Task 5 does not add a Dockerfile or docker-compose file.

Current repo status:

| File | Status |
|---|---|
| `backend/services/auth-service/Dockerfile` | Not present |
| root `docker-compose.yml` | Not present |
| dependency Docker examples | Already documented in previous dependency guide |

Refer:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 8. Docker Setup
Subsection: Docker Compose example for dependencies
```

Task 5 uses the same dependency containers:

| Container/service | Port | Why |
|---|---:|---|
| MySQL | `3306` | Stores `otp_challenges` |
| Redis | `6379` | Stores OTP rate limit keys |
| Notification Service | `8084` example | Receives OTP delivery request |

If running Auth Service itself in Docker later, make sure:

- `AUTH_MYSQL_DSN` uses Docker network host `mysql`, not `127.0.0.1`.
- `AUTH_REDIS_ADDR` uses `redis:6379`.
- `NOTIFICATION_OTP_ENDPOINT` uses internal service DNS, for example `http://notification-service:8084/internal/v1/notifications/otp`.
- JWT key files are mounted into the container.
- Peppers are injected as secrets, not baked into image.

---

## 9. Local Development Setup

### Flow A: Run OTP package/usecase tests only

This is the fastest Task 5 check and does not need MySQL, Redis, Notification Service, or JWT keys.

```bash
cd backend/services/auth-service
go test ./internal/security/otp ./internal/usecase
```

What this verifies:

- OTP code length and numeric format.
- Email/phone normalization.
- HMAC hashing uses challenge ID and pepper.
- Wrong OTP increments attempts.
- Correct OTP marks verified.
- Replay and expiry behavior.
- Notification client is called by the usecase through a fake notifier.

### Flow B: Prepare DB for real OTP flow

Follow baseline MySQL setup first:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 9. Local Development Setup
Steps 4 and 5
```

Then ensure Task 5 related migrations are applied:

```bash
mysql -u root -p < backend/services/auth-service/migrations/001_create_auth_tables.up.sql
mysql -u root -p < backend/services/auth-service/migrations/003_add_otp_account_fk.up.sql
```

### Flow C: Prepare Redis

Follow:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 6. Redis / Queue / External Services
Section: 8. Docker Setup
```

Verify:

```bash
redis-cli PING
```

Expected:

```text
PONG
```

### Flow D: Prepare Notification endpoint

For real `/api/v1/auth/otp/send`, `NOTIFICATION_OTP_ENDPOINT` must accept POST and return `2xx`.

If Notification Service is not running:

- Auth Service can still start if the URL is syntactically valid.
- OTP send will fail at request time with `OTP_DELIVERY_UNAVAILABLE`.
- Use the real Notification Service or a local mock endpoint before testing `/otp/send`.

Verify configured endpoint:

```bash
curl -i -X POST "$NOTIFICATION_OTP_ENDPOINT" \
  -H 'Content-Type: application/json' \
  -d '{"target":"dev@example.com","channel":"email","purpose":"signup","otp":"123456","challenge_id":"otp_chal_manualtest","expires_in_seconds":300}'
```

Expected for a working mock/service:

```text
HTTP/1.1 2xx
```

### Flow E: Run current full Auth Service

Use full baseline instructions:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 9. Local Development Setup
Section: 10. Running the Project
```

Quick version:

```bash
cd backend/services/auth-service
set -a
source .env
set +a
go run ./cmd/server
```

Verify:

```bash
curl http://localhost:8081/healthz
```

Expected:

```json
{"status":"ok"}
```

### Flow F: Manual OTP API test

Send OTP:

```bash
curl -i -X POST http://localhost:8081/api/v1/auth/otp/send \
  -H 'Content-Type: application/json' \
  -d '{"target":"user@example.com","channel":"email","purpose":"signup"}'
```

Expected successful response:

```json
{
  "challenge_id": "otp_chal_...",
  "expires_at": "2026-05-21T...",
  "expires_in": 300,
  "resend_after_seconds": 60
}
```

Verify OTP:

```bash
curl -i -X POST http://localhost:8081/api/v1/auth/otp/verify \
  -H 'Content-Type: application/json' \
  -d '{"challenge_id":"PASTE_CHALLENGE_ID","otp":"PASTE_OTP_FROM_NOTIFICATION"}'
```

Expected:

```json
{"success":true}
```

Important: Auth Service response never returns the OTP. OTP is only sent to Notification Service.

---

## 10. Running the Project

### Normal run

Use the existing full run flow:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 10. Running the Project
```

Command:

```bash
cd backend/services/auth-service
set -a
source .env
set +a
go run ./cmd/server
```

### Task 5 endpoints

| Endpoint | Method | Purpose | External dependency |
|---|---|---|---|
| `/api/v1/auth/otp/send` | `POST` | Create OTP challenge and request delivery | MySQL, Redis, Notification endpoint |
| `/api/v1/auth/otp/verify` | `POST` | Verify challenge ID + OTP code | MySQL, Redis |
| `/api/v1/auth/password/forgot` | `POST` | Create password reset OTP challenge | MySQL, Redis, Notification endpoint |
| `/healthz` | `GET` | Basic service health | Service must already be started |

### Ports and networking

| Service | Port | Purpose | Reused or new |
|---|---:|---|---|
| Auth Service HTTP | `8081` | OTP HTTP APIs and health endpoint | Reused from Task 1 |
| MySQL | `3306` | `auth_db.otp_challenges` | Reused from Task 1/2 |
| Redis | `6379` | OTP rate limits and cooldowns | Reused from Task 1 |
| Notification Service | `8084` example | OTP email/SMS delivery | Reused dependency, required for real send |

Port conflict fixes are already documented in:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 11. Ports & Networking
```

---

## 11. Common Errors & Fixes

Baseline common errors already documented:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section: 12. Common Errors & Fixes

TaskImplementation/Auth Service/task4_Dependency.md
Section: 11. Common Errors & Fixes
```

Task 5 specific troubleshooting:

| Error/symptom | Likely cause | Fix |
|---|---|---|
| `OTP_HASH_PEPPER cannot be empty` | Required secret missing | Set `OTP_HASH_PEPPER` in `.env` and source it |
| `OTP_RATE_LIMIT_PEPPER cannot be empty` | Both rate pepper and hash pepper missing | Set separate `OTP_RATE_LIMIT_PEPPER`, or at least set `OTP_HASH_PEPPER` |
| `otp length must be between 6 and 10` | `OTP_LENGTH` invalid | Use `OTP_LENGTH=6` for local |
| `NOTIFICATION_OTP_ENDPOINT must be an absolute URL` | URL missing scheme/host | Use `http://localhost:8084/internal/v1/notifications/otp` |
| `/api/v1/auth/otp/send` returns `OTP_DELIVERY_UNAVAILABLE` | Notification Service down or returns non-2xx | Start Notification Service/mock and verify endpoint with curl |
| `/api/v1/auth/otp/send` returns `OTP_RATE_LIMITED` | Cooldown/window/daily Redis limit hit | Wait for `Retry-After`, or clear local Redis keys only in dev |
| `/api/v1/auth/otp/verify` returns `INVALID_OTP` | Wrong code, expired code, replay, unknown challenge | Request a new OTP and use latest code |
| `/api/v1/auth/otp/verify` returns `OTP_RATE_LIMITED` | Too many verify attempts from same source/IP | Wait for throttle window to expire |
| `/api/v1/auth/otp/verify` returns `OTP_RATE_LIMIT_UNAVAILABLE` | Redis unavailable during verify | Start Redis, fix `AUTH_REDIS_ADDR`/password |
| MySQL time scan error | DSN missing `parseTime=true` | Add `parseTime=true` to `AUTH_MYSQL_DSN` |
| Duplicate FK on migration `003` | Migration already applied | Verify constraints instead of re-running |
| OTP send works once then repeated sends fail | Redis cooldown active | Wait `OTP_RESEND_COOLDOWN` or inspect Redis TTL |

Dev-only Redis cleanup:

```bash
redis-cli --scan --pattern 'otp:*'
```

To delete local OTP keys manually, copy exact keys from scan output and run:

```bash
redis-cli DEL 'PASTE_EXACT_OTP_KEY'
```

Do this only in local/dev. Never manually delete production rate-limit keys unless incident owner approves.

---

## 12. Security & Best Practices

### Task 5 security checklist

| Rule | Status in current implementation | Why it matters |
|---|---|---|
| Use `crypto/rand` for OTP | Implemented | OTP must not be predictable |
| Store only OTP hash | Implemented | DB leak should not reveal OTP |
| Use HMAC pepper | Implemented | 6 digit OTP hashes need server secret protection |
| Include `challenge_id` in hash input | Implemented | Same OTP across challenges should not produce same hash |
| Limit OTP length range | Implemented, 6 to 10 | Prevent invalid policy config |
| Expire old active challenges | Implemented | Latest OTP should be the active one |
| Verify with transaction and row lock | Implemented | Prevent double-use race conditions |
| Track failed attempts | Implemented | Reduce brute-force risk |
| Redis cooldown and quotas | Implemented | Reduce spam and provider abuse |
| Hash Redis target/source lookup keys | Implemented | Avoid raw PII in Redis keys |
| Generic invalid OTP response | Implemented | Avoid leaking challenge existence |

### Beginner best practices

- Keep `OTP_HASH_PEPPER` long and secret.
- Use a different value for `OTP_RATE_LIMIT_PEPPER` so DB hash and Redis lookup secrets are separated.
- Never print raw OTP, `otp_hash`, JWT, refresh token, password, or peppers in logs.
- Keep OTP TTL short. Default `5m` is fine for local and common production flows.
- Keep resend cooldown enabled. Default `1m` prevents spam and SMS/email cost spikes.
- Use `SESSION_LINK_MODE=disabled` for first local beginner run unless outbox/event publishing is configured.
- For production, authenticate internal Notification Service calls. Current HTTP client sends no auth header.

### Configuration audit from current implementation

| Finding | Impact | Recommendation |
|---|---|---|
| `backend/services/auth-service/.env` exists as untracked local file | Secrets can accidentally be committed | Add/verify `.gitignore` for `.env` and secret files before committing |
| `backend/services/auth-service/secrets/` contains local JWT PEM files | Private key risk if committed | Keep private key outside git, or ignore `secrets/*.pem` |
| Notification HTTP client has no auth header | Any reachable endpoint could receive OTP payload if URL is misconfigured | Add service-to-service auth before staging/prod |
| Notification request contains plain OTP | Necessary for delivery but sensitive | Ensure Notification Service does not log request body |
| OTP cleanup job not visible in current repo | Expired/verified OTP rows may accumulate | Add scheduled cleanup or DB retention task |
| No Dockerfile/docker-compose in repo | Onboarding depends on docs/manual commands | Add compose file later for repeatable local setup |
| Redis rate limiting fails closed | Safer security posture, but Redis outage blocks OTP | Monitor Redis and alert on `OTP_RATE_LIMIT_UNAVAILABLE` |

---

## 13. Missing or Misconfigured Things

Task 5 setup gaps to be aware of:

| Missing/misconfigured item | Why it matters | Suggested fix |
|---|---|---|
| No committed `.env.example` | Beginners may not know exact env names | Add sanitized `.env.example` later, with fake values only |
| No Docker Compose file in repo | MySQL/Redis setup is manual | Add compose for local dependencies |
| No Notification Service implementation in this repo path | `/otp/send` cannot succeed without endpoint | Run real Notification Service or a local mock |
| No OTP cleanup migration/job | Old rows may grow over time | Add scheduled cleanup for expired/verified rows |
| No service auth for Notification HTTP call | Internal endpoint security gap | Add mTLS, internal token, or signed service request |
| `task5.md` mentions packages not in current `go.mod` | New devs may install unnecessary dependencies | Follow current imports and this dependency guide |

Example cleanup SQL for local/dev or future scheduled job:

```sql
DELETE FROM otp_challenges
WHERE expires_at < UTC_TIMESTAMP() - INTERVAL 30 DAY
   OR verified_at < UTC_TIMESTAMP() - INTERVAL 30 DAY;
```

Do not run cleanup blindly in production until retention requirements are confirmed.

---

## 14. References to Previous Dependency Files

Use these instead of duplicating setup:

| Topic | Refer |
|---|---|
| Clone repo, Go install, base commands | `TaskImplementation/Auth Service/task1_Dependency.md`, sections `3`, `4`, `9` |
| Full MySQL install/Docker/DSN/migrations | `TaskImplementation/Auth Service/task1_Dependency.md`, section `5` |
| Base schema and `otp_challenges` table | `TaskImplementation/Auth Service/task2_Dependency.md`, section `5` |
| Password hashing and credential setup | `TaskImplementation/Auth Service/task3_Dependency.md`, sections `4`, `7`, `12` |
| JWT keys and refresh token config | `TaskImplementation/Auth Service/task4_Dependency.md`, sections `7`, `9`, `12` |
| Redis installation and Docker setup | `TaskImplementation/Auth Service/task1_Dependency.md`, sections `6`, `8` |
| Complete local `.env` | `TaskImplementation/Auth Service/task1_Dependency.md`, section `7` |
| Ports and networking | `TaskImplementation/Auth Service/task1_Dependency.md`, section `11` |
| Common baseline errors | `TaskImplementation/Auth Service/task1_Dependency.md`, section `12` |

---

## 15. Final Checklist

### For OTP unit tests

- [ ] Go version is compatible with `backend/services/auth-service/go.mod`.
- [ ] Dependencies are downloaded with `go mod download`.
- [ ] OTP package tests pass:

```bash
cd backend/services/auth-service
go test ./internal/security/otp
```

- [ ] OTP usecase tests pass:

```bash
cd backend/services/auth-service
go test ./internal/usecase
```

### For full OTP HTTP flow

- [ ] MySQL is running.
- [ ] Redis is running.
- [ ] Migration `001_create_auth_tables.up.sql` is applied.
- [ ] Migration `003_add_otp_account_fk.up.sql` is applied or verified.
- [ ] `AUTH_MYSQL_DSN` includes `parseTime=true`.
- [ ] `OTP_HASH_PEPPER` is set.
- [ ] `OTP_RATE_LIMIT_PEPPER` is set to a separate long secret where possible.
- [ ] `NOTIFICATION_OTP_ENDPOINT` is an absolute URL.
- [ ] Notification Service or mock returns `2xx`.
- [ ] JWT key env variables are set for full server startup.
- [ ] `SESSION_LINK_MODE=disabled` is used for first local run unless outbox is configured.
- [ ] Auth Service starts successfully.
- [ ] `/healthz` returns `{"status":"ok"}`.
- [ ] `/api/v1/auth/otp/send` returns a challenge response.
- [ ] `/api/v1/auth/otp/verify` accepts the OTP from Notification Service.

### Quick Task 5 commands

```bash
cd backend/services/auth-service
go test ./internal/security/otp ./internal/usecase
set -a
source .env
set +a
go run ./cmd/server
```

Then:

```bash
curl http://localhost:8081/healthz
```

Final mental model:

```text
Client asks for OTP
  -> Auth checks Redis cooldown/quota
  -> Auth generates secure OTP
  -> Auth stores only OTP hash in MySQL
  -> Auth sends plain OTP to Notification Service
  -> Client submits challenge_id + OTP
  -> Auth checks Redis verify throttle
  -> Auth locks MySQL challenge row
  -> Auth checks expiry, replay, attempts
  -> Auth hashes submitted OTP and compares
  -> Auth marks challenge verified
```

