# Project Dependency & Setup Guide

This guide is for **Auth Service - Task 3: Password Hashing**.

Input implementation guide:

```text
TaskImplementation/Auth Service/task3.md
```

Output dependency guide:

```text
TaskImplementation/Auth Service/task3_Dependency.md
```

Simple goal: ek beginner developer ko clearly samajh aaye ki Task 3 ke password hashing feature ke liye kaunsi dependency, config, database column, env variable, aur local setup chahiye.

Important reuse rule: Auth Service ke previous dependency guides me Go setup, MySQL installation, Redis, Docker, full `.env`, ports, and full service run flow already documented hai. Is file me wahi content repeat nahi kiya gaya. Jahan setup same hai, wahan exact reference diya gaya hai.

Current repository note: `task3.md` originally ek structured implementation guide hai, but current repo me password hashing code already present hai under `backend/services/auth-service/internal/security/password/`. This dependency guide reflects the current repository state.

---

## 1. Project Overview

Task 3 ka focus sirf **password hashing and credential validation setup** hai.

Is task me project ko ye non-code requirements chahiye:

| Requirement | Why needed |
|---|---|
| Go module dependency `golang.org/x/crypto` | Argon2id and bcrypt password hashing ke liye |
| MySQL `credentials` table | Password hash, algorithm, failed attempts, and lockout data store karne ke liye |
| Password-related env variables | Hash cost, password length, failed-attempt lockout tune karne ke liye |
| Secure random source | Har password ke liye unique salt generate karne ke liye |
| Full Auth Service baseline setup | Full server run karne ke liye MySQL, Redis, JWT keys, and env vars still required |

Task 3 does **not** introduce:

- New database engine
- New Redis feature
- Kafka/RabbitMQ/NATS
- Dockerfile
- docker-compose file
- New application port
- Third-party SaaS credentials

---

## 2. Tech Stack

### Task 3 specific technologies

| Technology | Required? | Where used | Beginner-friendly explanation |
|---|---:|---|---|
| Go | Yes | Auth Service backend code | Go ek compiled backend language hai. Is project me APIs, repositories, password hashing, and service startup Go me likhe gaye hain. |
| Go modules | Yes | `go.mod`, `go.sum` | Go modules dependency manager hai. Ye batata hai kaunsi external library ka kaunsa version use hoga. |
| `golang.org/x/crypto` | Yes | `internal/security/password/argon2id.go`, `bcrypt.go` | Ye official Go extended crypto package hai. Isme Argon2id and bcrypt implementations milte hain. |
| Argon2id | Yes, preferred | New password hashes | Argon2id modern password hashing algorithm hai. Ye slow and memory-hard hota hai, isliye brute force attacks expensive ho jate hain. |
| bcrypt | Yes, fallback | Legacy/fallback password verification | bcrypt purana but trusted password hashing algorithm hai. Legacy hashes verify karne ke liye useful hai. |
| `crypto/rand` | Yes | Salt generation | Secure random bytes generate karta hai. Salt predictable nahi hona chahiye. |
| `crypto/subtle` | Yes | Hash comparison | Constant-time comparison ke liye use hota hai, taaki timing attack risk kam ho. |
| MySQL 8+ | Yes for integration/full server | `credentials` table | MySQL relational database hai. Isme password hash and lockout state durable way me store hota hai. |
| Standard `net/http` | Existing service tech | HTTP endpoints | Project Gin/Fiber/NestJS use nahi kar raha. Current Auth Service Go standard library HTTP server use karta hai. |

### Already documented baseline technologies

Do not duplicate setup from previous docs. Use these references:

| Topic | Already explained in |
|---|---|
| Full Go installation and Go command explanation | `TaskImplementation/Auth Service/task1_Dependency.md`, section `4. Dependency Management` |
| MySQL install, Docker run, migrations, DSN | `TaskImplementation/Auth Service/task1_Dependency.md`, section `5. Database Setup` |
| Task 2 base schema and `credentials` table migration | `TaskImplementation/Auth Service/task2_Dependency.md`, section `5. Database Setup` |
| Redis setup for full server startup | `TaskImplementation/Auth Service/task1_Dependency.md`, section `6. Redis / Queue / External Services` |
| Full `.env` example | `TaskImplementation/Auth Service/task1_Dependency.md`, section `7. Environment Variables` |
| Docker dependency containers | `TaskImplementation/Auth Service/task1_Dependency.md`, section `8. Docker Setup` |

---

## 3. Required Software

### For Task 3 password package tests only

| Software | Required? | Purpose | Verify command |
|---|---:|---|---|
| Go | Yes | Build and test password hashing code | `go version` |
| Internet access or module cache | Usually yes | Download `golang.org/x/crypto` if not already cached | `go mod download` |

Task 3 password package tests do **not** need MySQL, Redis, Docker, JWT keys, or Notification Service.

### For Task 3 database integration or full server run

| Software/service | Required? | Purpose |
|---|---:|---|
| MySQL 8+ | Yes | `credentials` table stores hash and lockout fields |
| Redis | Required for current full server startup | Server pings Redis on startup even if you only test password endpoints |
| JWT RSA keys | Required for current full server startup | Current `main.go` loads JWT keys during startup |
| Docker | Optional but recommended | Easy local MySQL/Redis setup |
| MySQL CLI | Recommended | Verify migration and table columns |

Refer:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Sections:
3. Required Software
5. Database Setup
6. Redis / Queue / External Services
8. Docker Setup
```

---

## 4. Dependency Management

### Language-specific dependency system: Go modules

This project is a Go project. Go dependency files are:

| File | Meaning |
|---|---|
| `backend/services/auth-service/go.mod` | Module name, Go version, and direct dependencies |
| `backend/services/auth-service/go.sum` | Cryptographic checksums for downloaded modules |

Simple Hinglish:

- `go.mod` dependency list hai.
- `go.sum` dependency integrity lock hai.
- `go mod download` dependencies download karta hai.
- `go mod tidy` missing dependency add karta hai and unused dependency remove karta hai.
- `go test` code compile karke tests run karta hai.

Full beginner explanation already exists in:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section:
4. Dependency Management
```

### Task 3 new/important dependency

Current `backend/services/auth-service/go.mod` includes:

```go
require (
    github.com/go-sql-driver/mysql v1.9.3
    golang.org/x/crypto v0.32.0
)
```

Task 3 password hashing uses `golang.org/x/crypto` for:

| Package | Used for |
|---|---|
| `golang.org/x/crypto/argon2` | Argon2id password hashing |
| `golang.org/x/crypto/bcrypt` | bcrypt fallback and legacy verification |

### Commands

From repo root:

```bash
cd backend/services/auth-service
go mod download
go test ./internal/security/password
```

If `golang.org/x/crypto` is missing in a fresh branch:

```bash
cd backend/services/auth-service
go get golang.org/x/crypto
go mod tidy
```

Note: current repo already has this dependency, so normally `go get` is not needed.

### Why dependencies fail

| Error/symptom | Common reason | Fix |
|---|---|---|
| `no required module provides package golang.org/x/crypto/argon2` | Dependency missing from `go.mod` or module not downloaded | Run `go mod download`; if still missing, run `go get golang.org/x/crypto` |
| `go: go.mod requires go >= 1.26.3` | Installed Go version old hai | Install Go version compatible with `go.mod` |
| `checksum mismatch` | Go module cache corrupted | Run `go clean -modcache`, then `go mod download` |
| Download fails behind office network | Proxy/firewall blocks Go module proxy | Configure `GOPROXY` or company proxy |
| Password tests slow | Argon2 memory/time cost intentionally expensive | Use focused tests; do not lower production config without benchmarking |

---

## 5. Database Setup

### A. Database detected

Task 3 does not add a new database. It reuses:

```text
MySQL 8+
```

### B. What MySQL is

MySQL ek relational database hai. Data tables, rows, and columns me store hota hai. Password hash ke saath account relation, failed login attempts, and lockout state consistent rehna chahiye, isliye MySQL use hota hai.

### C. Why Task 3 uses MySQL

Task 3 password hashing flow `credentials` table use karta hai:

| Column | Purpose |
|---|---|
| `account_id` | Credential ko account se link karta hai |
| `password_hash` | Plain password nahi, only encoded hash store hota hai |
| `password_algo` | `argon2id` or `bcrypt` identify karta hai |
| `password_changed_at` | Last password change timestamp |
| `failed_attempts` | Wrong password attempts count |
| `locked_until` | Account-level temporary lockout timestamp |

Current migration file:

```text
backend/services/auth-service/migrations/001_create_auth_tables.up.sql
```

### D. Required or optional

| Scenario | MySQL needed? | Explanation |
|---|---:|---|
| Run `go test ./internal/security/password` | No | Pure hashing tests DB touch nahi karte |
| Test password usecase with fake repo | No | `internal/usecase` tests fake repository use kar sakte hain |
| Verify real credential persistence | Yes | `credentials` table chahiye |
| Run full Auth Service | Yes | `main.go` MySQL `PingContext` karta hai |

### E. Local installation

MySQL installation and local setup is unchanged from previous dependency docs.

Refer:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section:
5. Database Setup

TaskImplementation/Auth Service/task2_Dependency.md
Sections:
5. Database Setup
5.H. Apply Task 2 migration
```

### F. Docker setup

Task 3 introduces no new MySQL Docker settings. Use the same MySQL container setup from:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section:
8. Docker Setup
```

### G. Start commands

For MySQL start commands, reuse:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Sections:
5.F. Start commands
8. Docker Setup
```

### H. Verify running and verify Task 3 columns

After applying migration `001`, verify the credential table:

```bash
cd backend/services/auth-service
mysql -h 127.0.0.1 -P 3306 -u auth_user -p auth_db
```

Inside MySQL:

```sql
DESCRIBE credentials;
SHOW COLUMNS FROM credentials LIKE 'password_hash';
SHOW COLUMNS FROM credentials LIKE 'password_algo';
SHOW COLUMNS FROM credentials LIKE 'failed_attempts';
SHOW COLUMNS FROM credentials LIKE 'locked_until';
```

Expected high-level result:

| Check | Expected |
|---|---|
| `password_hash` | Exists, `VARCHAR(255)`, not null |
| `password_algo` | Exists, `VARCHAR(32)`, default `argon2id` |
| `failed_attempts` | Exists, integer, default `0` |
| `locked_until` | Exists, nullable timestamp |

### I. Default port

MySQL default port is:

```text
3306
```

This is reused from previous setup. If you map Docker MySQL to another host port, update `AUTH_MYSQL_DSN`.

### J. Connection string format

Same DSN format as previous docs:

```env
AUTH_MYSQL_DSN='auth_user:<password>@tcp(127.0.0.1:3306)/auth_db?parseTime=true&charset=utf8mb4&loc=UTC'
```

Important:

- Put DB username/password inside `AUTH_MYSQL_DSN`.
- Keep credentials in `backend/services/auth-service/.env` for local development.
- In production, keep credentials in secret manager or Kubernetes Secret.
- `parseTime=true` is required because Go scans MySQL timestamps into `time.Time`.

Full DSN explanation:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section:
5.I. Connection string format
```

---

## 6. Redis / Queue / External Services

Task 3 password hashing does **not** introduce Redis, Kafka, RabbitMQ, NATS, MinIO, Elasticsearch, SMTP, Stripe, Twilio, OAuth, Docker, or Kubernetes.

### Task 3 service dependency matrix

| Service | Task 3 hashing tests | Full Auth Service startup | Notes |
|---|---:|---:|---|
| MySQL | No | Yes | Needed for real credential persistence |
| Redis | No | Yes | Current `main.go` pings Redis at startup |
| Kafka/RabbitMQ/NATS | No | No direct dependency | Current Auth Service uses DB outbox plus optional HTTP publisher, not direct broker client |
| Notification Service | No | Config exists | Password-only endpoints do not need OTP delivery |
| JWT key files | No | Yes | Current full server loads JWT keys |
| Docker | Optional | Optional | Helpful for MySQL/Redis only |

### Redis reference

If you run only:

```bash
go test ./internal/security/password
```

Redis is not required.

If you run:

```bash
go run ./cmd/server
```

Redis is required because current server startup creates and pings Redis.

Refer:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section:
6. Redis / Queue / External Services
```

---

## 7. Environment Variables

Only Task 3-specific variables are listed here. Full `.env` is already documented in:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section:
7. Environment Variables
```

### Task 3 `.env` variables

Add these to:

```text
backend/services/auth-service/.env
```

```env
# Password policy and hashing
AUTH_PASSWORD_MIN_LENGTH=8
AUTH_PASSWORD_MAX_LENGTH=128
AUTH_PASSWORD_ARGON2_MEMORY_KIB=65536
AUTH_PASSWORD_ARGON2_ITERATIONS=3
AUTH_PASSWORD_ARGON2_PARALLELISM=2
AUTH_PASSWORD_SALT_LENGTH=16
AUTH_PASSWORD_KEY_LENGTH=32
AUTH_PASSWORD_BCRYPT_COST=12
AUTH_PASSWORD_MAX_FAILED_ATTEMPTS=5
AUTH_PASSWORD_LOCKOUT_DURATION=15m
```

### Variable explanation

| Variable | Required? | Default in code/docs | Purpose | Beginner note |
|---|---:|---|---|---|
| `AUTH_PASSWORD_MIN_LENGTH` | Optional | `8` | Minimum password length | Kam value weak password allow karegi. |
| `AUTH_PASSWORD_MAX_LENGTH` | Optional | `128` | Maximum password length | Very large password input se abuse/memory pressure avoid hota hai. |
| `AUTH_PASSWORD_ARGON2_MEMORY_KIB` | Optional | `65536` | Argon2id memory cost in KiB | `65536` means 64 MiB per hash operation. |
| `AUTH_PASSWORD_ARGON2_ITERATIONS` | Optional | `3` | Argon2id time cost | Higher value stronger but slower hoti hai. |
| `AUTH_PASSWORD_ARGON2_PARALLELISM` | Optional | `2` | Argon2id lanes/parallelism | CPU resources ke according tune hota hai. |
| `AUTH_PASSWORD_SALT_LENGTH` | Optional | `16` | Salt length in bytes | Non-zero hona chahiye. 16 bytes local/prod ke liye reasonable hai. |
| `AUTH_PASSWORD_KEY_LENGTH` | Optional | `32` | Derived hash key length in bytes | Non-zero hona chahiye. |
| `AUTH_PASSWORD_BCRYPT_COST` | Optional | `12` | bcrypt work factor | Legacy bcrypt hashes ke liye used. |
| `AUTH_PASSWORD_MAX_FAILED_ATTEMPTS` | Optional | `5` | Wrong password attempts before lockout | DB-level brute-force protection. |
| `AUTH_PASSWORD_LOCKOUT_DURATION` | Optional | `15m` | Lock duration after too many failures | Example values: `5m`, `15m`, `1h`. |

### Credentials placement

Password hashing config itself is not a secret, but it should still live with service config:

| Item | Local placement | Production placement |
|---|---|---|
| Password tuning variables | `backend/services/auth-service/.env` | Deployment env/config map |
| DB username/password | `AUTH_MYSQL_DSN` in `.env` | Secret manager or Kubernetes Secret |
| Plain user passwords | Never store in config | Never store anywhere |
| Password hashes | MySQL `credentials.password_hash` | MySQL/Auth DB |

Warning:

- Never commit `.env`.
- Never put real plain passwords in docs, logs, shell history examples, or SQL seed files.
- Current repo has a local `.env` file in working tree. Treat it as local-only and do not commit it.

---

## 8. Docker Setup

Task 3 introduces **no new Docker container**.

### Current repo status

Current repo does not contain:

- `Dockerfile`
- `docker-compose.yml`
- `.dockerignore`

So for Task 3:

| Docker item | Needed for password tests? | Needed for full server? | Explanation |
|---|---:|---:|---|
| App container | No | Yes | Auth Service Dockerfile and root Compose service are available |
| MySQL container | No | Yes/recommended | Full server needs DB |
| Redis container | No | Yes/recommended | Full server pings Redis |
| Persistent MySQL volume | No | Recommended | Keeps `credentials` data across restarts |
| Docker network | No | Useful later | Needed if app is containerized with dependencies |

Reuse previous Docker docs:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section:
8. Docker Setup
```

### Task 3 DevOps note: Argon2 memory

Argon2id uses memory intentionally. Default config:

```text
AUTH_PASSWORD_ARGON2_MEMORY_KIB=65536
```

That means one hash operation can use roughly 64 MiB memory. In production/container deployment:

- Do not set tiny container memory limits.
- Benchmark login/signup concurrency.
- Keep API/gateway rate limits so attackers cannot force many expensive hash operations.
- Avoid lowering Argon2 cost without security review.

---

## 9. Local Development Setup

### Flow A: Run Task 3 password hashing tests only

Use this when you only want to verify hashing, verification, salt uniqueness, bcrypt fallback, and policy validation.

```bash
cd backend/services/auth-service
go mod download
go test ./internal/security/password
```

Expected:

```text
ok   github.com/example/ecommerce-platform/backend/services/auth-service/internal/security/password
```

This flow does not require MySQL/Redis/Docker.

### Flow B: Run password usecase tests

Usecase tests validate that plain password is hashed before storing, failed attempts are recorded, and rehash can happen.

```bash
cd backend/services/auth-service
go test ./internal/usecase -run 'Password|Credential'
```

This flow uses fake repositories in tests, so MySQL is not required.

### Flow C: Verify real DB integration

Use this when you want real MySQL `credentials` table verification.

1. Start MySQL using previous docs.
2. Apply migration `001_create_auth_tables.up.sql`.
3. Verify `credentials` columns from section `5.H` above.
4. Use full server flow from Task 1 if you want to call internal credential endpoints.

Refer:

```text
TaskImplementation/Auth Service/task2_Dependency.md
Sections:
5.H. Apply Task 2 migration
5.I. Verify Task 2 migration
```

### Flow D: Run current full Auth Service

Full service startup needs more than Task 3:

- MySQL running
- Redis running
- migrations `001` to `005` applied
- JWT RSA keys available
- complete `.env` loaded
- required secret peppers configured

Refer:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Sections:
9. Local Development Setup
10. Running the Project
```

Minimal command shape:

```bash
cd backend/services/auth-service
set -a
source .env
set +a
go run ./cmd/server
```

---

## 10. Running the Project

Task 3 itself is not a standalone service. It is a feature inside Auth Service.

### Task 3 related runtime endpoints

Current HTTP handler includes these password-related internal endpoints:

| Method | Endpoint | Purpose |
|---|---|---|
| `POST` | `/internal/v1/auth/credentials` | Create credential and store hashed password |
| `POST` | `/internal/v1/auth/password/verify` | Verify identifier/password |
| `POST` | `/internal/v1/auth/password/reset` | Store a new password hash |
| `POST` | `/api/v1/auth/login` | Public login flow uses password verification internally |

Important:

- Internal endpoints should not be exposed publicly.
- Run full server only after following Task 1 full setup.
- Do not insert plain passwords manually into `credentials.password_hash`.

### Ports and networking

Task 3 introduces no new ports.

| Service | Port | Purpose | Task 3 status |
|---|---:|---|---|
| Auth Service HTTP | `8081` | Password endpoints and full Auth API | Reused |
| MySQL | `3306` | `credentials` table | Reused |
| Redis | `6379` | Required by current full server startup | Reused, not password-specific |
| Notification Service | `8084` | OTP delivery integration | Not used by Task 3 password hashing |

Full port explanation:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section:
11. Ports & Networking
```

---

## 11. Common Errors & Fixes

Only Task 3-specific issues are listed here. General Go, MySQL, Redis, Docker, JWT, and full service errors already exist in:

```text
TaskImplementation/Auth Service/task1_Dependency.md
Section:
12. Common Errors & Fixes
```

### Password dependency and config errors

| Error/symptom | Likely reason | Fix |
|---|---|---|
| `no required module provides package golang.org/x/crypto/argon2` | `golang.org/x/crypto` missing or not downloaded | Run `go mod download`; if missing from `go.mod`, run `go get golang.org/x/crypto` |
| `auth.password.hasher_init_failed` | Password env values invalid | Check `AUTH_PASSWORD_*` variables |
| `invalid argon2id config: memory must be greater than zero` | `AUTH_PASSWORD_ARGON2_MEMORY_KIB=0` | Set `AUTH_PASSWORD_ARGON2_MEMORY_KIB=65536` |
| `invalid argon2id config: iterations must be greater than zero` | `AUTH_PASSWORD_ARGON2_ITERATIONS=0` | Set `AUTH_PASSWORD_ARGON2_ITERATIONS=3` |
| `invalid argon2id config: parallelism must be greater than zero` | `AUTH_PASSWORD_ARGON2_PARALLELISM=0` | Set `AUTH_PASSWORD_ARGON2_PARALLELISM=2` |
| `bcrypt cost must be within bcrypt supported range` | Cost too low/high | Use `AUTH_PASSWORD_BCRYPT_COST=12` |
| `password must be at least 8 characters` | Password shorter than policy | Use longer password or change policy intentionally |
| `password cannot be blank` | Password empty or only spaces | Send a real password |
| `password exceeds bcrypt 72 byte limit` | bcrypt fallback cannot handle long byte input | Prefer Argon2id for new hashes; avoid long passwords for legacy bcrypt reset |

### Runtime and database errors

| Error/symptom | Likely reason | Fix |
|---|---|---|
| Login always returns `INVALID_CREDENTIALS` after manual seed | Plain password was inserted into `password_hash` | Create credential through password usecase/internal endpoint so it stores a hash |
| `invalid password hash format` | DB has malformed `password_hash` | Recreate/reset credential using proper hasher |
| `unsupported password algorithm` | `password_algo` not `argon2id` or `bcrypt` | Fix stored algorithm value or migrate credential |
| Account locked during local testing | Too many wrong password attempts updated `failed_attempts`/`locked_until` | Wait for lockout duration or reset local DB row during development |
| Full server fails even though password tests pass | Full server also needs MySQL, Redis, JWT keys, OTP peppers, etc. | Follow `task1_Dependency.md` full setup |
| Hashing is slow under load | Argon2id cost is intentionally high | Benchmark and tune carefully; add rate limits before lowering cost |

### Important config behavior

Current config helper falls back to defaults when a numeric env value cannot be parsed. Example: if `AUTH_PASSWORD_ARGON2_ITERATIONS=abc`, code may silently use the default instead of failing.

Best practice:

- Use numeric values only.
- Add config validation tests or stricter parsing later.
- Keep a reviewed `.env.example` so beginners do not mistype variable values.

---

## 12. Security & Best Practices

### Task 3 security checklist

- Never store plain passwords in MySQL.
- Never log request password, password hash, salt, or reset password value.
- Use Argon2id for all new password hashes.
- Keep bcrypt only for fallback/legacy verification.
- Rehash bcrypt/old Argon2id params after successful login when `NeedsRehash=true`.
- Keep password errors generic for login: `Invalid credentials`.
- Use account lockout plus API/gateway IP rate limiting.
- Use HTTPS in any environment where real passwords are sent.
- Keep Auth Service internal endpoints private.
- Do not expose `/internal/v1/auth/*` directly to the public internet.
- Benchmark Argon2id settings before production launch.
- Use MySQL app user with least privilege for runtime.
- Use migration user separately for DDL in production.

### Beginner best practices

| Practice | Why it matters |
|---|---|
| Use long test passwords like `DevPassword123!` | Short passwords fail policy and confuse local testing |
| Do not manually write hash strings | Hash format has algorithm, version, params, salt, and key |
| Keep one local `.env` source of truth | Mismatched env values create hard-to-debug startup issues |
| Run focused tests first | Password package tests are faster to debug than full service startup |
| Treat `.env` as secret | It contains DB/Redis/JWT/pepper values in real local setup |

### Security audit from current implementation

| Finding | Risk | Recommendation |
|---|---|---|
| Repository `.gitignore` covers `.env`, PEM, secrets, and binaries | Accidental secret/build artifact commits are reduced | Keep ignore rules and secret scanning in CI |
| Sanitized `.env.example` is present | Local configuration has safe placeholders | Keep real local/production credentials untracked |
| Dockerfile and root Compose are present | Repeatable local deployment is available | Keep container build and smoke checks in CI |
| Internal password endpoints are registered on same HTTP server | If service is exposed, internal credential APIs can be abused | Protect internal routes using network policy, gateway rules, mTLS, or service auth |
| Numeric env parse errors can fall back to defaults | Mistyped env values may be hidden | Use strict env parsing or log warnings for invalid values |
| No dedicated IP/login rate limiter in Task 3 | DB lockout protects account, but attackers can still cause load | Add gateway/Auth middleware rate limiting |
| `/healthz` is liveness-only and `/readyz` checks MySQL/Redis | Probes distinguish process health from dependency readiness | Use `/readyz` for deployment rollouts |
| Compose `migrate-auth` job is present | SQL migrations run before Auth startup | Keep the migration job and schema compatibility checks in CI |

---

## 13. Missing or Misconfigured Things

Task 3 does not block on all of these for local unit tests, but they matter for smooth onboarding and deployment.

| Missing/misconfigured item | Impact | Suggested fix |
|---|---|---|
| `.env.example` | Sanitized template is present | Keep it aligned with validated config |
| `.gitignore` | Env, secrets, PEM, cache, and binaries are ignored | Keep secret scanning enabled |
| Dockerfile | Multi-stage Auth image is present | Keep build verification in CI |
| docker-compose file | Root Compose starts dependencies and Auth | Keep health/dependency ordering current |
| Migration runner/tracking | Compose `migrate-auth` uses `golang-migrate` | Monitor dirty versions and test rollback procedures |
| Argon2 benchmark guide | Production cost tuning unclear | Add benchmark command/script for signup/login hashing |
| Internal endpoint auth | Internal APIs rely on deployment/network safety | Add middleware or service-to-service authentication |
| Strict env parsing | Invalid numeric env can silently fallback | Return config error when env value is malformed |

---

## 14. References to Previous Dependency Files

Use these instead of duplicating setup instructions:

| Need | Refer |
|---|---|
| Full clone/install/run flow | `TaskImplementation/Auth Service/task1_Dependency.md`, section `9. Local Development Setup` |
| Go module explanation | `TaskImplementation/Auth Service/task1_Dependency.md`, section `4. Dependency Management` |
| MySQL installation | `TaskImplementation/Auth Service/task1_Dependency.md`, section `5. Database Setup` |
| Base schema migration | `TaskImplementation/Auth Service/task2_Dependency.md`, section `5. Database Setup` |
| MySQL Docker setup | `TaskImplementation/Auth Service/task1_Dependency.md`, section `8. Docker Setup` |
| Redis setup | `TaskImplementation/Auth Service/task1_Dependency.md`, section `6. Redis / Queue / External Services` |
| Full `.env` example | `TaskImplementation/Auth Service/task1_Dependency.md`, section `7. Environment Variables` |
| Ports and networking | `TaskImplementation/Auth Service/task1_Dependency.md`, section `11. Ports & Networking` |
| General common errors | `TaskImplementation/Auth Service/task1_Dependency.md`, section `12. Common Errors & Fixes` |
| Baseline security checklist | `TaskImplementation/Auth Service/task1_Dependency.md`, section `13. Security & Best Practices` |

---

## 15. Final Checklist

### For Task 3 hashing tests

- [ ] Go version is compatible with `backend/services/auth-service/go.mod`.
- [ ] `golang.org/x/crypto` exists in `go.mod`.
- [ ] `go mod download` completed.
- [ ] `go test ./internal/security/password` passes.
- [ ] Password policy values are understood: min length `8`, max length `128`.
- [ ] Argon2id default parameters are understood before changing them.

### For Task 3 DB/full-service verification

- [ ] MySQL is running if testing real credential persistence.
- [ ] Migration `001_create_auth_tables.up.sql` is applied.
- [ ] `credentials.password_hash` column exists.
- [ ] `credentials.password_algo` column exists and supports `argon2id`/`bcrypt`.
- [ ] `credentials.failed_attempts` and `credentials.locked_until` exist.
- [ ] `AUTH_PASSWORD_*` variables are present in local `.env` if overriding defaults.
- [ ] Full Auth Service setup from `task1_Dependency.md` is complete before `go run ./cmd/server`.
- [ ] Redis is running before full server startup.
- [ ] JWT keys and required peppers are configured before full server startup.
- [ ] `.env` and key files are not committed.
- [ ] Internal password endpoints are not exposed publicly.

### Quick Task 3 commands

```bash
cd backend/services/auth-service
go mod download
go test ./internal/security/password
go test ./internal/usecase -run 'Password|Credential'
```
